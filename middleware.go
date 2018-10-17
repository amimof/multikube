package multikube

import (
	"context"
	"crypto/x509"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/opentracing/opentracing-go"
	"log"
	"net/http"
)

type MiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc
type Middleware func(http.Handler) http.Handler

// responseWriter implements http.ResponseWriter and adds status code
// so that WithLogging middleware can log response status codes
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// WithEmpty is an empty handler that does nothing
func WithEmpty(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// WithEmpty is a middleware that starts a new span and populates the context
func WithTracing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := opentracing.GlobalTracer().StartSpan("hello")
		ctx := opentracing.ContextWithSpan(r.Context(), span)
		defer span.Finish()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WithLogging applies access log style logging to the HTTP server
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(lrw, r)
		log.Printf("%s %s %s %s %s %d", r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr, r.Proto, lrw.status)
	})
}

// WithValidate validates JWT tokens in the request. For example Bearer-tokens
func WithValidate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		cert := r.Context().Value("rs256PublicKey").(*x509.Certificate)

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

		c, ok := t.Claims().Get("ctx").(string)
		if !ok {
			c = ""
		}

		s, ok := t.Claims().Get("sub").(string)
		if !ok {
			s = ""
		}

		ctx := context.WithValue(r.Context(), "Context", c)
		ctx = context.WithValue(ctx, "Subject", s)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}
