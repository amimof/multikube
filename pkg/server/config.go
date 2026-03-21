package server

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strconv"
	"strings"
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

	// Apply HTTPS listener.
	if cfg.Server.HTTPS != nil {
		s.EnabledListeners = append(s.EnabledListeners, "https")
		if err := applyHTTPSListener(s, cfg.Server.HTTPS, &cfg.Server); err != nil {
			return nil, fmt.Errorf("server.https: %w", err)
		}
	}

	// Apply Unix listener.
	if cfg.Server.Unix != nil {
		s.EnabledListeners = append(s.EnabledListeners, "unix")
		s.SocketPath = cfg.Server.Unix.Path
	}

	return s, nil
}

// applyHTTPSListener maps an HTTPS listener config to Server fields and builds
// the TLS configuration from already-resolved crypto materials.
func applyHTTPSListener(s *Server, hl *config.HTTPSListenerConfig, sc *config.ServerConfig) error {
	host, port, err := parseAddress(hl.Address)
	if err != nil {
		return fmt.Errorf("parsing address: %w", err)
	}
	s.TLSHost = host
	s.TLSPort = port
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

	if hl.TLS != nil {
		s.TLSConfig = &tls.Config{
			Certificates:             hl.TLS.Certificates,
			PreferServerCipherSuites: true,
			CurvePreferences:         []tls.CurveID{tls.CurveP256},
			NextProtos:               []string{"h2", "http/1.1"},
			MinVersion:               hl.TLS.MinVersion,
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			ClientAuth: hl.TLS.ClientAuth,
			ClientCAs:  hl.TLS.ClientCA,
		}
	}

	return nil
}

// parseAddress splits a Go net.Listen-style address (e.g. ":8443",
// "0.0.0.0:8443") into host string and port int.
func parseAddress(addr string) (string, int, error) {
	if addr == "" {
		return "", 0, nil
	}
	// Handle bare port like ":8443"
	if strings.HasPrefix(addr, ":") {
		port, err := strconv.Atoi(addr[1:])
		if err != nil {
			return "", 0, fmt.Errorf("invalid port in %q: %w", addr, err)
		}
		return "", port, nil
	}
	// Split host:port
	lastColon := strings.LastIndex(addr, ":")
	if lastColon < 0 {
		return "", 0, fmt.Errorf("no port in address %q", addr)
	}
	host := addr[:lastColon]
	port, err := strconv.Atoi(addr[lastColon+1:])
	if err != nil {
		return "", 0, fmt.Errorf("invalid port in %q: %w", addr, err)
	}
	return host, port, nil
}
