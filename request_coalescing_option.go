package ratelimiter

// CoalescingOption configures a CoalescingGroup.
type CoalescingOption func(*coalescingConfig)

type coalescingConfig struct {
	keyFunc KeyFunc
}

// WithCoalescingKeyFunc sets a custom KeyFunc used to derive the coalescing
// key from an incoming request. Defaults to the IP-based key function.
func WithCoalescingKeyFunc(fn KeyFunc) CoalescingOption {
	return func(c *coalescingConfig) {
		c.keyFunc = fn
	}
}

// NewCoalescingGroupWithOptions creates a CoalescingGroup and returns both the
// group and the resolved KeyFunc so callers can wire them into middleware.
func NewCoalescingGroupWithOptions(opts ...CoalescingOption) (*CoalescingGroup, KeyFunc) {
	cfg := &coalescingConfig{
		keyFunc: defaultKeyFunc,
	}
	for _, o := range opts {
		o(cfg)
	}
	return NewCoalescingGroup(), cfg.keyFunc
}
