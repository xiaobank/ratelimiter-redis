package ratelimiter_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	ratelimiter "github.com/example/ratelimiter-redis"
)

func cacheOKHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(body)) //nolint:errcheck
	})
}

func TestResponseCache_MissOnFirstRequest(t *testing.T) {
	cache := ratelimiter.NewResponseCache(5 * time.Second)
	mw := ratelimiter.WithResponseCache(cache, nil)(cacheOKHandler("hello"))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if rec.Header().Get("X-Cache") != "MISS" {
		t.Errorf("expected X-Cache: MISS, got %s", rec.Header().Get("X-Cache"))
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestResponseCache_HitOnSecondRequest(t *testing.T) {
	cache := ratelimiter.NewResponseCache(5 * time.Second)
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("cached-body")) //nolint:errcheck
	})
	mw := ratelimiter.WithResponseCache(cache, nil)(handler)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/resource", nil)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
		body, _ := io.ReadAll(rec.Body)
		if string(body) != "cached-body" {
			t.Errorf("unexpected body: %s", body)
		}
	}
	if calls != 1 {
		t.Errorf("expected handler called once, got %d", calls)
	}
}

func TestResponseCache_DoesNotCacheNonGET(t *testing.T) {
	cache := ratelimiter.NewResponseCache(5 * time.Second)
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	})
	mw := ratelimiter.WithResponseCache(cache, nil)(handler)

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodPost, "/resource", nil)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)
	}
	if calls != 3 {
		t.Errorf("expected handler called 3 times for POST, got %d", calls)
	}
}

func TestResponseCache_ExpiresEntries(t *testing.T) {
	cache := ratelimiter.NewResponseCache(50 * time.Millisecond)
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data")) //nolint:errcheck
	})
	mw := ratelimiter.WithResponseCache(cache, nil)(handler)

	req1 := httptest.NewRequest(http.MethodGet, "/expire", nil)
	mw.ServeHTTP(httptest.NewRecorder(), req1)

	time.Sleep(80 * time.Millisecond)

	req2 := httptest.NewRequest(http.MethodGet, "/expire", nil)
	rec2 := httptest.NewRecorder()
	mw.ServeHTTP(rec2, req2)

	if calls != 2 {
		t.Errorf("expected 2 handler calls after expiry, got %d", calls)
	}
	if rec2.Header().Get("X-Cache") != "MISS" {
		t.Errorf("expected MISS after expiry")
	}
}

func TestResponseCache_PanicsOnNilCache(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nil cache")
		}
	}()
	ratelimiter.WithResponseCache(nil, nil)
}

func TestNewResponseCacheWithOptions_DefaultTTL(t *testing.T) {
	cache, keyFunc := ratelimiter.NewResponseCacheWithOptions()
	if cache == nil {
		t.Fatal("expected non-nil cache")
	}
	if keyFunc == nil {
		t.Fatal("expected non-nil keyFunc")
	}
}
