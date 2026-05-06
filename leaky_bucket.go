package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// LeakyBucketCounter implements a leaky bucket rate limiting algorithm.
// Requests are processed at a constant rate; excess requests are rejected.
type LeakyBucketCounter struct {
	client   *redis.Client
	rate     int
	window   time.Duration
	leakRate time.Duration
}

// NewLeakyBucketCounter creates a new leaky bucket counter.
// rate is the maximum number of requests allowed per window.
func NewLeakyBucketCounter(client *redis.Client, rate int, window time.Duration) *LeakyBucketCounter {
	return &LeakyBucketCounter{
		client:   client,
		rate:     rate,
		window:   window,
		leakRate: window / time.Duration(rate),
	}
}

// Allow checks whether a new request is permitted under the leaky bucket algorithm.
// It returns (allowed bool, remaining int, err error).
func (l *LeakyBucketCounter) Allow(ctx context.Context, key string) (bool, int, error) {
	now := time.Now().UnixNano()
	bucketKey := fmt.Sprintf("leaky:%s", key)
	lastLeakKey := fmt.Sprintf("leaky:%s:last", key)

	pipe := l.client.TxPipeline()
	getCount := pipe.Get(ctx, bucketKey)
	getLast := pipe.Get(ctx, lastLeakKey)
	_, _ = pipe.Exec(ctx)

	count, _ := getCount.Int()
	lastLeak, err := getLast.Int64()
	if err != nil {
		lastLeak = now
	}

	elapsed := time.Duration(now - lastLeak)
	leaked := int(elapsed / l.leakRate)
	if leaked > count {
		leaked = count
	}
	count -= leaked
	if count < 0 {
		count = 0
	}

	if count >= l.rate {
		return false, 0, nil
	}

	count++
	pipe = l.client.TxPipeline()
	pipe.Set(ctx, bucketKey, count, l.window)
	pipe.Set(ctx, lastLeakKey, now, l.window)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, fmt.Errorf("leaky bucket pipeline exec: %w", err)
	}

	remaining := l.rate - count
	return true, remaining, nil
}
