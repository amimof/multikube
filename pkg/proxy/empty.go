package proxy

import (
	"net/http"
)

// WithEmpty is an empty handler that does nothing
func WithEmpty() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
