package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ratelimiter "github.com/example/ratelimiter-redis"
)

func TestWithRouteConfigs_AppliesLimitToMatchedRoute(t *testing.T) {
	store := ratelimiter.NewMemoryStore()

	configs := ratelimiter.RouteConfigMap{
		"/api/login": {
			Limit:    2,
			Window:   10 * time.Second,
			Strategy: "fixed_window",
		},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := ratelimiter.WithRouteConfigs(store, configs)(next)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i+1, rec.Code)
		}
	}

	// Third request should be blocked.
	req := httptest.NewRequest(http.MethodGet, "/api/login", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on 3rd request, got %d", rec.Code)
	}
}

func TestWithRouteConfigs_PassesThroughUnmatchedRoute(t *testing.T) {
	store := ratelimiter.NewMemoryStore()

	configs := ratelimiter.RouteConfigMap{
		"/api/login": {
			Limit:  1,
			Window: 10 * time.Second,
		},
	}

	called := 0
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusOK)
	})

	mw := ratelimiter.WithRouteConfigs(store, configs)(next)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/other", nil)
		req.RemoteAddr = "10.0.0.1:1234"
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("request %d: expected 200 for unmatched route, got %d", i+1, rec.Code)
		}
	}

	if called != 5 {
		t.Fatalf("expected next handler called 5 times, got %d", called)
	}
}

func TestWithRouteConfigs_PerRouteCustomHandler(t *testing.T) {
	store := ratelimiter.NewMemoryStore()

	customHandlerCalled := false
	configs := ratelimiter.RouteConfigMap{
		"/api/strict": {
			Limit:  1,
			Window: 10 * time.Second,
			LimitExceededHandler: func(w http.ResponseWriter, r *http.Request) {
				customHandlerCalled = true
				w.WriteHeader(http.StatusForbidden)
			},
		},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mw := ratelimiter.WithRouteConfigs(store, configs)(next)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/strict", nil)
		req.RemoteAddr = "10.0.0.2:5678"
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
	}

	if !customHandlerCalled {
		t.Fatal("expected custom limit exceeded handler to be called")
	}
}
