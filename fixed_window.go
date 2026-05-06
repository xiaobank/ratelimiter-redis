package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// FixedWindowCounter implements a fixed window rate limiting strategy.
// Within each window period, requests are counted and limited to the max.
// The counter resets at the start of each new window.
type FixedWindowCounter struct {
	client    *redis.Client
	window    time.Duration
	max       int
}

// NewFixedWindowCounter creates a new FixedWindowCounter backed by Redis.
func NewFixedWindowCounter(client *redis.Client, window time.Duration, max int) *FixedWindowCounter {
	return &FixedWindowCounter{
		client: client,
		window: window,
		max:    max,
	}
}

// Allow checks whether a request identified by key is permitted under the
// fixed window policy. It returns (allowed, remaining, resetAt, error).
func (f *FixedWindowCounter) Allow(ctx context.Context, key string) (bool, int, time.Time, error) {
	now := time.Now().UTC()
	windowStart := now.Truncate(f.window)
	resetAt := windowStart.Add(f.window)

	redisKey := fmt.Sprintf("fw:%s:%d", key, windowStart.Unix())

	pipe := f.client.Pipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.ExpireAt(ctx, redisKey, resetAt)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, resetAt, fmt.Errorf("fixed window pipeline: %w", err)
	}

	count := int(incr.Val())
	remaining := f.max - count
	if remaining < 0 {
		remaining = 0
	}

	if count > f.max {
		return false, remaining, resetAt, nil
	}

	return true, remaining, resetAt, nil
}
