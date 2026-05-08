package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

func requestIDOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestID_SetsHeaderWhenAbsent(t *testing.T) {
	handler := ratelimiter.WithRequestID()(http.HandlerFunc(requestIDOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get(ratelimiter.DefaultRequestIDHeader) == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestWithRequestID_PreservesExistingID(t *testing.T) {
	handler := ratelimiter.WithRequestID()(http.HandlerFunc(requestIDOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(ratelimiter.DefaultRequestIDHeader, "existing-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(ratelimiter.DefaultRequestIDHeader); got != "existing-id" {
		t.Fatalf("expected existing-id, got %s", got)
	}
}

func TestWithRequestID_OverwriteOption(t *testing.T) {
	handler := ratelimiter.WithRequestID(
		ratelimiter.WithRequestIDOverwrite(true),
	)(http.HandlerFunc(requestIDOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(ratelimiter.DefaultRequestIDHeader, "old-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(ratelimiter.DefaultRequestIDHeader); got == "old-id" {
		t.Fatal("expected overwrite to replace the existing ID")
	}
}

func TestWithRequestID_CustomHeader(t *testing.T) {
	const customHeader = "X-Trace-ID"
	handler := ratelimiter.WithRequestID(
		ratelimiter.WithRequestIDHeader(customHeader),
	)(http.HandlerFunc(requestIDOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get(customHeader) == "" {
		t.Fatalf("expected %s header to be set", customHeader)
	}
	if rec.Header().Get(ratelimiter.DefaultRequestIDHeader) != "" {
		t.Fatal("default header should not be set when custom header is used")
	}
}

func TestWithRequestID_CustomGenerator(t *testing.T) {
	const fixedID = "fixed-test-id"
	handler := ratelimiter.WithRequestID(
		ratelimiter.WithRequestIDGenerator(func() string { return fixedID }),
	)(http.HandlerFunc(requestIDOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get(ratelimiter.DefaultRequestIDHeader); got != fixedID {
		t.Fatalf("expected %s, got %s", fixedID, got)
	}
}

func TestNewRequestIDConfig_Defaults(t *testing.T) {
	cfg := ratelimiter.NewRequestIDConfig()
	if cfg.Header() != ratelimiter.DefaultRequestIDHeader {
		t.Fatalf("expected default header %s, got %s", ratelimiter.DefaultRequestIDHeader, cfg.Header())
	}
	if cfg.Overwrite() {
		t.Fatal("expected overwrite to be false by default")
	}
}
