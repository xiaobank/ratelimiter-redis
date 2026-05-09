package ratelimiter

import (
	"net/http"
	"strconv"
)

// PriorityLevel represents the priority tier of a request.
type PriorityLevel int

const (
	PriorityLow    PriorityLevel = 1
	PriorityNormal PriorityLevel = 5
	PriorityHigh   PriorityLevel = 10
)

// PriorityKeyFunc extracts a priority level from an HTTP request.
type PriorityKeyFunc func(r *http.Request) PriorityLevel

// PriorityConfig holds configuration for priority-based rate limiting.
type PriorityConfig struct {
	keyFunc        PriorityKeyFunc
	highLimitMult  float64
	lowLimitMult   float64
	exceededHandler http.HandlerFunc
}

// defaultPriorityKeyFunc reads priority from the X-Request-Priority header.
func defaultPriorityKeyFunc(r *http.Request) PriorityLevel {
	v := r.Header.Get("X-Request-Priority")
	if v == "" {
		return PriorityNormal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 1 {
		return PriorityNormal
	}
	return PriorityLevel(n)
}

func defaultPriorityExceededHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
}

// WithRequestPriority wraps a rate-limited handler and adjusts effective limits
// based on the priority of each incoming request. High-priority requests receive
// an increased limit multiplier; low-priority requests receive a reduced one.
func WithRequestPriority(limit int, store Store, opts ...func(*PriorityConfig)) http.Handler {
	cfg := &PriorityConfig{
		keyFunc:        defaultPriorityKeyFunc,
		highLimitMult:  2.0,
		lowLimitMult:   0.5,
		exceededHandler: defaultPriorityExceededHandler,
	}
	for _, o := range opts {
		o(cfg)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		priority := cfg.keyFunc(r)

		var effectiveLimit int
		switch {
		case priority >= PriorityHigh:
			effectiveLimit = int(float64(limit) * cfg.highLimitMult)
		case priority <= PriorityLow:
			effectiveLimit = int(float64(limit) * cfg.lowLimitMult)
		default:
			effectiveLimit = limit
		}
		if effectiveLimit < 1 {
			effectiveLimit = 1
		}

		key := "priority:" + r.RemoteAddr
		count, _ := store.Increment(r.Context(), key, windowTTL(r))

		w.Header().Set("X-RateLimit-Limit", strconv.Itoa(effectiveLimit))
		w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(max(0, effectiveLimit-int(count))))

		if int(count) > effectiveLimit {
			cfg.exceededHandler(w, r)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
}

func windowTTL(_ *http.Request) int64 { return 60 }

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
