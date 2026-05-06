package ratelimiter

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Strategy defines the interface for a rate limiting algorithm.
// Any struct implementing Allow can be used as a pluggable strategy.
type Strategy interface {
	// Allow checks whether the request identified by key is permitted.
	// Returns: allowed, remaining tokens/requests, window reset time, error.
	Allow(ctx context.Context, key string) (allowed bool, remaining int, resetAt time.Time, err error)
}

// StrategyType enumerates the built-in rate limiting strategies.
type StrategyType string

const (
	// StrategyFixedWindow uses a fixed window counter.
	StrategyFixedWindow StrategyType = "fixed_window"
	// StrategySlidingWindow uses a sliding window counter.
	StrategySlidingWindow StrategyType = "sliding_window"
)

// NewStrategy is a convenience factory that returns a Strategy based on the
// provided StrategyType, Redis client, window duration, and request limit.
func NewStrategy(st StrategyType, client *redis.Client, window time.Duration, max int) Strategy {
	switch st {
	case StrategyFixedWindow:
		return NewFixedWindowCounter(client, window, max)
	case StrategySlidingWindow:
		return NewSlidingWindowCounter(client, window, max)
	default:
		return NewSlidingWindowCounter(client, window, max)
	}
}
