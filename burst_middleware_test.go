package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func TestWithBurst_AllowsRequestsWithinBurst(t *testing.T) {
	store := NewMemoryStore()
	counter := NewBurstCounter(store, 2, 3, time.Minute)
	mw := WithBurst(BurstConfig{Counter: counter})
	handler := mw(okHandler())

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}
}

func TestWithBurst_BlocksOverBurst(t *testing.T) {
	store := NewMemoryStore()
	counter := NewBurstCounter(store, 2, 1, time.Minute)
	mw := WithBurst(BurstConfig{Counter: counter})
	handler := mw(okHandler())

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.0.0.2:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.2:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}

func TestWithBurst_SetsBurstHeaders(t *testing.T) {
	store := NewMemoryStore()
	counter := NewBurstCounter(store, 5, 3, time.Minute)
	mw := WithBurst(BurstConfig{Counter: counter})
	handler := mw(okHandler())

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.3:1234"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("X-RateLimit-Limit") != strconv.Itoa(8) {
		t.Fatalf("expected X-RateLimit-Limit=8, got %s", rec.Header().Get("X-RateLimit-Limit"))
	}
	if rec.Header().Get("X-RateLimit-Burst") != strconv.Itoa(3) {
		t.Fatalf("expected X-RateLimit-Burst=3, got %s", rec.Header().Get("X-RateLimit-Burst"))
	}
}
