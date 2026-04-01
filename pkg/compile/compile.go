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

// NewCompiler returns a new Compiler2.
func NewCompiler() *Compiler {
	return &Compiler{}
}

// Compile converts the contents of a State into a *proxy.RuntimeConfig that
// the proxy can use to match and forward requests.
func (c *Compiler) Compile(st *State) (*proxy.RuntimeConfig, error) {
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

	// compile backends into BackendRuntimes and per-backend Forwarders.
	backends, forwarders, err := compileBackends2(st.Backends, caPools, tlsCerts, compiledCreds)
	if err != nil {
		return nil, fmt.Errorf("compile backends: %w", err)
	}

	// compile routes into CompiledRoutes.
	routes, err := compileRoutes2(st.Routes, backends, forwarders)
	if err != nil {
		return nil, fmt.Errorf("compile routes: %w", err)
	}

	return &proxy.RuntimeConfig{
		Version:  c.version.Add(1),
		Backends: backends,
		Routes:   routes,
		Policies: compilePolicies(st.Policies),
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

// compileBackends2 builds a BackendRuntime and a Forwarder for every healthy
// backend. Unhealthy backends are silently skipped.
func compileBackends2(
	backends map[string]*backendv1.Backend,
	caPools map[string]*x509.CertPool,
	tlsCerts map[string]tls.Certificate,
	credentials map[string]compiledCredential,
) (map[string]*proxy.BackendRuntime, map[string]*proxy.Forwarder, error) {
	out := make(map[string]*proxy.BackendRuntime, len(backends))
	fwds := make(map[string]*proxy.Forwarder, len(backends))

	for name, be := range backends {

		// TODO: Check for healthyness but only if it's cheap
		// if !be.GetStatus().GetHealthy() {
		// 	continue
		// }

		br, fwd, err := compileBackend2(be, caPools, tlsCerts, credentials)
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
func compileBackend2(
	be *backendv1.Backend,
	caPools map[string]*x509.CertPool,
	tlsCerts map[string]tls.Certificate,
	credentials map[string]compiledCredential,
) (*proxy.BackendRuntime, *proxy.Forwarder, error) {
	serverURL, err := url.Parse(be.GetConfig().GetServer())
	if err != nil {
		return nil, nil, fmt.Errorf("parsing server URL %q: %w", be.GetConfig().GetServer(), err)
	}

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

	return br, fwd, nil
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

// compileRoutes2 classifies each route into the correct CompiledRoutes bucket
// and builds an http.Handler for it.
func compileRoutes2(
	routes map[string]*routev1.Route,
	backends map[string]*proxy.BackendRuntime,
	forwarders map[string]*proxy.Forwarder,
) (proxy.CompiledRoutes, error) {
	cr := proxy.CompiledRoutes{
		SNIExact: make(map[string][]*proxy.RouteRuntime),
	}

	for name, route := range routes {
		ref := route.GetConfig().GetBackendRef()

		br, ok := backends[ref]
		if !ok {
			// Backend missing or unhealthy, skip silently.
			continue
		}

		fwd, ok := forwarders[ref]
		if !ok {
			continue
		}

		// Build a single-target BackendPool so the Forwarder can pick a target.
		pool := backendPoolFromRuntime(br)
		handler := fwd.Handler(pool)

		rr := &proxy.RouteRuntime{
			Name:        name,
			BackendPool: pool,
			Handler:     handler,
		}

		match := route.GetConfig().GetMatch()

		switch {
		case match.GetHeader().GetName() != "":
			hm := match.GetHeader()
			rr.Kind = proxy.RouteMatchKindHeader
			rr.Header = &proxy.HeaderRuntime{
				Name:      hm.GetName(),
				Canonical: textproto.CanonicalMIMEHeaderKey(hm.GetName()),
				Value:     hm.GetValue(),
			}
			cr.Headers = append(cr.Headers, rr)

		case match.GetPath() != "":
			rr.Kind = proxy.RouteMatchKindPath
			rr.Path = match.GetPath()
			cr.Paths = append(cr.Paths, rr)

		case match.GetPathPrefix() != "":
			rr.Kind = proxy.RouteMatchKindPathPrefix
			rr.PathPrefix = match.GetPathPrefix()
			cr.PathPrefixes = append(cr.PathPrefixes, rr)

		case match.GetSni() != "":
			rr.Kind = proxy.RouteMatchKindSNI
			rr.SNI = match.GetSni()
			cr.SNIExact[rr.SNI] = append(cr.SNIExact[rr.SNI], rr)

		case match.GetJwt().GetClaim() != "":
			jm := match.GetJwt()
			rr.Kind = proxy.RouteMatchKindJWT
			rr.JWT = &proxy.JWTRuntime{
				Claim: jm.GetClaim(),
				Value: jm.GetValue(),
			}
			cr.JWT = append(cr.JWT, rr)
		default:
			if cr.Default != nil {
				return proxy.CompiledRoutes{}, fmt.Errorf(
					"route %q: multiple default routes not allowed (conflicts with %q)",
					name, cr.Default.Name,
				)
			}
			cr.Default = rr
		}
	}

	// Sort path-prefix routes by descending length so the most specific prefix wins.
	sort.Slice(cr.PathPrefixes, func(i, j int) bool {
		return len(cr.PathPrefixes[i].PathPrefix) > len(cr.PathPrefixes[j].PathPrefix)
	})

	return cr, nil
}

// backendPoolFromRuntime wraps a BackendRuntime in a single-target BackendPool
// compatible with the existing Forwarder.Handler signature.
func backendPoolFromRuntime(br *proxy.BackendRuntime) *proxy.BackendPool {
	return &proxy.BackendPool{
		Name:    br.Name,
		Targets: []*proxy.BackendRuntime{br},
	}
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
