package middleware

import (
	"net/http"
	"github.com/amimof/multikube/pkg/proxy"
)

type EmptyMiddleware struct {
	Handler http.Handler
}

// WithEmpty is an empty handler that does nothing
func WithEmpty(c *proxy.Proxy, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}