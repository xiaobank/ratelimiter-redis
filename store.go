package ratelimiter

import (
	"context"
	"time"
)

// Store defines the interface for a rate limiter backend.
// Any Redis client wrapper or mock must implement this interface.
type Store interface {
	// Increment increments the counter for key by 1 and sets expiry if the key is new.
	// Returns the new count and any error.
	Increment(ctx context.Context, key string, window time.Duration) (int64, error)

	// Get returns the current count for the given key.
	Get(ctx context.Context, key string) (int64, error)

	// Set sets an arbitrary integer value for a key with expiry.
	Set(ctx context.Context, key string, value int64, expiry time.Duration) error

	// SetNX sets a key only if it does not exist. Returns true if the key was set.
	SetNX(ctx context.Context, key string, value int64, expiry time.Duration) (bool, error)

	// TTL returns the remaining time-to-live for a key.
	TTL(ctx context.Context, key string) (time.Duration, error)

	// Delete removes a key from the store.
	Delete(ctx context.Context, key string) error

	// Ping checks connectivity to the store.
	Ping(ctx context.Context) error
}
