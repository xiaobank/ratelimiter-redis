package ratelimiter

import (
	"net/http"
	"time"
)

// TracingConfig holds configuration for the request tracing middleware.
type TracingConfig struct {
	traceIDHeader  string
	spanIDHeader   string
	latencyHeader  string
	generator      func() string
	onTrace        func(r *http.Request, traceID, spanID string, latency time.Duration)
}

// WithTracingTraceIDHeader sets the header name used to propagate the trace ID.
func WithTracingTraceIDHeader(h string) func(*TracingConfig) {
	return func(c *TracingConfig) { c.traceIDHeader = h }
}

// WithTracingSpanIDHeader sets the header name used to propagate the span ID.
func WithTracingSpanIDHeader(h string) func(*TracingConfig) {
	return func(c *TracingConfig) { c.spanIDHeader = h }
}

// WithTracingLatencyHeader sets the response header name for elapsed time.
func WithTracingLatencyHeader(h string) func(*TracingConfig) {
	return func(c *TracingConfig) { c.latencyHeader = h }
}

// WithTracingGenerator overrides the default ID generator.
func WithTracingGenerator(g func() string) func(*TracingConfig) {
	return func(c *TracingConfig) { c.generator = g }
}

// WithTracingOnTrace registers a callback invoked after each request.
func WithTracingOnTrace(fn func(r *http.Request, traceID, spanID string, latency time.Duration)) func(*TracingConfig) {
	return func(c *TracingConfig) { c.onTrace = fn }
}

// NewTracingConfig builds a TracingConfig with the given options applied.
func NewTracingConfig(opts ...func(*TracingConfig)) *TracingConfig {
	cfg := &TracingConfig{
		traceIDHeader: "X-Trace-ID",
		spanIDHeader:  "X-Span-ID",
		latencyHeader: "X-Trace-Latency-Ms",
		generator:     defaultTracingIDGenerator,
	}
	for _, o := range opts {
		o(cfg)
	}
	return cfg
}
