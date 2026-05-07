package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/you/ratelimiter-redis"
)

// TestWithHeaderRewrite_ChainedWithRateLimiter verifies that header renaming
// works end-to-end when composed with the core rate-limiter middleware.
func TestWithHeaderRewrite_ChainedWithRateLimiter(t *testing.T) {
	store := ratelimiter.NewMemoryStore()

	rl := ratelimiter.New(
		ratelimiter.WithStore(store),
		ratelimiter.WithLimit(5),
	)

	cfg := ratelimiter.NewHeaderRewriteConfig(
		ratelimiter.WithLimitHeader("RateLimit-Limit"),
		ratelimiter.WithRemainingHeader("RateLimit-Remaining"),
		ratelimiter.WithResetHeader("RateLimit-Reset"),
	)

	handler := ratelimiter.WithHeaderRewrite(cfg)(rl(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.1:1234"
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if v := rec.Header().Get("RateLimit-Limit"); v == "" {
		t.Error("expected RateLimit-Limit header to be present")
	}
	if v := rec.Header().Get("RateLimit-Remaining"); v == "" {
		t.Error("expected RateLimit-Remaining header to be present")
	}
	if v := rec.Header().Get("X-RateLimit-Limit"); v != "" {
		t.Errorf("expected old X-RateLimit-Limit to be absent, got %q", v)
	}
}
