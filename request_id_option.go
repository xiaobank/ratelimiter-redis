package ratelimiter

// NewRequestIDConfig creates a RequestIDConfig with the provided options applied.
// This is useful when you want to inspect or reuse the config outside of middleware.
func NewRequestIDConfig(opts ...requestIDOption) *RequestIDConfig {
	cfg := &RequestIDConfig{
		header:    DefaultRequestIDHeader,
		generator: defaultRequestIDGenerator,
		overwrite: false,
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}

// Header returns the configured request ID header name.
func (c *RequestIDConfig) Header() string {
	return c.header
}

// Overwrite returns whether the middleware will overwrite an existing request ID.
func (c *RequestIDConfig) Overwrite() bool {
	return c.overwrite
}
