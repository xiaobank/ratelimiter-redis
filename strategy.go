package ratelimiter

import (
	"fmt"
	"time"
)

// Strategy names understood by NewStrategy.
const (
	StrategyFixedWindow   = "fixed_window"
	StrategySlidingWindow = "sliding_window"
	StrategyTokenBucket   = "token_bucket"
	StrategyLeakyBucket   = "leaky_bucket"
	StrategyBurst         = "burst"
)

// ValidStrategies lists all supported strategy identifiers.
var ValidStrategies = []string{
	StrategyFixedWindow,
	StrategySlidingWindow,
	StrategyTokenBucket,
	StrategyLeakyBucket,
	StrategyBurst,
}

// Counter is the common interface for all rate-limiting strategies.
type Counter interface {
	Allow(ctx interface{ Value(interface{}) interface{} }, key string) (bool, int, error)
}

// StrategyConfig holds parameters for building a Counter via NewStrategy.
type StrategyConfig struct {
	Strategy string
	Store    Store
	Limit    int
	Burst    int // only used by burst strategy
	Window   time.Duration
}

// NewStrategy constructs the appropriate Counter for the given strategy name.
// Defaults to sliding_window when strategy is empty or unrecognised.
func NewStrategy(cfg StrategyConfig) (interface{}, error) {
	switch cfg.Strategy {
	case StrategyFixedWindow:
		return NewFixedWindowCounter(cfg.Store, cfg.Limit, cfg.Window), nil
	case StrategySlidingWindow, "":
		return NewSlidingWindowCounter(cfg.Store, cfg.Limit, cfg.Window), nil
	case StrategyTokenBucket:
		return NewTokenBucketCounter(cfg.Store, cfg.Limit, cfg.Window), nil
	case StrategyLeakyBucket:
		return NewLeakyBucketCounter(cfg.Store, cfg.Limit, cfg.Window), nil
	case StrategyBurst:
		burst := cfg.Burst
		if burst <= 0 {
			burst = cfg.Limit / 2
			if burst < 1 {
				burst = 1
			}
		}
		return NewBurstCounter(cfg.Store, cfg.Limit, burst, cfg.Window), nil
	default:
		return nil, fmt.Errorf("unknown strategy %q; valid strategies: %v", cfg.Strategy, ValidStrategies)
	}
}
