package proxy

import (
	"context"
	"crypto/ecdsa"
	"net/http"
	"strconv"
	"time"

	"github.com/amimof/multikube/pkg/audit"
)

type Proxy struct {
	runtime   *RuntimeStore
	pubKey    *ecdsa.PublicKey
	publisher audit.Publisher
}

// NewProxy creates a Proxy that serves requests from the given runtime store.
// If pubKey is non-nil, JWT extraction and policy enforcement are enabled.
func NewProxy(runtime *RuntimeStore, opts ...ProxyOption) *Proxy {
	p := &Proxy{runtime: runtime}
	for _, opt := range opts {
		opt(p)
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

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt := p.runtime.Load()

	ctx := r.Context()

	// JWT extraction. If a public key is configured, we require and validate a JWT.
	var principal *Principal
	if p.pubKey != nil {
		var flat map[string]any
		var err error
		principal, flat, err = ExtractJWT(r, p.pubKey)
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		ctx = WithPrincipal(ctx, principal)
		ctx = WithJWTClaims(ctx, flat)
		r = r.WithContext(ctx)
	}

	// Route matching
	route, ok := rt.Match(r)
	if !ok {
		http.NotFound(w, r)
		return
	}

	// Parse K8s from request
	k8sReq := ParseK8sRequest(ctx, r)

	// Policy enforcement. Only enforce when policies are present and a public key is configured.
	if p.pubKey != nil && len(rt.Policies) > 0 {
		if EvalPolicies(rt.Policies, principal, route.BackendPool, k8sReq) == EvalDeny {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	// Dispatch
	handler := route.Handler
	if route.Timeout > 0 {
		handler = timeoutMiddleware(route.Timeout)(handler)
	}
	handler = withRuntimeVersion(rt.Version)(handler)

	handler.ServeHTTP(w, r)
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
