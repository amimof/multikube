// Package compile transforms API proto objects into a proxy runtime configuration.
package compile

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
	proxy "github.com/amimof/multikube/pkg/proxyv2"
)

// State holds the current desired state of all API resources.
// It is usually populated by controllers and consumed by the Compiler.
type State struct {
	Backends               map[string]*backendv1.Backend
	Routes                 map[string]*routev1.Route
	Certificates           map[string]*certificatev1.Certificate
	CertificateAuthorities map[string]*cav1.CertificateAuthority
}

// Compiler compiles a State into a proxy Runtime.
// It holds no shared state and each Compile call is self-contained.
type Compiler struct{}

// NewCompiler returns a new Compiler.
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Compile converts the contents of a ControllerCache into a [*proxy.Runtime] that
// the proxy can use to match and forward requests.
func (c *Compiler) Compile(st *State) (*proxy.Runtime, error) {
	// compile TLS client certificates first — CAs may reference them
	tlsCerts, err := compileCerts(st.Certificates)
	if err != nil {
		return nil, fmt.Errorf("compile certs: %w", err)
	}

	// compile CA certificates into x509 cert pools
	caPools, err := compileCAs(st.CertificateAuthorities, st.Certificates)
	if err != nil {
		return nil, fmt.Errorf("compile CAs: %w", err)
	}

	// compile backends into BackendPools, one Forwarder per backend
	pools, forwarders, err := compileBackends(st.Backends, caPools, tlsCerts)
	if err != nil {
		return nil, fmt.Errorf("compile backends: %w", err)
	}

	// compile routes into per-SNI VirtualHostRuntimes
	exactHosts, wildcardHosts, err := compileRoutes(st.Routes, pools, forwarders)
	if err != nil {
		return nil, fmt.Errorf("compile routes: %w", err)
	}

	listener := &proxy.ListenerRuntime{
		Name:          "default",
		ExactHosts:    exactHosts,
		WildcardHosts: wildcardHosts,
	}

	return &proxy.Runtime{
		Version: 1,
		Listeners: map[string]*proxy.ListenerRuntime{
			"default": listener,
		},
	}, nil
}

func compileCAs(cas map[string]*cav1.CertificateAuthority, certs map[string]*certificatev1.Certificate) (map[string]*x509.CertPool, error) {
	out := make(map[string]*x509.CertPool, len(cas))
	for name, ca := range cas {
		pool, err := compileCA(ca, certs)
		if err != nil {
			return nil, fmt.Errorf("CA %q: %w", name, err)
		}
		out[name] = pool
	}
	return out, nil
}

// compileCA builds an *x509.CertPool from a CertificateAuthority object.
func compileCA(ca *cav1.CertificateAuthority, certs map[string]*certificatev1.Certificate) (*x509.CertPool, error) {
	var pemBytes []byte

	switch {
	case ca.GetConfig().GetCertificate() != "":
		ref := ca.GetConfig().GetCertificate()
		certObj, ok := certs[ref]
		if !ok {
			return nil, fmt.Errorf("certificate ref %q not found", ref)
		}
		inline := certObj.GetConfig().GetCertificate()
		if inline == "" {
			return nil, fmt.Errorf("certificate ref %q has no inline certificate data", ref)
		}
		pemBytes = []byte(inline)
	case ca.GetConfig().GetCertificateData() != "":
		pemBytes = []byte(ca.GetConfig().GetCertificateData())
	default:
		return nil, fmt.Errorf("neither certificate ref nor certificate_data provided")
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("no valid certificates found in PEM data")
	}
	return pool, nil
}

func compileCerts(certs map[string]*certificatev1.Certificate) (map[string]tls.Certificate, error) {
	out := make(map[string]tls.Certificate, len(certs))
	for name, cert := range certs {
		tlsCert, err := compileCert(cert)
		if err != nil {
			return nil, fmt.Errorf("certificate %q: %w", name, err)
		}
		out[name] = tlsCert
	}
	return out, nil
}

// compileCert builds a tls.Certificate from a Certificate object.
func compileCert(cert *certificatev1.Certificate) (tls.Certificate, error) {
	certPEM := cert.GetConfig().GetCertificate()
	if certPEM == "" {
		return tls.Certificate{}, fmt.Errorf("certificate has no inline PEM data")
	}

	keyPEM := cert.GetConfig().GetKey()
	if keyPEM == "" {
		return tls.Certificate{}, fmt.Errorf("key has no inline PEM data")
	}

	tlsCert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("X509KeyPair: %w", err)
	}
	return tlsCert, nil
}

// compileBackends returns two parallel maps: BackendPool and Forwarder, keyed
// by backend name. Each backend gets its own Forwarder built around a
// per-backend TLS transport so that mTLS can be configured independently.
func compileBackends(
	backends map[string]*backendv1.Backend,
	caPools map[string]*x509.CertPool,
	tlsCerts map[string]tls.Certificate,
) (map[string]*proxy.BackendPool, map[string]*proxy.Forwarder, error) {
	pools := make(map[string]*proxy.BackendPool, len(backends))
	forwarders := make(map[string]*proxy.Forwarder, len(backends))

	for name, be := range backends {
		pool, fwd, err := compileBackend(be, caPools, tlsCerts)
		if err != nil {
			return nil, nil, fmt.Errorf("backend %q: %w", name, err)
		}
		// nil pool means the backend is unhealthy and was skipped.
		if pool == nil {
			continue
		}
		pools[name] = pool
		forwarders[name] = fwd
	}

	return pools, forwarders, nil
}

// compileBackend compiles a single Backend into a BackendPool and its dedicated
// Forwarder. Returns (nil, nil, nil) for unhealthy backends so they are skipped.
func compileBackend(
	be *backendv1.Backend,
	caPools map[string]*x509.CertPool,
	tlsCerts map[string]tls.Certificate,
) (*proxy.BackendPool, *proxy.Forwarder, error) {
	if !be.GetStatus().GetHealthy() {
		return nil, nil, nil
	}

	serverURL, err := url.Parse(be.GetConfig().GetServer())
	if err != nil {
		return nil, nil, fmt.Errorf("parsing server URL %q: %w", be.GetConfig().GetServer(), err)
	}

	// Build per-backend TLS configuration.
	tlsCfg := &tls.Config{
		InsecureSkipVerify: be.GetConfig().GetInsecureSkipTlsVerify(),
	}

	// Resolve optional CA reference.
	if ref := be.GetConfig().GetCaRef(); ref != "" {
		pool, ok := caPools[ref]
		if !ok {
			return nil, nil, fmt.Errorf("ca_ref %q not found", ref)
		}
		tlsCfg.RootCAs = pool
	}

	// Resolve optional client certificate reference.
	if ref := be.GetConfig().GetAuthRef(); ref != "" {
		cert, ok := tlsCerts[ref]
		if !ok {
			return nil, nil, fmt.Errorf("auth_ref %q not found", ref)
		}
		tlsCfg.Certificates = []tls.Certificate{cert}
	}

	transport := buildTLSTransport(tlsCfg)
	fwd := proxy.NewForwarder(transport)

	target := &proxy.BackendTarget{
		ID:     be.GetMeta().GetName(),
		URL:    serverURL,
		Weight: 1,
	}
	target.Healthy.Store(true)

	pool := &proxy.BackendPool{
		Name:    be.GetMeta().GetName(),
		Targets: []*proxy.BackendTarget{target},
	}

	return pool, fwd, nil
}

// compileRoutes builds per-SNI VirtualHostRuntimes from the route list.
// It returns the ExactHosts map and WildcardHosts slice ready for a ListenerRuntime.
func compileRoutes(
	routes map[string]*routev1.Route,
	pools map[string]*proxy.BackendPool,
	forwarders map[string]*proxy.Forwarder,
) (map[string]*proxy.VirtualHostRuntime, []proxy.WildcardHostRuntime, error) {
	exactHosts := map[string]*proxy.VirtualHostRuntime{}
	// wildcardVHosts accumulates wildcard vhosts by suffix before converting to a slice.
	wildcardVHosts := map[string]*proxy.VirtualHostRuntime{}

	for name, route := range routes {
		ref := route.GetConfig().GetBackendRef()
		pool, ok := pools[ref]
		if !ok {
			// Backend not found or unhealthy – skip this route.
			continue
		}
		fwd, ok := forwarders[ref]
		if !ok {
			continue
		}

		match := route.GetConfig().GetMatch()
		pathPrefix := match.GetPathPrefix()
		pathType := compilePathType(pathPrefix)

		// Base handler: forward to the backend pool.
		handler := fwd.Handler(pool)

		// Wrap with JWT middleware if a JWT match is configured.
		if jwt := match.GetJwt(); jwt != nil && jwt.GetClaim() != "" {
			handler = jwtMiddleware(jwt.GetClaim(), jwt.GetValue())(handler)
		}

		compiled := &proxy.CompiledRoute{
			Name:        name,
			PathType:    pathType,
			Path:        pathPrefix,
			BackendPool: pool,
			Handler:     handler,
		}

		// Compile optional header match.
		if hm := match.GetHeader(); hm != nil && hm.GetName() != "" {
			compiled.Headers = []proxy.HeaderMatch{
				{Name: hm.GetName(), Value: hm.GetValue()},
			}
		}

		// Determine which virtual host this route belongs to based on SNI.
		key, isWildcard := sniKey(match.GetSni())

		if isWildcard {
			vh := getOrCreateVHost(wildcardVHosts, key)
			placeRoute(vh, compiled)
		} else {
			vh := getOrCreateVHost(exactHosts, key)
			placeRoute(vh, compiled)
		}
	}

	// Sort prefix paths in every virtual host by descending path length
	// so that the most specific prefix wins.
	for _, vh := range exactHosts {
		sortPrefixPaths(vh)
	}
	for _, vh := range wildcardVHosts {
		sortPrefixPaths(vh)
	}

	// Convert wildcardVHosts map into the slice expected by ListenerRuntime.
	wildcardSlice := make([]proxy.WildcardHostRuntime, 0, len(wildcardVHosts))
	for suffix, vh := range wildcardVHosts {
		wildcardSlice = append(wildcardSlice, proxy.WildcardHostRuntime{
			Suffix: suffix,
			VHost:  vh,
		})
	}
	// Sort wildcard entries by descending suffix length for deterministic behaviour.
	sort.Slice(wildcardSlice, func(i, j int) bool {
		return len(wildcardSlice[i].Suffix) > len(wildcardSlice[j].Suffix)
	})

	return exactHosts, wildcardSlice, nil
}

// placeRoute inserts a CompiledRoute into the correct bucket of a VirtualHostRuntime.
func placeRoute(vh *proxy.VirtualHostRuntime, route *proxy.CompiledRoute) {
	switch route.PathType {
	case proxy.PathExact:
		vh.ExactPaths[route.Path] = append(vh.ExactPaths[route.Path], route)
	case proxy.PathPrefix:
		vh.PrefixPaths = append(vh.PrefixPaths, route)
	default: // PathAny / catch-all
		vh.CatchAll = append(vh.CatchAll, route)
	}
}

// sortPrefixPaths sorts a VirtualHostRuntime's PrefixPaths by descending path
// length so that the longest (most specific) prefix matches first.
func sortPrefixPaths(vh *proxy.VirtualHostRuntime) {
	sort.Slice(vh.PrefixPaths, func(i, j int) bool {
		return len(vh.PrefixPaths[i].Path) > len(vh.PrefixPaths[j].Path)
	})
}

// buildTLSTransport constructs an *http.Transport using the supplied tls.Config.
func buildTLSTransport(cfg *tls.Config) http.RoundTripper {
	return &http.Transport{
		TLSClientConfig:     cfg,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
}

// compilePathType maps a path_prefix string to a PathMatchType.
// A non-empty path prefix means prefix matching; an empty one is a catch-all.
func compilePathType(pathPrefix string) proxy.PathMatchType {
	if pathPrefix != "" {
		return proxy.PathPrefix
	}
	return proxy.PathAny
}

// jwtMiddleware returns an HTTP middleware that validates that the JWT Bearer
// token in the Authorization header contains the expected claim value.
// Token signature is NOT verified. Only claim presence and value are checked.
func jwtMiddleware(claim, value string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, bearerPrefix)
			parts := strings.Split(tokenStr, ".")
			if len(parts) != 3 {
				http.Error(w, "malformed JWT", http.StatusUnauthorized)
				return
			}

			// Decode the payload (second segment); add padding as required.
			seg := parts[1]
			if rem := len(seg) % 4; rem != 0 {
				seg += strings.Repeat("=", 4-rem)
			}
			payloadBytes, err := base64.URLEncoding.DecodeString(seg)
			if err != nil {
				http.Error(w, "invalid JWT payload encoding", http.StatusUnauthorized)
				return
			}

			var claims map[string]any
			if err := json.Unmarshal(payloadBytes, &claims); err != nil {
				http.Error(w, "invalid JWT payload JSON", http.StatusUnauthorized)
				return
			}

			got, ok := claims[claim]
			if !ok {
				http.Error(w, fmt.Sprintf("JWT claim %q not found", claim), http.StatusUnauthorized)
				return
			}

			// Normalise to string for comparison; encoding/json uses float64 for numbers.
			if fmt.Sprintf("%v", got) != value {
				http.Error(w, fmt.Sprintf("JWT claim %q value mismatch", claim), http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// sniKey classifies an SNI string:
func sniKey(sni string) (key string, isWildcard bool) {
	if sni == "" {
		return "*", false
	}
	if strings.HasPrefix(sni, "*.") {
		return sni[1:], true // "*.example.com" → ".example.com"
	}
	return sni, false
}

// getOrCreateVHost returns the VirtualHostRuntime stored at key in m, creating
// and inserting a new empty one if absent.
func getOrCreateVHost(m map[string]*proxy.VirtualHostRuntime, key string) *proxy.VirtualHostRuntime {
	vh, ok := m[key]
	if !ok {
		vh = &proxy.VirtualHostRuntime{
			ExactPaths:  map[string][]*proxy.CompiledRoute{},
			PrefixPaths: []*proxy.CompiledRoute{},
			CatchAll:    []*proxy.CompiledRoute{},
		}
		m[key] = vh
	}
	return vh
}
