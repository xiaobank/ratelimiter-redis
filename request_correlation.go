package ratelimiter

import (
	"net/http"
	"strings"
)

// CorrelationConfig holds configuration for correlation ID propagation middleware.
type CorrelationConfig struct {
	// IncomingHeader is the header to read the correlation ID from (default: "X-Correlation-ID").
	IncomingHeader string
	// OutgoingHeader is the header to write the correlation ID to downstream (default: same as IncomingHeader).
	OutgoingHeader string
	// Generator is called when no correlation ID is present in the request.
	Generator func() string
	// Overwrite forces a new ID to be generated even if one is already present.
	Overwrite bool
}

// WithCorrelationIncomingHeader sets the header name to read from incoming requests.
func WithCorrelationIncomingHeader(h string) func(*CorrelationConfig) {
	return func(c *CorrelationConfig) {
		c.IncomingHeader = h
	}
}

// WithCorrelationOutgoingHeader sets the header name to write to outgoing responses.
func WithCorrelationOutgoingHeader(h string) func(*CorrelationConfig) {
	return func(c *CorrelationConfig) {
		c.OutgoingHeader = h
	}
}

// WithCorrelationGenerator sets the function used to generate new correlation IDs.
func WithCorrelationGenerator(fn func() string) func(*CorrelationConfig) {
	return func(c *CorrelationConfig) {
		c.Generator = fn
	}
}

// WithCorrelationOverwrite forces generation of a new correlation ID on every request.
func WithCorrelationOverwrite(c *CorrelationConfig) {
	c.Overwrite = true
}

// NewCorrelationConfig creates a CorrelationConfig with defaults applied and any provided options.
func NewCorrelationConfig(opts ...func(*CorrelationConfig)) *CorrelationConfig {
	c := &CorrelationConfig{
		IncomingHeader: "X-Correlation-ID",
		OutgoingHeader: "X-Correlation-ID",
		Generator:      defaultRequestIDGenerator,
	}
	for _, o := range opts {
		o(c)
	}
	if c.OutgoingHeader == "" {
		c.OutgoingHeader = c.IncomingHeader
	}
	return c
}

// WithRequestCorrelation returns middleware that reads or generates a correlation ID,
// attaches it to the request header so downstream handlers can use it, and echoes it
// back on the response.
func WithRequestCorrelation(cfg *CorrelationConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = NewCorrelationConfig()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := strings.TrimSpace(r.Header.Get(cfg.IncomingHeader))
			if id == "" || cfg.Overwrite {
				id = cfg.Generator()
				// Propagate to request so downstream middleware/handlers see it.
				r = r.Clone(r.Context())
				r.Header.Set(cfg.IncomingHeader, id)
			}
			w.Header().Set(cfg.OutgoingHeader, id)
			next.ServeHTTP(w, r)
		})
	}
}
