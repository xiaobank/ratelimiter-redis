package ratelimiter_test

import (
	"context"
	"testing"
	"time"

	ratelimiter "github.com/yourusername/ratelimiter-redis"
)

func TestLeakyBucketCounter_AllowsUnderLimit(t *testing.T) {
	client := newTestClient(t)
	counter := ratelimiter.NewLeakyBucketCounter(client, 5, time.Second)
	ctx := context.Background()
	key := "test:leaky:under"

	for i := 0; i < 5; i++ {
		allowed, _, err := counter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}
}

func TestLeakyBucketCounter_BlocksOverLimit(t *testing.T) {
	client := newTestClient(t)
	counter := ratelimiter.NewLeakyBucketCounter(client, 3, time.Second)
	ctx := context.Background()
	key := "test:leaky:over"

	for i := 0; i < 3; i++ {
		allowed, _, err := counter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	allowed, remaining, err := counter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected 4th request to be blocked")
	}
	if remaining != 0 {
		t.Fatalf("expected remaining=0, got %d", remaining)
	}
}

func TestLeakyBucketCounter_LeaksOverTime(t *testing.T) {
	client := newTestClient(t)
	rate := 3
	window := 300 * time.Millisecond
	counter := ratelimiter.NewLeakyBucketCounter(client, rate, window)
	ctx := context.Background()
	key := "test:leaky:leaks"

	for i := 0; i < rate; i++ {
		_, _, err := counter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	allowed, _, _ := counter.Allow(ctx, key)
	if allowed {
		t.Fatal("expected bucket to be full before leak")
	}

	// Wait for at least one token to leak
	time.Sleep(window/time.Duration(rate) + 20*time.Millisecond)

	allowed, _, err := counter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error after leak: %v", err)
	}
	if !allowed {
		t.Fatal("expected request to be allowed after leak interval")
	}
}

func TestLeakyBucketCounter_Remaining(t *testing.T) {
	client := newTestClient(t)
	counter := ratelimiter.NewLeakyBucketCounter(client, 5, time.Second)
	ctx := context.Background()
	key := "test:leaky:remaining"

	_, remaining, err := counter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 4 {
		t.Fatalf("expected remaining=4, got %d", remaining)
	}
}
