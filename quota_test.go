package ratelimiter_test

import (
	"context"
	"testing"
	"time"

	ratelimiter "github.com/example/ratelimiter-redis"
)

func TestQuotaManager_AllowsWhenUnderAllQuotas(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	qm := ratelimiter.NewQuotaManager(store, nil,
		ratelimiter.Quota{Name: "hourly", Limit: 100, Window: time.Hour},
		ratelimiter.Quota{Name: "daily", Limit: 1000, Window: 24 * time.Hour},
	)

	exceeded, err := qm.CheckAll(context.Background(), "user:1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exceeded != nil {
		t.Errorf("expected no quota exceeded, got %q", exceeded.Name)
	}
}

func TestQuotaManager_BlocksWhenFirstQuotaExceeded(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	qm := ratelimiter.NewQuotaManager(store, nil,
		ratelimiter.Quota{Name: "burst", Limit: 2, Window: time.Minute},
	)

	ctx := context.Background()
	for i := 0; i < 2; i++ {
		if _, err := qm.CheckAll(ctx, "user:2"); err != nil {
			t.Fatalf("unexpected error on request %d: %v", i+1, err)
		}
	}

	exceeded, err := qm.CheckAll(ctx, "user:2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exceeded == nil {
		t.Fatal("expected quota to be exceeded")
	}
	if exceeded.Name != "burst" {
		t.Errorf("expected quota name %q, got %q", "burst", exceeded.Name)
	}
}

func TestQuotaManager_Remaining(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	qm := ratelimiter.NewQuotaManager(store, nil,
		ratelimiter.Quota{Name: "monthly", Limit: 10, Window: 30 * 24 * time.Hour},
	)

	ctx := context.Background()
	_, _ = qm.CheckAll(ctx, "user:3")
	_, _ = qm.CheckAll(ctx, "user:3")

	remaining, err := qm.Remaining(ctx, "monthly", "user:3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if remaining != 8 {
		t.Errorf("expected remaining 8, got %d", remaining)
	}
}

func TestQuotaManager_RemainingUnknownQuota(t *testing.T) {
	store := ratelimiter.NewMemoryStore()
	qm := ratelimiter.NewQuotaManager(store, nil)

	_, err := qm.Remaining(context.Background(), "nonexistent", "user:4")
	if err == nil {
		t.Fatal("expected error for unknown quota name")
	}
}
