package ratelimiter

import (
	"net/http"
)

// defaultGeoBlockedHandler responds with 403 Forbidden when a request is
// blocked by the geo filter.
func defaultGeoBlockedHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Forbidden", http.StatusForbidden)
}

// WithGeoFilter returns middleware that enforces a GeoFilter. Requests that
// are not allowed by the filter are handled by blockedHandler. If
// blockedHandler is nil the default 403 handler is used.
func WithGeoFilter(gf *GeoFilter, blockedHandler http.HandlerFunc) func(http.Handler) http.Handler {
	if gf == nil {
		panic("ratelimiter: GeoFilter must not be nil")
	}
	if blockedHandler == nil {
		blockedHandler = defaultGeoBlockedHandler
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !gf.Allowed(r) {
				blockedHandler(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
