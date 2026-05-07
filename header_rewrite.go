package ratelimiter

import (
	"net/http"
	"strconv"
)

// HeaderRewriteConfig holds configuration for rewriting rate limit response headers.
type HeaderRewriteConfig struct {
	LimitHeader     string
	RemainingHeader string
	ResetHeader     string
	RetryAfterHeader string
}

// DefaultHeaderRewriteConfig returns a config using standard header names.
func DefaultHeaderRewriteConfig() HeaderRewriteConfig {
	return HeaderRewriteConfig{
		LimitHeader:      "X-RateLimit-Limit",
		RemainingHeader:  "X-RateLimit-Remaining",
		ResetHeader:      "X-RateLimit-Reset",
		RetryAfterHeader: "Retry-After",
	}
}

// headerRewriter applies a HeaderRewriteConfig to rename outgoing rate limit headers.
type headerRewriter struct {
	cfg HeaderRewriteConfig
	next http.Handler
}

func (h *headerRewriter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rw := &headerCapture{ResponseWriter: w, headers: make(http.Header)}
	h.next.ServeHTTP(rw, r)

	rewrite := map[string]string{
		"X-RateLimit-Limit":     h.cfg.LimitHeader,
		"X-RateLimit-Remaining": h.cfg.RemainingHeader,
		"X-RateLimit-Reset":     h.cfg.ResetHeader,
		"Retry-After":           h.cfg.RetryAfterHeader,
	}

	for orig, mapped := range rewrite {
		if v := rw.headers.Get(orig); v != "" && mapped != "" && mapped != orig {
			w.Header().Del(orig)
			w.Header().Set(mapped, v)
		}
	}

	for k, vs := range rw.headers {
		if w.Header().Get(k) == "" {
			for _, v := range vs {
				w.Header().Set(k, v)
			}
		}
	}

	w.WriteHeader(rw.status)
	if rw.body != nil {
		_, _ = w.Write(rw.body)
	}
}

// headerCapture buffers the response so headers can be rewritten before sending.
type headerCapture struct {
	http.ResponseWriter
	headers http.Header
	status  int
	body    []byte
}

func (c *headerCapture) Header() http.Header {
	return c.headers
}

func (c *headerCapture) WriteHeader(code int) {
	c.status = code
}

func (c *headerCapture) Write(b []byte) (int, error) {
	c.body = append(c.body, b...)
	return len(b), nil
}

// WithHeaderRewrite returns middleware that renames rate limit headers according to cfg.
func WithHeaderRewrite(cfg HeaderRewriteConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &headerRewriter{cfg: cfg, next: next}
	}
}

// setRateLimitHeaders is a helper used by counters to write standard headers.
func setRateLimitHeaders(w http.ResponseWriter, limit, remaining, resetAt int64) {
	w.Header().Set("X-RateLimit-Limit", strconv.FormatInt(limit, 10))
	w.Header().Set("X-RateLimit-Remaining", strconv.FormatInt(remaining, 10))
	w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt, 10))
}
