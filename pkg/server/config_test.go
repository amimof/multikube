package server

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"testing"
	"time"

	"github.com/amimof/multikube/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers for generating test certs ---

// generateTestCert creates a self-signed tls.Certificate for use in tests.
func generateTestCert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-1 * time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)
	return tlsCert
}

// generateTestCAPool creates a CA cert pool for use in tests.
func generateTestCAPool(t *testing.T) *x509.CertPool {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(100),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             time.Now().Add(-1 * time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	caPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	pool := x509.NewCertPool()
	require.True(t, pool.AppendCertsFromPEM(caPEM))
	return pool
}

var noopHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

// --- Tests ---

func TestNewServerFromConfig_UnixOnly(t *testing.T) {
	cfg := &config.RuntimeConfig{
		Server: config.ServerConfig{
			Unix: &config.UnixListenerConfig{
				Path: "/tmp/test.sock",
			},
		},
	}

	s, err := NewServerFromConfig(cfg, noopHandler)
	require.NoError(t, err)

	assert.Equal(t, []string{"unix"}, s.EnabledListeners)
	assert.Equal(t, "/tmp/test.sock", s.SocketPath)
}

func TestNewServerFromConfig_HTTPSWithCerts(t *testing.T) {
	tlsCert := generateTestCert(t)

	cfg := &config.RuntimeConfig{
		Server: config.ServerConfig{
			HTTPS: &config.HTTPSListenerConfig{
				Address: "0.0.0.0:8443",
				TLS: &config.ListenerTLS{
					Certificates: []tls.Certificate{tlsCert},
					ClientAuth:   tls.NoClientCert,
					MinVersion:   tls.VersionTLS12,
				},
			},
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			KeepAlive:    2 * time.Minute,
		},
	}

	s, err := NewServerFromConfig(cfg, noopHandler)
	require.NoError(t, err)

	assert.Equal(t, []string{"https"}, s.EnabledListeners)
	assert.Equal(t, "0.0.0.0", s.TLSHost)
	assert.Equal(t, 8443, s.TLSPort)
	assert.Equal(t, 10*time.Second, s.TLSReadTimeout)
	assert.Equal(t, 10*time.Second, s.TLSWriteTimeout)
	assert.Equal(t, 2*time.Minute, s.TLSKeepAlive)

	require.NotNil(t, s.TLSConfig)
	assert.Len(t, s.TLSConfig.Certificates, 1)
	assert.Equal(t, uint16(tls.VersionTLS12), s.TLSConfig.MinVersion)
	assert.Equal(t, tls.NoClientCert, s.TLSConfig.ClientAuth)
	assert.Nil(t, s.TLSConfig.ClientCAs)
}

func TestNewServerFromConfig_HTTPSWithMultipleCerts(t *testing.T) {
	cert1 := generateTestCert(t)
	cert2 := generateTestCert(t)

	cfg := &config.RuntimeConfig{
		Server: config.ServerConfig{
			HTTPS: &config.HTTPSListenerConfig{
				Address: "0.0.0.0:8443",
				TLS: &config.ListenerTLS{
					Certificates: []tls.Certificate{cert1, cert2},
					MinVersion:   tls.VersionTLS12,
				},
			},
		},
	}

	s, err := NewServerFromConfig(cfg, noopHandler)
	require.NoError(t, err)

	require.NotNil(t, s.TLSConfig)
	assert.Len(t, s.TLSConfig.Certificates, 2)
}

func TestNewServerFromConfig_HTTPSWithClientAuth(t *testing.T) {
	tlsCert := generateTestCert(t)
	caPool := generateTestCAPool(t)

	cfg := &config.RuntimeConfig{
		Server: config.ServerConfig{
			HTTPS: &config.HTTPSListenerConfig{
				Address: "0.0.0.0:8443",
				TLS: &config.ListenerTLS{
					Certificates: []tls.Certificate{tlsCert},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					ClientCA:     caPool,
					MinVersion:   tls.VersionTLS12,
				},
			},
		},
	}

	s, err := NewServerFromConfig(cfg, noopHandler)
	require.NoError(t, err)

	require.NotNil(t, s.TLSConfig)
	assert.Equal(t, tls.RequireAndVerifyClientCert, s.TLSConfig.ClientAuth)
	assert.NotNil(t, s.TLSConfig.ClientCAs)
}

func TestNewServerFromConfig_HTTPSClientAuthRequest(t *testing.T) {
	tlsCert := generateTestCert(t)

	cfg := &config.RuntimeConfig{
		Server: config.ServerConfig{
			HTTPS: &config.HTTPSListenerConfig{
				Address: "0.0.0.0:8443",
				TLS: &config.ListenerTLS{
					Certificates: []tls.Certificate{tlsCert},
					ClientAuth:   tls.RequestClientCert,
					MinVersion:   tls.VersionTLS12,
				},
			},
		},
	}

	s, err := NewServerFromConfig(cfg, noopHandler)
	require.NoError(t, err)

	require.NotNil(t, s.TLSConfig)
	assert.Equal(t, tls.RequestClientCert, s.TLSConfig.ClientAuth)
}

func TestNewServerFromConfig_AllProtocols(t *testing.T) {
	tlsCert := generateTestCert(t)

	cfg := &config.RuntimeConfig{
		Server: config.ServerConfig{
			HTTPS: &config.HTTPSListenerConfig{
				Address: "0.0.0.0:8443",
				TLS: &config.ListenerTLS{
					Certificates: []tls.Certificate{tlsCert},
					MinVersion:   tls.VersionTLS12,
				},
			},
			Unix: &config.UnixListenerConfig{
				Path: "/tmp/mk.sock",
			},
			ReadTimeout:         5 * time.Second,
			WriteTimeout:        5 * time.Second,
			KeepAlive:           1 * time.Minute,
			ShutdownGracePeriod: 3 * time.Second,
		},
	}

	s, err := NewServerFromConfig(cfg, noopHandler)
	require.NoError(t, err)

	assert.Equal(t, []string{"https", "unix"}, s.EnabledListeners)
	assert.Equal(t, "0.0.0.0", s.TLSHost)
	assert.Equal(t, 8443, s.TLSPort)
	assert.Equal(t, "/tmp/mk.sock", s.SocketPath)
	assert.NotNil(t, s.TLSConfig)
}

func TestNewServerFromConfig_TLSMinVersions(t *testing.T) {
	tests := []struct {
		version uint16
		name    string
	}{
		{tls.VersionTLS10, "TLS_1.0"},
		{tls.VersionTLS11, "TLS_1.1"},
		{tls.VersionTLS12, "TLS_1.2"},
		{tls.VersionTLS13, "TLS_1.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tlsCert := generateTestCert(t)

			cfg := &config.RuntimeConfig{
				Server: config.ServerConfig{
					HTTPS: &config.HTTPSListenerConfig{
						Address: "0.0.0.0:8443",
						TLS: &config.ListenerTLS{
							Certificates: []tls.Certificate{tlsCert},
							MinVersion:   tt.version,
						},
					},
				},
			}

			s, err := NewServerFromConfig(cfg, noopHandler)
			require.NoError(t, err)
			require.NotNil(t, s.TLSConfig)
			assert.Equal(t, tt.version, s.TLSConfig.MinVersion)
		})
	}
}

func TestNewServerFromConfig_Defaults(t *testing.T) {
	tlsCert := generateTestCert(t)

	cfg := &config.RuntimeConfig{
		Server: config.ServerConfig{
			HTTPS: &config.HTTPSListenerConfig{
				Address: "127.0.0.1:8443",
				TLS: &config.ListenerTLS{
					Certificates: []tls.Certificate{tlsCert},
					MinVersion:   tls.VersionTLS12,
				},
			},
			// All timeouts left at zero — should get defaults.
		},
	}

	s, err := NewServerFromConfig(cfg, noopHandler)
	require.NoError(t, err)

	assert.Equal(t, uint64(1000000), s.MaxHeaderSize)
	assert.Equal(t, 10*time.Second, s.CleanupTimeout)
	assert.Equal(t, 3*time.Minute, s.TLSKeepAlive)
	assert.Equal(t, 30*time.Second, s.TLSReadTimeout)
	assert.Equal(t, 30*time.Second, s.TLSWriteTimeout)
}

func TestNewServerFromConfig_HandlerIsSet(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	tlsCert := generateTestCert(t)

	cfg := &config.RuntimeConfig{
		Server: config.ServerConfig{
			HTTPS: &config.HTTPSListenerConfig{
				Address: "127.0.0.1:8443",
				TLS: &config.ListenerTLS{
					Certificates: []tls.Certificate{tlsCert},
					MinVersion:   tls.VersionTLS12,
				},
			},
		},
	}

	s, err := NewServerFromConfig(cfg, handler)
	require.NoError(t, err)
	assert.NotNil(t, s.Handler)

	// Verify it's the same handler.
	s.Handler.ServeHTTP(nil, nil)
	assert.True(t, called)
}

func TestNewServerFromConfig_BarePort(t *testing.T) {
	tlsCert := generateTestCert(t)

	cfg := &config.RuntimeConfig{
		Server: config.ServerConfig{
			HTTPS: &config.HTTPSListenerConfig{
				Address: ":8443",
				TLS: &config.ListenerTLS{
					Certificates: []tls.Certificate{tlsCert},
					MinVersion:   tls.VersionTLS12,
				},
			},
		},
	}

	s, err := NewServerFromConfig(cfg, noopHandler)
	require.NoError(t, err)

	assert.Equal(t, "", s.TLSHost)
	assert.Equal(t, 8443, s.TLSPort)
}
