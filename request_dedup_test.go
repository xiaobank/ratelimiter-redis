package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ratelimiter "ratelimiter-redis"
)

func dedupOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestDedupStore_FirstRequestAllowed(t *testing.T) {
	store := ratelimiter.NewDedupStore(5 * time.Second)
	if store.IsDuplicate("fp-1") {
		t.Fatal("expected first request to be allowed")
	}
}

func TestDedupStore_SecondRequestBlocked(t *testing.T) {
	store := ratelimiter.NewDedupStore(5 * time.Second)
	store.IsDuplicate("fp-2") // record
	if !store.IsDuplicate("fp-2") {
		t.Fatal("expected duplicate to be blocked")
	}
}

func TestDedupStore_ExpiresAfterTTL(t *testing.T) {
	store := ratelimiter.NewDedupStore(50 * time.Millisecond)
	store.IsDuplicate("fp-3")
	time.Sleep(120 * time.Millisecond)
	if store.IsDuplicate("fp-3") {
		t.Fatal("expected fingerprint to have expired")
	}
}

func TestWithRequestDedup_AllowsFirstRequest(t *testing.T) {
	store := ratelimiter.NewDedupStore(5 * time.Second)
	mw := ratelimiter.WithRequestDedup(store, nil)
	h := mw(http.HandlerFunc(dedupOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithRequestDedup_BlocksDuplicateRequest(t *testing.T) {
	store := ratelimiter.NewDedupStore(5 * time.Second)
	mw := ratelimiter.WithRequestDedup(store, nil)
	h := mw(http.HandlerFunc(dedupOKHandler))

	req := httptest.NewRequest(http.MethodPost, "/submit", nil)
	req.RemoteAddr = "10.0.0.2:5678"

	h.ServeHTTP(httptest.NewRecorder(), req)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for duplicate, got %d", rec.Code)
	}
}

func TestWithRequestDedup_PanicsOnNilStore(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on nil store")
		}
	}()
	ratelimiter.WithRequestDedup(nil, nil)
}

func TestNewDedupWithOptions_CustomHandler(t *testing.T) {
	called := false
	mw := ratelimiter.NewDedupWithOptions(
		ratelimiter.WithDedupTTL(5*time.Second),
		ratelimiter.WithDedupHandler(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusTooManyRequests)
		}),
	)
	h := mw(http.HandlerFunc(dedupOKHandler))

	req := httptest.NewRequest(http.MethodGet, "/data", nil)
	req.RemoteAddr = "192.168.1.1:9999"
	h.ServeHTTP(httptest.NewRecorder(), req)

	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected custom duplicate handler to be called")
	}
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
}
