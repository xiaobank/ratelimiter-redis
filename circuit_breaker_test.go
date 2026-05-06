package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCircuitBreaker_InitiallyClosedState(t *testing.T) {
	cb := NewCircuitBreaker()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected Closed, got %v", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	cb := NewCircuitBreaker(WithTripThreshold(3))
	for i := 0; i < 3; i++ {
		cb.RecordBlocked()
	}
	if cb.State() != CircuitOpen {
		t.Fatalf("expected Open after threshold, got %v", cb.State())
	}
}

func TestCircuitBreaker_ResetsOnAllowed(t *testing.T) {
	cb := NewCircuitBreaker(WithTripThreshold(3))
	cb.RecordBlocked()
	cb.RecordBlocked()
	cb.RecordAllowed()
	if cb.State() != CircuitClosed {
		t.Fatalf("expected Closed after allowed, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenAfterReset(t *testing.T) {
	cb := NewCircuitBreaker(
		WithTripThreshold(1),
		WithResetAfter(10*time.Millisecond),
	)
	cb.RecordBlocked()
	time.Sleep(20 * time.Millisecond)
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected HalfOpen after reset window, got %v", cb.State())
	}
}

func TestCircuitBreaker_OnTripCallback(t *testing.T) {
	tripped := make(chan struct{}, 1)
	cb := NewCircuitBreaker(
		WithTripThreshold(2),
		WithOnTrip(func() { tripped <- struct{}{} }),
	)
	cb.RecordBlocked()
	cb.RecordBlocked()
	select {
	case <-tripped:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("onTrip callback was not called")
	}
}

func TestWithCircuitBreakerMiddleware_BlocksWhenOpen(t *testing.T) {
	cb := NewCircuitBreaker(WithTripThreshold(1))
	cb.RecordBlocked() // trip immediately

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := WithCircuitBreakerMiddleware(inner, cb)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", rec.Code)
	}
}

func TestWithCircuitBreakerMiddleware_RecordsBlocked429(t *testing.T) {
	cb := NewCircuitBreaker(WithTripThreshold(3))

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	})
	handler := WithCircuitBreakerMiddleware(inner, cb)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	if cb.State() != CircuitOpen {
		t.Fatalf("expected circuit Open after 3 blocked responses, got %v", cb.State())
	}
}
