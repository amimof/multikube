package proxy

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync/atomic"
	"time"

	policyv1 "github.com/amimof/multikube/api/policy/v1"
)

type RuntimeConfig struct {
	Version uint64

	Routes   CompiledRoutes
	Backends map[string]*BackendRuntime
	Policies []*policyv1.Policy
}

type CompiledRoutes struct {
	Paths        []*RouteRuntime
	PathPrefixes []*RouteRuntime
	Headers      []*RouteRuntime
	SNIExact     map[string][]*RouteRuntime
	JWT          []*RouteRuntime
}

type RouteRuntime struct {
	Name    string
	Kind    RouteMatchKind
	Timeout time.Duration

	Path       string
	PathPrefix string
	Header     *HeaderRuntime
	SNI        string
	JWT        *JWTRuntime

	// Backend *BackendRuntime

	Handler     http.Handler
	BackendPool *BackendPool
}

type BackendPool struct {
	Name    string
	Targets []*BackendRuntime
	rr      atomic.Uint32
}

func (p *BackendPool) Next(r *http.Request) (*BackendRuntime, bool) {
	targets := p.healthyTargets()
	if len(targets) == 0 {
		return nil, false
	}

	n := uint64(p.rr.Add(1))
	idx := int(n % uint64(len(targets)))
	return targets[idx], true
}

func (p *BackendPool) healthyTargets() []*BackendRuntime {
	out := make([]*BackendRuntime, 0, len(p.Targets))
	out = append(out, p.Targets...)
	return out
}

type BackendTarget struct {
	ID      string
	URL     *url.URL
	Healthy atomic.Bool
	Weight  int
}

type RouteMatchKind uint8

const (
	RouteMatchKindPathPrefix RouteMatchKind = iota + 1
	RouteMatchKindPath
	RouteMatchKindHeader
	RouteMatchKindSNI
	RouteMatchKindJWT
)

type HeaderRuntime struct {
	Name      string
	Canonical string
	Value     string
}

type JWTRuntime struct {
	Claim string
	Value string
}

type BackendRuntime struct {
	Name   string
	Labels map[string]string

	URL *url.URL

	CacheTTL time.Duration

	TLSConfig *tls.Config
	Transport http.RoundTripper

	AuthInjector RequestAuthInjector
}

type RequestAuthInjector interface {
	Apply(req *http.Request) error
}

func (rc *RuntimeConfig) Match(r *http.Request) (*RouteRuntime, bool) {
	if route, ok := rc.Routes.matchPath(r); ok {
		return route, true
	}
	if route, ok := rc.Routes.matchPathPrefix(r); ok {
		return route, true
	}
	if route, ok := rc.Routes.matchHeader(r); ok {
		return route, true
	}
	if route, ok := rc.Routes.matchJWT(r); ok {
		return route, true
	}
	if route, ok := rc.Routes.matchSNI(r); ok {
		return route, true
	}
	return nil, false
}

func (cr *CompiledRoutes) matchPath(r *http.Request) (*RouteRuntime, bool) {
	reqPath := r.URL.Path
	for _, route := range cr.Paths {
		match, err := path.Match(route.Path, reqPath)
		if err != nil {
			return nil, false
		}
		if match {
			return route, true
		}

	}
	return nil, false
}

func (cr *CompiledRoutes) matchPathPrefix(r *http.Request) (*RouteRuntime, bool) {
	path := r.URL.Path
	for _, route := range cr.PathPrefixes {
		if strings.HasPrefix(path, route.PathPrefix) {
			return route, true
		}
	}
	return nil, false
}

func (cr *CompiledRoutes) matchHeader(r *http.Request) (*RouteRuntime, bool) {
	for _, route := range cr.Headers {
		if route.Header == nil {
			continue
		}
		if r.Header.Get(route.Header.Canonical) == route.Header.Value {
			return route, true
		}
	}
	return nil, false
}

func (cr *CompiledRoutes) matchJWT(r *http.Request) (*RouteRuntime, bool) {
	claims, ok := JWTClaimsFromContext(r.Context())
	if !ok {
		return nil, false
	}

	for _, route := range cr.JWT {
		if route.JWT == nil {
			continue
		}
		if value, ok := claims[route.JWT.Claim]; ok && strings.EqualFold(fmt.Sprintf("%v", value), route.JWT.Value) {
			return route, true
		}
	}
	return nil, false
}

func (cr *CompiledRoutes) matchSNI(r *http.Request) (*RouteRuntime, bool) {
	sni, ok := SNIFromContext(r.Context())
	if !ok || sni == "" {
		return nil, false
	}

	routes := cr.SNIExact[sni]
	for _, route := range routes {
		return route, true
	}
	return nil, false
}

type contextKey string

const (
	ctxKeyJWTClaims contextKey = "jwt_claims"
	ctxKeySNI       contextKey = "sni"
	ctxKeyPrincipal contextKey = "principal"
)

func JWTClaimsFromContext(ctx context.Context) (map[string]any, bool) {
	v := ctx.Value(ctxKeyJWTClaims)
	claims, ok := v.(map[string]any)
	return claims, ok
}

func SNIFromContext(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeySNI)
	sni, ok := v.(string)
	return sni, ok
}

// Principal represents the authenticated identity extracted from a JWT.
type Principal struct {
	Subject         string
	User            string
	Groups          []string
	Issuer          string
	Audience        []string
	ServiceAccounts []string
	Claims          map[string]any
	ExpiresAt       time.Time
}

func PrincipalFromContext(ctx context.Context) (*Principal, bool) {
	v := ctx.Value(ctxKeyPrincipal)
	p, ok := v.(*Principal)
	return p, ok
}

func WithPrincipal(ctx context.Context, p *Principal) context.Context {
	if ev, ok := EventFromContext(ctx); ok {
		ev.Subject = p.Subject
		ev.Username = p.User
		ev.Groups = p.Groups
		ev.Issuer = p.Issuer
	}
	return context.WithValue(ctx, ctxKeyPrincipal, p)
}

func WithJWTClaims(ctx context.Context, claims map[string]any) context.Context {
	return context.WithValue(ctx, ctxKeyJWTClaims, claims)
}
