package server

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-openapi/swag"
	"github.com/tylerb/graceful"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
	schemeUnix  = "unix"
)

// Server for the multikube API
type Server struct {
	EnabledListeners  []string
	Host              string
	Port              int
	ListenLimit       int
	TLSHost           string
	TLSPort           int
	TLSListenLimit    int
	TLSCertificate    string
	TLSCertificateKey string
	TLSCACertificate  string
	SocketPath        string
	Name              string
	KeepAlive         time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	TLSKeepAlive      time.Duration
	TLSReadTimeout    time.Duration
	TLSWriteTimeout   time.Duration
	CleanupTimeout    time.Duration
	MaxHeaderSize     uint64
	Handler           http.Handler

	shutdown      chan struct{}
	httpServerL   net.Listener
	httpsServerL  net.Listener
	domainSocketL net.Listener
	hasListeners  bool
	shuttingDown  int32
}

// NewServer returns a default non-tls server
func NewServer() *Server {
	return &Server{
		EnabledListeners: []string{"http"},
		CleanupTimeout:   10 * time.Second,
		MaxHeaderSize:    1000000,
		Host:             "127.0.0.1",
		Port:             8080,
		ListenLimit:      0,
		KeepAlive:        3 * time.Minute,
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
	}
}

// NewServerTLS returns a default TLS enabled server
func NewServerTLS() *Server {
	return &Server{
		EnabledListeners:  []string{"https"},
		CleanupTimeout:    10 * time.Second,
		MaxHeaderSize:     1000000,
		TLSHost:           "127.0.0.1",
		TLSPort:           8443,
		TLSCertificate:    "",
		TLSCertificateKey: "",
		TLSCACertificate:  "",
		TLSListenLimit:    0,
		TLSKeepAlive:      3 * time.Minute,
		TLSReadTimeout:    30 * time.Second,
		TLSWriteTimeout:   30 * time.Second,
	}
}

func (s *Server) hasScheme(scheme string) bool {
	for _, v := range s.EnabledListeners {
		if v == scheme {
			return true
		}
	}
	return false
}

// Listen configures server listeners
func (s *Server) Listen() error {

	if s.shutdown == nil {
		s.shutdown = make(chan struct{})
	}

	if s.hasListeners { // already done this
		return nil
	}

	if s.hasScheme(schemeHTTPS) {
		// Use http host if https host wasn't defined
		if s.TLSHost == "" {
			s.TLSHost = s.Host
		}
		// Use http listen limit if https listen limit wasn't defined
		if s.TLSListenLimit == 0 {
			s.TLSListenLimit = s.ListenLimit
		}
		// Use http tcp keep alive if https tcp keep alive wasn't defined
		if int64(s.TLSKeepAlive) == 0 {
			s.TLSKeepAlive = s.KeepAlive
		}
		// Use http read timeout if https read timeout wasn't defined
		if int64(s.TLSReadTimeout) == 0 {
			s.TLSReadTimeout = s.ReadTimeout
		}
		// Use http write timeout if https write timeout wasn't defined
		if int64(s.TLSWriteTimeout) == 0 {
			s.TLSWriteTimeout = s.WriteTimeout
		}
	}

	if s.hasScheme(schemeUnix) {
		domSockListener, err := net.Listen("unix", string(s.SocketPath))
		if err != nil {
			return err
		}
		s.domainSocketL = domSockListener
	}

	if s.hasScheme(schemeHTTP) {
		listener, err := net.Listen("tcp", net.JoinHostPort(s.Host, strconv.Itoa(s.Port)))
		if err != nil {
			return err
		}

		h, p, err := swag.SplitHostPort(listener.Addr().String())
		if err != nil {
			return err
		}
		s.Host = h
		s.Port = p
		s.httpServerL = listener
	}

	if s.hasScheme(schemeHTTPS) {
		tlsListener, err := net.Listen("tcp", net.JoinHostPort(s.TLSHost, strconv.Itoa(s.TLSPort)))
		if err != nil {
			return err
		}

		sh, sp, err := swag.SplitHostPort(tlsListener.Addr().String())
		if err != nil {
			return err
		}
		s.TLSHost = sh
		s.TLSPort = sp
		s.httpsServerL = tlsListener
	}

	s.hasListeners = true
	return nil
}

// Serve the api
func (s *Server) Serve() error {

	if !s.hasListeners {
		err := s.Listen()
		if err != nil {
			return err
		}
	}

	if s.Name == "" {
		s.Name = "multikube"
	}

	var wg sync.WaitGroup

	if s.hasScheme(schemeUnix) {
		domainSocket := &graceful.Server{Server: new(http.Server)}
		domainSocket.MaxHeaderBytes = int(s.MaxHeaderSize)
		if int64(s.CleanupTimeout) > 0 {
			domainSocket.Timeout = s.CleanupTimeout
		}

		domainSocket.Handler = s.Handler

		wg.Add(2)
		log.Printf("Serving %s at unix://%s", s.Name, s.SocketPath)
		go func(l net.Listener) {
			defer wg.Done()
			if err := domainSocket.Serve(l); err != nil {
				log.Fatalf("%v", err)
			}
			log.Printf("Stopped serving %s at unix://%s", s.Name, s.SocketPath)
		}(s.domainSocketL)
		go s.handleShutdown(&wg, domainSocket)
	}

	if s.hasScheme(schemeHTTP) {
		httpServer := &graceful.Server{Server: new(http.Server)}
		httpServer.MaxHeaderBytes = int(s.MaxHeaderSize)
		httpServer.ReadTimeout = s.ReadTimeout
		httpServer.WriteTimeout = s.WriteTimeout
		httpServer.SetKeepAlivesEnabled(int64(s.KeepAlive) > 0)
		httpServer.TCPKeepAlive = s.KeepAlive
		if s.ListenLimit > 0 {
			httpServer.ListenLimit = s.ListenLimit
		}

		if int64(s.CleanupTimeout) > 0 {
			httpServer.Timeout = s.CleanupTimeout
		}

		httpServer.Handler = s.Handler

		wg.Add(2)
		log.Printf("Serving %s at http://%s", s.Name, s.httpServerL.Addr())
		go func(l net.Listener) {
			defer wg.Done()
			if err := httpServer.Serve(l); err != nil {
				log.Printf("%v", err)
			}
			log.Printf("Stopped serving %s at http://%s", s.Name, l.Addr())
		}(s.httpServerL)
		go s.handleShutdown(&wg, httpServer)
	}

	if s.hasScheme(schemeHTTPS) {

		srv := http.Server{}
		httpsServer := &graceful.Server{Server: &srv}
		httpsServer.MaxHeaderBytes = int(s.MaxHeaderSize)
		//httpsServer.ReadTimeout = s.TLSReadTimeout
		//httpsServer.WriteTimeout = s.TLSWriteTimeout
		httpsServer.SetKeepAlivesEnabled(int64(s.TLSKeepAlive) > 0)
		httpsServer.TCPKeepAlive = s.TLSKeepAlive
		if s.TLSListenLimit > 0 {
			httpsServer.ListenLimit = s.TLSListenLimit
		}
		if int64(s.CleanupTimeout) > 0 {
			httpsServer.Timeout = s.CleanupTimeout
		}

		httpsServer.Handler = s.Handler

		// Inspired by https://blog.bracebin.com/achieving-perfect-ssl-labs-score-with-go
		httpsServer.TLSConfig = &tls.Config{
			// Causes servers to use Go's default ciphersuite preferences,
			// which are tuned to avoid attacks. Does nothing on clients.
			PreferServerCipherSuites: true,
			// Only use curves which have assembly implementations
			// https://github.com/golang/go/tree/master/src/crypto/elliptic
			CurvePreferences: []tls.CurveID{tls.CurveP256},
			// Use modern tls mode https://wiki.mozilla.org/Security/Server_Side_TLS#Modern_compatibility
			NextProtos: []string{"h2", "http/1.1"},
			// https://www.owasp.org/index.php/Transport_Layer_Protection_Cheat_Sheet#Rule_-_Only_Support_Strong_Protocols
			MinVersion: tls.VersionTLS12,
			// These ciphersuites support Forward Secrecy: https://en.wikipedia.org/wiki/Forward_secrecy
			CipherSuites: []uint16{
				tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			},
			ClientAuth: tls.RequestClientCert,
		}

		if s.TLSCertificate != "" && s.TLSCertificateKey != "" {
			httpsServer.TLSConfig.Certificates = make([]tls.Certificate, 1)
			cert, err := tls.LoadX509KeyPair(s.TLSCertificate, s.TLSCertificateKey)
			if err != nil {
				return err
			}
			httpsServer.TLSConfig.Certificates[0] = cert
		}

		if s.TLSCACertificate != "" {
			caCert, err := ioutil.ReadFile(s.TLSCACertificate)
			if err != nil {
				return err
			}
			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)
			httpsServer.TLSConfig.ClientCAs = caCertPool
			httpsServer.TLSConfig.ClientAuth = tls.RequireAndVerifyClientCert
		}

		httpsServer.TLSConfig.BuildNameToCertificate()

		if len(httpsServer.TLSConfig.Certificates) == 0 {
			if s.TLSCertificate == "" {
				if s.TLSCertificateKey == "" {
					log.Fatalf("the required flags `--tls-certificate` and `--tls-key` were not specified")
				}
				log.Printf("the required flag `--tls-certificate` was not specified")
			}
			if s.TLSCertificateKey == "" {
				log.Fatalf("the required flag `--tls-key` was not specified")
			}
		}

		wg.Add(2)
		log.Printf("Serving %s at https://%s", s.Name, s.httpsServerL.Addr())
		go func(l net.Listener) {
			defer wg.Done()
			if err := httpsServer.Serve(l); err != nil {
				log.Fatalf("%v", err)
			}
			log.Printf("Stopped serving %s at https://%s", s.Name, l.Addr())
		}(tls.NewListener(s.httpsServerL, httpsServer.TLSConfig))
		go s.handleShutdown(&wg, httpsServer)
	}

	wg.Wait()
	return nil
}

func (s *Server) handleShutdown(wg *sync.WaitGroup, server *graceful.Server) {
	defer wg.Done()
	for {
		select {
		case <-s.shutdown:
			atomic.AddInt32(&s.shuttingDown, 1)
			server.Stop(s.CleanupTimeout)
			<-server.StopChan()
			log.Printf("Shutting down")
			return
		case <-server.StopChan():
			atomic.AddInt32(&s.shuttingDown, 1)
			log.Printf("Shutting down")
			return
		}
	}
}
