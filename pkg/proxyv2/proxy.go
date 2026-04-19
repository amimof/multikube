package proxy

import (
	"context"
	"crypto/ecdsa"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/amimof/multikube/pkg/audit"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

type Proxy struct {
	runtime   *RuntimeStore
	pubKey    *ecdsa.PublicKey
	publisher audit.Publisher
	metrics   *ProxyMetrics
	meter     metric.Meter
}

// NewProxy creates a Proxy that serves requests from the given runtime store.
// If pubKey is non-nil, JWT extraction and policy enforcement are enabled.
func NewProxy(runtime *RuntimeStore, opts ...ProxyOption) *Proxy {
	p := &Proxy{
		runtime: runtime,
		meter:   otel.GetMeterProvider().Meter("multikube_proxy"),
	}
	for _, opt := range opts {
		opt(p)
	}

	if p.metrics == nil {
		m, err := InitMetrics(p.meter)
		if err != nil {
			log.Printf("failed to initialize proxy metrics: %v", err)
		}
		p.metrics = m
	}

	return p
}

type ProxyOption func(*Proxy)

// WithPublicKey enables JWT extraction and policy enforcement using the given
// ECDSA public key.
func WithPublicKey(key *ecdsa.PublicKey) ProxyOption {
	return func(p *Proxy) {
		p.pubKey = key
	}
}

func WithPublisher(pub audit.Publisher) ProxyOption {
	return func(p *Proxy) {
		p.publisher = pub
	}
}

func WithMeter(m metric.Meter) ProxyOption {
	return func(p *Proxy) {
		p.meter = m
	}
}

func WithMetrics(m *ProxyMetrics) ProxyOption {
	return func(p *Proxy) {
		p.metrics = m
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt := p.runtime.Load()
	ctx := r.Context()
	start := time.Now()

	// Track active requests
	if p.metrics != nil {
		p.metrics.ActiveRequests.Add(ctx, 1)
		defer p.metrics.ActiveRequests.Add(ctx, -1)

		if r.ContentLength > 0 {
			p.metrics.RequestSizeBytes.Record(ctx, r.ContentLength,
				metric.WithAttributes(attribute.String("method", r.Method)))
		}
	}

	// Wrap response writer to capture status code and bytes written
	mw := &metricsResponseWriter{ResponseWriter: w}

	// JWT extraction
	var principal *Principal
	if p.pubKey != nil {
		var flat map[string]any
		var err error
		principal, flat, err = ExtractJWT(r, p.pubKey)
		if err != nil {
			if p.metrics != nil {
				result := "invalid"
				if bearerToken(r) == "" {
					result = "missing"
				}
				p.metrics.AuthRequestsTotal.Inc(ctx, 1,
					metric.WithAttributes(attribute.String("result", result)))
			}
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, err.Error(), http.StatusUnauthorized)
			p.recordRequestMetrics(ctx, start, r.Method, "", http.StatusUnauthorized, mw)
			return
		}
		if p.metrics != nil {
			p.metrics.AuthRequestsTotal.Inc(ctx, 1,
				metric.WithAttributes(attribute.String("result", "success")))
		}
		ctx = WithPrincipal(ctx, principal)
		ctx = WithJWTClaims(ctx, flat)
		r = r.WithContext(ctx)
	}

	// Route matching
	route, ok := rt.Match(r)
	if !ok {
		if p.metrics != nil {
			p.metrics.RouteNoMatchTotal.Inc(ctx, 1)
		}
		http.NotFound(w, r)
		p.recordRequestMetrics(ctx, start, r.Method, "", http.StatusNotFound, mw)
		return
	}

	// Record route match
	if p.metrics != nil {
		p.metrics.RouteMatchesTotal.Inc(ctx, 1,
			metric.WithAttributes(attribute.String("match_kind", matchKindString(route.Kind))))
	}

	ctx = WithMatchedRoute(ctx, route)
	r = r.WithContext(ctx)

	// Parse K8s from request
	k8sReq := ParseK8sRequest(ctx, r)

	// Policy enforcement
	if p.pubKey != nil && len(rt.Policies) > 0 {
		result := EvalPolicies(rt.Policies, principal, route.BackendPool, k8sReq)
		if p.metrics != nil {
			evalResult := "allow"
			if result == EvalDeny {
				evalResult = "deny"
			}
			p.metrics.PolicyEvaluationsTotal.Inc(ctx, 1,
				metric.WithAttributes(
					attribute.String("result", evalResult),
					attribute.String("route", route.Name),
				))
		}
		if result == EvalDeny {
			http.Error(w, "Forbidden", http.StatusForbidden)
			p.recordRequestMetrics(ctx, start, r.Method, route.Name, http.StatusForbidden, mw)
			return
		}
	}

	// Dispatch
	handler := route.Handler
	if route.Timeout > 0 {
		handler = timeoutMiddleware(route.Timeout)(handler)
	}
	handler = withRuntimeVersion(rt.Version)(handler)

	handler.ServeHTTP(mw, r)

	p.recordRequestMetrics(ctx, start, r.Method, route.Name, mw.statusCode, mw)
}

func (p *Proxy) recordRequestMetrics(ctx context.Context, start time.Time, method, route string, statusCode int, mw *metricsResponseWriter) {
	if p.metrics == nil {
		return
	}
	duration := time.Since(start).Seconds()
	code := statusCode
	if code == 0 {
		code = http.StatusOK
	}
	attrs := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("route", route),
		attribute.Int("status_code", code),
	)
	p.metrics.RequestsTotal.Inc(ctx, 1, attrs)
	p.metrics.RequestDuration.Record(ctx, duration, attrs)
	if mw.bytesWritten > 0 {
		p.metrics.ResponseSizeBytes.Record(ctx, mw.bytesWritten,
			metric.WithAttributes(
				attribute.String("method", method),
				attribute.String("route", route),
			))
	}
}

// metricsResponseWriter captures status code and bytes written for metrics.
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int64
}

func (w *metricsResponseWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *metricsResponseWriter) Write(p []byte) (int, error) {
	if w.statusCode == 0 {
		w.statusCode = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytesWritten += int64(n)
	return n, err
}

func (w *metricsResponseWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func matchKindString(k RouteMatchKind) string {
	switch k {
	case RouteMatchKindPath:
		return "path"
	case RouteMatchKindPathPrefix:
		return "path_prefix"
	case RouteMatchKindHeader:
		return "header"
	case RouteMatchKindSNI:
		return "sni"
	case RouteMatchKindJWT:
		return "jwt"
	default:
		return "unknown"
	}
}

func timeoutMiddleware(d time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), d)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func withRuntimeVersion(version uint64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Proxy-Config-Version", itoa(version))
			next.ServeHTTP(w, r)
		})
	}
}

func itoa(v uint64) string {
	return strconv.FormatUint(v, 10)
}
