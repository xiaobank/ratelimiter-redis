package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/example/ratelimiter-redis"
)

func correlationOKHandler(w http.ResponseWriter, r *http.Request) {
	// Echo back whatever correlation ID the request carries so tests can assert it.
	w.Header().Set("X-Seen-ID", r.Header.Get("X-Correlation-ID"))
	w.WriteHeader(http.StatusOK)
}

func TestWithRequestCorrelation_SetsIDWhenAbsent(t *testing.T) {
	handler := ratelimiter.WithRequestCorrelation(nil)(http.HandlerFunc(correlationOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	id := rec.Header().Get("X-Correlation-ID")
	if id == "" {
		t.Fatal("expected a correlation ID to be set on the response, got empty string")
	}
	if rec.Header().Get("X-Seen-ID") != id {
		t.Errorf("downstream handler did not receive correlation ID: got %q want %q",
			rec.Header().Get("X-Seen-ID"), id)
	}
}

func TestWithRequestCorrelation_PreservesExistingID(t *testing.T) {
	handler := ratelimiter.WithRequestCorrelation(nil)(http.HandlerFunc(correlationOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "existing-id-123")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Correlation-ID"); got != "existing-id-123" {
		t.Errorf("expected existing ID to be preserved, got %q", got)
	}
}

func TestWithRequestCorrelation_OverwriteOption(t *testing.T) {
	cfg := ratelimiter.NewCorrelationConfig(ratelimiter.WithCorrelationOverwrite)
	handler := ratelimiter.WithRequestCorrelation(cfg)(http.HandlerFunc(correlationOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Correlation-ID", "old-id")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Correlation-ID"); got == "old-id" || got == "" {
		t.Errorf("expected a fresh ID to be generated, got %q", got)
	}
}

func TestWithRequestCorrelation_CustomHeaders(t *testing.T) {
	cfg := ratelimiter.NewCorrelationConfig(
		ratelimiter.WithCorrelationIncomingHeader("X-Request-Trace"),
		ratelimiter.WithCorrelationOutgoingHeader("X-Trace-Response"),
	)
	handler := ratelimiter.WithRequestCorrelation(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Seen-Trace", r.Header.Get("X-Request-Trace"))
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-Trace-Response") == "" {
		t.Error("expected X-Trace-Response header to be set on response")
	}
	if rec.Header().Get("X-Seen-Trace") == "" {
		t.Error("expected downstream handler to receive X-Request-Trace header")
	}
}

func TestWithRequestCorrelation_CustomGenerator(t *testing.T) {
	cfg := ratelimiter.NewCorrelationConfig(
		ratelimiter.WithCorrelationGenerator(func() string { return "static-test-id" }),
	)
	handler := ratelimiter.WithRequestCorrelation(cfg)(http.HandlerFunc(correlationOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("X-Correlation-ID"); got != "static-test-id" {
		t.Errorf("expected custom generator ID %q, got %q", "static-test-id", got)
	}
}
