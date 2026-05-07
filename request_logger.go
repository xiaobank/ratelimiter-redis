package ratelimiter

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// LogEntry holds information about a single rate-limited request.
type LogEntry struct {
	Key       string
	Path      string
	Method    string
	Allowed   bool
	Remaining int
	Latency   time.Duration
}

// LogFunc is a callback invoked after every request processed by WithRequestLogger.
type LogFunc func(entry LogEntry)

// defaultLogFunc writes a structured log line using the standard logger.
func defaultLogFunc(e LogEntry) {
	status := "allowed"
	if !e.Allowed {
		status = "blocked"
	}
	log.Printf("[ratelimiter] %s %s key=%s status=%s remaining=%d latency=%s",
		e.Method, e.Path, e.Key, status, e.Remaining, e.Latency)
}

// loggerResponseWriter wraps http.ResponseWriter to capture the status code.
type loggerResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lw *loggerResponseWriter) WriteHeader(code int) {
	lw.statusCode = code
	lw.ResponseWriter.WriteHeader(code)
}

// WithRequestLogger wraps a handler and logs each request via logFn.
// keyFunc extracts the rate-limit key from the request for logging.
// If logFn is nil the default logger is used.
func WithRequestLogger(next http.Handler, keyFunc KeyFunc, logFn LogFunc) http.Handler {
	if logFn == nil {
		logFn = defaultLogFunc
	}
	if keyFunc == nil {
		keyFunc = defaultKeyFunc
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggerResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		key, _ := keyFunc(r)
		remaining := 0
		if v := lrw.Header().Get("X-RateLimit-Remaining"); v != "" {
			fmt.Sscanf(v, "%d", &remaining) //nolint:errcheck
		}
		allowed := lrw.statusCode != http.StatusTooManyRequests
		logFn(LogEntry{
			Key:       key,
			Path:      r.URL.Path,
			Method:    r.Method,
			Allowed:   allowed,
			Remaining: remaining,
			Latency:   time.Since(start),
		})
	})
}
