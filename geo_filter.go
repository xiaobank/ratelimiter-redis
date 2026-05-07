package ratelimiter

import (
	"net/http"
)

// GeoFilter holds a set of allowed or blocked country codes and a function
// to resolve the country code from a request.
type GeoFilter struct {
	codes      map[string]struct{}
	block      bool
	resolveCC  func(r *http.Request) string
}

// GeoFilterOption configures a GeoFilter.
type GeoFilterOption func(*GeoFilter)

// WithBlockedCountries creates a GeoFilter that blocks requests from the given
// ISO 3166-1 alpha-2 country codes.
func WithBlockedCountries(codes ...string) *GeoFilter {
	g := &GeoFilter{
		codes:     make(map[string]struct{}, len(codes)),
		block:     true,
		resolveCC: defaultCountryCodeFunc,
	}
	for _, c := range codes {
		g.codes[c] = struct{}{}
	}
	return g
}

// WithAllowedCountries creates a GeoFilter that only allows requests from the
// given ISO 3166-1 alpha-2 country codes.
func WithAllowedCountries(codes ...string) *GeoFilter {
	g := &GeoFilter{
		codes:     make(map[string]struct{}, len(codes)),
		block:     false,
		resolveCC: defaultCountryCodeFunc,
	}
	for _, c := range codes {
		g.codes[c] = struct{}{}
	}
	return g
}

// SetCountryCodeFunc overrides the function used to resolve a country code
// from a request (e.g. read a header set by a CDN).
func (g *GeoFilter) SetCountryCodeFunc(fn func(r *http.Request) string) {
	if fn != nil {
		g.resolveCC = fn
	}
}

// Allowed reports whether the request should be allowed through.
func (g *GeoFilter) Allowed(r *http.Request) bool {
	cc := g.resolveCC(r)
	_, found := g.codes[cc]
	if g.block {
		return !found
	}
	return found
}

// defaultCountryCodeFunc reads the CF-IPCountry header, which is commonly set
// by Cloudflare. Falls back to the X-Country-Code header.
func defaultCountryCodeFunc(r *http.Request) string {
	if cc := r.Header.Get("CF-IPCountry"); cc != "" {
		return cc
	}
	return r.Header.Get("X-Country-Code")
}
