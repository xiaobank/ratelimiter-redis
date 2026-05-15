package ratelimiter

import "net/http"

// ShadowOption is a functional option for ShadowConfig.
type ShadowOption func(*ShadowConfig)

// WithShadowOnBlock sets the callback invoked when a request would have been
// blocked by the limiter in shadow mode.
func WithShadowOnBlock(fn func(r *http.Request, key string)) ShadowOption {
	return func(cfg *ShadowConfig) {
		cfg.OnBlock = fn
	}
}

// WithShadowKeyFunc sets the key extraction function used to identify the
// offending key when a shadow block occurs.
func WithShadowKeyFunc(fn func(r *http.Request) string) ShadowOption {
	return func(cfg *ShadowConfig) {
		cfg.KeyFunc = fn
	}
}

// NewShadowConfig builds a ShadowConfig from the provided options.
func NewShadowConfig(opts ...ShadowOption) ShadowConfig {
	cfg := ShadowConfig{}
	for _, o := range opts {
		o(&cfg)
	}
	return cfg
}
