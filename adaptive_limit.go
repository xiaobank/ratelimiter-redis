package ratelimiter

import (
	"net/http"
	"sync"
)

// AdaptiveConfig holds configuration for adaptive rate limiting.
type AdaptiveConfig struct {
	BaseLimit    int
	MinLimit     int
	MaxLimit     int
	LoadFunc     func() float64 // returns a value between 0.0 and 1.0
	HighLoadMark float64
	LowLoadMark  float64
}

// AdaptiveLimiter adjusts the effective rate limit based on system load.
type AdaptiveLimiter struct {
	mu           sync.RWMutex
	cfg          AdaptiveConfig
	currentLimit int
}

// NewAdaptiveLimiter creates an AdaptiveLimiter with the given config.
func NewAdaptiveLimiter(cfg AdaptiveConfig) *AdaptiveLimiter {
	if cfg.MinLimit <= 0 {
		cfg.MinLimit = 1
	}
	if cfg.MaxLimit <= 0 {
		cfg.MaxLimit = cfg.BaseLimit * 2
	}
	if cfg.HighLoadMark == 0 {
		cfg.HighLoadMark = 0.8
	}
	if cfg.LowLoadMark == 0 {
		cfg.LowLoadMark = 0.4
	}
	return &AdaptiveLimiter{
		cfg:          cfg,
		currentLimit: cfg.BaseLimit,
	}
}

// CurrentLimit returns the current effective rate limit.
func (a *AdaptiveLimiter) CurrentLimit() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.currentLimit
}

// Recalculate updates the current limit based on the latest load reading.
func (a *AdaptiveLimiter) Recalculate() {
	if a.cfg.LoadFunc == nil {
		return
	}
	load := a.cfg.LoadFunc()
	a.mu.Lock()
	defer a.mu.Unlock()
	switch {
	case load >= a.cfg.HighLoadMark:
		a.currentLimit = a.cfg.MinLimit
	case load <= a.cfg.LowLoadMark:
		a.currentLimit = a.cfg.MaxLimit
	default:
		// linear interpolation between MinLimit and MaxLimit
		range_ := a.cfg.HighLoadMark - a.cfg.LowLoadMark
		ratio := (load - a.cfg.LowLoadMark) / range_
		span := float64(a.cfg.MaxLimit - a.cfg.MinLimit)
		a.currentLimit = a.cfg.MaxLimit - int(ratio*span)
	}
}

// WithAdaptiveLimit returns middleware that recalculates the limit on each
// request and injects the current limit as a request-context header so that
// downstream middleware (e.g. the core rate-limiter) can read it.
func WithAdaptiveLimit(al *AdaptiveLimiter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		al.Recalculate()
		w.Header().Set("X-RateLimit-Adaptive-Limit", itoa(al.CurrentLimit()))
		next.ServeHTTP(w, r)
	})
}

// itoa is a tiny helper to avoid importing strconv elsewhere in this file.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 10)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
