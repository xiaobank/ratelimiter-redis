package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

func enrichOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestEnrichment_SetsRequestTimeHeader(t *testing.T) {
	cfg := ratelimiter.DefaultEnrichmentConfig()
	handler := ratelimiter.WithRequestEnrichment(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Header.Get("X-Request-Time")
		if val == "" {
			t.Error("expected X-Request-Time header to be set on request")
		}
		_, err := time.Parse(time.RFC3339Nano, val)
		if err != nil {
			t.Errorf("X-Request-Time is not valid RFC3339Nano: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)
}

func TestWithRequestEnrichment_SetsLatencyHeader(t *testing.T) {
	cfg := ratelimiter.DefaultEnrichmentConfig()
	handler := ratelimiter.WithRequestEnrichment(cfg)(http.HandlerFunc(enrichOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	val := rec.Header().Get("X-Response-Latency-Ms")
	if val == "" {
		t.Fatal("expected X-Response-Latency-Ms header on response")
	}
	ms, err := strconv.Atoi(val)
	if err != nil {
		t.Fatalf("latency header is not a valid integer: %v", err)
	}
	if ms < 0 {
		t.Errorf("latency should be non-negative, got %d", ms)
	}
}

func TestWithRequestEnrichment_SetsServerHeader(t *testing.T) {
	cfg := ratelimiter.DefaultEnrichmentConfig()
	cfg.ServerHeader = "node-1"
	handler := ratelimiter.WithRequestEnrichment(cfg)(http.HandlerFunc(enrichOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Served-By"); got != "node-1" {
		t.Errorf("expected X-Served-By=node-1, got %q", got)
	}
}

func TestWithRequestEnrichment_NilConfigUsesDefaults(t *testing.T) {
	handler := ratelimiter.WithRequestEnrichment(nil)(http.HandlerFunc(enrichOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Response-Latency-Ms") == "" {
		t.Error("expected default latency header to be set")
	}
}

func TestWithRequestEnrichment_NoServerHeaderWhenEmpty(t *testing.T) {
	cfg := ratelimiter.DefaultEnrichmentConfig()
	// ServerHeader intentionally left empty
	handler := ratelimiter.WithRequestEnrichment(cfg)(http.HandlerFunc(enrichOKHandler))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Served-By"); got != "" {
		t.Errorf("expected no X-Served-By header, got %q", got)
	}
}
