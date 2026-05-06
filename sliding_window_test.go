package ratelimiter

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestSlidingWindowCounter_AllowsUnderLimit(t *testing.T) {
	client := newTestClient(t)
	sw := NewSlidingWindowCounter(client)
	ctx := context.Background()
	key := fmt.Sprintf("test:sliding:under:%d", time.Now().UnixNano())

	for i := 0; i < 5; i++ {
		allowed, count, err := sw.Allow(ctx, key, 10, time.Minute)
		if err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
		if !allowed {
			t.Errorf("request %d should be allowed, count=%d", i+1, count)
		}
	}
}

func TestSlidingWindowCounter_BlocksOverLimit(t *testing.T) {
	client := newTestClient(t)
	sw := NewSlidingWindowCounter(client)
	ctx := context.Background()
	key := fmt.Sprintf("test:sliding:over:%d", time.Now().UnixNano())
	max := 3

	for i := 0; i < max; i++ {
		allowed, _, err := sw.Allow(ctx, key, max, time.Minute)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Errorf("request %d should be allowed", i+1)
		}
	}

	allowed, count, err := sw.Allow(ctx, key, max, time.Minute)
	if err != nil {
		t.Fatalf("unexpected error on excess request: %v", err)
	}
	if allowed {
		t.Errorf("excess request should be blocked, count=%d", count)
	}
}

func TestSlidingWindowCounter_ResetsAfterWindow(t *testing.T) {
	client := newTestClient(t)
	sw := NewSlidingWindowCounter(client)
	ctx := context.Background()
	key := fmt.Sprintf("test:sliding:reset:%d", time.Now().UnixNano())
	window := 200 * time.Millisecond

	allowed, _, err := sw.Allow(ctx, key, 1, window)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("first request should be allowed")
	}

	// Wait for the window to expire
	time.Sleep(window + 50*time.Millisecond)

	allowed, _, err = sw.Allow(ctx, key, 1, window)
	if err != nil {
		t.Fatalf("unexpected error after reset: %v", err)
	}
	if !allowed {
		t.Error("request after window reset should be allowed")
	}
}
