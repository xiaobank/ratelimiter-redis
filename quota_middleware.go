package ratelimiter

import (
	"fmt"
	"net/http"
)

// QuotaLimitExceededHandler is called when any quota is exceeded.
type QuotaLimitExceededHandler func(w http.ResponseWriter, r *http.Request, exceeded *Quota)

func defaultQuotaLimitExceededHandler(w http.ResponseWriter, r *http.Request, q *Quota) {
	w.Header().Set("X-Quota-Exceeded", q.Name)
	http.Error(w, fmt.Sprintf("quota %q exceeded", q.Name), http.StatusTooManyRequests)
}

// WithQuotaManager wraps an http.Handler and enforces all quotas managed by qm.
// If any quota is exceeded the onExceeded handler is called instead of next.
func WithQuotaManager(qm *QuotaManager, onExceeded QuotaLimitExceededHandler) func(http.Handler) http.Handler {
	if onExceeded == nil {
		onExceeded = defaultQuotaLimitExceededHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestKey := qm.keyFn(r)

			exceeded, err := qm.CheckAll(r.Context(), requestKey)
			if err != nil {
				http.Error(w, "internal rate limit error", http.StatusInternalServerError)
				return
			}

			if exceeded != nil {
				onExceeded(w, r, exceeded)
				return
			}

			// Attach remaining counts for each quota as response headers.
			for _, q := range qm.quotas {
				remaining, err := qm.Remaining(r.Context(), q.Name, requestKey)
				if err == nil {
					w.Header().Set(
						fmt.Sprintf("X-Quota-%s-Remaining", q.Name),
						fmt.Sprintf("%d", remaining),
					)
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}
