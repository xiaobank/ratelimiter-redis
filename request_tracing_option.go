package ratelimiter

// DefaultTracingConfig returns a TracingConfig populated with sensible defaults.
//
//	Trace-ID header : X-Trace-ID
//	Span-ID  header : X-Span-ID
//	Latency  header : X-Trace-Latency-Ms
func DefaultTracingConfig() *TracingConfig {
	return NewTracingConfig()
}

// TracingWithB3Headers returns a TracingConfig pre-configured for the B3
// propagation format used by Zipkin / Brave.
func TracingWithB3Headers() *TracingConfig {
	return NewTracingConfig(
		WithTracingTraceIDHeader("X-B3-TraceId"),
		WithTracingSpanIDHeader("X-B3-SpanId"),
		WithTracingLatencyHeader("X-B3-Latency-Ms"),
	)
}

// TracingWithW3CHeaders returns a TracingConfig pre-configured for the W3C
// Trace Context propagation format.
func TracingWithW3CHeaders() *TracingConfig {
	return NewTracingConfig(
		WithTracingTraceIDHeader("Traceparent"),
		WithTracingSpanIDHeader("Tracestate"),
		WithTracingLatencyHeader("X-Trace-Latency-Ms"),
	)
}
