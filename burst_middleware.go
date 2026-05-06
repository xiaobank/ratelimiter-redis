package ratelimiter

import (
	"net/http"
	"strconv"
)

// BurstConfig holds configuration for the burst-aware middleware.
type BurstConfig struct {
	Counter        *BurstCounter
	KeyFunc        KeyFunc
	LimitExceeded  http.HandlerFunc
}

// WithBurst returns HTTP middleware that enforces a burst-capable rate limit.
// Requests within limit+burst are allowed; excess requests receive 429.
func WithBurst(cfg BurstConfig) func(http.Handler) http.Handler {
	keyFunc := cfg.KeyFunc
	if keyFunc == nil {
		keyFunc = defaultKeyFunc
	}

	limitExceeded := cfg.LimitExceeded
	if limitExceeded == nil {
		limitExceeded = defaultLimitExceededHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key, err := keyFunc(r)
			if err != nil {
				http.Error(w, "rate limiter key error", http.StatusInternalServerError)
				return
			}

			allowed, remaining, err := cfg.Counter.Allow(r.Context(), key)
			if err != nil {
				http.Error(w, "rate limiter error", http.StatusInternalServerError)
				return
			}

			effective := cfg.Counter.Limit() + cfg.Counter.Burst()
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(effective))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Burst", strconv.Itoa(cfg.Counter.Burst()))

			if !allowed {
				limitExceeded(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
