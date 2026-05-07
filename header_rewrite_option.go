package ratelimiter

// HeaderRewriteOption is a functional option for HeaderRewriteConfig.
type HeaderRewriteOption func(*HeaderRewriteConfig)

// WithLimitHeader overrides the header name used for the rate limit ceiling.
func WithLimitHeader(name string) HeaderRewriteOption {
	return func(c *HeaderRewriteConfig) {
		c.LimitHeader = name
	}
}

// WithRemainingHeader overrides the header name used for remaining requests.
func WithRemainingHeader(name string) HeaderRewriteOption {
	return func(c *HeaderRewriteConfig) {
		c.RemainingHeader = name
	}
}

// WithResetHeader overrides the header name used for the window reset timestamp.
func WithResetHeader(name string) HeaderRewriteOption {
	return func(c *HeaderRewriteConfig) {
		c.ResetHeader = name
	}
}

// WithRetryAfterHeader overrides the header name used for retry-after seconds.
func WithRetryAfterHeader(name string) HeaderRewriteOption {
	return func(c *HeaderRewriteConfig) {
		c.RetryAfterHeader = name
	}
}

// NewHeaderRewriteConfig builds a HeaderRewriteConfig starting from defaults
// and applying the provided options.
func NewHeaderRewriteConfig(opts ...HeaderRewriteOption) HeaderRewriteConfig {
	cfg := DefaultHeaderRewriteConfig()
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}
