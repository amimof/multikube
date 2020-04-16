package proxy

import (
	"context"
	"github.com/SermoDigital/jose/jws"
	"net/http"
)

// WithJWT is a middleware that parses a JWT token from the requests and propagates
// the request context with a claim value.
func WithJWT() MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			// Get the JWT from the request
			t, err := jws.ParseJWTFromRequest(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// Check if request has empty credentials
			if t == nil {
				http.Error(w, "No valid access token", http.StatusUnauthorized)
				return
			}

			// Set context
			//username, ok := t.Claims().Get(c.OIDCUsernameClaim).(string)
			username, ok := t.Claims().Get("sub").(string)
			if !ok {
				username = ""
			}

			ctx := context.WithValue(r.Context(), subjectKey, username)

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
