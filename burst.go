package ratelimiter

import (
	"context"
	"fmt"
	"time"
)

// BurstCounter allows a configurable burst above the base rate limit
// for short periods, backed by a Redis store.
type BurstCounter struct {
	store     Store
	limit     int
	burst     int
	window    time.Duration
}

// NewBurstCounter creates a rate limiter that permits up to limit requests
// per window, with an additional burst allowance of burst extra requests.
func NewBurstCounter(store Store, limit, burst int, window time.Duration) *BurstCounter {
	return &BurstCounter{
		store:  store,
		limit:  limit,
		burst:  burst,
		window: window,
	}
}

// Allow checks whether the request identified by key is within the
// combined limit+burst quota for the current window.
func (b *BurstCounter) Allow(ctx context.Context, key string) (bool, int, error) {
	effectiveLimit := b.limit + b.burst
	windowKey := fmt.Sprintf("%s:%d", key, time.Now().Truncate(b.window).Unix())

	count, err := b.store.Increment(ctx, windowKey, b.window)
	if err != nil {
		return false, 0, fmt.Errorf("burst counter increment: %w", err)
	}

	remaining := effectiveLimit - int(count)
	if remaining < 0 {
		remaining = 0
	}

	if int(count) > effectiveLimit {
		return false, remaining, nil
	}
	return true, remaining, nil
}

// Limit returns the base rate limit (excluding burst).
func (b *BurstCounter) Limit() int { return b.limit }

// Burst returns the additional burst allowance.
func (b *BurstCounter) Burst() int { return b.burst }
