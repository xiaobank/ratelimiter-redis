package ratelimiter

import (
	"net/http"
	"time"
)

// EnrichmentConfig holds configuration for the request enrichment middleware.
type EnrichmentConfig struct {
	// RequestTimeHeader is the header name to set with the time the request was received.
	RequestTimeHeader string
	// LatencyHeader is the header name to set on the response with the total processing latency.
	LatencyHeader string
	// ServerHeader is an optional static value to set on the response as a server identifier.
	ServerHeader string
	// ServerHeaderName is the header name used when ServerHeader is set.
	ServerHeaderName string
}

// DefaultEnrichmentConfig returns an EnrichmentConfig with sensible defaults.
func DefaultEnrichmentConfig() *EnrichmentConfig {
	return &EnrichmentConfig{
		RequestTimeHeader: "X-Request-Time",
		LatencyHeader:     "X-Response-Latency-Ms",
		ServerHeaderName:  "X-Served-By",
	}
}

// enrichmentResponseWriter wraps http.ResponseWriter to allow header injection after the handler runs.
type enrichmentResponseWriter struct {
	http.ResponseWriter
	latencyHeader string
	start         time.Time
	serverHeader  string
	serverName    string
	wroteHeader   bool
}

func (w *enrichmentResponseWriter) WriteHeader(code int) {
	if !w.wroteHeader {
		w.wroteHeader = true
		latency := time.Since(w.start).Milliseconds()
		w.ResponseWriter.Header().Set(w.latencyHeader, itoa(int(latency)))
		if w.serverName != "" {
			w.ResponseWriter.Header().Set(w.serverHeader, w.serverName)
		}
	}
	w.ResponseWriter.WriteHeader(code)
}

func (w *enrichmentResponseWriter) Write(b []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(b)
}

// WithRequestEnrichment returns middleware that stamps each request and response
// with timing and optional server identity headers.
func WithRequestEnrichment(cfg *EnrichmentConfig) func(http.Handler) http.Handler {
	if cfg == nil {
		cfg = DefaultEnrichmentConfig()
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			now := time.Now().UTC()
			if cfg.RequestTimeHeader != "" {
				r.Header.Set(cfg.RequestTimeHeader, now.Format(time.RFC3339Nano))
			}
			erw := &enrichmentResponseWriter{
				ResponseWriter: w,
				latencyHeader:  cfg.LatencyHeader,
				start:          now,
				serverHeader:   cfg.ServerHeaderName,
				serverName:     cfg.ServerHeader,
			}
			next.ServeHTTP(erw, r)
			if !erw.wroteHeader {
				erw.WriteHeader(http.StatusOK)
			}
		})
	}
}
