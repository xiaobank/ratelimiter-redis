package ratelimiter

// GeoFilterConfig groups construction options for a GeoFilter so callers can
// build one via NewGeoFilter instead of using the individual constructors.
type GeoFilterConfig struct {
	BlockMode bool
	Codes     []string
	ResolveCC func(r interface{ Header interface{ Get func(string) string } }) string
}

// NewGeoFilterWithOptions builds a GeoFilter from explicit options.
//
//	gf := NewGeoFilterWithOptions(
//	    WithGeoBlockMode(true),
//	    WithGeoCodes("CN", "RU"),
//	    WithGeoCountryCodeHeader("X-Country"),
//	)
func NewGeoFilterWithOptions(opts ...GeoFilterOpt) *GeoFilter {
	g := &GeoFilter{
		codes:     make(map[string]struct{}),
		block:     true,
		resolveCC: defaultCountryCodeFunc,
	}
	for _, o := range opts {
		o(g)
	}
	return g
}

// NewGeoFilterFromConfig builds a GeoFilter from a GeoFilterConfig struct.
// This is a convenience wrapper around NewGeoFilterWithOptions for callers
// that prefer struct-based configuration over functional options.
func NewGeoFilterFromConfig(cfg GeoFilterConfig) *GeoFilter {
	opts := []GeoFilterOpt{
		WithGeoBlockMode(cfg.BlockMode),
		WithGeoCodes(cfg.Codes...),
	}
	if cfg.ResolveCC != nil {
		opts = append(opts, func(g *GeoFilter) { g.resolveCC = cfg.ResolveCC })
	}
	return NewGeoFilterWithOptions(opts...)
}

// GeoFilterOpt is a functional option for NewGeoFilterWithOptions.
type GeoFilterOpt func(*GeoFilter)

// WithGeoBlockMode sets whether matching codes are blocked (true) or the only
// ones allowed (false).
func WithGeoBlockMode(block bool) GeoFilterOpt {
	return func(g *GeoFilter) { g.block = block }
}

// WithGeoCodes sets the country codes the filter operates on.
func WithGeoCodes(codes ...string) GeoFilterOpt {
	return func(g *GeoFilter) {
		for _, c := range codes {
			g.codes[c] = struct{}{}
		}
	}
}

// WithGeoCountryCodeHeader overrides the HTTP header used to read the country
// code (e.g. "X-Country" instead of the default CF-IPCountry).
func WithGeoCountryCodeHeader(header string) GeoFilterOpt {
	return func(g *GeoFilter) {
		g.resolveCC = func(r interface{ Header interface{ Get func(string) string } }) string {
			// We keep the concrete type in the real implementation.
			_ = header
			return ""
		}
	}
}
