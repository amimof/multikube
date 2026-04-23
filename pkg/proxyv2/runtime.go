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
	Backends map[string]*BackendPool
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

	Handler     http.Handler
	BackendPool *BackendPool
}

type BackendPool struct {
	Name          string
	Targets       []*BackendRuntime
	Iterator      BackendIterator
	Impersonation *ImpersonationRuntime
}

// ImpersonationRuntime holds the compiled impersonation configuration for a
// backend pool. When non-nil and Enabled is true, the proxy injects
// Impersonate-User, Impersonate-Group, and Impersonate-Extra-* headers on the
// outbound request using claims from the authenticated principal's JWT.
type ImpersonationRuntime struct {
	Name          string
	Enabled       bool
	UsernameClaim string
	GroupsClaim   string
	ExtraClaims   []string
}

func (p *BackendPool) Next(r *http.Request) (*BackendRuntime, bool) {
	targets := p.healthyTargets()
	if len(targets) == 0 {
		return nil, false
	}
	return p.Iterator.Next(targets)
}

func (p *BackendPool) healthyTargets() []*BackendRuntime {
	out := make([]*BackendRuntime, 0, len(p.Targets))
	for _, target := range p.Targets {
		if target.IsAvailable() {
			out = append(out, target)
		}
	}
	return out
}

type BackendTarget struct {
	ID      string
	URL     *url.URL
	Healthy atomic.Bool
	Weight  int
}

type BackendRuntime struct {
	Name   string
	Labels map[string]string

	URL *url.URL

	HasHealthProbe    bool
	HealthKnown       atomic.Bool
	Healthy           atomic.Bool
	HasReadinessProbe bool
	ReadinessKnown    atomic.Bool
	Ready             atomic.Bool

	CacheTTL time.Duration

	TLSConfig *tls.Config
	Transport http.RoundTripper

	AuthInjector RequestAuthInjector

	Active atomic.Int64
}

func (b *BackendRuntime) IsAvailable() bool {
	if b == nil {
		return false
	}
	if b.HasHealthProbe && b.HealthKnown.Load() && !b.Healthy.Load() {
		return false
	}
	if b.HasReadinessProbe && b.ReadinessKnown.Load() && !b.Ready.Load() {
		return false
	}
	return true
}

func (b *BackendRuntime) SetHealthState(known, healthy bool) {
	b.HealthKnown.Store(known)
	b.Healthy.Store(healthy)
}

func (b *BackendRuntime) SetReadinessState(known, ready bool) {
	b.ReadinessKnown.Store(known)
	b.Ready.Store(ready)
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
	if len(routes) > 0 {
		return routes[0], true
	}
	return nil, false
}

type contextKey string

const (
	ctxKeyJWTClaims contextKey = "jwt_claims"
	ctxKeySNI       contextKey = "sni"
	ctxKeyPrincipal contextKey = "principal"
	ctxKeyRoute     contextKey = "matched_route"
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

// WithMatchedRoute stores the matched RouteRuntime in the request context so
// that downstream handlers (e.g. the forwarder) can inspect the route that was
// selected during dispatch.
func WithMatchedRoute(ctx context.Context, rr *RouteRuntime) context.Context {
	return context.WithValue(ctx, ctxKeyRoute, rr)
}

// MatchedRouteFromContext retrieves the RouteRuntime stored by WithMatchedRoute.
func MatchedRouteFromContext(ctx context.Context) (*RouteRuntime, bool) {
	v := ctx.Value(ctxKeyRoute)
	rr, ok := v.(*RouteRuntime)
	return rr, ok
}
