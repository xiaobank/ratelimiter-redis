package ratelimiter

import "net/http"

// BlacklistOption configures a Blacklist.
type BlacklistOption func(*Blacklist)

// WithBlacklistKeyFunc sets a custom key extraction function on the blacklist.
func WithBlacklistKeyFunc(fn func(*http.Request) string) BlacklistOption {
	return func(bl *Blacklist) {
		if fn != nil {
			bl.keyFunc = fn
		}
	}
}

// WithBlacklistEntries adds initial entries to the blacklist.
func WithBlacklistEntries(keys ...string) BlacklistOption {
	return func(bl *Blacklist) {
		bl.mu.Lock()
		defer bl.mu.Unlock()
		for _, k := range keys {
			bl.keys[k] = struct{}{}
		}
	}
}

// NewBlacklistWithOptions creates a Blacklist applying the given options.
func NewBlacklistWithOptions(opts ...BlacklistOption) *Blacklist {
	bl := &Blacklist{
		keys:    make(map[string]struct{}),
		keyFunc: defaultKeyFunc,
	}
	for _, opt := range opts {
		opt(bl)
	}
	return bl
}
