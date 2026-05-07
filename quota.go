package ratelimiter

import (
	"context"
	"fmt"
	"time"
)

// Quota represents a named rate limit quota with a key, limit, and window.
type Quota struct {
	Name   string
	Limit  int64
	Window time.Duration
}

// QuotaManager tracks multiple named quotas against a shared store.
type QuotaManager struct {
	store  Store
	keyFn  KeyFunc
	quotas []Quota
}

// NewQuotaManager creates a QuotaManager with the given store, key function, and quotas.
func NewQuotaManager(store Store, keyFn KeyFunc, quotas ...Quota) *QuotaManager {
	if keyFn == nil {
		keyFn = defaultKeyFunc
	}
	return &QuotaManager{
		store:  store,
		keyFn:  keyFn,
		quotas: quotas,
	}
}

// CheckAll verifies all quotas for the given request key.
// It returns the first quota that is exceeded, or nil if all pass.
func (qm *QuotaManager) CheckAll(ctx context.Context, requestKey string) (*Quota, error) {
	for i := range qm.quotas {
		q := &qm.quotas[i]
		key := fmt.Sprintf("quota:%s:%s", q.Name, requestKey)

		count, err := qm.store.Increment(ctx, key, q.Window)
		if err != nil {
			return nil, fmt.Errorf("quota %s: %w", q.Name, err)
		}

		if count > q.Limit {
			// Decrement to avoid overcounting on a blocked request.
			_ = qm.store.Decrement(ctx, key)
			return q, nil
		}
	}
	return nil, nil
}

// Remaining returns the remaining count for a named quota and request key.
func (qm *QuotaManager) Remaining(ctx context.Context, quotaName, requestKey string) (int64, error) {
	for _, q := range qm.quotas {
		if q.Name != quotaName {
			continue
		}
		key := fmt.Sprintf("quota:%s:%s", q.Name, requestKey)
		count, err := qm.store.Get(ctx, key)
		if err != nil {
			return 0, err
		}
		remaining := q.Limit - count
		if remaining < 0 {
			remaining = 0
		}
		return remaining, nil
	}
	return 0, fmt.Errorf("quota %q not found", quotaName)
}
