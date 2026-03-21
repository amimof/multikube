package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/url"
	"os"
	"time"

	"buf.build/go/protovalidate"
	types "github.com/amimof/multikube/api/config/v1"
)

// Convert validates the external protobuf Config and converts it into a
// fully-resolved RuntimeConfig. All string references are resolved, crypto
// materials loaded/parsed, URLs parsed, and durations converted.
func Convert(cfg *types.Config) (*RuntimeConfig, error) {
	if err := Validate(cfg); err != nil {
		return nil, err
	}
	return convert(cfg)
}

// Validate performs validation on the external protobuf Config. It first runs
// protovalidate to check structural constraints expressed via buf/validate
// annotations, then performs cross-reference and semantic checks that cannot
// be expressed in proto annotations (unique names, reference resolution,
// protocol-conditional rules, etc.).
func Validate(cfg *types.Config) error {
	if err := protovalidate.Validate(cfg); err != nil {
		return err
	}
	if err := validateUniqueNames(cfg); err != nil {
		return err
	}
	if err := validateCrossReferences(cfg); err != nil {
		return err
	}
	if err := validateListenerSemantics(cfg); err != nil {
		return err
	}
	if err := validateCertificateMixing(cfg); err != nil {
		return err
	}
	if err := validateMatchSemantics(cfg); err != nil {
		return err
	}
	return nil
}

// validateUniqueNames ensures no duplicate names within each collection.
func validateUniqueNames(cfg *types.Config) error {
	if err := checkUnique("certificates", cfg.Certificates, func(c *types.Certificate) string { return c.Name }); err != nil {
		return err
	}
	if err := checkUnique("certificate_authorities", cfg.CertificateAuthorities, func(c *types.CertificateAuthority) string { return c.Name }); err != nil {
		return err
	}
	if err := checkUnique("credentials", cfg.Credentials, func(c *types.Credential) string { return c.Name }); err != nil {
		return err
	}
	if err := checkUnique("backends", cfg.Backends, func(b *types.Backend) string { return b.Name }); err != nil {
		return err
	}
	if err := checkUnique("routes", cfg.Routes, func(r *types.Route) string { return r.Name }); err != nil {
		return err
	}
	return nil
}

func checkUnique[T any](collection string, items []*T, getName func(*T) string) error {
	seen := make(map[string]bool, len(items))
	for i, item := range items {
		name := getName(item)
		if seen[name] {
			return fmt.Errorf("%s[%d]: duplicate name %q", collection, i, name)
		}
		seen[name] = true
	}
	return nil
}

// validateCrossReferences ensures all _ref fields resolve to existing names.
func validateCrossReferences(cfg *types.Config) error {
	certNames := protoNameSet(cfg.Certificates, func(c *types.Certificate) string { return c.Name })
	caNames := protoNameSet(cfg.CertificateAuthorities, func(ca *types.CertificateAuthority) string { return ca.Name })
	credNames := protoNameSet(cfg.Credentials, func(c *types.Credential) string { return c.Name })
	backendNames := protoNameSet(cfg.Backends, func(b *types.Backend) string { return b.Name })

	// Credential → Certificate
	for i, cred := range cfg.Credentials {
		if cred.ClientCertificateRef != "" && !certNames[cred.ClientCertificateRef] {
			return fmt.Errorf("credentials[%d] %q: client_certificate_ref references unknown certificate %q", i, cred.Name, cred.ClientCertificateRef)
		}
	}

	// Backend → CA, Credential
	for i, b := range cfg.Backends {
		if b.CaRef != "" && !caNames[b.CaRef] {
			return fmt.Errorf("backends[%d] %q: ca_ref references unknown certificate_authority %q", i, b.Name, b.CaRef)
		}
		if b.AuthRef != "" && !credNames[b.AuthRef] {
			return fmt.Errorf("backends[%d] %q: auth_ref references unknown credential %q", i, b.Name, b.AuthRef)
		}
	}

	// Route → Backend
	for i, r := range cfg.Routes {
		if !backendNames[r.BackendRef] {
			return fmt.Errorf("routes[%d] %q: backend_ref references unknown backend %q", i, r.Name, r.BackendRef)
		}
	}

	return nil
}

// validateListenerSemantics checks protocol-conditional rules.
func validateListenerSemantics(cfg *types.Config) error {
	if cfg.Server == nil {
		return nil
	}

	// Unix listener requires a non-empty path.
	if cfg.Server.Unix != nil && cfg.Server.Unix.Path == "" {
		return fmt.Errorf("server.unix: path is required")
	}
	return nil
}

// validateCertificateMixing ensures certificates don't mix file paths with inline data.
func validateCertificateMixing(cfg *types.Config) error {
	for i, c := range cfg.Certificates {
		hasFile := c.Certificate != "" || c.Key != ""
		hasInline := c.CertificateData != "" || c.KeyData != ""
		if hasFile && hasInline {
			return fmt.Errorf("certificates[%d] %q: cannot mix file paths (certificate/key) and inline data (certificate_data/key_data)", i, c.Name)
		}
	}
	return nil
}

// validateMatchSemantics ensures match blocks have at least one condition.
func validateMatchSemantics(cfg *types.Config) error {
	for i, r := range cfg.Routes {
		if r.Match == nil {
			continue
		}
		m := r.Match
		hasCondition := m.Sni != "" || m.PathPrefix != "" || m.Header != nil || m.Jwt != nil
		if !hasCondition {
			return fmt.Errorf("routes[%d] %q: match block is present but has no conditions; remove it to create a default route", i, r.Name)
		}
	}
	return nil
}

// Converts proto types to runtime types
func convert(cfg *types.Config) (*RuntimeConfig, error) {
	// Build indexes for cross-reference resolution.
	certIdx := make(map[string]int, len(cfg.Certificates))
	for i, c := range cfg.Certificates {
		certIdx[c.Name] = i
	}
	caIdx := make(map[string]int, len(cfg.CertificateAuthorities))
	for i, ca := range cfg.CertificateAuthorities {
		caIdx[ca.Name] = i
	}
	credIdx := make(map[string]int, len(cfg.Credentials))
	for i, c := range cfg.Credentials {
		credIdx[c.Name] = i
	}
	backendIdx := make(map[string]int, len(cfg.Backends))
	for i, b := range cfg.Backends {
		backendIdx[b.Name] = i
	}

	// Convert certificates (load crypto).
	certs := make([]Certificate, len(cfg.Certificates))
	for i, c := range cfg.Certificates {
		tlsCert, err := loadCertificate(c)
		if err != nil {
			return nil, fmt.Errorf("certificates[%d] %q: %w", i, c.Name, err)
		}
		certs[i] = Certificate{Name: c.Name, TLS: tlsCert}
	}

	// Convert certificate authorities (load crypto).
	cas := make([]CertificateAuthority, len(cfg.CertificateAuthorities))
	for i, ca := range cfg.CertificateAuthorities {
		pool, err := loadCertificateAuthority(ca)
		if err != nil {
			return nil, fmt.Errorf("certificate_authorities[%d] %q: %w", i, ca.Name, err)
		}
		cas[i] = CertificateAuthority{Name: ca.Name, Pool: pool}
	}

	// Convert credentials (resolve client_certificate_ref).
	creds := make([]Credential, len(cfg.Credentials))
	for i, c := range cfg.Credentials {
		cred := Credential{Name: c.Name}
		if c.ClientCertificateRef != "" {
			idx := certIdx[c.ClientCertificateRef]
			cred.ClientCertificate = &certs[idx].TLS
		}
		if c.Token != "" {
			cred.Token = c.Token
		}
		if c.Basic != nil {
			cred.Username = c.Basic.Username
			cred.Password = c.Basic.Password
		}
		creds[i] = cred
	}

	// Convert backends (resolve ca_ref, auth_ref, parse URL).
	backends := make([]Backend, len(cfg.Backends))
	for i, b := range cfg.Backends {
		u, err := url.Parse(b.Server)
		if err != nil {
			return nil, fmt.Errorf("backends[%d] %q: invalid server URL: %w", i, b.Name, err)
		}
		be := Backend{
			Name:                  b.Name,
			Server:                u,
			InsecureSkipTLSVerify: b.InsecureSkipTlsVerify,
		}
		if b.CacheTtl != nil {
			be.CacheTTL = b.CacheTtl.AsDuration()
		}
		if b.CaRef != "" {
			idx := caIdx[b.CaRef]
			be.CA = &cas[idx]
		}
		if b.AuthRef != "" {
			idx := credIdx[b.AuthRef]
			be.Auth = &creds[idx]
		}
		backends[i] = be
	}

	// Convert routes (resolve backend_ref).
	routes := make([]Route, len(cfg.Routes))
	for i, r := range cfg.Routes {
		idx := backendIdx[r.BackendRef]
		rt := Route{
			Name:    r.Name,
			Backend: &backends[idx],
		}
		if r.Match != nil {
			rt.Match = convertMatch(r.Match)
		}
		routes[i] = rt
	}

	// Convert server config.
	sc, err := convertServer(cfg.Server)
	if err != nil {
		return nil, err
	}

	rc := &RuntimeConfig{
		Certificates:           certs,
		CertificateAuthorities: cas,
		Credentials:            creds,
		Backends:               backends,
		Routes:                 routes,
		Server:                 sc,
	}

	// Convert auth.
	if cfg.Auth != nil {
		rc.Auth = convertAuth(cfg.Auth)
	}

	// Convert cache.
	if cfg.Cache != nil && cfg.Cache.Ttl != nil {
		rc.Cache = &CacheConfig{
			TTL: cfg.Cache.Ttl.AsDuration(),
		}
	}

	return rc, nil
}

func convertMatch(m *types.Match) *Match {
	rm := &Match{
		SNI:        m.Sni,
		PathPrefix: m.PathPrefix,
	}
	if m.Header != nil {
		rm.Header = &HeaderMatch{
			Name:  m.Header.Name,
			Value: m.Header.Value,
		}
	}
	if m.Jwt != nil {
		rm.JWT = &JWTMatch{
			Claim: m.Jwt.Claim,
			Value: m.Jwt.Value,
		}
	}
	return rm
}

func convertServer(s *types.Server) (ServerConfig, error) {
	sc := ServerConfig{
		MaxHeaderSize: s.GetMaxHeaderSize(),
	}
	if s.GetReadTimeout() != nil {
		sc.ReadTimeout = s.ReadTimeout.AsDuration()
	}
	if s.GetWriteTimeout() != nil {
		sc.WriteTimeout = s.WriteTimeout.AsDuration()
	}
	if s.GetKeepAlive() != nil {
		sc.KeepAlive = s.KeepAlive.AsDuration()
	}
	if s.GetShutdownGracePeriod() != nil {
		sc.ShutdownGracePeriod = s.ShutdownGracePeriod.AsDuration()
	}

	// Convert HTTPS listener.
	if s.GetHttps() != nil {
		hl, err := convertHTTPSListener(s.Https)
		if err != nil {
			return ServerConfig{}, fmt.Errorf("server.https: %w", err)
		}
		sc.HTTPS = hl
	}

	// Convert Unix listener.
	if s.GetUnix() != nil {
		sc.Unix = &UnixListenerConfig{
			Path: s.Unix.Path,
		}
	}

	// Convert Metrics listener.
	if s.GetMetrics() != nil {
		sc.Metrics = &MetricsListenerConfig{
			Address: s.Metrics.Address,
		}
	}

	return sc, nil
}

// convertHTTPSListener converts the proto HTTPSListener to a runtime
// HTTPSListenerConfig. It loads TLS cert/key from file or inline data
// (or auto-generates if neither is provided), loads CA, and parses
// client_auth and min_version.
func convertHTTPSListener(h *types.HTTPSListener) (*HTTPSListenerConfig, error) {
	lt := &ListenerTLS{}

	// Load server certificate/key from oneof fields, or auto-generate.
	cert, key := h.GetCert(), h.GetKey()
	certData, keyData := h.GetCertData(), h.GetKeyData()

	switch {
	case cert != "" && key != "":
		tlsCert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, fmt.Errorf("loading cert/key files: %w", err)
		}
		lt.Certificates = []tls.Certificate{tlsCert}
	case certData != "" && keyData != "":
		tlsCert, err := tls.X509KeyPair([]byte(certData), []byte(keyData))
		if err != nil {
			return nil, fmt.Errorf("parsing cert_data/key_data: %w", err)
		}
		lt.Certificates = []tls.Certificate{tlsCert}
	default:
		// No cert/key provided; auto-generate a self-signed certificate.
		tlsCert, err := generateCertificates()
		if err != nil {
			return nil, fmt.Errorf("generating self-signed certificate: %w", err)
		}
		lt.Certificates = []tls.Certificate{tlsCert}
	}

	// Load CA from oneof fields (for client certificate verification).
	ca, caData := h.GetCa(), h.GetCaData()
	switch {
	case ca != "":
		data, err := os.ReadFile(ca)
		if err != nil {
			return nil, fmt.Errorf("reading ca file: %w", err)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("failed to parse CA PEM from file %q", ca)
		}
		lt.ClientCA = pool
	case caData != "":
		// Try base64 decode first (kubeconfig-style), fall back to raw PEM.
		var pemData []byte
		decoded, err := base64.StdEncoding.DecodeString(caData)
		if err == nil {
			pemData = decoded
		} else {
			pemData = []byte(caData)
		}
		pool := x509.NewCertPool()
		if !pool.AppendCertsFromPEM(pemData) {
			return nil, fmt.Errorf("failed to parse CA PEM from ca_data")
		}
		lt.ClientCA = pool
	}

	// Parse client_auth.
	switch h.ClientAuth {
	case "", "none":
		lt.ClientAuth = tls.NoClientCert
	case "request":
		lt.ClientAuth = tls.RequestClientCert
	case "require":
		lt.ClientAuth = tls.RequireAndVerifyClientCert
	}

	// If client CA is set but client_auth was not explicitly set, default to require.
	if lt.ClientCA != nil && h.ClientAuth == "" {
		lt.ClientAuth = tls.RequireAndVerifyClientCert
	}

	// Parse min_version (default to TLS 1.2).
	switch h.MinVersion {
	case "1.0":
		lt.MinVersion = tls.VersionTLS10
	case "1.1":
		lt.MinVersion = tls.VersionTLS11
	case "1.3":
		lt.MinVersion = tls.VersionTLS13
	default:
		lt.MinVersion = tls.VersionTLS12
	}

	return &HTTPSListenerConfig{
		Address: h.Address,
		TLS:     lt,
	}, nil
}

func convertAuth(a *types.Authentication) *AuthConfig {
	if a.JwtAuth == nil {
		return nil
	}
	ac := &AuthConfig{
		JWT: &JWTAuthConfig{},
	}
	if a.JwtAuth.Rs256 != nil {
		ac.JWT.RS256 = &RS256AuthConfig{
			PublicKey: a.JwtAuth.Rs256.PublicKey,
		}
	}
	if a.JwtAuth.Oidc != nil {
		oc := &OIDCAuthConfig{
			IssuerURL:             a.JwtAuth.Oidc.IssuerUrl,
			UsernameClaim:         a.JwtAuth.Oidc.UsernameClaim,
			CAFile:                a.JwtAuth.Oidc.CaFile,
			InsecureSkipTLSVerify: a.JwtAuth.Oidc.InsecureSkipTlsVerify,
		}
		if a.JwtAuth.Oidc.PollInterval != nil {
			oc.PollInterval = a.JwtAuth.Oidc.PollInterval.AsDuration()
		}
		ac.JWT.OIDC = oc
	}
	return ac
}

func generateCertificates() (tls.Certificate, error) {
	cert := tls.Certificate{}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return cert, err
	}

	// Valid for 1 year
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour) // Valid for 1 year

	parent := &x509.Certificate{
		KeyUsage: x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageServerAuth,
			x509.ExtKeyUsageClientAuth,
		},
		IsCA:                  false,
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost", "voiyd-node"},
		IPAddresses: []net.IP{
			net.IPv4(127, 0, 0, 1),
		},
		Subject: pkix.Name{
			CommonName:         "voiyd-node",
			Country:            []string{"SE"},
			Province:           []string{"Halland"},
			Locality:           []string{"Varberg"},
			Organization:       []string{"voiyd-node"},
			OrganizationalUnit: []string{"voiyd"},
		},
		SerialNumber: serial,
		NotAfter:     notAfter,
		NotBefore:    notBefore,
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return cert, err
	}

	certData, err := x509.CreateCertificate(rand.Reader, parent, parent, &key.PublicKey, key)
	if err != nil {
		return cert, err
	}

	cert = tls.Certificate{
		Certificate: [][]byte{certData}, // Raw DER bytes from x509.CreateCertificate
		PrivateKey:  key,                // *ecdsa.PrivateKey directly
	}

	log.Println("generated x509 key pair")
	return cert, nil
}

// loadCertificate loads a TLS certificate from file paths or inline PEM data.
func loadCertificate(c *types.Certificate) (tls.Certificate, error) {
	if c.Certificate != "" && c.Key != "" {
		return tls.LoadX509KeyPair(c.Certificate, c.Key)
	}
	if c.CertificateData != "" && c.KeyData != "" {
		return tls.X509KeyPair([]byte(c.CertificateData), []byte(c.KeyData))
	}
	log.Println("generating certs")
	return generateCertificates()
}

// loadCertificateAuthority loads a CA certificate pool from a file path or
// inline PEM data.
func loadCertificateAuthority(ca *types.CertificateAuthority) (*x509.CertPool, error) {
	pool := x509.NewCertPool()
	var pemData []byte

	if ca.Certificate != "" {
		data, err := os.ReadFile(ca.Certificate)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
		pemData = data
	} else if ca.CertificateData != "" {
		// Try base64 decode first (kubeconfig-style), fall back to raw PEM.
		decoded, err := base64.StdEncoding.DecodeString(ca.CertificateData)
		if err == nil {
			pemData = decoded
		} else {
			pemData = []byte(ca.CertificateData)
		}
	} else {
		return nil, fmt.Errorf("no certificate data")
	}

	if !pool.AppendCertsFromPEM(pemData) {
		return nil, fmt.Errorf("failed to parse PEM data")
	}
	return pool, nil
}

// protoNameSet builds a name→bool map from a slice of proto messages.
func protoNameSet[T any](items []*T, getName func(*T) string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, item := range items {
		s[getName(item)] = true
	}
	return s
}
