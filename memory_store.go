package ratelimiter

import (
	"context"
	"sync"
	"time"
)

type memEntry struct {
	value  int64
	expiry time.Time
}

// MemoryStore is an in-memory Store implementation intended for testing.
type MemoryStore struct {
	mu   sync.Mutex
	data map[string]memEntry
}

// NewMemoryStore creates a new empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{data: make(map[string]memEntry)}
}

func (m *MemoryStore) get(key string) (memEntry, bool) {
	e, ok := m.data[key]
	if !ok || (!e.expiry.IsZero() && time.Now().After(e.expiry)) {
		delete(m.data, key)
		return memEntry{}, false
	}
	return e, true
}

func (m *MemoryStore) Increment(_ context.Context, key string, window time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	e, ok := m.get(key)
	if !ok {
		e = memEntry{value: 0, expiry: time.Now().Add(window)}
	}
	e.value++
	m.data[key] = e
	return e.value, nil
}

func (m *MemoryStore) Get(_ context.Context, key string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.get(key); ok {
		return e.value, nil
	}
	return 0, nil
}

func (m *MemoryStore) Set(_ context.Context, key string, value int64, expiry time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = memEntry{value: value, expiry: time.Now().Add(expiry)}
	return nil
}

func (m *MemoryStore) SetNX(_ context.Context, key string, value int64, expiry time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.get(key); ok {
		return false, nil
	}
	m.data[key] = memEntry{value: value, expiry: time.Now().Add(expiry)}
	return true, nil
}

func (m *MemoryStore) TTL(_ context.Context, key string) (time.Duration, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if e, ok := m.get(key); ok {
		return time.Until(e.expiry), nil
	}
	return -1, nil
}

func (m *MemoryStore) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
	return nil
}

func (m *MemoryStore) Ping(_ context.Context) error { return nil }
