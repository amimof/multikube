package compile

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/textproto"
	"net/url"
	"sort"
	"strings"
	"sync/atomic"
	"time"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	policyv1 "github.com/amimof/multikube/api/policy/v1"
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
	Policies               map[string]*policyv1.Policy
	Credentials            map[string]*credentialv1.Credential
}

// Compiler compiles a State into a proxy Runtime.
// It holds no shared state and each Compile call is self-contained.
type Compiler struct {
	version atomic.Uint64
}

const (
	RoutePhaseReady    = "READY"
	RoutePhaseInvalid  = "INVALID"
	RoutePhaseConflict = "CONFLICT"
)

type RouteCompileStatus struct {
	Phase  string
	Reason string
}

type CompileResult struct {
	Runtime       *proxy.RuntimeConfig
	RouteStatuses map[string]RouteCompileStatus
}

// NewCompiler returns a new Compiler2.
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Compile converts the contents of a State into a runtime snapshot and route statuses.
func (c *Compiler) Compile(st *State) (*CompileResult, error) {
	// compile TLS client certificates first; CAs may reference them.
	tlsCerts, err := compileCerts(st.Certificates)
	if err != nil {
		return nil, fmt.Errorf("compile certs: %w", err)
	}

	// compile CA certificate pools.
	caPools, err := compileCAs(st.CertificateAuthorities, st.Certificates)
	if err != nil {
		return nil, fmt.Errorf("compile CAs: %w", err)
	}

	compiledCreds, err := compileCredentials(st.Credentials)
	if err != nil {
		return nil, fmt.Errorf("compile credentials: %w", err)
	}

	// Compile backends into a BackendPool
	backendPools, forwarders, err := compileBackendsPools(st.Backends, caPools, tlsCerts, compiledCreds)
	if err != nil {
		return nil, fmt.Errorf("compile backend pools: %w", err)
	}

	// compile routes into CompiledRoutes.
	routes, statuses := compileRoutes2(st.Routes, backendPools, forwarders)

	for name := range st.Routes {
		if _, ok := statuses[name]; !ok {
			statuses[name] = RouteCompileStatus{Phase: RoutePhaseReady}
		}
	}

	rt := &proxy.RuntimeConfig{
		Version:  c.version.Add(1),
		Backends: backendPools,
		Routes:   routes,
		Policies: compilePolicies(st.Policies),
	}

	return &CompileResult{
		Runtime:       rt,
		RouteStatuses: statuses,
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
	certPEM := cert.GetConfig().GetCertificateData()
	if certPEM == "" {
		certPEM = cert.GetConfig().GetCertificate()
	}
	if certPEM == "" {
		return tls.Certificate{}, fmt.Errorf("certificate has no inline PEM data")
	}

	keyPEM := cert.GetConfig().GetKeyData()
	if keyPEM == "" {
		keyPEM = cert.GetConfig().GetKey()
	}
	if keyPEM == "" {
		return tls.Certificate{}, fmt.Errorf("key has no inline PEM data")
	}

	tlsCert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("X509KeyPair: %w", err)
	}
	return tlsCert, nil
}

type compiledCredential struct {
	clientCertificateRef string
	authInjector         proxy.RequestAuthInjector
}

func compileCredentials(credentials map[string]*credentialv1.Credential) (map[string]compiledCredential, error) {
	if credentials == nil {
		return map[string]compiledCredential{}, nil
	}
	out := make(map[string]compiledCredential, len(credentials))
	for name, credential := range credentials {
		compiled, err := compileCredential(credential)
		if err != nil {
			return nil, fmt.Errorf("credential %q: %w", name, err)
		}
		out[name] = compiled
	}
	return out, nil
}

func compileCredential(credential *credentialv1.Credential) (compiledCredential, error) {
	config := credential.GetConfig()
	if config == nil {
		return compiledCredential{}, fmt.Errorf("missing config")
	}

	switch {
	case config.GetClientCertificateRef() != "":
		return compiledCredential{clientCertificateRef: config.GetClientCertificateRef()}, nil
	case config.GetToken() != "":
		return compiledCredential{authInjector: bearerTokenInjector{token: config.GetToken()}}, nil
	case config.GetBasic() != nil:
		basic := config.GetBasic()
		return compiledCredential{authInjector: basicAuthInjector{username: basic.GetUsername(), password: basic.GetPassword()}}, nil
	default:
		return compiledCredential{}, fmt.Errorf("credential has no supported auth material")
	}
}

func compileBackendsPools(
	backends map[string]*backendv1.Backend,
	caPools map[string]*x509.CertPool,
	tlsCerts map[string]tls.Certificate,
	credentials map[string]compiledCredential,
) (map[string]*proxy.BackendPool, map[string]*proxy.Forwarder, error) {
	out := make(map[string]*proxy.BackendPool, len(backends))
	fwds := make(map[string]*proxy.Forwarder, len(backends))

	for name, be := range backends {

		// TODO: Check for healthyness but only if it's cheap
		// if !be.GetStatus().GetHealthy() {
		// 	continue
		// }

		br, fwd, err := compileBackendPool(be, caPools, tlsCerts, credentials)
		if err != nil {
			return nil, nil, fmt.Errorf("backend %q: %w", name, err)
		}

		out[name] = br
		fwds[name] = fwd
	}

	return out, fwds, nil
}

// compileBackend2 converts a single Backend proto into a BackendRuntime and its
// dedicated Forwarder.
func compileBackendPool(
	be *backendv1.Backend,
	caPools map[string]*x509.CertPool,
	tlsCerts map[string]tls.Certificate,
	credentials map[string]compiledCredential,
) (*proxy.BackendPool, *proxy.Forwarder, error) {
	tlsCfg := &tls.Config{
		InsecureSkipVerify: be.GetConfig().GetInsecureSkipTlsVerify(), //nolint:gosec // user-controlled
	}

	if ref := be.GetConfig().GetCaRef(); ref != "" {
		pool, ok := caPools[ref]
		if !ok {
			return nil, nil, fmt.Errorf("ca_ref %q not found", ref)
		}
		tlsCfg.RootCAs = pool
	}

	if ref := be.GetConfig().GetAuthRef(); ref != "" {
		credential, ok := credentials[ref]
		if !ok {
			return nil, nil, fmt.Errorf("auth_ref %q not found", ref)
		}
		if credential.clientCertificateRef != "" {
			cert, ok := tlsCerts[credential.clientCertificateRef]
			if !ok {
				return nil, nil, fmt.Errorf("auth_ref %q references missing certificate %q", ref, credential.clientCertificateRef)
			}
			tlsCfg.Certificates = []tls.Certificate{cert}
		}
	}

	var cacheTTL time.Duration
	if pb := be.GetConfig().GetCacheTtl(); pb != nil {
		cacheTTL = pb.AsDuration()
	}

	transport := buildTLSTransport(tlsCfg)
	fwd := proxy.NewForwarder(transport)

	out := []*proxy.BackendRuntime{}

	for _, server := range be.GetConfig().GetServers() {
		serverURL, err := url.Parse(server)
		if err != nil {
			return nil, nil, fmt.Errorf("parsing server URL %q: %w", server, err)
		}
		if len(server) == 0 {
			return nil, nil, fmt.Errorf("server URL is empty")
		}
		br := &proxy.BackendRuntime{
			Name:      be.GetMeta().GetName(),
			Labels:    be.GetMeta().GetLabels(),
			URL:       serverURL,
			CacheTTL:  cacheTTL,
			TLSConfig: tlsCfg,
			Transport: transport,
		}
		if ref := be.GetConfig().GetAuthRef(); ref != "" {
			br.AuthInjector = credentials[ref].authInjector
		}
		out = append(out, br)
	}

	pool := &proxy.BackendPool{
		Name:    be.GetConfig().GetName(),
		Targets: out,
	}

	return pool, fwd, nil
}

type bearerTokenInjector struct {
	token string
}

func (i bearerTokenInjector) Apply(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+i.token)
	return nil
}

type basicAuthInjector struct {
	username string
	password string
}

func (i basicAuthInjector) Apply(req *http.Request) error {
	encoded := base64.StdEncoding.EncodeToString([]byte(i.username + ":" + i.password))
	req.Header.Set("Authorization", "Basic "+encoded)
	return nil
}

type routeMatchKey struct {
	kind  proxy.RouteMatchKind
	value string
}

// compileRoutes2 classifies each route into the correct CompiledRoutes bucket
// and builds an http.Handler for it.
func compileRoutes2(
	routes map[string]*routev1.Route,
	backends map[string]*proxy.BackendPool,
	forwarders map[string]*proxy.Forwarder,
) (proxy.CompiledRoutes, map[string]RouteCompileStatus) {
	cr := proxy.CompiledRoutes{
		SNIExact: make(map[string][]*proxy.RouteRuntime),
	}
	statuses := make(map[string]RouteCompileStatus, len(routes))
	matchOwners := map[routeMatchKey]string{}

	for name, route := range routes {
		rr, matchKey, status := compileRoute(name, route, backends, forwarders)
		if status.Phase != "" {
			statuses[name] = status
			continue
		}

		if owner, exists := matchOwners[matchKey]; exists {
			reason := fmt.Sprintf("route matcher conflicts with %q", owner)
			statuses[name] = RouteCompileStatus{Phase: RoutePhaseConflict, Reason: reason}
			statuses[owner] = RouteCompileStatus{Phase: RoutePhaseConflict, Reason: fmt.Sprintf("route matcher conflicts with %q", name)}
			removeCompiledRoute(&cr, matchKey, owner)
			continue
		}
		matchOwners[matchKey] = name
		appendCompiledRoute(&cr, rr)
	}

	// Sort path-prefix routes by descending length so the most specific prefix wins.
	sort.Slice(cr.PathPrefixes, func(i, j int) bool {
		return len(cr.PathPrefixes[i].PathPrefix) > len(cr.PathPrefixes[j].PathPrefix)
	})

	return cr, statuses
}

func compileRoute(
	name string,
	route *routev1.Route,
	backends map[string]*proxy.BackendPool,
	forwarders map[string]*proxy.Forwarder,
) (*proxy.RouteRuntime, routeMatchKey, RouteCompileStatus) {
	config := route.GetConfig()
	if config == nil {
		return nil, routeMatchKey{}, RouteCompileStatus{Phase: RoutePhaseInvalid, Reason: "missing route config"}
	}

	match := config.GetMatch()
	if match == nil {
		return nil, routeMatchKey{}, RouteCompileStatus{Phase: RoutePhaseInvalid, Reason: "route matcher is required"}
	}

	ref := config.GetBackendRef()

	br, ok := backends[ref]
	if !ok {
		return nil, routeMatchKey{}, RouteCompileStatus{Phase: RoutePhaseInvalid, Reason: fmt.Sprintf("backend_ref %q not found", ref)}
	}

	fwd, ok := forwarders[ref]
	if !ok {
		return nil, routeMatchKey{}, RouteCompileStatus{Phase: RoutePhaseInvalid, Reason: fmt.Sprintf("backend_ref %q has no forwarder", ref)}
	}

	// Build a single-target BackendPool so the Forwarder can pick a target.
	handler := fwd.Handler(br)

	rr := &proxy.RouteRuntime{
		Name:        name,
		BackendPool: br,
		Handler:     handler,
	}

	switch {
	case match.GetHeader().GetName() != "":
		hm := match.GetHeader()
		rr.Kind = proxy.RouteMatchKindHeader
		rr.Header = &proxy.HeaderRuntime{
			Name:      hm.GetName(),
			Canonical: textproto.CanonicalMIMEHeaderKey(hm.GetName()),
			Value:     hm.GetValue(),
		}
		return rr, routeMatchKey{kind: rr.Kind, value: textproto.CanonicalMIMEHeaderKey(hm.GetName()) + "=" + hm.GetValue()}, RouteCompileStatus{}

	case match.GetPath() != "":
		rr.Kind = proxy.RouteMatchKindPath
		rr.Path = match.GetPath()
		return rr, routeMatchKey{kind: rr.Kind, value: rr.Path}, RouteCompileStatus{}

	case match.GetPathPrefix() != "":
		rr.Kind = proxy.RouteMatchKindPathPrefix
		rr.PathPrefix = match.GetPathPrefix()
		return rr, routeMatchKey{kind: rr.Kind, value: rr.PathPrefix}, RouteCompileStatus{}

	case match.GetSni() != "":
		rr.Kind = proxy.RouteMatchKindSNI
		rr.SNI = match.GetSni()
		return rr, routeMatchKey{kind: rr.Kind, value: strings.ToLower(rr.SNI)}, RouteCompileStatus{}

	case match.GetJwt().GetClaim() != "":
		jm := match.GetJwt()
		rr.Kind = proxy.RouteMatchKindJWT
		rr.JWT = &proxy.JWTRuntime{
			Claim: jm.GetClaim(),
			Value: jm.GetValue(),
		}
		return rr, routeMatchKey{kind: rr.Kind, value: jm.GetClaim() + "=" + jm.GetValue()}, RouteCompileStatus{}
	default:
		return nil, routeMatchKey{}, RouteCompileStatus{Phase: RoutePhaseInvalid, Reason: "route matcher is required"}
	}
}

func appendCompiledRoute(cr *proxy.CompiledRoutes, rr *proxy.RouteRuntime) {
	switch rr.Kind {
	case proxy.RouteMatchKindPath:
		cr.Paths = append(cr.Paths, rr)
	case proxy.RouteMatchKindPathPrefix:
		cr.PathPrefixes = append(cr.PathPrefixes, rr)
	case proxy.RouteMatchKindHeader:
		cr.Headers = append(cr.Headers, rr)
	case proxy.RouteMatchKindSNI:
		cr.SNIExact[rr.SNI] = append(cr.SNIExact[rr.SNI], rr)
	case proxy.RouteMatchKindJWT:
		cr.JWT = append(cr.JWT, rr)
	}
}

func removeCompiledRoute(cr *proxy.CompiledRoutes, key routeMatchKey, name string) {
	switch key.kind {
	case proxy.RouteMatchKindPath:
		cr.Paths = removeRouteRuntime(cr.Paths, name)
	case proxy.RouteMatchKindPathPrefix:
		cr.PathPrefixes = removeRouteRuntime(cr.PathPrefixes, name)
	case proxy.RouteMatchKindHeader:
		cr.Headers = removeRouteRuntime(cr.Headers, name)
	case proxy.RouteMatchKindSNI:
		for sni, routes := range cr.SNIExact {
			updated := removeRouteRuntime(routes, name)
			if len(updated) == 0 {
				delete(cr.SNIExact, sni)
				continue
			}
			cr.SNIExact[sni] = updated
		}
	case proxy.RouteMatchKindJWT:
		cr.JWT = removeRouteRuntime(cr.JWT, name)
	}
}

func removeRouteRuntime(routes []*proxy.RouteRuntime, name string) []*proxy.RouteRuntime {
	for i, route := range routes {
		if route.Name == name {
			return append(routes[:i], routes[i+1:]...)
		}
	}
	return routes
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

// compilePolicies converts a map of Policy proto objects into a slice for the
// RuntimeConfig. Policy evaluation is done at request time from this slice.
func compilePolicies(policies map[string]*policyv1.Policy) []*policyv1.Policy {
	out := make([]*policyv1.Policy, 0, len(policies))
	for _, p := range policies {
		out = append(out, p)
	}
	return out
}
