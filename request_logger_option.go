package ratelimiter

import "net/http"

// LoggerOption configures NewRequestLogger behaviour.
type LoggerOption func(*loggerConfig)

type loggerConfig struct {
	keyFunc KeyFunc
	logFn   LogFunc
}

// WithLogKeyFunc sets a custom KeyFunc used to extract the key for log entries.
func WithLogKeyFunc(kf KeyFunc) LoggerOption {
	return func(c *loggerConfig) {
		c.keyFunc = kf
	}
}

// WithLogFunc sets a custom LogFunc to handle log entries.
func WithLogFunc(fn LogFunc) LoggerOption {
	return func(c *loggerConfig) {
		c.logFn = fn
	}
}

// NewRequestLogger builds a request-logging middleware from functional options.
func NewRequestLogger(next http.Handler, opts ...LoggerOption) http.Handler {
	cfg := &loggerConfig{}
	for _, o := range opts {
		o(cfg)
	}
	return WithRequestLogger(next, cfg.keyFunc, cfg.logFn)
}
