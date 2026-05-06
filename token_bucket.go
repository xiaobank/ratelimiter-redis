package ratelimiter

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBucketCounter implements a token bucket rate limiting strategy.
// Tokens are replenished continuously over time up to a maximum burst size.
type TokenBucketCounter struct {
	client    *redis.Client
	rate      float64 // tokens per second
	burst     int     // maximum bucket capacity
	window    time.Duration
}

// NewTokenBucketCounter creates a new token bucket counter.
// rate is requests per window, burst is the maximum burst size.
func NewTokenBucketCounter(client *redis.Client, rate int, burst int, window time.Duration) *TokenBucketCounter {
	tokensPerSecond := float64(rate) / window.Seconds()
	return &TokenBucketCounter{
		client: client,
		rate:   tokensPerSecond,
		burst:  burst,
		window: window,
	}
}

// Allow checks whether a request identified by key is permitted.
// It returns true if a token is available, false otherwise.
func (t *TokenBucketCounter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now().UnixMilli()
	bucketKey := fmt.Sprintf("tb:%s", key)
	tokensKey := fmt.Sprintf("tb:%s:tokens", key)
	timestampKey := fmt.Sprintf("tb:%s:ts", key)

	script := redis.NewScript(`
		local tokens_key = KEYS[1]
		local ts_key = KEYS[2]
		local rate = tonumber(ARGV[1])
		local burst = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local ttl = tonumber(ARGV[4])

		local last_tokens = tonumber(redis.call('get', tokens_key))
		if last_tokens == nil then last_tokens = burst end

		local last_ts = tonumber(redis.call('get', ts_key))
		if last_ts == nil then last_ts = now end

		local elapsed = math.max(0, now - last_ts) / 1000.0
		local new_tokens = math.min(burst, last_tokens + elapsed * rate)

		if new_tokens < 1 then
			return 0
		end

		new_tokens = new_tokens - 1
		redis.call('set', tokens_key, new_tokens, 'PX', ttl)
		redis.call('set', ts_key, now, 'PX', ttl)
		return 1
	`)

	_ = bucketKey
	ttlMs := int64(t.window.Milliseconds() * 2)
	result, err := script.Run(ctx, t.client,
		[]string{tokensKey, timestampKey},
		t.rate, t.burst, now, ttlMs,
	).Int()
	if err != nil {
		return false, fmt.Errorf("token bucket allow: %w", err)
	}

	return result == 1, nil
}

// Remaining returns an approximate count of available tokens for the given key.
func (t *TokenBucketCounter) Remaining(ctx context.Context, key string) (int, error) {
	tokensKey := fmt.Sprintf("tb:%s:tokens", key)
	val, err := t.client.Get(ctx, tokensKey).Float64()
	if err == redis.Nil {
		return t.burst, nil
	}
	if err != nil {
		return 0, fmt.Errorf("token bucket remaining: %w", err)
	}
	return int(math.Floor(val)), nil
}
