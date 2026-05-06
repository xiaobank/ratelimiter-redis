package ratelimiter_test

import (
	"context"
	"testing"
	"time"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

func TestNewStrategy_ReturnsFixedWindow(t *testing.T) {
	client := newTestClient(t)
	c, err := ratelimiter.NewStrategy(client, ratelimiter.StrategyFixedWindow, 10, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil counter")
	}
}

func TestNewStrategy_ReturnsSlidingWindow(t *testing.T) {
	client := newTestClient(t)
	c, err := ratelimiter.NewStrategy(client, ratelimiter.StrategySlidingWindow, 10, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil counter")
	}
}

func TestNewStrategy_ReturnsTokenBucket(t *testing.T) {
	client := newTestClient(t)
	c, err := ratelimiter.NewStrategy(client, ratelimiter.StrategyTokenBucket, 10, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil counter")
	}
}

func TestNewStrategy_ReturnsLeakyBucket(t *testing.T) {
	client := newTestClient(t)
	c, err := ratelimiter.NewStrategy(client, ratelimiter.StrategyLeakyBucket, 10, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil counter")
	}
}

func TestNewStrategy_DefaultsToSlidingWindow(t *testing.T) {
	client := newTestClient(t)
	c, err := ratelimiter.NewStrategy(client, "", 10, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil counter")
	}
}

func TestNewStrategy_ReturnsErrorForUnknown(t *testing.T) {
	client := newTestClient(t)
	_, err := ratelimiter.NewStrategy(client, "unknown_algo", 10, time.Second)
	if err == nil {
		t.Fatal("expected error for unknown strategy")
	}
}

func TestStrategy_LeakyBucketAllow(t *testing.T) {
	client := newTestClient(t)
	c, err := ratelimiter.NewStrategy(client, ratelimiter.StrategyLeakyBucket, 5, time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := context.Background()
	allowed, _, err := c.Allow(ctx, "strategy:leaky:test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}
}
