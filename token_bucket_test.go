package ratelimiter_test

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucketCounter_AllowsUnderLimit(t *testing.T) {
	client := newTestClient(t)
	counter := NewTokenBucketCounter(client, 10, 5, time.Minute)
	ctx := context.Background()
	key := "test:tb:under"

	for i := 0; i < 5; i++ {
		allowed, err := counter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}
}

func TestTokenBucketCounter_BlocksOverLimit(t *testing.T) {
	client := newTestClient(t)
	counter := NewTokenBucketCounter(client, 10, 3, time.Minute)
	ctx := context.Background()
	key := "test:tb:over"

	for i := 0; i < 3; i++ {
		allowed, err := counter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	allowed, err := counter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected request to be blocked after burst exhausted")
	}
}

func TestTokenBucketCounter_ReplenishesTokens(t *testing.T) {
	client := newTestClient(t)
	// 10 tokens/sec rate, burst of 1
	counter := NewTokenBucketCounter(client, 10, 1, time.Second)
	ctx := context.Background()
	key := "test:tb:replenish"

	allowed, err := counter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}

	allowed, err = counter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected second immediate request to be blocked")
	}

	// Wait for token replenishment (100ms should yield ~1 token at 10/sec)
	time.Sleep(150 * time.Millisecond)

	allowed, err = counter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Fatal("expected request to be allowed after replenishment")
	}
}

func TestTokenBucketCounter_Remaining(t *testing.T) {
	client := newTestClient(t)
	counter := NewTokenBucketCounter(client, 10, 5, time.Minute)
	ctx := context.Background()
	key := "test:tb:remaining"

	remaining, err := counter.Remaining(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 5 {
		t.Fatalf("expected 5 remaining tokens for new key, got %d", remaining)
	}

	_, _ = counter.Allow(ctx, key)
	_, _ = counter.Allow(ctx, key)

	remaining, err = counter.Remaining(ctx, key)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining > 3 {
		t.Fatalf("expected at most 3 remaining tokens after 2 requests, got %d", remaining)
	}
}
