package proxy

import (
	"context"
	"crypto/rsa"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"math/rand"
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

			ctxName := ParseContextFromRequest(r, false)
			rs256ReqsTotal.WithLabelValues(ctxName).Inc()

			t, err := jws.ParseJWTFromRequest(r)
			if err != nil {
				rs256ReqsUnauthorized.WithLabelValues(ctxName).Inc()
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			if t == nil {
				rs256ReqsUnauthorized.WithLabelValues(ctxName).Inc()
				http.Error(w, "No token in request", http.StatusUnauthorized)
				return
			}

			err = t.Validate(c.PublicKey, crypto.SigningMethodRS256)
			if err != nil {
				rs256ReqsUnauthorized.WithLabelValues(ctxName).Inc()
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			// For security purposes a random value is set to username if no username claim is found or if the claim returns an empty string.
			// Kubernetes API will ignore user impersonation if this value is empty or nil.
			username, ok := t.Claims().Get("sub").(string)
			if !ok || username == "" {
				username = randomStr(10)
			}

			rs256ReqsAuthorized.WithLabelValues(ctxName).Inc()

			ctx := context.WithValue(r.Context(), subjectKey, username)
			next.ServeHTTP(w, r.WithContext(ctx))

		})
	}
}

// Returns a random fixed length string
func randomStr(n int) string {
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
