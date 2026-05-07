package ratelimiter_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	ratelimiter "github.com/you/ratelimiter-redis"
)

func headerOKHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-RateLimit-Limit", "100")
	w.Header().Set("X-RateLimit-Remaining", "42")
	w.Header().Set("X-RateLimit-Reset", "1700000000")
	w.Header().Set("Retry-After", "30")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func TestWithHeaderRewrite_DefaultConfig_PassesThrough(t *testing.T) {
	cfg := ratelimiter.DefaultHeaderRewriteConfig()
	mw := ratelimiter.WithHeaderRewrite(cfg)(http.HandlerFunc(headerOKHandler))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-RateLimit-Limit") != "100" {
		t.Errorf("expected X-RateLimit-Limit=100, got %q", rec.Header().Get("X-RateLimit-Limit"))
	}
}

func TestWithHeaderRewrite_RenamesLimitHeader(t *testing.T) {
	cfg := ratelimiter.NewHeaderRewriteConfig(
		ratelimiter.WithLimitHeader("RateLimit-Limit"),
	)
	mw := ratelimiter.WithHeaderRewrite(cfg)(http.HandlerFunc(headerOKHandler))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("RateLimit-Limit"); got != "100" {
		t.Errorf("expected RateLimit-Limit=100, got %q", got)
	}
	if got := rec.Header().Get("X-RateLimit-Limit"); got != "" {
		t.Errorf("expected old header to be absent, got %q", got)
	}
}

func TestWithHeaderRewrite_RenamesRemainingHeader(t *testing.T) {
	cfg := ratelimiter.NewHeaderRewriteConfig(
		ratelimiter.WithRemainingHeader("RateLimit-Remaining"),
	)
	mw := ratelimiter.WithHeaderRewrite(cfg)(http.HandlerFunc(headerOKHandler))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("RateLimit-Remaining"); got != "42" {
		t.Errorf("expected RateLimit-Remaining=42, got %q", got)
	}
}

func TestWithHeaderRewrite_RenamesRetryAfterHeader(t *testing.T) {
	cfg := ratelimiter.NewHeaderRewriteConfig(
		ratelimiter.WithRetryAfterHeader("X-Retry-After"),
	)
	mw := ratelimiter.WithHeaderRewrite(cfg)(http.HandlerFunc(headerOKHandler))

	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if got := rec.Header().Get("X-Retry-After"); got != "30" {
		t.Errorf("expected X-Retry-After=30, got %q", got)
	}
	if got := rec.Header().Get("Retry-After"); got != "" {
		t.Errorf("expected original Retry-After to be absent, got %q", got)
	}
}

func TestNewHeaderRewriteConfig_AppliesMultipleOptions(t *testing.T) {
	cfg := ratelimiter.NewHeaderRewriteConfig(
		ratelimiter.WithLimitHeader("RL-Limit"),
		ratelimiter.WithRemainingHeader("RL-Remaining"),
		ratelimiter.WithResetHeader("RL-Reset"),
		ratelimiter.WithRetryAfterHeader("RL-Retry-After"),
	)

	if cfg.LimitHeader != "RL-Limit" {
		t.Errorf("LimitHeader: got %q", cfg.LimitHeader)
	}
	if cfg.RemainingHeader != "RL-Remaining" {
		t.Errorf("RemainingHeader: got %q", cfg.RemainingHeader)
	}
	if cfg.ResetHeader != "RL-Reset" {
		t.Errorf("ResetHeader: got %q", cfg.ResetHeader)
	}
	if cfg.RetryAfterHeader != "RL-Retry-After" {
		t.Errorf("RetryAfterHeader: got %q", cfg.RetryAfterHeader)
	}
}
