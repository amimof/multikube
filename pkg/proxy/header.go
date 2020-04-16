package proxy

import (
	"context"
	"net/http"
)

// WithHeader is a middleware that reads the value of the HTTP header "Multikube-Context"
// in the request and, if found, sets it's value in the request context.
func WithHeader() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			req := r
			header := r.Header.Get("Multikube-Context")
			if header != "" {
				ctx := context.WithValue(r.Context(), contextKey, header)
				req = r.WithContext(ctx)
			}
			next.ServeHTTP(w, req)
		})
	}
}
