package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	ratelimiter "."
)

func priorityRequest(priority int) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:9999"
	if priority > 0 {
		req.Header.Set("X-Request-Priority", strconv.Itoa(priority))
	}
	return req
}

func TestWithRequestPriority_NormalPriorityUsesBaseLimit(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	h := ratelimiter.WithRequestPriority(3, store)

	for i := 1; i <= 3; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, priorityRequest(5))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
	}
}

func TestWithRequestPriority_HighPriorityGetsMoreRequests(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	// base limit 3, high mult 2.0 => effective 6
	h := ratelimiter.WithRequestPriority(3, store,
		ratelimiter.WithHighLimitMultiplier(2.0),
	)

	for i := 1; i <= 6; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, priorityRequest(10))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, priorityRequest(10))
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exceeding high-priority limit, got %d", w.Code)
	}
}

func TestWithRequestPriority_LowPriorityGetsFewerRequests(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	// base limit 4, low mult 0.5 => effective 2
	h := ratelimiter.WithRequestPriority(4, store,
		ratelimiter.WithLowLimitMultiplier(0.5),
	)

	for i := 1; i <= 2; i++ {
		w := httptest.NewRecorder()
		h.ServeHTTP(w, priorityRequest(1))
		if w.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, w.Code)
		}
	}

	w := httptest.NewRecorder()
	h.ServeHTTP(w, priorityRequest(1))
	if w.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after exceeding low-priority limit, got %d", w.Code)
	}
}

func TestWithRequestPriority_SetsRateLimitHeaders(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	h := ratelimiter.WithRequestPriority(5, store)

	w := httptest.NewRecorder()
	h.ServeHTTP(w, priorityRequest(5))

	if w.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("expected X-RateLimit-Limit header to be set")
	}
	if w.Header().Get("X-RateLimit-Remaining") == "" {
		t.Error("expected X-RateLimit-Remaining header to be set")
	}
}

func TestWithRequestPriority_CustomExceededHandler(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	customCalled := false
	h := ratelimiter.WithRequestPriority(1, store,
		ratelimiter.WithPriorityExceededHandler(func(w http.ResponseWriter, r *http.Request) {
			customCalled = true
			w.WriteHeader(http.StatusServiceUnavailable)
		}),
	)

	h.ServeHTTP(httptest.NewRecorder(), priorityRequest(5))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, priorityRequest(5))

	if !customCalled {
		t.Error("expected custom exceeded handler to be called")
	}
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}
