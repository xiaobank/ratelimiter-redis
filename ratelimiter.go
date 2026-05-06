package ratelimiter

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config holds the configuration for the rate limiter.
type Config struct {
	// Limit is the maximum number of requests allowed within the Window.
	Limit int
	// Window is the time duration for the rate limit window.
	Window time.Duration
	// KeyFunc extracts a unique key from the request (e.g., IP address or user ID).
	KeyFunc func(r *http.Request) string
	// LimitExceededHandler is called when the rate limit is exceeded.
	// Defaults to a 429 Too Many Requests response.
	LimitExceededHandler http.Handler
}

// RateLimiter is a Redis-backed HTTP rate limiter middleware.
type RateLimiter struct {
	client *redis.Client
	cfg    Config
}

// New creates a new RateLimiter with the given Redis client and config.
func New(client *redis.Client, cfg Config) *RateLimiter {
	if cfg.KeyFunc == nil {
		cfg.KeyFunc = defaultKeyFunc
	}
	if cfg.LimitExceededHandler == nil {
		cfg.LimitExceededHandler = http.HandlerFunc(defaultLimitExceededHandler)
	}
	return &RateLimiter{client: client, cfg: cfg}
}

// Middleware returns an http.Handler that enforces the rate limit.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := fmt.Sprintf("rl:%s", rl.cfg.KeyFunc(r))
		ctx := context.Background()

		count, err := rl.increment(ctx, key)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.cfg.Limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", max(0, rl.cfg.Limit-count)))

		if count > rl.cfg.Limit {
			rl.cfg.LimitExceededHandler.ServeHTTP(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// increment atomically increments the request counter and sets TTL on first request.
func (rl *RateLimiter) increment(ctx context.Context, key string) (int, error) {
	pipe := rl.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, rl.cfg.Window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return int(incr.Val()), nil
}

func defaultKeyFunc(r *http.Request) string {
	return r.RemoteAddr
}

func defaultLimitExceededHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
