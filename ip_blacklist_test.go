package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/example/ratelimiter-redis"
)

func blacklistOKHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestBlacklist_ContainsReturnsTrueForKnownKey(t *testing.T) {
	bl := ratelimiter.NewBlacklist(nil, "1.2.3.4")
	req := newRequest("1.2.3.4:9000", "", "")
	if !bl.Contains(req) {
		t.Fatal("expected blacklist to contain 1.2.3.4")
	}
}

func TestBlacklist_ContainsReturnsFalseForUnknownKey(t *testing.T) {
	bl := ratelimiter.NewBlacklist(nil, "1.2.3.4")
	req := newRequest("9.9.9.9:9000", "", "")
	if bl.Contains(req) {
		t.Fatal("expected blacklist not to contain 9.9.9.9")
	}
}

func TestBlacklist_AddAndRemove(t *testing.T) {
	bl := ratelimiter.NewBlacklist(nil)
	req := newRequest("5.5.5.5:1234", "", "")

	if bl.Contains(req) {
		t.Fatal("should not contain key before Add")
	}

	bl.Add("5.5.5.5")
	if !bl.Contains(req) {
		t.Fatal("should contain key after Add")
	}

	bl.Remove("5.5.5.5")
	if bl.Contains(req) {
		t.Fatal("should not contain key after Remove")
	}
}

func TestWithBlacklist_BlocksBlacklistedRequest(t *testing.T) {
	bl := ratelimiter.NewBlacklist(nil, "10.0.0.1")
	mw := ratelimiter.WithBlacklist(bl)
	handler := mw(http.HandlerFunc(blacklistOKHandler))

	req := newRequest("10.0.0.1:1234", "", "")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
}

func TestWithBlacklist_PassesThroughNonBlacklistedRequest(t *testing.T) {
	bl := ratelimiter.NewBlacklist(nil, "10.0.0.1")
	mw := ratelimiter.WithBlacklist(bl)
	handler := mw(http.HandlerFunc(blacklistOKHandler))

	req := newRequest("192.168.1.1:5678", "", "")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestBlacklist_CustomKeyFunc(t *testing.T) {
	keyFunc := func(r *http.Request) string {
		return r.Header.Get("X-User-ID")
	}
	bl := ratelimiter.NewBlacklist(keyFunc, "banned-user")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-User-ID", "banned-user")

	if !bl.Contains(req) {
		t.Fatal("expected custom key func to match banned-user")
	}

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-User-ID", "allowed-user")
	if bl.Contains(req2) {
		t.Fatal("expected allowed-user not to be blacklisted")
	}
}
