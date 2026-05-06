package ratelimiter

import (
	"net/http"
)

// MetricsMiddleware wraps an existing http.Handler and records rate-limit
// metrics by inspecting the response status code written downstream.
//
// Usage:
//
//	m := &Metrics{}
//	handler = WithMetrics(handler, m)
func WithMetrics(next http.Handler, metrics *Metrics) http.Handler {
	if metrics == nil {
		panic("ratelimiter: metrics must not be nil")
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rw := &metricsResponseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rw, r)
		if rw.code == http.StatusTooManyRequests {
			metrics.recordBlocked()
		} else {
			metrics.recordAllowed()
		}
	})
}

// metricsResponseWriter captures the status code written by the handler.
type metricsResponseWriter struct {
	http.ResponseWriter
	code    int
	wrote   bool
}

func (rw *metricsResponseWriter) WriteHeader(code int) {
	if !rw.wrote {
		rw.code = code
		rw.wrote = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *metricsResponseWriter) Write(b []byte) (int, error) {
	if !rw.wrote {
		rw.code = http.StatusOK
		rw.wrote = true
	}
	return rw.ResponseWriter.Write(b)
}
