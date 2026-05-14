package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

func hedgingFastHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("fast"))
}

func TestWithRequestHedging_SingleResponseReturned(t *testing.T) {
	handler := ratelimiter.WithRequestHedging(
		http.HandlerFunc(hedgingFastHandler),
		ratelimiter.WithHedgeDelay(5*time.Millisecond),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "fast" {
		t.Fatalf("expected body 'fast', got %q", rec.Body.String())
	}
}

func TestWithRequestHedging_HedgeFiresAfterDelay(t *testing.T) {
	var calls int64
	slow := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		time.Sleep(30 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	handler := ratelimiter.WithRequestHedging(
		slow,
		ratelimiter.WithHedgeDelay(10*time.Millisecond),
		ratelimiter.WithMaxHedges(1),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Both the original and the hedge should have been called.
	if atomic.LoadInt64(&calls) < 2 {
		t.Fatalf("expected at least 2 handler calls (original + hedge), got %d", calls)
	}
}

func TestWithRequestHedging_PredicateSkipsHedge(t *testing.T) {
	var calls int64
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&calls, 1)
		w.WriteHeader(http.StatusOK)
	})

	neverHedge := func(_ *http.Request) bool { return false }

	handler := ratelimiter.WithRequestHedging(
		h,
		ratelimiter.WithHedgeDelay(5*time.Millisecond),
		ratelimiter.WithHedgePredicate(neverHedge),
	)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	time.Sleep(20 * time.Millisecond) // allow any spurious hedge to fire

	if atomic.LoadInt64(&calls) != 1 {
		t.Fatalf("expected exactly 1 handler call when hedging disabled, got %d", calls)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestHedging_DefaultDelay(t *testing.T) {
	handler := ratelimiter.WithRequestHedging(http.HandlerFunc(hedgingFastHandler))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
