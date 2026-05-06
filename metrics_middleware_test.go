package ratelimiter

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithMetrics_CountsAllowedRequests(t *testing.T) {
	m := &Metrics{}
	handler := WithMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), m)

	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	if m.Allowed() != 5 {
		t.Fatalf("expected allowed=5, got %d", m.Allowed())
	}
	if m.Blocked() != 0 {
		t.Fatalf("expected blocked=0, got %d", m.Blocked())
	}
	if m.Total() != 5 {
		t.Fatalf("expected total=5, got %d", m.Total())
	}
}

func TestWithMetrics_CountsBlockedRequests(t *testing.T) {
	m := &Metrics{}
	handler := WithMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}), m)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if m.Blocked() != 1 {
		t.Fatalf("expected blocked=1, got %d", m.Blocked())
	}
	if m.Allowed() != 0 {
		t.Fatalf("expected allowed=0, got %d", m.Allowed())
	}
}

func TestWithMetrics_MixedRequests(t *testing.T) {
	m := &Metrics{}
	calls := 0
	handler := WithMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls%2 == 0 {
			w.WriteHeader(http.StatusTooManyRequests)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}), m)

	for i := 0; i < 4; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}

	if m.Allowed() != 2 {
		t.Fatalf("expected allowed=2, got %d", m.Allowed())
	}
	if m.Blocked() != 2 {
		t.Fatalf("expected blocked=2, got %d", m.Blocked())
	}
	if m.Total() != 4 {
		t.Fatalf("expected total=4, got %d", m.Total())
	}
}

func TestWithMetrics_PanicsOnNilMetrics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil metrics")
		}
	}()
	WithMetrics(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}), nil)
}
