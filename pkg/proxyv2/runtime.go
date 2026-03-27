package proxy

import (
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

type PathMatchType string

const (
	PathExact  PathMatchType = "Exact"
	PathPrefix PathMatchType = "Prefix"
	PathAny    PathMatchType = "Any"
)

type Runtime struct {
	Version   uint64
	Listeners map[string]*ListenerRuntime
}

type ListenerRuntime struct {
	Name          string
	ExactHosts    map[string]*VirtualHostRuntime
	WildcardHosts []WildcardHostRuntime
}

type WildcardHostRuntime struct {
	Suffix string
	VHost  *VirtualHostRuntime
}

type VirtualHostRuntime struct {
	ExactPaths  map[string][]*CompiledRoute
	PrefixPaths []*CompiledRoute
	CatchAll    []*CompiledRoute
}

type CompiledRoute struct {
	Name        string
	PathType    PathMatchType
	Path        string
	Methods     map[string]struct{}
	Headers     []HeaderMatch
	BackendPool *BackendPool
	Timeout     time.Duration
	Handler     http.Handler
}

type HeaderMatch struct {
	Name  string
	Value string
}

type BackendPool struct {
	Name    string
	Targets []*BackendTarget
	rr      atomic.Uint32
}

type BackendTarget struct {
	ID      string
	URL     *url.URL
	Healthy atomic.Bool
	Weight  int
}

func (rt *Runtime) Match(r *http.Request) (*CompiledRoute, bool) {
	ln, ok := rt.Listeners["default"]
	if !ok {
		return nil, false
	}
	return ln.Match(r)
}

func (l *ListenerRuntime) Match(r *http.Request) (*CompiledRoute, bool) {
	host := normalizeHost(r.Host)

	if vh, ok := l.ExactHosts[host]; ok {
		if route, ok := vh.Match(r); ok {
			return route, true
		}
	}

	for _, wh := range l.WildcardHosts {
		if strings.HasSuffix(host, wh.Suffix) {
			if route, ok := wh.VHost.Match(r); ok {
				return route, true
			}
		}
	}

	return nil, false
}

func (vh *VirtualHostRuntime) Match(r *http.Request) (*CompiledRoute, bool) {
	path := r.URL.Path

	if routes := vh.ExactPaths[path]; len(routes) > 0 {
		for _, route := range routes {
			if route.MatchRequest(r) {
				return route, true
			}
		}
	}

	for _, route := range vh.PrefixPaths {
		if routeMatchesPrefix(route, path) && route.MatchRequest(r) {
			return route, true
		}
	}

	for _, route := range vh.CatchAll {
		if route.MatchRequest(r) {
			return route, true
		}
	}

	return nil, false
}

func (cr *CompiledRoute) MatchRequest(r *http.Request) bool {
	if len(cr.Methods) > 0 {
		if _, ok := cr.Methods[r.Method]; !ok {
			return false
		}
	}

	for _, hm := range cr.Headers {
		if r.Header.Get(hm.Name) != hm.Value {
			return false
		}
	}

	return true
}

func (p *BackendPool) Next(r *http.Request) (*BackendTarget, bool) {
	targets := p.healthyTargets()
	if len(targets) == 0 {
		return nil, false
	}

	n := uint64(p.rr.Add(1))
	idx := int(n % uint64(len(targets)))
	return targets[idx], true
}

func (p *BackendPool) healthyTargets() []*BackendTarget {
	out := make([]*BackendTarget, 0, len(p.Targets))
	for _, t := range p.Targets {
		if t.Healthy.Load() {
			out = append(out, t)
		}
	}
	return out
}

func routePathPrefix(route *CompiledRoute) (string, bool) {
	return route.Path, route.PathType == PathPrefix
}

func routeMatchesPrefix(route *CompiledRoute, path string) bool {
	pfx, ok := routePathPrefix(route)
	if !ok {
		return false
	}
	return strings.HasPrefix(path, pfx)
}

func normalizeHost(h string) string {
	host, _, err := net.SplitHostPort(h)
	if err == nil {
		return strings.ToLower(host)
	}
	return strings.ToLower(h)
}
