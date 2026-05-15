package ratelimiter_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	ratelimiter "ratelimiter-redis"
)

// ExampleWithRequestTracing demonstrates how to attach distributed tracing
// headers to every request passing through the middleware.
func ExampleWithRequestTracing() {
	cfg := ratelimiter.NewTracingConfig(
		ratelimiter.WithTracingTraceIDHeader("X-Trace-ID"),
		ratelimiter.WithTracingSpanIDHeader("X-Span-ID"),
		ratelimiter.WithTracingLatencyHeader("X-Trace-Latency-Ms"),
		ratelimiter.WithTracingGenerator(func() string { return "fixed-id" }),
		ratelimiter.WithTracingOnTrace(func(r *http.Request, traceID, spanID string, latency time.Duration) {
			// emit to your tracing backend here
		}),
	)

	handler := ratelimiter.WithRequestTracing(cfg, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	handler.ServeHTTP(rec, req)

	fmt.Println(rec.Header().Get("X-Trace-ID"))
	fmt.Println(rec.Header().Get("X-Span-ID"))
	// Output:
	// fixed-id
	// fixed-id
}
