package proxy

import (
	"context"
	"crypto/rsa"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"net/http"
)

// RS256Config is configuration for RS256 middleware
type RS256Config struct {
	PublicKey *rsa.PublicKey
}

// WithRS256 is a middleware that validates a JWT token in the http request using RS256 signing method.
// It will do so using a rsa public key provided in Config
func WithRS256(c RS256Config) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			t, err := jws.ParseJWTFromRequest(r)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if t == nil {
				http.Error(w, "No token in request", http.StatusUnauthorized)
				return
			}

			err = t.Validate(c.PublicKey, crypto.SigningMethodRS256)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// Set context
			//username, ok := t.Claims().Get(c.OIDCUsernameClaim).(string)
			username, ok := t.Claims().Get("").(string)
			if !ok {
				username = ""
			}

			ctx := context.WithValue(r.Context(), subjectKey, username)

			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}
