package ratelimiter

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisStore is a Store implementation backed by a Redis client.
type RedisStore struct {
	client *redis.Client
}

// NewRedisStore creates a new RedisStore wrapping the provided redis.Client.
func NewRedisStore(client *redis.Client) *RedisStore {
	return &RedisStore{client: client}
}

// Increment atomically increments the counter and sets the expiry on first creation.
func (r *RedisStore) Increment(ctx context.Context, key string, window time.Duration) (int64, error) {
	pipe := r.client.TxPipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// Get returns the current integer value stored at key, or 0 if not found.
func (r *RedisStore) Get(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Get(ctx, key).Int64()
	if errors.Is(err, redis.Nil) {
		return 0, nil
	}
	return val, err
}

// Set stores value at key with the given expiry duration.
func (r *RedisStore) Set(ctx context.Context, key string, value int64, expiry time.Duration) error {
	return r.client.Set(ctx, key, value, expiry).Err()
}

// SetNX sets key to value only if it does not already exist.
func (r *RedisStore) SetNX(ctx context.Context, key string, value int64, expiry time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, expiry).Result()
}

// TTL returns the remaining TTL of the key.
func (r *RedisStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	return r.client.TTL(ctx, key).Result()
}

// Delete removes the key from Redis.
func (r *RedisStore) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Ping checks that the Redis connection is alive.
func (r *RedisStore) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
