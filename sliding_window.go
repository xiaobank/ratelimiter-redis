package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// SlidingWindowCounter implements a sliding window rate limiting algorithm
// using a Redis sorted set to track request timestamps.
type SlidingWindowCounter struct {
	client *redis.Client
}

// NewSlidingWindowCounter creates a new SlidingWindowCounter backed by the given Redis client.
func NewSlidingWindowCounter(client *redis.Client) *SlidingWindowCounter {
	return &SlidingWindowCounter{client: client}
}

// Allow checks whether a request identified by key is allowed under the sliding window
// rate limit of max requests per window duration. It returns (allowed, current count, error).
func (s *SlidingWindowCounter) Allow(ctx context.Context, key string, max int, window time.Duration) (bool, int64, error) {
	now := time.Now()
	windowStart := now.Add(-window)

	pipe := s.client.TxPipeline()

	// Remove entries outside the current window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixMilli()))

	// Add current request with timestamp as score
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixMilli()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Count requests in the current window
	countCmd := pipe.ZCard(ctx, key)

	// Set key expiry to avoid stale keys
	pipe.Expire(ctx, key, window+time.Second)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, fmt.Errorf("sliding window pipeline exec: %w", err)
	}

	count := countCmd.Val()
	if count > int64(max) {
		return false, count, nil
	}

	return true, count, nil
}
