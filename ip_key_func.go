package ratelimiter

import (
	"net"
	"net/http"
	"strings"
)

// IPKeyFunc returns a KeyFunc that uses the client's IP address as the rate
// limit key. It respects X-Forwarded-For and X-Real-IP headers when
// trustProxies is true.
func IPKeyFunc(trustProxies bool) KeyFunc {
	return func(r *http.Request) string {
		if trustProxies {
			if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
				// X-Forwarded-For may contain a comma-separated list; use the first entry.
				parts := strings.SplitN(xff, ",", 2)
				ip := strings.TrimSpace(parts[0])
				if net.ParseIP(ip) != nil {
					return ip
				}
			}

			if xri := r.Header.Get("X-Real-IP"); xri != "" {
				ip := strings.TrimSpace(xri)
				if net.ParseIP(ip) != nil {
					return ip
				}
			}
		}

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			// RemoteAddr may already be a bare IP (e.g. in tests).
			return r.RemoteAddr
		}
		return ip
	}
}

// RouteKeyFunc returns a KeyFunc that combines the client IP with the request
// path, allowing per-route rate limiting.
func RouteKeyFunc(trustProxies bool) KeyFunc {
	ipFn := IPKeyFunc(trustProxies)
	return func(r *http.Request) string {
		return ipFn(r) + ":" + r.URL.Path
	}
}
