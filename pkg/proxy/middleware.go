package proxy

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"github.com/SermoDigital/jose/crypto"
	"github.com/SermoDigital/jose/jws"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gopkg.in/square/go-jose.v2/jwt"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"strings"
)

type ctxKey string

var (
	contextKey = ctxKey("Context")
	subjectKey = ctxKey("Subject")
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

// Middleware represents a multikube middleware
type Middleware func(*Config, http.Handler) http.Handler

// responseWriter implements http.ResponseWriter and adds status code
// so that WithLogging middleware can log response status codes
type responseWriter struct {
	http.ResponseWriter
	status int
}

// responseError satisfies the error interface
type responseError struct {
	Status int      `json:"status"`
	Errs   []string `json:"errors"`
}

func init() {
	prometheus.MustRegister(frontendGauge, frontendCounter, frontendHistogram, responseSize)
}

// getTokenFromRequest returns a []byte representation of JWT from an HTTP Authorization Bearer header
func getTokenFromRequest(req *http.Request) []byte {
	if ah := req.Header.Get("Authorization"); len(ah) > 7 && strings.EqualFold(ah[0:7], "BEARER ") {
		return []byte(ah[7:])
	}
	return nil
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

// WithMetrics is an empty handler that does nothing
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

// WithTracing is a middleware that starts a new span and populates the context
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

// WithJWT is a middleware that parses a JWT token from the requests and propagates
// the request context with a claim value.
func WithJWT(c *Config, next http.Handler) http.Handler {
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
		username, ok := t.Claims().Get(c.OIDCUsernameClaim).(string)
		if !ok {
			username = ""
		}

		ctx := context.WithValue(r.Context(), subjectKey, username)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// WithX509Validation is a middleware that validates a JWT token in the http request using RS256 signing method.
// It will do so using a x509 certificate provided in c
func WithX509Validation(c *Config, next http.Handler) http.Handler {
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

		err = t.Validate(c.RS256PublicKey.PublicKey, crypto.SigningMethodRS256)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Set context
		username, ok := t.Claims().Get(c.OIDCUsernameClaim).(string)
		if !ok {
			username = ""
		}

		ctx := context.WithValue(r.Context(), subjectKey, username)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// WithJWKValidation is a middleware that validates a JWT token in the http request using RS256 signing method.
// It will do so using a JWK (Json Web Key) provided in c
func WithJWKValidation(c *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		t, err := jws.ParseJWTFromRequest(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		raw := string(getTokenFromRequest(r))
		tok, err := jwt.ParseSigned(raw)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Try to find a JWK using the kid
		kid := tok.Headers[0].KeyID
		jwk := c.JWKS.Find(kid)
		if jwk == nil {
			http.Error(w, "key id invalid", http.StatusUnauthorized)
			return
		}
		if jwk.Kty != "RSA" {
			http.Error(w, fmt.Sprintf("Invalid key type. Expected 'RSA' got '%s'", jwk.Kty), http.StatusUnauthorized)
			return
		}

		// decode the base64 bytes for n
		nb, err := base64.RawURLEncoding.DecodeString(jwk.N)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		// Check if E is big-endian int
		if jwk.E != "AQAB" && jwk.E != "AAEAAQ" {
			http.Error(w, fmt.Sprintf("Expected E to be one of 'AQAB' and 'AAEAAQ' but got '%s'", jwk.E), http.StatusUnauthorized)
			return
		}

		pk := &rsa.PublicKey{
			N: new(big.Int).SetBytes(nb),
			E: 65537,
		}

		err = t.Validate(pk, crypto.SigningMethodRS256)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		username, ok := t.Claims().Get(c.OIDCUsernameClaim).(string)
		if !ok {
			username = ""
		}

		ctx := context.WithValue(r.Context(), subjectKey, username)

		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

// WithHeader is a middleware that reads the value of the HTTP header "Multikube-Context"
// in the request and, if found, sets it's value in the request context.
func WithHeader(c *Config, next http.Handler) http.Handler {
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

// WithCtxRoot is a middleware that reads the url path params in the request and
// tries to determine which kubeconfig context to use for upstream api server requests.
// If a context is found in the URL path params, the request-context is populated with the value
// so that other handlers and middlewares may use the information
func WithCtxRoot(c *Config, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req := r
		c, rem := getCtxFromURL(r.URL)
		if c != "" {
			ctx := context.WithValue(r.Context(), contextKey, c)
			req = r.WithContext(ctx)
			if rem != "" {
				req.URL.Path = rem
			}
		}
		next.ServeHTTP(w, req)
	})
}

// getCtxFromURL reads path params from u and returns the kubeconfig context
// as well as the path params used for upstream communication
func getCtxFromURL(u *url.URL) (string, string) {
	val := ""
	rem := []string{}
	if vals := strings.Split(u.Path, "/"); len(vals) > 1 {
		val = vals[1]
		rem = vals[2:]
	}
	return val, fmt.Sprintf("/%s", strings.Join(rem, "/"))
}
