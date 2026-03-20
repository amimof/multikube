package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"time"

	"github.com/amimof/multikube/pkg/config"
)

// NewServerFromConfig constructs a Server from a fully-resolved RuntimeConfig.
// The handler is the HTTP handler chain (proxy + middlewares).
//
// All crypto materials, URL parsing, reference resolution, and duration
// conversion have already been performed by config.Convert(). This function
// simply maps the resolved values into Server struct fields.
func NewServerFromConfig(cfg *config.RuntimeConfig, handler http.Handler) (*Server, error) {
	s := &Server{
		Handler: handler,
	}

	// Apply server-level settings with defaults.
	if cfg.Server.MaxHeaderSize != 0 {
		s.MaxHeaderSize = cfg.Server.MaxHeaderSize
	} else {
		s.MaxHeaderSize = 1000000
	}
	if cfg.Server.ShutdownGracePeriod != 0 {
		s.CleanupTimeout = cfg.Server.ShutdownGracePeriod
	} else {
		s.CleanupTimeout = 10 * time.Second
	}

	// Track which protocols we've seen to detect duplicates.
	seen := make(map[string]bool)

	for i, l := range cfg.Server.Listeners {
		if seen[l.Protocol] {
			return nil, fmt.Errorf("server.listeners[%d]: duplicate protocol %q; the server currently supports at most one listener per protocol", i, l.Protocol)
		}
		seen[l.Protocol] = true
		s.EnabledListeners = append(s.EnabledListeners, l.Protocol)

		switch l.Protocol {
		case "http":
			applyHTTPListener(s, &l, &cfg.Server)
		case "https":
			applyHTTPSListener(s, &l, &cfg.Server)
		case "unix":
			s.SocketPath = l.SocketPath
		}
	}

	return s, nil
}

// applyHTTPListener maps an HTTP listener config to Server fields.
func applyHTTPListener(s *Server, l *config.Listener, sc *config.ServerConfig) {
	s.Host = l.Address
	s.Port = l.Port
	s.ListenLimit = 0

	if sc.KeepAlive != 0 {
		s.KeepAlive = sc.KeepAlive
	} else {
		s.KeepAlive = 3 * time.Minute
	}
	if sc.ReadTimeout != 0 {
		s.ReadTimeout = sc.ReadTimeout
	} else {
		s.ReadTimeout = 30 * time.Second
	}
	if sc.WriteTimeout != 0 {
		s.WriteTimeout = sc.WriteTimeout
	} else {
		s.WriteTimeout = 30 * time.Second
	}
}

// applyHTTPSListener maps an HTTPS listener config to Server fields and builds
// the TLS configuration from already-resolved crypto materials.
func applyHTTPSListener(s *Server, l *config.Listener, sc *config.ServerConfig) {
	s.TLSHost = l.Address
	s.TLSPort = l.Port
	s.TLSListenLimit = 0

	if sc.KeepAlive != 0 {
		s.TLSKeepAlive = sc.KeepAlive
	} else {
		s.TLSKeepAlive = 3 * time.Minute
	}
	if sc.ReadTimeout != 0 {
		s.TLSReadTimeout = sc.ReadTimeout
	} else {
		s.TLSReadTimeout = 30 * time.Second
	}
	if sc.WriteTimeout != 0 {
		s.TLSWriteTimeout = sc.WriteTimeout
	} else {
		s.TLSWriteTimeout = 30 * time.Second
	}

	// Clear file-path fields so the existing Serve() code path in
	// server.go skips its own certificate loading logic.
	s.TLSCertificate = ""
	s.TLSCertificateKey = ""
	s.TLSCACertificate = ""

	if l.TLS != nil {
		s.TLSConfig = &tls.Config{
			Certificates:             l.TLS.Certificates,
			PreferServerCipherSuites: true,
			CurvePreferences:         []tls.CurveID{tls.CurveP256},
			NextProtos:               []string{"h2", "http/1.1"},
			MinVersion:               l.TLS.MinVersion,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			ClientAuth: l.TLS.ClientAuth,
			ClientCAs:  l.TLS.ClientCA,
		}
	}
}
