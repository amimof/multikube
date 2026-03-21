package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helper: generate self-signed cert + key PEM bytes, optionally write to files
// ---------------------------------------------------------------------------

func generateSelfSignedCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})
	return certPEM, keyPEM
}

func generateCACert(t *testing.T) []byte {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
}

func writeCertFiles(t *testing.T, dir, name string) (certPath, keyPath string) {
	t.Helper()
	certPEM, keyPEM := generateSelfSignedCert(t)
	certPath = filepath.Join(dir, name+".crt")
	keyPath = filepath.Join(dir, name+".key")
	require.NoError(t, os.WriteFile(certPath, certPEM, 0644))
	require.NoError(t, os.WriteFile(keyPath, keyPEM, 0600))
	return certPath, keyPath
}

func writeCAFile(t *testing.T, dir, name string) string {
	t.Helper()
	caPEM := generateCACert(t)
	caPath := filepath.Join(dir, name+".crt")
	require.NoError(t, os.WriteFile(caPath, caPEM, 0644))
	return caPath
}

// ---------------------------------------------------------------------------
// YAML fixtures for Validate() tests (use file paths that won't be loaded)
// ---------------------------------------------------------------------------

var validYAML = `
certificates:
  - name: frontend-cert
    certificate: /etc/multikube/tls/server.crt
    key: /etc/multikube/tls/server.key

  - name: frontend-cert-wildcard
    certificate: /etc/multikube/tls/wildcard.crt
    key: /etc/multikube/tls/wildcard.key

  - name: staging-client-cert
    certificate: /etc/multikube/tls/staging-client.crt
    key: /etc/multikube/tls/staging-client.key

  - name: prod-admin-cert
    certificate_data: |
      -----BEGIN CERTIFICATE-----
      dGVzdC1jZXJ0LWRhdGE=
      -----END CERTIFICATE-----
    key_data: |
      -----BEGIN RSA PRIVATE KEY-----
      dGVzdC1rZXktZGF0YQ==
      -----END RSA PRIVATE KEY-----

certificate_authorities:
  - name: prod-ca
    certificate: /etc/multikube/tls/prod-ca.crt

  - name: staging-ca
    certificate: /etc/multikube/tls/staging-ca.crt

  - name: dev-ca
    certificate_data: |
      -----BEGIN CERTIFICATE-----
      dGVzdC1jYS1kYXRh
      -----END CERTIFICATE-----

credentials:
  - name: prod-admin
    client_certificate_ref: prod-admin-cert

  - name: staging-token
    token: "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9"

  - name: dev-basic
    basic:
      username: admin
      password: secret

backends:
  - name: prod-cluster
    server: https://k8s-prod.internal:6443
    ca_ref: prod-ca
    auth_ref: prod-admin
    insecure_skip_tls_verify: false
    cache_ttl: 2s

  - name: staging-cluster
    server: https://k8s-staging.internal:6443
    ca_ref: staging-ca
    auth_ref: staging-token

  - name: dev-cluster
    server: https://k8s-dev.internal:6443
    ca_ref: dev-ca
    auth_ref: dev-basic
    insecure_skip_tls_verify: true
    cache_ttl: 0s

routes:
  - name: prod-by-sni
    match:
      sni: prod.api.mycluster.com
    backend_ref: prod-cluster

  - name: staging-by-sni
    match:
      sni: staging.api.mycluster.com
    backend_ref: staging-cluster

  - name: prod-by-header
    match:
      header:
        name: Multikube-Context
        value: prod-cluster
    backend_ref: prod-cluster

  - name: staging-by-path
    match:
      path_prefix: /staging-cluster
    backend_ref: staging-cluster

  - name: prod-by-jwt
    match:
      jwt:
        claim: kubernetes_cluster
        value: prod-cluster
    backend_ref: prod-cluster

  - name: dev-compound
    match:
      sni: dev.api.mycluster.com
      header:
        name: X-Environment
        value: development
    backend_ref: dev-cluster

  - name: default
    backend_ref: dev-cluster

server:
  https:
    address: "0.0.0.0:8443"
    cert: /etc/multikube/tls/server.crt
    key: /etc/multikube/tls/server.key
    client_auth: request
    ca: /etc/multikube/tls/prod-ca.crt
    min_version: "1.2"
  unix:
    path: /var/run/multikube.sock
  metrics:
    address: "0.0.0.0:8888"
  max_header_size: 1000000
  read_timeout: 30s
  write_timeout: 30s
  keep_alive: 180s
  shutdown_grace_period: 10s

auth:
  jwt_auth:
    rs256:
      public_key: /etc/multikube/keys/public.pem

cache:
  ttl: 1s
`

var minimalYAML = `
certificates:
  - name: server-cert
    certificate: /etc/multikube/tls/server.crt
    key: /etc/multikube/tls/server.key

certificate_authorities:
  - name: backend-ca
    certificate: /etc/multikube/tls/ca.crt

credentials:
  - name: backend-token
    token: "some-token"

backends:
  - name: default-cluster
    server: https://k8s.internal:6443
    ca_ref: backend-ca
    auth_ref: backend-token

routes:
  - name: default
    backend_ref: default-cluster

server:
  https:
    address: "0.0.0.0:8443"
`

// ---------------------------------------------------------------------------
// Tests: Load (YAML → proto)
// ---------------------------------------------------------------------------

func TestLoad_ValidConfig(t *testing.T) {
	cfg, err := Load([]byte(validYAML))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Certificates
	assert.Len(t, cfg.Certificates, 4)
	assert.Equal(t, "frontend-cert", cfg.Certificates[0].Name)
	assert.Equal(t, "/etc/multikube/tls/server.crt", cfg.Certificates[0].Certificate)
	assert.Equal(t, "/etc/multikube/tls/server.key", cfg.Certificates[0].Key)
	// Inline cert
	assert.Equal(t, "prod-admin-cert", cfg.Certificates[3].Name)
	assert.NotEmpty(t, cfg.Certificates[3].CertificateData)
	assert.NotEmpty(t, cfg.Certificates[3].KeyData)
	assert.Empty(t, cfg.Certificates[3].Certificate)
	assert.Empty(t, cfg.Certificates[3].Key)

	// Certificate Authorities
	assert.Len(t, cfg.CertificateAuthorities, 3)
	assert.Equal(t, "prod-ca", cfg.CertificateAuthorities[0].Name)
	assert.Equal(t, "/etc/multikube/tls/prod-ca.crt", cfg.CertificateAuthorities[0].Certificate)
	assert.Equal(t, "dev-ca", cfg.CertificateAuthorities[2].Name)
	assert.NotEmpty(t, cfg.CertificateAuthorities[2].CertificateData)

	// Credentials
	assert.Len(t, cfg.Credentials, 3)
	assert.Equal(t, "prod-admin", cfg.Credentials[0].Name)
	assert.Equal(t, "prod-admin-cert", cfg.Credentials[0].ClientCertificateRef)
	assert.Equal(t, "staging-token", cfg.Credentials[1].Name)
	assert.Equal(t, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9", cfg.Credentials[1].Token)
	assert.Equal(t, "dev-basic", cfg.Credentials[2].Name)
	require.NotNil(t, cfg.Credentials[2].Basic)
	assert.Equal(t, "admin", cfg.Credentials[2].Basic.Username)
	assert.Equal(t, "secret", cfg.Credentials[2].Basic.Password)

	// Backends
	assert.Len(t, cfg.Backends, 3)
	assert.Equal(t, "prod-cluster", cfg.Backends[0].Name)
	assert.Equal(t, "https://k8s-prod.internal:6443", cfg.Backends[0].Server)
	assert.Equal(t, "prod-ca", cfg.Backends[0].CaRef)
	assert.Equal(t, "prod-admin", cfg.Backends[0].AuthRef)
	assert.False(t, cfg.Backends[0].InsecureSkipTlsVerify)
	require.NotNil(t, cfg.Backends[0].CacheTtl)
	assert.Equal(t, 2*time.Second, cfg.Backends[0].CacheTtl.AsDuration())
	assert.True(t, cfg.Backends[2].InsecureSkipTlsVerify)

	// Routes
	assert.Len(t, cfg.Routes, 7)
	assert.Equal(t, "prod-by-sni", cfg.Routes[0].Name)
	require.NotNil(t, cfg.Routes[0].Match)
	assert.Equal(t, "prod.api.mycluster.com", cfg.Routes[0].Match.Sni)
	assert.Equal(t, "prod-cluster", cfg.Routes[0].BackendRef)
	// Default route (no match)
	assert.Equal(t, "default", cfg.Routes[6].Name)
	assert.Nil(t, cfg.Routes[6].Match)
	assert.Equal(t, "dev-cluster", cfg.Routes[6].BackendRef)

	// Server
	require.NotNil(t, cfg.Server)
	require.NotNil(t, cfg.Server.Https)
	assert.Equal(t, "0.0.0.0:8443", cfg.Server.Https.GetAddress())
	assert.Equal(t, "/etc/multikube/tls/server.crt", cfg.Server.Https.GetCert())
	assert.Equal(t, "/etc/multikube/tls/server.key", cfg.Server.Https.GetKey())
	assert.Equal(t, "request", cfg.Server.Https.GetClientAuth())
	assert.Equal(t, "/etc/multikube/tls/prod-ca.crt", cfg.Server.Https.GetCa())
	assert.Equal(t, "1.2", cfg.Server.Https.GetMinVersion())
	require.NotNil(t, cfg.Server.Unix)
	assert.Equal(t, "/var/run/multikube.sock", cfg.Server.Unix.Path)
	require.NotNil(t, cfg.Server.Metrics)
	assert.Equal(t, "0.0.0.0:8888", cfg.Server.Metrics.Address)
	assert.Equal(t, uint64(1000000), cfg.Server.MaxHeaderSize)
	assert.Equal(t, 30*time.Second, cfg.Server.ReadTimeout.AsDuration())
	assert.Equal(t, 30*time.Second, cfg.Server.WriteTimeout.AsDuration())
	assert.Equal(t, 3*time.Minute, cfg.Server.KeepAlive.AsDuration())
	assert.Equal(t, 10*time.Second, cfg.Server.ShutdownGracePeriod.AsDuration())

	// Auth
	require.NotNil(t, cfg.Auth)
	require.NotNil(t, cfg.Auth.JwtAuth)
	require.NotNil(t, cfg.Auth.JwtAuth.Rs256)
	assert.Equal(t, "/etc/multikube/keys/public.pem", cfg.Auth.JwtAuth.Rs256.PublicKey)
	assert.Nil(t, cfg.Auth.JwtAuth.Oidc)

	// Cache
	require.NotNil(t, cfg.Cache)
	assert.Equal(t, 1*time.Second, cfg.Cache.Ttl.AsDuration())
}

func TestLoad_Minimal(t *testing.T) {
	cfg, err := Load([]byte(minimalYAML))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Len(t, cfg.Certificates, 1)
	assert.Len(t, cfg.CertificateAuthorities, 1)
	assert.Len(t, cfg.Credentials, 1)
	assert.Len(t, cfg.Backends, 1)
	assert.Len(t, cfg.Routes, 1)
	require.NotNil(t, cfg.Server)
	require.NotNil(t, cfg.Server.Https)
	assert.Equal(t, "0.0.0.0:8443", cfg.Server.Https.GetAddress())
	assert.Nil(t, cfg.Routes[0].Match)
	assert.Nil(t, cfg.Auth)
}

func TestLoadFromFile_Valid(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(tmpFile, []byte(minimalYAML), 0644))

	cfg, err := LoadFromFile(tmpFile)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Len(t, cfg.Certificates, 1)
}

func TestLoadFromFile_FileNotFound(t *testing.T) {
	_, err := LoadFromFile(filepath.Join(t.TempDir(), "nonexistent.yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "reading config file")
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "bad.yaml")
	require.NoError(t, os.WriteFile(tmpFile, []byte("{{{{not yaml"), 0644))

	_, err := LoadFromFile(tmpFile)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parsing YAML")
}

// ---------------------------------------------------------------------------
// Tests: Validate — protovalidate (structural constraints via annotations)
// ---------------------------------------------------------------------------

func TestValidate_ValidConfig(t *testing.T) {
	cfg, err := Load([]byte(validYAML))
	require.NoError(t, err)
	assert.NoError(t, Validate(cfg))
}

func TestValidate_ProtoAnnotations(t *testing.T) {
	// Each subtest verifies that a proto annotation catches a specific
	// structural error. These are all handled by buf/validate, not Go code.
	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "CertificateNameRequired",
			yaml: `
certificates:
  - certificate: /a.crt
    key: /a.key
server:
  https:
    address: ":8443"
`,
			wantErr: "name: value length must be at least 1 characters",
		},
		{
			name: "CertificateNoCertOrData",
			yaml: `
certificates:
  - name: empty-cert
server:
  https:
    address: ":8443"
`,
			wantErr: "one of certificate, certificate_data must be set",
		},
		{
			name: "CertificateNoKeyOrKeyData",
			yaml: `
certificates:
  - name: no-key
    certificate: /a.crt
server:
  https:
    address: ":8443"
`,
			wantErr: "one of key, key_data must be set",
		},
		{
			name: "CANameRequired",
			yaml: `
certificate_authorities:
  - certificate: /ca.crt
server:
  https:
    address: ":8443"
`,
			wantErr: "name: value length must be at least 1 characters",
		},
		{
			name: "CANoData",
			yaml: `
certificate_authorities:
  - name: ca
server:
  https:
    address: ":8443"
`,
			wantErr: "one of certificate, certificate_data must be set",
		},
		{
			name: "CABothFileAndInline",
			yaml: `
certificate_authorities:
  - name: ca
    certificate: /ca.crt
    certificate_data: "data"
server:
  https:
    address: ":8443"
`,
			wantErr: "only one of certificate, certificate_data can be set",
		},
		{
			name: "CredentialNameRequired",
			yaml: `
credentials:
  - token: tok
server:
  https:
    address: ":8443"
`,
			wantErr: "name: value length must be at least 1 characters",
		},
		{
			name: "CredentialNoMethod",
			yaml: `
credentials:
  - name: empty-cred
server:
  https:
    address: ":8443"
`,
			wantErr: "one of client_certificate_ref, token, basic must be set",
		},
		{
			name: "CredentialMultipleMethods",
			yaml: `
credentials:
  - name: mixed
    token: tok
    basic:
      username: admin
      password: secret
server:
  https:
    address: ":8443"
`,
			wantErr: "only one of client_certificate_ref, token, basic can be set",
		},
		{
			name: "CredentialBasicMissingPassword",
			yaml: `
credentials:
  - name: partial
    basic:
      username: admin
server:
  https:
    address: ":8443"
`,
			wantErr: "password: value length must be at least 1 characters",
		},
		{
			name: "BackendNameRequired",
			yaml: `
backends:
  - server: https://k8s:6443
server:
  https:
    address: ":8443"
`,
			wantErr: "name: value length must be at least 1 characters",
		},
		{
			name: "BackendServerRequired",
			yaml: `
backends:
  - name: cluster
server:
  https:
    address: ":8443"
`,
			wantErr: "server: value length must be at least 1 characters",
		},
		{
			name: "RouteNameRequired",
			yaml: `
routes:
  - backend_ref: cluster
server:
  https:
    address: ":8443"
`,
			wantErr: "name: value length must be at least 1 characters",
		},
		{
			name: "RouteBackendRefRequired",
			yaml: `
routes:
  - name: route
server:
  https:
    address: ":8443"
`,
			wantErr: "backend_ref: value length must be at least 1 characters",
		},
		{
			name: "HeaderMatchNameRequired",
			yaml: `
routes:
  - name: bad
    backend_ref: cluster
    match:
      header:
        value: something
server:
  https:
    address: ":8443"
`,
			wantErr: "header.name: value length must be at least 1 characters",
		},
		{
			name: "HeaderMatchValueRequired",
			yaml: `
routes:
  - name: bad
    backend_ref: cluster
    match:
      header:
        name: X-Cluster
server:
  https:
    address: ":8443"
`,
			wantErr: "header.value: value length must be at least 1 characters",
		},
		{
			name: "JWTMatchClaimRequired",
			yaml: `
routes:
  - name: bad
    backend_ref: cluster
    match:
      jwt:
        value: something
server:
  https:
    address: ":8443"
`,
			wantErr: "jwt.claim: value length must be at least 1 characters",
		},
		{
			name: "JWTMatchValueRequired",
			yaml: `
routes:
  - name: bad
    backend_ref: cluster
    match:
      jwt:
        claim: cluster
server:
  https:
    address: ":8443"
`,
			wantErr: "jwt.value: value length must be at least 1 characters",
		},
		{
			name: "InvalidClientAuth",
			yaml: `
server:
  https:
    address: ":8443"
    client_auth: maybe
`,
			wantErr: "client_auth: value must be in list",
		},
		{
			name: "InvalidMinVersion",
			yaml: `
server:
  https:
    address: ":8443"
    min_version: "1.5"
`,
			wantErr: "min_version: value must be in list",
		},
		{
			name: "AuthBothRS256andOIDC",
			yaml: `
server:
  https:
    address: ":8443"
auth:
  jwt_auth:
    rs256:
      public_key: /key.pem
    oidc:
      issuer_url: https://accounts.google.com
`,
			wantErr: "only one of rs256, oidc can be set",
		},
		{
			name: "RS256MissingPublicKey",
			yaml: `
server:
  https:
    address: ":8443"
auth:
  jwt_auth:
    rs256: {}
`,
			wantErr: "public_key: value length must be at least 1 characters",
		},
		{
			name: "OIDCMissingIssuerURL",
			yaml: `
server:
  https:
    address: ":8443"
auth:
  jwt_auth:
    oidc:
      username_claim: email
`,
			wantErr: "issuer_url: value length must be at least 1 characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := Load([]byte(tt.yaml))
			require.NoError(t, err)
			err = Validate(ext)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

// ---------------------------------------------------------------------------
// Tests: Validate — cross-reference and semantic checks (Go code)
// ---------------------------------------------------------------------------

// baseYAML returns a minimal valid config YAML fragment. Tests can override
// specific sections to introduce targeted errors.
func baseYAML(overrides map[string]string) string {
	sections := map[string]string{
		"certificates": `
certificates:
  - name: cert
    certificate: /a.crt
    key: /a.key`,
		"certificate_authorities": `
certificate_authorities:
  - name: ca
    certificate: /ca.crt`,
		"credentials": `
credentials:
  - name: cred
    token: tok`,
		"backends": `
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: cred`,
		"routes": `
routes:
  - name: default
    backend_ref: cluster`,
		"server": `
server:
  https:
    address: ":8443"`,
	}
	for k, v := range overrides {
		sections[k] = v
	}
	return sections["certificates"] + "\n" +
		sections["certificate_authorities"] + "\n" +
		sections["credentials"] + "\n" +
		sections["backends"] + "\n" +
		sections["routes"] + "\n" +
		sections["server"]
}

func TestValidate_DuplicateCertificateNames(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"certificates": `
certificates:
  - name: my-cert
    certificate: /a.crt
    key: /a.key
  - name: my-cert
    certificate: /b.crt
    key: /b.key`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `duplicate name "my-cert"`)
}

func TestValidate_DuplicateCANames(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"certificate_authorities": `
certificate_authorities:
  - name: my-ca
    certificate: /ca1.crt
  - name: my-ca
    certificate: /ca2.crt`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `duplicate name "my-ca"`)
}

func TestValidate_DuplicateCredentialNames(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"credentials": `
credentials:
  - name: my-cred
    token: tok1
  - name: my-cred
    token: tok2`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `duplicate name "my-cred"`)
}

func TestValidate_DuplicateBackendNames(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"backends": `
backends:
  - name: cluster
    server: https://k8s-1:6443
    ca_ref: ca
    auth_ref: cred
  - name: cluster
    server: https://k8s-2:6443
    ca_ref: ca
    auth_ref: cred`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `duplicate name "cluster"`)
}

func TestValidate_DuplicateRouteNames(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"routes": `
routes:
  - name: my-route
    match:
      sni: foo.example.com
    backend_ref: cluster
  - name: my-route
    backend_ref: cluster`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `duplicate name "my-route"`)
}

func TestValidate_CredentialBadCertRef(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"credentials": `
credentials:
  - name: bad
    client_certificate_ref: nonexistent-cert`,
		"backends": `
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: bad`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `client_certificate_ref references unknown certificate "nonexistent-cert"`)
}

func TestValidate_BackendBadCARef(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"certificate_authorities": `
certificate_authorities: []`,
		"backends": `
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: nonexistent-ca
    auth_ref: cred`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `ca_ref references unknown certificate_authority "nonexistent-ca"`)
}

func TestValidate_BackendBadAuthRef(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"backends": `
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: nonexistent-cred`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `auth_ref references unknown credential "nonexistent-cred"`)
}

func TestValidate_RouteBadBackendRef(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"routes": `
routes:
  - name: bad-route
    backend_ref: nonexistent-cluster`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `backend_ref references unknown backend "nonexistent-cluster"`)
}

func TestValidate_UnixListenerNoPath(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"server": `
server:
  https:
    address: ":8443"
  unix:
    path: ""`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "server.unix.path")
}

func TestValidate_CertMixed(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"certificates": `
certificates:
  - name: cert
    certificate: /a.crt
    key_data: "data"`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot mix file paths")
}

func TestValidate_EmptyMatch(t *testing.T) {
	cfg := baseYAML(map[string]string{
		"routes": `
routes:
  - name: empty-match
    match: {}
    backend_ref: cluster`,
	})
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	err = Validate(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "match block is present but has no conditions")
}

// ---------------------------------------------------------------------------
// Tests: Convert (end-to-end with real crypto materials)
// ---------------------------------------------------------------------------

func TestConvert_FullConfig(t *testing.T) {
	dir := t.TempDir()

	// Generate certs and CAs
	certPath, keyPath := writeCertFiles(t, dir, "frontend")
	clientCertPath, clientKeyPath := writeCertFiles(t, dir, "client")
	caPath := writeCAFile(t, dir, "prod-ca")
	stagingCAPath := writeCAFile(t, dir, "staging-ca")

	inlineCertPEM, inlineKeyPEM := generateSelfSignedCert(t)
	inlineCAPEM := generateCACert(t)

	cfg := `
certificates:
  - name: frontend-cert
    certificate: ` + certPath + `
    key: ` + keyPath + `

  - name: client-cert
    certificate: ` + clientCertPath + `
    key: ` + clientKeyPath + `

  - name: inline-cert
    certificate_data: |
` + indent(string(inlineCertPEM), 6) + `
    key_data: |
` + indent(string(inlineKeyPEM), 6) + `

certificate_authorities:
  - name: prod-ca
    certificate: ` + caPath + `

  - name: staging-ca
    certificate: ` + stagingCAPath + `

  - name: inline-ca
    certificate_data: |
` + indent(string(inlineCAPEM), 6) + `

credentials:
  - name: client-cert-cred
    client_certificate_ref: client-cert

  - name: token-cred
    token: "my-secret-token"

  - name: basic-cred
    basic:
      username: admin
      password: secret

backends:
  - name: prod-cluster
    server: https://k8s-prod.internal:6443
    ca_ref: prod-ca
    auth_ref: client-cert-cred
    cache_ttl: 2s

  - name: staging-cluster
    server: https://k8s-staging.internal:6443
    ca_ref: staging-ca
    auth_ref: token-cred
    insecure_skip_tls_verify: true

  - name: dev-cluster
    server: https://k8s-dev.internal:6443
    ca_ref: inline-ca
    auth_ref: basic-cred

routes:
  - name: prod-by-sni
    match:
      sni: prod.api.example.com
    backend_ref: prod-cluster

  - name: staging-by-header
    match:
      header:
        name: X-Cluster
        value: staging
    backend_ref: staging-cluster

  - name: prod-by-jwt
    match:
      jwt:
        claim: cluster
        value: prod
    backend_ref: prod-cluster

  - name: staging-by-path
    match:
      path_prefix: /staging
    backend_ref: staging-cluster

  - name: default
    backend_ref: dev-cluster

server:
  https:
    address: "0.0.0.0:8443"
    cert: ` + certPath + `
    key: ` + keyPath + `
    client_auth: request
    ca: ` + caPath + `
    min_version: "1.2"
  unix:
    path: /var/run/multikube.sock
  metrics:
    address: "0.0.0.0:8888"
  max_header_size: 1000000
  read_timeout: 30s
  write_timeout: 30s
  keep_alive: 180s
  shutdown_grace_period: 10s

auth:
  jwt_auth:
    rs256:
      public_key: /etc/multikube/keys/public.pem

cache:
  ttl: 5s
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)

	rc, err := Convert(ext)
	require.NoError(t, err)
	require.NotNil(t, rc)

	// Certificates are loaded
	assert.Len(t, rc.Certificates, 3)
	assert.Equal(t, "frontend-cert", rc.Certificates[0].Name)
	assert.NotEmpty(t, rc.Certificates[0].TLS.Certificate)
	assert.Equal(t, "inline-cert", rc.Certificates[2].Name)
	assert.NotEmpty(t, rc.Certificates[2].TLS.Certificate)

	// CAs are loaded
	assert.Len(t, rc.CertificateAuthorities, 3)
	assert.Equal(t, "prod-ca", rc.CertificateAuthorities[0].Name)
	assert.NotNil(t, rc.CertificateAuthorities[0].Pool)
	assert.Equal(t, "inline-ca", rc.CertificateAuthorities[2].Name)
	assert.NotNil(t, rc.CertificateAuthorities[2].Pool)

	// Credentials are resolved
	assert.Len(t, rc.Credentials, 3)
	assert.Equal(t, "client-cert-cred", rc.Credentials[0].Name)
	require.NotNil(t, rc.Credentials[0].ClientCertificate)
	assert.NotEmpty(t, rc.Credentials[0].ClientCertificate.Certificate)
	assert.Equal(t, "token-cred", rc.Credentials[1].Name)
	assert.Equal(t, "my-secret-token", rc.Credentials[1].Token)
	assert.Equal(t, "basic-cred", rc.Credentials[2].Name)
	assert.Equal(t, "admin", rc.Credentials[2].Username)
	assert.Equal(t, "secret", rc.Credentials[2].Password)

	// Backends are resolved
	assert.Len(t, rc.Backends, 3)
	assert.Equal(t, "prod-cluster", rc.Backends[0].Name)
	assert.Equal(t, "https", rc.Backends[0].Server.Scheme)
	assert.Equal(t, "k8s-prod.internal:6443", rc.Backends[0].Server.Host)
	require.NotNil(t, rc.Backends[0].CA)
	assert.Equal(t, "prod-ca", rc.Backends[0].CA.Name)
	require.NotNil(t, rc.Backends[0].Auth)
	assert.Equal(t, "client-cert-cred", rc.Backends[0].Auth.Name)
	assert.Equal(t, 2*time.Second, rc.Backends[0].CacheTTL)
	assert.False(t, rc.Backends[0].InsecureSkipTLSVerify)

	assert.True(t, rc.Backends[1].InsecureSkipTLSVerify)
	assert.Equal(t, "my-secret-token", rc.Backends[1].Auth.Token)

	// Routes are resolved
	assert.Len(t, rc.Routes, 5)
	assert.Equal(t, "prod-by-sni", rc.Routes[0].Name)
	require.NotNil(t, rc.Routes[0].Match)
	assert.Equal(t, "prod.api.example.com", rc.Routes[0].Match.SNI)
	require.NotNil(t, rc.Routes[0].Backend)
	assert.Equal(t, "prod-cluster", rc.Routes[0].Backend.Name)

	assert.Equal(t, "staging-by-header", rc.Routes[1].Name)
	require.NotNil(t, rc.Routes[1].Match.Header)
	assert.Equal(t, "X-Cluster", rc.Routes[1].Match.Header.Name)
	assert.Equal(t, "staging", rc.Routes[1].Match.Header.Value)

	assert.Equal(t, "prod-by-jwt", rc.Routes[2].Name)
	require.NotNil(t, rc.Routes[2].Match.JWT)
	assert.Equal(t, "cluster", rc.Routes[2].Match.JWT.Claim)

	assert.Equal(t, "staging-by-path", rc.Routes[3].Name)
	assert.Equal(t, "/staging", rc.Routes[3].Match.PathPrefix)

	assert.Equal(t, "default", rc.Routes[4].Name)
	assert.Nil(t, rc.Routes[4].Match)
	assert.Equal(t, "dev-cluster", rc.Routes[4].Backend.Name)

	// Server config — HTTPS listener
	require.NotNil(t, rc.Server.HTTPS)
	assert.Equal(t, "0.0.0.0:8443", rc.Server.HTTPS.Address)
	require.NotNil(t, rc.Server.HTTPS.TLS)
	assert.Len(t, rc.Server.HTTPS.TLS.Certificates, 1)
	assert.Equal(t, tls.RequestClientCert, rc.Server.HTTPS.TLS.ClientAuth)
	assert.NotNil(t, rc.Server.HTTPS.TLS.ClientCA)
	assert.Equal(t, uint16(tls.VersionTLS12), rc.Server.HTTPS.TLS.MinVersion)
	// Unix listener
	require.NotNil(t, rc.Server.Unix)
	assert.Equal(t, "/var/run/multikube.sock", rc.Server.Unix.Path)
	// Metrics listener
	require.NotNil(t, rc.Server.Metrics)
	assert.Equal(t, "0.0.0.0:8888", rc.Server.Metrics.Address)

	assert.Equal(t, uint64(1000000), rc.Server.MaxHeaderSize)
	assert.Equal(t, 30*time.Second, rc.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, rc.Server.WriteTimeout)
	assert.Equal(t, 3*time.Minute, rc.Server.KeepAlive)
	assert.Equal(t, 10*time.Second, rc.Server.ShutdownGracePeriod)

	// Auth
	require.NotNil(t, rc.Auth)
	require.NotNil(t, rc.Auth.JWT)
	require.NotNil(t, rc.Auth.JWT.RS256)
	assert.Equal(t, "/etc/multikube/keys/public.pem", rc.Auth.JWT.RS256.PublicKey)
	assert.Nil(t, rc.Auth.JWT.OIDC)

	// Cache
	require.NotNil(t, rc.Cache)
	assert.Equal(t, 5*time.Second, rc.Cache.TTL)
}

func TestConvert_MinimalConfig(t *testing.T) {
	dir := t.TempDir()
	caPath := writeCAFile(t, dir, "ca")

	cfg := `
certificate_authorities:
  - name: backend-ca
    certificate: ` + caPath + `

credentials:
  - name: backend-token
    token: "some-token"

backends:
  - name: default-cluster
    server: https://k8s.internal:6443
    ca_ref: backend-ca
    auth_ref: backend-token

routes:
  - name: default
    backend_ref: default-cluster

server:
  https:
    address: "0.0.0.0:8443"
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)

	rc, err := Convert(ext)
	require.NoError(t, err)
	require.NotNil(t, rc)

	assert.Len(t, rc.CertificateAuthorities, 1)
	assert.Len(t, rc.Credentials, 1)
	assert.Len(t, rc.Backends, 1)
	assert.Len(t, rc.Routes, 1)
	assert.Nil(t, rc.Routes[0].Match)
	require.NotNil(t, rc.Routes[0].Backend)
	assert.Equal(t, "default-cluster", rc.Routes[0].Backend.Name)
	assert.Nil(t, rc.Auth)
	assert.Nil(t, rc.Cache)

	// HTTPS listener should have auto-generated cert
	require.NotNil(t, rc.Server.HTTPS)
	require.NotNil(t, rc.Server.HTTPS.TLS)
	assert.Len(t, rc.Server.HTTPS.TLS.Certificates, 1)
}

func TestConvert_InlineCertificates(t *testing.T) {
	dir := t.TempDir()
	certPEM, keyPEM := generateSelfSignedCert(t)
	caPEM := generateCACert(t)
	caPath := writeCAFile(t, dir, "ca")

	cfg := `
certificates:
  - name: inline-cert
    certificate_data: |
` + indent(string(certPEM), 6) + `
    key_data: |
` + indent(string(keyPEM), 6) + `

certificate_authorities:
  - name: ca
    certificate: ` + caPath + `

credentials:
  - name: cred
    token: tok

backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: cred

routes:
  - name: default
    backend_ref: cluster

server:
  https:
    address: ":8443"
`
	_ = caPEM // generated but used through file
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)

	rc, err := Convert(ext)
	require.NoError(t, err)
	assert.NotEmpty(t, rc.Certificates[0].TLS.Certificate)
}

func TestConvert_InlineCA(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writeCertFiles(t, dir, "cert")
	caPEM := generateCACert(t)

	cfg := `
certificates:
  - name: cert
    certificate: ` + certPath + `
    key: ` + keyPath + `

certificate_authorities:
  - name: inline-ca
    certificate_data: |
` + indent(string(caPEM), 6) + `

credentials:
  - name: cred
    token: tok

backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: inline-ca
    auth_ref: cred

routes:
  - name: default
    backend_ref: cluster

server:
  https:
    address: ":8443"
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)

	rc, err := Convert(ext)
	require.NoError(t, err)
	require.NotNil(t, rc.CertificateAuthorities[0].Pool)
	require.NotNil(t, rc.Backends[0].CA)
	assert.Equal(t, "inline-ca", rc.Backends[0].CA.Name)
}

func TestConvert_TLSMinVersions(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writeCertFiles(t, dir, "cert")
	caPath := writeCAFile(t, dir, "ca")

	tests := []struct {
		version  string
		expected uint16
	}{
		{"1.0", tls.VersionTLS10},
		{"1.1", tls.VersionTLS11},
		{"1.2", tls.VersionTLS12},
		{"1.3", tls.VersionTLS13},
	}

	for _, tt := range tests {
		t.Run("TLS "+tt.version, func(t *testing.T) {
			cfg := `
certificates:
  - name: cert
    certificate: ` + certPath + `
    key: ` + keyPath + `
certificate_authorities:
  - name: ca
    certificate: ` + caPath + `
credentials:
  - name: cred
    token: tok
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: cred
routes:
  - name: default
    backend_ref: cluster
server:
  https:
    address: ":8443"
    cert: ` + certPath + `
    key: ` + keyPath + `
    min_version: "` + tt.version + `"
`
			ext, err := Load([]byte(cfg))
			require.NoError(t, err)
			rc, err := Convert(ext)
			require.NoError(t, err)
			require.NotNil(t, rc.Server.HTTPS)
			require.NotNil(t, rc.Server.HTTPS.TLS)
			assert.Equal(t, tt.expected, rc.Server.HTTPS.TLS.MinVersion)
		})
	}
}

func TestConvert_DefaultMinVersionTLS12(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writeCertFiles(t, dir, "cert")
	caPath := writeCAFile(t, dir, "ca")

	cfg := `
certificates:
  - name: cert
    certificate: ` + certPath + `
    key: ` + keyPath + `
certificate_authorities:
  - name: ca
    certificate: ` + caPath + `
credentials:
  - name: cred
    token: tok
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: cred
routes:
  - name: default
    backend_ref: cluster
server:
  https:
    address: ":8443"
    cert: ` + certPath + `
    key: ` + keyPath + `
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	rc, err := Convert(ext)
	require.NoError(t, err)
	require.NotNil(t, rc.Server.HTTPS)
	require.NotNil(t, rc.Server.HTTPS.TLS)
	assert.Equal(t, uint16(tls.VersionTLS12), rc.Server.HTTPS.TLS.MinVersion)
}

func TestConvert_ClientCAImpliesRequireAuth(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writeCertFiles(t, dir, "cert")
	caPath := writeCAFile(t, dir, "ca")

	cfg := `
certificates:
  - name: cert
    certificate: ` + certPath + `
    key: ` + keyPath + `
certificate_authorities:
  - name: ca
    certificate: ` + caPath + `
credentials:
  - name: cred
    token: tok
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: cred
routes:
  - name: default
    backend_ref: cluster
server:
  https:
    address: ":8443"
    cert: ` + certPath + `
    key: ` + keyPath + `
    ca: ` + caPath + `
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	rc, err := Convert(ext)
	require.NoError(t, err)
	require.NotNil(t, rc.Server.HTTPS)
	require.NotNil(t, rc.Server.HTTPS.TLS)
	assert.Equal(t, tls.RequireAndVerifyClientCert, rc.Server.HTTPS.TLS.ClientAuth)
}

func TestConvert_OIDCAuth(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writeCertFiles(t, dir, "cert")
	caPath := writeCAFile(t, dir, "ca")

	cfg := `
certificates:
  - name: cert
    certificate: ` + certPath + `
    key: ` + keyPath + `
certificate_authorities:
  - name: ca
    certificate: ` + caPath + `
credentials:
  - name: cred
    token: tok
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: cred
routes:
  - name: default
    backend_ref: cluster
server:
  https:
    address: ":8443"
    cert: ` + certPath + `
    key: ` + keyPath + `
auth:
  jwt_auth:
    oidc:
      issuer_url: https://accounts.google.com
      username_claim: email
      ca_file: /etc/oidc-ca.crt
      poll_interval: 300s
      insecure_skip_tls_verify: true
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	rc, err := Convert(ext)
	require.NoError(t, err)

	require.NotNil(t, rc.Auth)
	require.NotNil(t, rc.Auth.JWT)
	assert.Nil(t, rc.Auth.JWT.RS256)
	require.NotNil(t, rc.Auth.JWT.OIDC)
	assert.Equal(t, "https://accounts.google.com", rc.Auth.JWT.OIDC.IssuerURL)
	assert.Equal(t, "email", rc.Auth.JWT.OIDC.UsernameClaim)
	assert.Equal(t, "/etc/oidc-ca.crt", rc.Auth.JWT.OIDC.CAFile)
	assert.Equal(t, 5*time.Minute, rc.Auth.JWT.OIDC.PollInterval)
	assert.True(t, rc.Auth.JWT.OIDC.InsecureSkipTLSVerify)
}

func TestConvert_CertFileNotFound(t *testing.T) {
	cfg := `
certificates:
  - name: cert
    certificate: /nonexistent/cert.crt
    key: /nonexistent/cert.key
certificate_authorities:
  - name: ca
    certificate: /nonexistent/ca.crt
credentials:
  - name: cred
    token: tok
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: cred
routes:
  - name: default
    backend_ref: cluster
server:
  https:
    address: ":8443"
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	_, err = Convert(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `certificates[0] "cert"`)
}

func TestConvert_CAFileNotFound(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writeCertFiles(t, dir, "cert")

	cfg := `
certificates:
  - name: cert
    certificate: ` + certPath + `
    key: ` + keyPath + `
certificate_authorities:
  - name: ca
    certificate: /nonexistent/ca.crt
credentials:
  - name: cred
    token: tok
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: cred
routes:
  - name: default
    backend_ref: cluster
server:
  https:
    address: ":8443"
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	_, err = Convert(ext)
	require.Error(t, err)
	assert.Contains(t, err.Error(), `certificate_authorities[0] "ca"`)
	assert.Contains(t, err.Error(), "reading file")
}

func TestConvert_BackendNoOptionalRefs(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := writeCertFiles(t, dir, "cert")

	cfg := `
certificates:
  - name: cert
    certificate: ` + certPath + `
    key: ` + keyPath + `
credentials:
  - name: cred
    token: tok
backends:
  - name: cluster
    server: https://k8s:6443
routes:
  - name: default
    backend_ref: cluster
server:
  https:
    address: ":8443"
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	rc, err := Convert(ext)
	require.NoError(t, err)

	assert.Nil(t, rc.Backends[0].CA)
	assert.Nil(t, rc.Backends[0].Auth)
}

func TestConvert_AutoGenerateCert(t *testing.T) {
	dir := t.TempDir()
	caPath := writeCAFile(t, dir, "ca")

	cfg := `
certificate_authorities:
  - name: ca
    certificate: ` + caPath + `
credentials:
  - name: cred
    token: tok
backends:
  - name: cluster
    server: https://k8s:6443
    ca_ref: ca
    auth_ref: cred
routes:
  - name: default
    backend_ref: cluster
server:
  https:
    address: ":8443"
`
	ext, err := Load([]byte(cfg))
	require.NoError(t, err)
	rc, err := Convert(ext)
	require.NoError(t, err)

	// Should auto-generate a self-signed cert
	require.NotNil(t, rc.Server.HTTPS)
	require.NotNil(t, rc.Server.HTTPS.TLS)
	assert.Len(t, rc.Server.HTTPS.TLS.Certificates, 1)
	assert.NotEmpty(t, rc.Server.HTTPS.TLS.Certificates[0].Certificate)
}

// ---------------------------------------------------------------------------
// Helper
// ---------------------------------------------------------------------------

func indent(s string, spaces int) string {
	prefix := ""
	for i := 0; i < spaces; i++ {
		prefix += " "
	}
	result := ""
	for i, line := range splitLines(s) {
		if i > 0 {
			result += "\n"
		}
		if line != "" {
			result += prefix + line
		}
	}
	return result
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
