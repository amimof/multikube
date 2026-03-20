package config

import (
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"time"
)

// RuntimeConfig is the fully-resolved internal configuration.
// All string references have been resolved to direct pointers/values,
// crypto materials have been loaded and parsed, and durations converted
// to native Go types.
type RuntimeConfig struct {
	Certificates           []Certificate
	CertificateAuthorities []CertificateAuthority
	Credentials            []Credential
	Backends               []Backend
	Routes                 []Route
	Server                 ServerConfig
	Auth                   *AuthConfig
	Metrics                *MetricsConfig
	Cache                  *CacheConfig
}

// Certificate holds a loaded TLS certificate key pair, ready to use.
type Certificate struct {
	Name string
	TLS  tls.Certificate
}

// CertificateAuthority holds a parsed CA certificate pool, ready to use.
type CertificateAuthority struct {
	Name string
	Pool *x509.CertPool
}

// Credential holds resolved authentication material for a backend.
// Exactly one of ClientCertificate, Token, or Basic (Username+Password)
// is populated.
type Credential struct {
	Name              string
	ClientCertificate *tls.Certificate // resolved from client_certificate_ref
	Token             string
	Username          string
	Password          string
}

// Backend is a fully-resolved backend target.
type Backend struct {
	Name                  string
	Server                *url.URL
	CA                    *CertificateAuthority // resolved from ca_ref, may be nil
	Auth                  *Credential           // resolved from auth_ref, may be nil
	InsecureSkipTLSVerify bool
	CacheTTL              time.Duration
}

// Route maps incoming requests to a resolved backend.
type Route struct {
	Name    string
	Match   *Match
	Backend *Backend // resolved from backend_ref
}

// Match specifies request matching criteria for a route.
type Match struct {
	SNI        string
	Header     *HeaderMatch
	PathPrefix string
	JWT        *JWTMatch
}

// HeaderMatch matches on an HTTP header name/value pair.
type HeaderMatch struct {
	Name  string
	Value string
}

// JWTMatch matches on a JWT claim name/value pair.
type JWTMatch struct {
	Claim string
	Value string
}

// ServerConfig holds server listener and timeout configuration.
type ServerConfig struct {
	Listeners           []Listener
	MaxHeaderSize       uint64
	ReadTimeout         time.Duration
	WriteTimeout        time.Duration
	KeepAlive           time.Duration
	ShutdownGracePeriod time.Duration
}

// Listener defines a single network listener.
type Listener struct {
	Protocol   string
	Address    string
	Port       int
	SocketPath string
	TLS        *ListenerTLS
}

// ListenerTLS holds fully-resolved TLS configuration for a listener.
type ListenerTLS struct {
	Certificates []tls.Certificate  // resolved from certificate_refs
	ClientAuth   tls.ClientAuthType // parsed from string
	ClientCA     *x509.CertPool     // resolved from client_ca_ref
	MinVersion   uint16             // parsed TLS version constant
}

// AuthConfig holds authentication configuration.
type AuthConfig struct {
	JWT *JWTAuthConfig
}

// JWTAuthConfig holds JWT authentication configuration.
// Exactly one of RS256 or OIDC is populated.
type JWTAuthConfig struct {
	RS256 *RS256AuthConfig
	OIDC  *OIDCAuthConfig
}

// RS256AuthConfig holds RS256 JWT verification configuration.
type RS256AuthConfig struct {
	PublicKey string
}

// OIDCAuthConfig holds OIDC JWT verification configuration.
type OIDCAuthConfig struct {
	IssuerURL             string
	UsernameClaim         string
	CAFile                string
	PollInterval          time.Duration
	InsecureSkipTLSVerify bool
}

// MetricsConfig holds Prometheus metrics endpoint configuration.
type MetricsConfig struct {
	Address string
	Port    int
}

// CacheConfig holds cache configuration.
type CacheConfig struct {
	TTL time.Duration
}
