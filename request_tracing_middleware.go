package ratelimiter

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

func defaultTracingIDGenerator() string {
	return fmt.Sprintf("%016x", rand.Int63())
}

// WithRequestTracing attaches trace and span IDs to every request and records
// per-request latency. If an incoming trace ID header is already present it is
// preserved; a fresh span ID is always generated.
func WithRequestTracing(cfg *TracingConfig, next http.Handler) http.Handler {
	if cfg == nil {
		cfg = NewTracingConfig()
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		traceID := r.Header.Get(cfg.traceIDHeader)
		if traceID == "" {
			traceID = cfg.generator()
		}
		spanID := cfg.generator()

		// Propagate IDs on the outbound request headers so downstream services
		// can join the trace.
		r.Header.Set(cfg.traceIDHeader, traceID)
		r.Header.Set(cfg.spanIDHeader, spanID)

		// Expose IDs on the response so clients can correlate.
		w.Header().Set(cfg.traceIDHeader, traceID)
		w.Header().Set(cfg.spanIDHeader, spanID)

		next.ServeHTTP(w, r)

		latency := time.Since(start)
		w.Header().Set(cfg.latencyHeader, strconv.FormatInt(latency.Milliseconds(), 10))

		if cfg.onTrace != nil {
			cfg.onTrace(r, traceID, spanID, latency)
		}
	})
}
