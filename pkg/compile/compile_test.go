package compile

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certificatev1 "github.com/amimof/multikube/api/certificate/v1"
	credentialv1 "github.com/amimof/multikube/api/credential/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
	proxy "github.com/amimof/multikube/pkg/proxyv2"
)

// ---------------------------------------------------------------------------
// Helpers — self-signed cert + key generation
// ---------------------------------------------------------------------------

// selfSignedPEM returns a self-signed certificate PEM and its private-key PEM.
func selfSignedPEM(t *testing.T) (certPEM, keyPEM string) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Minute),
		NotAfter:     time.Now().Add(time.Hour),
		IsCA:         true,
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}

	certPEM = string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))

	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	keyPEM = string(pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}))
	return
}

// ---------------------------------------------------------------------------
// Fixture constructors
// ---------------------------------------------------------------------------

func newBackend(name, server string) *backendv1.Backend {
	return &backendv1.Backend{
		Meta:   &metav1.Meta{Name: name},
		Config: &backendv1.BackendConfig{Servers: []string{server}, InsecureSkipTlsVerify: true},
	}
}

func requireSingleBackendTarget(t *testing.T, pool *proxy.BackendPool) *proxy.BackendRuntime {
	t.Helper()
	if pool == nil {
		t.Fatal("expected backend pool")
	}
	if len(pool.Targets) != 1 {
		t.Fatalf("expected 1 backend target, got %d", len(pool.Targets))
	}
	return pool.Targets[0]
}

func newRoute(name, backendRef string, match *routev1.Match) *routev1.Route {
	return &routev1.Route{
		Meta: &metav1.Meta{Name: name},
		Config: &routev1.RouteConfig{
			BackendRef: backendRef,
			Match:      match,
		},
	}
}

func newCertificate(name, certPEM, keyPEM string) *certificatev1.Certificate {
	return &certificatev1.Certificate{
		Meta: &metav1.Meta{Name: name},
		Config: &certificatev1.CertificateConfig{
			Certificate: certPEM,
			Key:         keyPEM,
		},
	}
}

func newCAFromRef(name, certRef string) *cav1.CertificateAuthority {
	return &cav1.CertificateAuthority{
		Meta: &metav1.Meta{Name: name},
		Config: &cav1.CertificateAuthorityConfig{
			Certificate: certRef,
		},
	}
}

func newCAInline(name, certPEM string) *cav1.CertificateAuthority {
	return &cav1.CertificateAuthority{
		Meta: &metav1.Meta{Name: name},
		Config: &cav1.CertificateAuthorityConfig{
			CertificateData: certPEM,
		},
	}
}

func newTokenCredential(name, token string) *credentialv1.Credential {
	return &credentialv1.Credential{
		Meta: &metav1.Meta{Name: name},
		Config: &credentialv1.CredentialConfig{
			Token: token,
		},
	}
}

func newBasicCredential(name, username, password string) *credentialv1.Credential {
	return &credentialv1.Credential{
		Meta: &metav1.Meta{Name: name},
		Config: &credentialv1.CredentialConfig{
			Basic: &credentialv1.CredentialBasic{
				Username: username,
				Password: password,
			},
		},
	}
}

func newClientCertCredential(name, certRef string) *credentialv1.Credential {
	return &credentialv1.Credential{
		Meta: &metav1.Meta{Name: name},
		Config: &credentialv1.CredentialConfig{
			ClientCertificateRef: certRef,
		},
	}
}

// ---------------------------------------------------------------------------
// Tests — Compiler.Compile happy paths
// ---------------------------------------------------------------------------

func TestCompile_EmptyState(t *testing.T) {
	c := NewCompiler()
	st := &State{
		Backends:               map[string]*backendv1.Backend{},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res == nil || res.Runtime == nil {
		t.Fatal("expected non-nil RuntimeConfig")
	}
	if res.Runtime.Version != 1 {
		t.Errorf("expected version 1, got %d", res.Runtime.Version)
	}
}

func TestCompile_VersionIncrement(t *testing.T) {
	c := NewCompiler()
	emptyState := func() *State {
		return &State{
			Backends:               map[string]*backendv1.Backend{},
			Routes:                 map[string]*routev1.Route{},
			Certificates:           map[string]*certificatev1.Certificate{},
			CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
			Credentials:            map[string]*credentialv1.Credential{},
		}
	}

	for i := uint64(1); i <= 3; i++ {
		res, err := c.Compile(emptyState())
		if err != nil {
			t.Fatalf("compile %d: %v", i, err)
		}
		if res.Runtime.Version != i {
			t.Errorf("compile %d: expected version %d, got %d", i, i, res.Runtime.Version)
		}
	}
}

func TestCompile_RouteWithoutMatcher_Invalid(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", srv.URL),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", nil),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Runtime.Routes.Paths)+len(res.Runtime.Routes.PathPrefixes)+len(res.Runtime.Routes.Headers)+len(res.Runtime.Routes.JWT)+len(res.Runtime.Routes.SNIExact) != 0 {
		t.Fatal("expected route without matcher to be excluded from runtime")
	}
	status := res.RouteStatuses["r"]
	if status.Phase != RoutePhaseInvalid {
		t.Fatalf("expected route phase %q, got %q", RoutePhaseInvalid, status.Phase)
	}
	if status.Reason == "" {
		t.Fatal("expected invalid route reason")
	}
}

func TestCompile_HeaderRoute(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", srv.URL),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", &routev1.Match{
				Header: &routev1.HeaderMatch{Name: "X-Tenant", Value: "acme"},
			}),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Runtime.Routes.Headers) != 1 {
		t.Fatalf("expected 1 header route, got %d", len(res.Runtime.Routes.Headers))
	}
	rr := res.Runtime.Routes.Headers[0]
	if rr.Kind != proxy.RouteMatchKindHeader {
		t.Errorf("expected kind Header, got %v", rr.Kind)
	}
	if rr.Header == nil || rr.Header.Name != "X-Tenant" || rr.Header.Value != "acme" {
		t.Errorf("unexpected header match: %+v", rr.Header)
	}
}

func TestCompile_PathRoute(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", srv.URL),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", &routev1.Match{Path: "/exact"}),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Runtime.Routes.Paths) != 1 {
		t.Fatalf("expected 1 path route, got %d", len(res.Runtime.Routes.Paths))
	}
	rr := res.Runtime.Routes.Paths[0]
	if rr.Kind != proxy.RouteMatchKindPath {
		t.Errorf("expected kind Path, got %v", rr.Kind)
	}
	if rr.Path != "/exact" {
		t.Errorf("expected path /exact, got %q", rr.Path)
	}
}

func TestCompile_PathPrefixRoute_SortedLongestFirst(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", srv.URL),
		},
		Routes: map[string]*routev1.Route{
			"short": newRoute("short", "be", &routev1.Match{PathPrefix: "/api"}),
			"long":  newRoute("long", "be", &routev1.Match{PathPrefix: "/api/v2"}),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Runtime.Routes.PathPrefixes) != 2 {
		t.Fatalf("expected 2 path-prefix routes, got %d", len(res.Runtime.Routes.PathPrefixes))
	}
	// Longest prefix must come first.
	if res.Runtime.Routes.PathPrefixes[0].PathPrefix != "/api/v2" {
		t.Errorf("expected longest prefix first, got %q", res.Runtime.Routes.PathPrefixes[0].PathPrefix)
	}
}

func TestCompile_SNIRoute(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", srv.URL),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", &routev1.Match{Sni: "myservice.example.com"}),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	routes, ok := res.Runtime.Routes.SNIExact["myservice.example.com"]
	if !ok || len(routes) != 1 {
		t.Fatalf("expected 1 SNI route for host, got %v", res.Runtime.Routes.SNIExact)
	}
	if routes[0].Kind != proxy.RouteMatchKindSNI {
		t.Errorf("expected kind SNI, got %v", routes[0].Kind)
	}
}

func TestCompile_JWTRoute(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", srv.URL),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", &routev1.Match{
				Jwt: &routev1.JWTMatch{Claim: "team", Value: "platform"},
			}),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Runtime.Routes.JWT) != 1 {
		t.Fatalf("expected 1 jwt route, got %d", len(res.Runtime.Routes.JWT))
	}
	rr := res.Runtime.Routes.JWT[0]
	if rr.Kind != proxy.RouteMatchKindJWT {
		t.Errorf("expected kind JWT, got %v", rr.Kind)
	}
	if rr.JWT == nil || rr.JWT.Claim != "team" || rr.JWT.Value != "platform" {
		t.Errorf("unexpected jwt match: %+v", rr.JWT)
	}
	if rr.Handler == nil {
		t.Fatal("expected handler to be set")
	}
}

// ---------------------------------------------------------------------------
// Tests — error conditions
// ---------------------------------------------------------------------------

func TestCompile_ConflictingRoutes_MarkedConflict(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", srv.URL),
		},
		Routes: map[string]*routev1.Route{
			"r1": newRoute("r1", "be", &routev1.Match{Path: "/same"}),
			"r2": newRoute("r2", "be", &routev1.Match{Path: "/same"}),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err == nil {
		if len(res.Runtime.Routes.Paths) != 0 {
			t.Fatal("expected conflicting routes to be excluded from runtime")
		}
		if res.RouteStatuses["r1"].Phase != RoutePhaseConflict {
			t.Fatalf("expected r1 conflict status, got %+v", res.RouteStatuses["r1"])
		}
		if res.RouteStatuses["r2"].Phase != RoutePhaseConflict {
			t.Fatalf("expected r2 conflict status, got %+v", res.RouteStatuses["r2"])
		}
	}
}

func TestCompile_MissingBackendRef_RouteInvalid(t *testing.T) {
	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{}, // no backends
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "missing-backend", &routev1.Match{Path: "/x"}),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(res.Runtime.Routes.Paths) != 0 {
		t.Fatal("expected missing-backend route to be excluded from runtime")
	}
	if res.RouteStatuses["r"].Phase != RoutePhaseInvalid {
		t.Fatalf("expected invalid phase, got %+v", res.RouteStatuses["r"])
	}
}

// ---------------------------------------------------------------------------
// Tests — BackendPool wiring
// ---------------------------------------------------------------------------

func TestCompile_BackendPool_SingleTarget(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", srv.URL),
		},
		Routes: map[string]*routev1.Route{
			"r": newRoute("r", "be", &routev1.Match{Path: "/pool"}),
		},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(res.Runtime.Routes.Paths) != 1 {
		t.Fatalf("expected 1 compiled route, got %d", len(res.Runtime.Routes.Paths))
	}
	pool := res.Runtime.Routes.Paths[0].BackendPool
	if pool == nil {
		t.Fatal("BackendPool is nil")
	}
	if len(pool.Targets) != 1 {
		t.Fatalf("expected 1 target in pool, got %d", len(pool.Targets))
	}
	if pool.Targets[0].Name != "be" {
		t.Errorf("expected target name %q, got %q", "be", pool.Targets[0].Name)
	}
}

// ---------------------------------------------------------------------------
// Tests — CA compilation
// ---------------------------------------------------------------------------

func TestCompile_CA_InlinePEM(t *testing.T) {
	certPEM, keyPEM := selfSignedPEM(t)

	c := NewCompiler()
	st := &State{
		Backends:     map[string]*backendv1.Backend{},
		Routes:       map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{
			"myca": newCAInline("myca", certPEM),
		},
		Credentials: map[string]*credentialv1.Credential{},
	}
	_ = keyPEM // only the cert PEM is needed for a CA pool

	rc, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = rc // successful compile is enough; CA pool is internal to compileBackends
}

func TestCompile_CA_CertificateRef(t *testing.T) {
	certPEM, keyPEM := selfSignedPEM(t)

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{},
		Routes:   map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{
			"mycert": newCertificate("mycert", certPEM, keyPEM),
		},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{
			"myca": newCAFromRef("myca", "mycert"),
		},
		Credentials: map[string]*credentialv1.Credential{},
	}

	rc, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_ = rc
}

func TestCompile_CA_MissingCertRef_Error(t *testing.T) {
	c := NewCompiler()
	st := &State{
		Backends:     map[string]*backendv1.Backend{},
		Routes:       map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{}, // empty — ref won't resolve
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{
			"myca": newCAFromRef("myca", "does-not-exist"),
		},
		Credentials: map[string]*credentialv1.Credential{},
	}

	_, err := c.Compile(st)
	if err == nil {
		t.Fatal("expected error for missing cert ref, got nil")
	}
}

func TestCompile_CA_InvalidPEM_Error(t *testing.T) {
	c := NewCompiler()
	st := &State{
		Backends:     map[string]*backendv1.Backend{},
		Routes:       map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{
			"myca": newCAInline("myca", "not-valid-pem"),
		},
		Credentials: map[string]*credentialv1.Credential{},
	}

	_, err := c.Compile(st)
	if err == nil {
		t.Fatal("expected error for invalid PEM, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests — Certificate compilation
// ---------------------------------------------------------------------------

func TestCompile_Certificate_MissingCert_Error(t *testing.T) {
	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{},
		Routes:   map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{
			"bad": {
				Meta:   &metav1.Meta{Name: "bad"},
				Config: &certificatev1.CertificateConfig{Certificate: "", Key: "somekey"},
			},
		},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	_, err := c.Compile(st)
	if err == nil {
		t.Fatal("expected error for missing certificate PEM, got nil")
	}
}

func TestCompile_Certificate_MissingKey_Error(t *testing.T) {
	certPEM, _ := selfSignedPEM(t)

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{},
		Routes:   map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{
			"bad": {
				Meta:   &metav1.Meta{Name: "bad"},
				Config: &certificatev1.CertificateConfig{Certificate: certPEM, Key: ""},
			},
		},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	_, err := c.Compile(st)
	if err == nil {
		t.Fatal("expected error for missing key PEM, got nil")
	}
}

func TestCompile_BackendTokenCredential_AttachesAuthInjector(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": {
				Meta: &metav1.Meta{Name: "be"},
				Config: &backendv1.BackendConfig{
					Servers:               []string{srv.URL},
					InsecureSkipTlsVerify: true,
					AuthRef:               "cred",
				},
			},
		},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials: map[string]*credentialv1.Credential{
			"cred": newTokenCredential("cred", "secret-token"),
		},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	target := requireSingleBackendTarget(t, res.Runtime.Backends["be"])
	if target.AuthInjector == nil {
		t.Fatal("expected auth injector")
	}

	req := httptest.NewRequest(http.MethodGet, srv.URL, nil)
	if err := target.AuthInjector.Apply(req); err != nil {
		t.Fatalf("apply auth injector: %v", err)
	}
	if got := req.Header.Get("Authorization"); got != "Bearer secret-token" {
		t.Fatalf("expected bearer auth header, got %q", got)
	}
}

func TestCompile_BackendBasicCredential_AttachesAuthInjector(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": {
				Meta: &metav1.Meta{Name: "be"},
				Config: &backendv1.BackendConfig{
					Servers:               []string{srv.URL},
					InsecureSkipTlsVerify: true,
					AuthRef:               "cred",
				},
			},
		},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials: map[string]*credentialv1.Credential{
			"cred": newBasicCredential("cred", "alice", "secret"),
		},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	target := requireSingleBackendTarget(t, res.Runtime.Backends["be"])
	if target.AuthInjector == nil {
		t.Fatal("expected auth injector")
	}

	req := httptest.NewRequest(http.MethodGet, srv.URL, nil)
	if err := target.AuthInjector.Apply(req); err != nil {
		t.Fatalf("apply auth injector: %v", err)
	}
	if got := req.Header.Get("Authorization"); len(got) < 6 || got[:6] != "Basic " {
		t.Fatalf("expected basic auth header, got %q", got)
	}
}

func TestCompile_BackendClientCertificateCredential_AttachesTLSCert(t *testing.T) {
	certPEM, keyPEM := selfSignedPEM(t)
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": {
				Meta: &metav1.Meta{Name: "be"},
				Config: &backendv1.BackendConfig{
					Servers:               []string{srv.URL},
					InsecureSkipTlsVerify: true,
					AuthRef:               "cred",
				},
			},
		},
		Routes: map[string]*routev1.Route{},
		Certificates: map[string]*certificatev1.Certificate{
			"client-cert": newCertificate("client-cert", certPEM, keyPEM),
		},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials: map[string]*credentialv1.Credential{
			"cred": newClientCertCredential("cred", "client-cert"),
		},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	target := requireSingleBackendTarget(t, res.Runtime.Backends["be"])
	if len(target.TLSConfig.Certificates) != 1 {
		t.Fatalf("expected one tls certificate, got %d", len(target.TLSConfig.Certificates))
	}
	if target.AuthInjector != nil {
		t.Fatal("expected no auth injector for client certificate credential")
	}
}

func TestCompile_BackendMissingCredentialRef_Error(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": {
				Meta: &metav1.Meta{Name: "be"},
				Config: &backendv1.BackendConfig{
					Servers:               []string{srv.URL},
					InsecureSkipTlsVerify: true,
					AuthRef:               "missing",
				},
			},
		},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	_, err := c.Compile(st)
	if err == nil {
		t.Fatal("expected error for missing credential ref, got nil")
	}
}

func TestCompile_BackendClientCredentialMissingCertificate_Error(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": {
				Meta: &metav1.Meta{Name: "be"},
				Config: &backendv1.BackendConfig{
					Servers:               []string{srv.URL},
					InsecureSkipTlsVerify: true,
					AuthRef:               "cred",
				},
			},
		},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials: map[string]*credentialv1.Credential{
			"cred": newClientCertCredential("cred", "missing-cert"),
		},
	}

	_, err := c.Compile(st)
	if err == nil {
		t.Fatal("expected error for missing certificate ref, got nil")
	}
}

// ---------------------------------------------------------------------------
// Tests — Impersonation config compilation
// ---------------------------------------------------------------------------

func TestCompile_ImpersonationConfig_DefaultWhenNil(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": newBackend("be", srv.URL), // no impersonation_config set
		},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pool := res.Runtime.Backends["be"]
	if pool == nil {
		t.Fatal("expected backend pool")
	}
	imp := pool.Impersonation
	if imp == nil {
		t.Fatal("expected default impersonation config")
	}
	if imp.Name != "default" {
		t.Errorf("name = %q, want %q", imp.Name, "default")
	}
	if !imp.Enabled {
		t.Error("expected enabled=true by default")
	}
	if imp.UsernameClaim != "sub" {
		t.Errorf("username_claim = %q, want %q", imp.UsernameClaim, "sub")
	}
	if imp.GroupsClaim != "groups" {
		t.Errorf("groups_claim = %q, want %q", imp.GroupsClaim, "groups")
	}
	if len(imp.ExtraClaims) != 0 {
		t.Errorf("extra_claims = %v, want empty", imp.ExtraClaims)
	}
}

func TestCompile_ImpersonationConfig_ExplicitEnabled(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": {
				Meta: &metav1.Meta{Name: "be"},
				Config: &backendv1.BackendConfig{
					Servers:               []string{srv.URL},
					InsecureSkipTlsVerify: true,
					ImpersonationConfig: &backendv1.ImpersonationConfig{
						Name:          "custom",
						Enabled:       true,
						UsernameClaim: "email",
						GroupsClaim:   "roles",
						ExtraClaims:   []string{"scopes", "tenant"},
					},
				},
			},
		},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	imp := res.Runtime.Backends["be"].Impersonation
	if imp == nil {
		t.Fatal("expected impersonation config")
	}
	if imp.Name != "custom" {
		t.Errorf("name = %q, want %q", imp.Name, "custom")
	}
	if !imp.Enabled {
		t.Error("expected enabled=true")
	}
	if imp.UsernameClaim != "email" {
		t.Errorf("username_claim = %q, want %q", imp.UsernameClaim, "email")
	}
	if imp.GroupsClaim != "roles" {
		t.Errorf("groups_claim = %q, want %q", imp.GroupsClaim, "roles")
	}
	if len(imp.ExtraClaims) != 2 || imp.ExtraClaims[0] != "scopes" || imp.ExtraClaims[1] != "tenant" {
		t.Errorf("extra_claims = %v, want [scopes tenant]", imp.ExtraClaims)
	}
}

func TestCompile_ImpersonationConfig_ExplicitDisabled(t *testing.T) {
	srv := httptest.NewServer(nil)
	defer srv.Close()

	c := NewCompiler()
	st := &State{
		Backends: map[string]*backendv1.Backend{
			"be": {
				Meta: &metav1.Meta{Name: "be"},
				Config: &backendv1.BackendConfig{
					Servers:               []string{srv.URL},
					InsecureSkipTlsVerify: true,
					ImpersonationConfig: &backendv1.ImpersonationConfig{
						Name:    "off",
						Enabled: false,
					},
				},
			},
		},
		Routes:                 map[string]*routev1.Route{},
		Certificates:           map[string]*certificatev1.Certificate{},
		CertificateAuthorities: map[string]*cav1.CertificateAuthority{},
		Credentials:            map[string]*credentialv1.Credential{},
	}

	res, err := c.Compile(st)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	imp := res.Runtime.Backends["be"].Impersonation
	if imp == nil {
		t.Fatal("expected impersonation config")
	}
	if imp.Enabled {
		t.Error("expected enabled=false")
	}
	// Empty claim fields should be defaulted.
	if imp.UsernameClaim != "sub" {
		t.Errorf("username_claim = %q, want %q (defaulted)", imp.UsernameClaim, "sub")
	}
	if imp.GroupsClaim != "groups" {
		t.Errorf("groups_claim = %q, want %q (defaulted)", imp.GroupsClaim, "groups")
	}
}
