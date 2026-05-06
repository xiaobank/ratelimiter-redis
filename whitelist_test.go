package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/example/ratelimiter-redis"
)

func TestWhitelist_ContainsReturnsTrueForKnownKey(t *testing.T) {
	wl := ratelimiter.NewWhitelist(nil, "127.0.0.1:0", "10.0.0.1:0")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:0"

	if !wl.Contains(req) {
		t.Fatal("expected key to be in whitelist")
	}
}

func TestWhitelist_ContainsReturnsFalseForUnknownKey(t *testing.T) {
	wl := ratelimiter.NewWhitelist(nil, "192.168.1.1:0")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:0"

	if wl.Contains(req) {
		t.Fatal("expected key to not be in whitelist")
	}
}

func TestWhitelist_AddAndRemove(t *testing.T) {
	wl := ratelimiter.NewWhitelist(nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "1.2.3.4:0"

	if wl.Contains(req) {
		t.Fatal("key should not be present before Add")
	}

	wl.Add("1.2.3.4:0")
	if !wl.Contains(req) {
		t.Fatal("key should be present after Add")
	}

	wl.Remove("1.2.3.4:0")
	if wl.Contains(req) {
		t.Fatal("key should not be present after Remove")
	}
}

func TestWhitelist_CustomKeyFunc(t *testing.T) {
	keyFunc := func(r *http.Request) string {
		return r.Header.Get("X-Api-Key")
	}
	wl := ratelimiter.NewWhitelist(keyFunc, "secret-token")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Api-Key", "secret-token")

	if !wl.Contains(req) {
		t.Fatal("expected custom key to be in whitelist")
	}
}

func TestWithWhitelist_BypassesForWhitelistedRequest(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	wl := ratelimiter.NewWhitelist(nil, "192.168.0.1:1234")
	handler := ratelimiter.WithWhitelist(wl, inner)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.0.1:1234"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected inner handler to be called for whitelisted request")
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestWithWhitelist_NilWhitelistPassesThrough(t *testing.T) {
	called := false
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	handler := ratelimiter.WithWhitelist(nil, inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected inner handler to be called when whitelist is nil")
	}
}
