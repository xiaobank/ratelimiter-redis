package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StrategyType represents the rate limiting algorithm to use.
type StrategyType string

const (
	StrategyFixedWindow   StrategyType = "fixed_window"
	StrategySlidingWindow StrategyType = "sliding_window"
	StrategyTokenBucket   StrategyType = "token_bucket"
	StrategyLeakyBucket   StrategyType = "leaky_bucket"
)

// validStrategies lists all supported strategy types for validation and documentation.
var validStrategies = []StrategyType{
	StrategyFixedWindow,
	StrategySlidingWindow,
	StrategyTokenBucket,
	StrategyLeakyBucket,
}

// Counter is the common interface implemented by all rate limiting strategies.
type Counter interface {
	Allow(ctx context.Context, key string) (allowed bool, remaining int, err error)
}

// NewStrategy returns a Counter for the given strategy type.
// Defaults to SlidingWindow if the strategy is unrecognised.
func NewStrategy(client *redis.Client, strategy StrategyType, rate int, window time.Duration) (Counter, error) {
	switch strategy {
	case StrategyFixedWindow:
		return NewFixedWindowCounter(client, rate, window), nil
	case StrategySlidingWindow:
		return NewSlidingWindowCounter(client, rate, window), nil
	case StrategyTokenBucket:
		return NewTokenBucketCounter(client, rate, window), nil
	case StrategyLeakyBucket:
		return NewLeakyBucketCounter(client, rate, window), nil
	case "":
		return NewSlidingWindowCounter(client, rate, window), nil
	default:
		return nil, fmt.Errorf("unknown strategy %q; valid options: fixed_window, sliding_window, token_bucket, leaky_bucket", strategy)
	}
}

// ValidStrategies returns a copy of all supported strategy types.
func ValidStrategies() []StrategyType {
	copy := make([]StrategyType, len(validStrategies))
	for i, s := range validStrategies {
		copy[i] = s
	}
	return copy
}
