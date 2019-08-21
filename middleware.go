package multikube

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/square/go-jose.v2/jwt"
	"log"
	"math/big"
	"net/http"
)

var frontendGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "multikube_frontend_live_requests",
	Help: "A gauge of live requests currently in flight from clients",
})

var frontendCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "multikube_frontend_requests_total",
		Help: "A counter for requests from clients",
	},
	[]string{"code", "method"},
)

var frontendHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "multikube_frontend_request_duration_seconds",
		Help:    "A histogram of request latencies from clients.",
		Buckets: prometheus.DefBuckets,
	},
	[]string{"handler", "method"},
)

// responseSize has no labels, making it a zero-dimensional
// ObserverVec.
var responseSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "multikube_frontend_response_size_bytes",
		Help:    "A histogram of response sizes for client requests",
		Buckets: []float64{200, 500, 900, 1500},
	},
	[]string{},
)

type MiddlewareFunc func(next http.HandlerFunc) http.HandlerFunc
type Middleware func(*Config, http.Handler) http.Handler

// responseWriter implements http.ResponseWriter and adds status code
// so that WithLogging middleware can log response status codes
type responseWriter struct {
	http.ResponseWriter
	status int
}

func init() {
	prometheus.MustRegister(frontendGauge, frontendCounter, frontendHistogram, responseSize)
}

// WriteHeader sends and sets an HTTP response header with the provided
// status code.
func (r *responseWriter) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

// WithEmpty is an empty handler that does nothing
func WithEmpty(c *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// WithEmpty is an empty handler that does nothing
func WithMetrics(c *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pushChain := promhttp.InstrumentHandlerInFlight(frontendGauge,
			promhttp.InstrumentHandlerDuration(frontendHistogram.MustCurryWith(prometheus.Labels{"handler": "push"}),
				promhttp.InstrumentHandlerCounter(frontendCounter,
					promhttp.InstrumentHandlerResponseSize(responseSize, next),
				),
			),
		)
		pushChain.ServeHTTP(w, r)
	})
}

// WithEmpty is a middleware that starts a new span and populates the context
func WithTracing(c *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		span := opentracing.GlobalTracer().StartSpan("hello")
		ctx := opentracing.ContextWithSpan(r.Context(), span)
		defer span.Finish()
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// WithLogging applies access log style logging to the HTTP server
func WithLogging(c *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := &responseWriter{w, http.StatusOK}
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s %s %s %d", r.Method, r.URL.Path, r.URL.RawQuery, r.RemoteAddr, r.Proto, lrw.status)
	})
}

// WithValidate validates JWT tokens in the request. For example Bearer-tokens
// TODO: Fix doc!
func WithRS256Validation(c *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		t, err := jws.ParseJWTFromRequest(r)
		if err != nil {
			log.Printf("ERROR parsing JWT from request: %s", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if t == nil {
			log.Printf("ERROR token: No token")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		err = t.Validate(c.RS256PublicKey.PublicKey, crypto.SigningMethodRS256)
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

// WithJWKValidation
// TODO: Fix doc!
func WithJWKValidation(c *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		t, err := jws.ParseJWTFromRequest(r)
		if err != nil {
			log.Printf("ERROR parsing JWT from request: %s", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		raw := string(getTokenFromRequest(r))
		tok, err := jwt.ParseSigned(raw)
		if err != nil {
			log.Printf("ERROR parsing JWS from request : %s", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Try to find a JWK using the kid
		kid := tok.Headers[0].KeyID
		jwk := c.JWKS.Find(kid)
		if jwk == nil {
			log.Printf("%s", "ERROR: Key ID mismatch parsing JWS from request : %s")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}
		if jwk.Kty != "RSA" {
			log.Printf("ERROR: Invalid key type. Expected 'RSA' got '%s'", jwk.Kty)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		// decode the base64 bytes for n
		nb, err := base64.RawURLEncoding.DecodeString(jwk.N)
		if err != nil {
			log.Printf("ERROR decoding N : %s", err)
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Check if E is big-endian int
		if jwk.E != "AQAB" && jwk.E != "AAEAAQ" {
			log.Printf("%s %s %s", "ERROR: Expected E to be a big-endian int", jwk.E)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		pk := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nb),
			E: 65537,
		}

		err = t.Validate(pk, crypto.SigningMethodRS256)
		if err != nil {
			log.Printf("ERROR validate: %s", err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		s, ok := t.Claims().Get("sub").(string)
		if !ok {
			s = ""
		}

		ctx := context.WithValue(r.Context(), "Context", "kladdis")
		ctx = context.WithValue(ctx, "Subject", s)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// WithValidate validates JWT tokens in the request. For example Bearer-tokens
func WithHeader(c *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r
		header := r.Header.Get("Multikube-Context")
		if header != "" {
			ctx := context.WithValue(r.Context(), "Context", header)
			req = r.WithContext(ctx)
		}
		next.ServeHTTP(w, req)
	})
}
