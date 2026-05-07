package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/example/ratelimiter-redis"
)

func logOKHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-RateLimit-Remaining", "9")
	w.WriteHeader(http.StatusOK)
}

func logBlockedHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTooManyRequests)
}

func TestWithRequestLogger_LogsAllowedRequest(t *testing.T) {
	var captured ratelimiter.LogEntry
	logFn := func(e ratelimiter.LogEntry) { captured = e }

	h := ratelimiter.WithRequestLogger(
		http.HandlerFunc(logOKHandler), nil, logFn,
	)

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if !captured.Allowed {
		t.Fatal("expected request to be logged as allowed")
	}
	if captured.Path != "/ping" {
		t.Fatalf("expected path /ping, got %s", captured.Path)
	}
	if captured.Remaining != 9 {
		t.Fatalf("expected remaining 9, got %d", captured.Remaining)
	}
	if captured.Latency <= 0 {
		t.Fatal("expected positive latency")
	}
}

func TestWithRequestLogger_LogsBlockedRequest(t *testing.T) {
	var captured ratelimiter.LogEntry
	logFn := func(e ratelimiter.LogEntry) { captured = e }

	h := ratelimiter.WithRequestLogger(
		http.HandlerFunc(logBlockedHandler), nil, logFn,
	)

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.RemoteAddr = "10.0.0.2:5678"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if captured.Allowed {
		t.Fatal("expected request to be logged as blocked")
	}
	if captured.Method != http.MethodPost {
		t.Fatalf("expected method POST, got %s", captured.Method)
	}
}

func TestNewRequestLogger_UsesOptions(t *testing.T) {
	var capturedKey string
	customKey := func(r *http.Request) (string, error) { return "custom-key", nil }
	logFn := func(e ratelimiter.LogEntry) { capturedKey = e.Key }

	h := ratelimiter.NewRequestLogger(
		http.HandlerFunc(logOKHandler),
		ratelimiter.WithLogKeyFunc(customKey),
		ratelimiter.WithLogFunc(logFn),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if capturedKey != "custom-key" {
		t.Fatalf("expected key 'custom-key', got %q", capturedKey)
	}
}

func TestWithRequestLogger_NilLogFnUsesDefault(t *testing.T) {
	// Should not panic when logFn is nil (falls back to default logger).
	h := ratelimiter.WithRequestLogger(http.HandlerFunc(logOKHandler), nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req) // no panic expected
}
