//go:build integration
// +build integration

package ratelimiter_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	ratelimiter "github.com/you/ratelimiter-redis"
)

func redisAddr() string {
	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		return addr
	}
	return "localhost:6379"
}

func newRedisStore(t *testing.T) *ratelimiter.RedisStore {
	t.Helper()
	client := redis.NewClient(&redis.Options{Addr: redisAddr()})
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		t.Skipf("Redis not available at %s: %v", redisAddr(), err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return ratelimiter.NewRedisStore(client)
}

func TestRedisStore_IncrementAndGet(t *testing.T) {
	s := newRedisStore(t)
	ctx := context.Background()
	key := "rl:test:incr"
	_ = s.Delete(ctx, key)

	v, err := s.Increment(ctx, key, time.Minute)
	if err != nil {
		t.Fatalf("Increment error: %v", err)
	}
	if v != 1 {
		t.Fatalf("expected 1, got %d", v)
	}

	v2, _ := s.Increment(ctx, key, time.Minute)
	if v2 != 2 {
		t.Fatalf("expected 2, got %d", v2)
	}
	_ = s.Delete(ctx, key)
}

func TestRedisStore_TTL(t *testing.T) {
	s := newRedisStore(t)
	ctx := context.Background()
	key := "rl:test:ttl"
	_ = s.Delete(ctx, key)

	_ = s.Set(ctx, key, 1, 30*time.Second)
	ttl, err := s.TTL(ctx, key)
	if err != nil {
		t.Fatalf("TTL error: %v", err)
	}
	if ttl <= 0 || ttl > 30*time.Second {
		t.Fatalf("unexpected TTL: %v", ttl)
	}
	_ = s.Delete(ctx, key)
}

func TestRedisStore_ImplementsStore(t *testing.T) {
	s := newRedisStore(t)
	var _ ratelimiter.Store = s
}
