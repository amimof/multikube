package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"net/url"
	"os"

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

	// Listener TLS → Certificate, CA
	if cfg.Server != nil {
		for i, l := range cfg.Server.Listeners {
			if l.Tls != nil {
				for _, ref := range l.Tls.CertificateRefs {
					if !certNames[ref] {
						return fmt.Errorf("server.listeners[%d]: tls.certificate_refs references unknown certificate %q", i, ref)
					}
				}
				if l.Tls.ClientCaRef != "" && !caNames[l.Tls.ClientCaRef] {
					return fmt.Errorf("server.listeners[%d]: tls.client_ca_ref references unknown certificate_authority %q", i, l.Tls.ClientCaRef)
				}
			}
		}
	}
	return nil
}

// validateListenerSemantics checks protocol-conditional rules.
func validateListenerSemantics(cfg *types.Config) error {
	if cfg.Server == nil {
		return nil
	}

	protocolSeen := make(map[string]bool)
	for i, l := range cfg.Server.Listeners {
		if protocolSeen[l.Protocol] {
			return fmt.Errorf("server.listeners[%d]: duplicate protocol %q; at most one listener per protocol is supported", i, l.Protocol)
		}
		protocolSeen[l.Protocol] = true

		switch l.Protocol {
		case "http":
			if l.Tls != nil {
				return fmt.Errorf("server.listeners[%d]: TLS must not be set for protocol http", i)
			}
		case "https":
			if l.Tls == nil {
				return fmt.Errorf("server.listeners[%d]: TLS is required for protocol https", i)
			}
		case "unix":
			if l.SocketPath == "" {
				return fmt.Errorf("server.listeners[%d]: socket_path is required for protocol unix", i)
			}
		}
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
	sc, err := convertServer(cfg.Server, certs, cas)
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

	// Convert metrics.
	if cfg.Metrics != nil {
		rc.Metrics = &MetricsConfig{
			Address: cfg.Metrics.Address,
			Port:    int(cfg.Metrics.Port),
		}
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

func convertServer(s *types.Server, certs []Certificate, cas []CertificateAuthority) (ServerConfig, error) {
	sc := ServerConfig{
		MaxHeaderSize: s.MaxHeaderSize,
	}
	if s.ReadTimeout != nil {
		sc.ReadTimeout = s.ReadTimeout.AsDuration()
	}
	if s.WriteTimeout != nil {
		sc.WriteTimeout = s.WriteTimeout.AsDuration()
	}
	if s.KeepAlive != nil {
		sc.KeepAlive = s.KeepAlive.AsDuration()
	}
	if s.ShutdownGracePeriod != nil {
		sc.ShutdownGracePeriod = s.ShutdownGracePeriod.AsDuration()
	}

	// Build cert/CA indexes for listener TLS resolution.
	certByName := make(map[string]int, len(certs))
	for i, c := range certs {
		certByName[c.Name] = i
	}
	caByName := make(map[string]int, len(cas))
	for i, ca := range cas {
		caByName[ca.Name] = i
	}

	for i, l := range s.Listeners {
		rl := Listener{
			Protocol:   l.Protocol,
			Address:    l.Address,
			Port:       int(l.Port),
			SocketPath: l.SocketPath,
		}

		if l.Tls != nil {
			lt, err := convertListenerTLS(l.Tls, certs, certByName, cas, caByName)
			if err != nil {
				return ServerConfig{}, fmt.Errorf("server.listeners[%d]: %w", i, err)
			}
			rl.TLS = lt
		}

		sc.Listeners = append(sc.Listeners, rl)
	}

	return sc, nil
}

func convertListenerTLS(
	t *types.ListenerTLS,
	certs []Certificate,
	certByName map[string]int,
	cas []CertificateAuthority,
	caByName map[string]int,
) (*ListenerTLS, error) {
	lt := &ListenerTLS{}

	// Resolve certificate_refs.
	for _, ref := range t.CertificateRefs {
		idx := certByName[ref]
		lt.Certificates = append(lt.Certificates, certs[idx].TLS)
	}

	// Parse client_auth.
	switch t.ClientAuth {
	case "", "none":
		lt.ClientAuth = tls.NoClientCert
	case "request":
		lt.ClientAuth = tls.RequestClientCert
	case "require":
		lt.ClientAuth = tls.RequireAndVerifyClientCert
	}

	// Resolve client_ca_ref.
	if t.ClientCaRef != "" {
		idx := caByName[t.ClientCaRef]
		lt.ClientCA = cas[idx].Pool
		// If client CA is set but client_auth was not explicitly set, default to require.
		if t.ClientAuth == "" {
			lt.ClientAuth = tls.RequireAndVerifyClientCert
		}
	}

	// Parse min_version (default to TLS 1.2).
	switch t.MinVersion {
	case "1.0":
		lt.MinVersion = tls.VersionTLS10
	case "1.1":
		lt.MinVersion = tls.VersionTLS11
	case "1.3":
		lt.MinVersion = tls.VersionTLS13
	default:
		lt.MinVersion = tls.VersionTLS12
	}

	return lt, nil
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

// loadCertificate loads a TLS certificate from file paths or inline PEM data.
func loadCertificate(c *types.Certificate) (tls.Certificate, error) {
	if c.Certificate != "" && c.Key != "" {
		return tls.LoadX509KeyPair(c.Certificate, c.Key)
	}
	if c.CertificateData != "" && c.KeyData != "" {
		return tls.X509KeyPair([]byte(c.CertificateData), []byte(c.KeyData))
	}
	return tls.Certificate{}, fmt.Errorf("missing certificate/key data")
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
