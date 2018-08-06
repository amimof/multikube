package multikube

import (
	"context"
	"crypto/x509"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"log"
	"net/http"
)

type MiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc
type Middleware func(http.Handler) http.Handler

// WithEmpty is an empty handler that does nothing
func WithEmpty(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// WithLogging applies access log style logging to the HTTP server
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s %s %s", r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr, r.Proto)
		next.ServeHTTP(w, r)
	})
}

// WithValidate validates JWT tokens in the request. For example Bearer-tokens
func WithValidate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cert := r.Context().Value("signer").(*x509.Certificate)

		t, err := jws.ParseJWTFromRequest(r)
		if err != nil {
			log.Printf("ERROR token: %s", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if t == nil {
			log.Printf("ERROR token: No token")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		err = t.Validate(cert.PublicKey, crypto.SigningMethodRS256)
		if err != nil {
			log.Printf("ERROR validate: %s", err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "Context", t.Claims().Get("ctx").(string))
		ctx = context.WithValue(ctx, "Subject", t.Claims().Get("sub").(string))

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
