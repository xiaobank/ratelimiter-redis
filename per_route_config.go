package ratelimiter

import (
	"net/http"
	"time"
)

// RouteConfig holds rate limiting configuration for a specific route or group.
type RouteConfig struct {
	// Limit is the maximum number of requests allowed within the Window.
	Limit int
	// Window is the duration of the rate limit window.
	Window time.Duration
	// Strategy overrides the global strategy for this route.
	// If empty, the global strategy is used.
	Strategy string
	// KeyFunc overrides the global key function for this route.
	// If nil, the global key function is used.
	KeyFunc func(r *http.Request) string
	// LimitExceededHandler overrides the global handler for this route.
	// If nil, the global handler is used.
	LimitExceededHandler http.HandlerFunc
}

// RouteConfigMap maps route identifiers (e.g. "/api/login") to their RouteConfig.
type RouteConfigMap map[string]RouteConfig

// WithRouteConfigs returns a new middleware that applies per-route rate limit
// configurations. Routes not present in the map fall through to the next handler
// without rate limiting applied by this middleware.
func WithRouteConfigs(store Store, configs RouteConfigMap) func(http.Handler) http.Handler {
	// Pre-build a limiter per route to avoid rebuilding on every request.
	limiters := make(map[string]*Limiter, len(configs))
	for route, cfg := range configs {
		cfg := cfg // capture

		opts := []Option{}
		if cfg.Strategy != "" {
			opts = append(opts, WithStrategy(cfg.Strategy))
		}
		if cfg.KeyFunc != nil {
			opts = append(opts, WithKeyFunc(cfg.KeyFunc))
		}
		if cfg.LimitExceededHandler != nil {
			opts = append(opts, WithLimitExceededHandler(cfg.LimitExceededHandler))
		}

		limiters[route] = New(store, cfg.Limit, cfg.Window, opts...)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter, ok := limiters[r.URL.Path]; ok {
				limiter.Handler(next).ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
