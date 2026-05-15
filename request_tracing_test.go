package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	ratelimiter "ratelimiter-redis"
)

func tracingOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestTracing_SetsTraceAndSpanHeaders(t *testing.T) {
	cfg := ratelimiter.NewTracingConfig()
	h := ratelimiter.WithRequestTracing(cfg, http.HandlerFunc(tracingOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Trace-ID") == "" {
		t.Error("expected X-Trace-ID to be set")
	}
	if rec.Header().Get("X-Span-ID") == "" {
		t.Error("expected X-Span-ID to be set")
	}
}

func TestWithRequestTracing_PreservesIncomingTraceID(t *testing.T) {
	cfg := ratelimiter.NewTracingConfig()
	h := ratelimiter.WithRequestTracing(cfg, http.HandlerFunc(tracingOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Trace-ID", "existing-trace-id")
	h.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Trace-ID"); got != "existing-trace-id" {
		t.Errorf("expected existing-trace-id, got %s", got)
	}
}

func TestWithRequestTracing_GeneratesUniqueSpanIDs(t *testing.T) {
	cfg := ratelimiter.NewTracingConfig()
	h := ratelimiter.WithRequestTracing(cfg, http.HandlerFunc(tracingOKHandler))

	ids := map[string]struct{}{}
	for i := 0; i < 10; i++ {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		h.ServeHTTP(rec, req)
		ids[rec.Header().Get("X-Span-ID")] = struct{}{}
	}
	if len(ids) < 9 {
		t.Errorf("expected unique span IDs, got %d distinct values", len(ids))
	}
}

func TestWithRequestTracing_SetsLatencyHeader(t *testing.T) {
	cfg := ratelimiter.NewTracingConfig()
	h := ratelimiter.WithRequestTracing(cfg, http.HandlerFunc(tracingOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Trace-Latency-Ms") == "" {
		t.Error("expected X-Trace-Latency-Ms to be set")
	}
}

func TestWithRequestTracing_OnTraceCallback(t *testing.T) {
	var (
		mu      sync.Mutex
		called  bool
		gotID   string
	)
	cfg := ratelimiter.NewTracingConfig(
		ratelimiter.WithTracingOnTrace(func(r *http.Request, traceID, spanID string, latency time.Duration) {
			mu.Lock()
			defer mu.Unlock()
			called = true
			gotID = traceID
		}),
	)
	h := ratelimiter.WithRequestTracing(cfg, http.HandlerFunc(tracingOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	mu.Lock()
	defer mu.Unlock()
	if !called {
		t.Error("expected onTrace callback to be called")
	}
	if gotID != rec.Header().Get("X-Trace-ID") {
		t.Errorf("callback traceID %q does not match response header %q", gotID, rec.Header().Get("X-Trace-ID"))
	}
}

func TestWithRequestTracing_NilConfigUsesDefaults(t *testing.T) {
	h := ratelimiter.WithRequestTracing(nil, http.HandlerFunc(tracingOKHandler))
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	h.ServeHTTP(rec, req)

	if rec.Header().Get("X-Trace-ID") == "" {
		t.Error("expected default X-Trace-ID header to be set")
	}
}
