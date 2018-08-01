package multikube

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-openapi/swag"
	"github.com/spf13/pflag"
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

var (
	enabledListeners []string
	cleanupTimout    time.Duration
	maxHeaderSize    uint64

	socketPath string

	host         string
	port         int
	listenLimit  int
	keepAlive    time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration

	tlsHost           string
	tlsPort           int
	tlsListenLimit    int
	tlsKeepAlive      time.Duration
	tlsReadTimeout    time.Duration
	tlsWriteTimeout   time.Duration
	tlsCertificate    string
	tlsCertificateKey string
	tlsCACertificate  string
)

// Server for the multikube API
type Server struct {
	EnabledListeners []string
	CleanupTimeout   time.Duration
	MaxHeaderSize    uint64

	SocketPath    string
	domainSocketL net.Listener

	Host         string
	Port         int
	ListenLimit  int
	KeepAlive    time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	httpServerL  net.Listener

	TLSHost           string
	TLSPort           int
	TLSCertificate    string
	TLSCertificateKey string
	TLSCACertificate  string
	TLSListenLimit    int
	TLSKeepAlive      time.Duration
	TLSReadTimeout    time.Duration
	TLSWriteTimeout   time.Duration
	httpsServerL      net.Listener

	handler      http.Handler
	hasListeners bool
	shutdown     chan struct{}
	shuttingDown int32
}

func init() {
	pflag.StringSliceVar(&enabledListeners, "scheme", []string{}, "the listeners to enable, this can be repeated and defaults to the schemes in the swagger spec")
	pflag.DurationVar(&cleanupTimout, "cleanup-timeout", 10*time.Second, "grace period for which to wait before shutting down the server")
	pflag.Uint64Var(&maxHeaderSize, "max-header-size", 1000000, "controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body")

	pflag.StringVar(&socketPath, "socket-path", "/tmp/multikube.sock", "the unix socket to listen on")

	pflag.StringVar(&host, "host", "localhost", "the IP to listen on")
	pflag.IntVar(&port, "port", 443, "the port to listen on for insecure connections, defaults to 443")
	pflag.IntVar(&listenLimit, "listen-limit", 0, "limit the number of outstanding requests")
	pflag.DurationVar(&keepAlive, "keep-alive", 3*time.Minute, "sets the TCP keep-alive timeouts on accepted connections. It prunes dead TCP connections ( e.g. closing laptop mid-download)")
	pflag.DurationVar(&readTimeout, "read-timeout", 30*time.Second, "maximum duration before timing out read of the request")
	pflag.DurationVar(&writeTimeout, "write-timeout", 30*time.Second, "maximum duration before timing out write of the response")

	pflag.StringVar(&tlsHost, "tls-host", "localhost", "the IP to listen on")
	pflag.IntVar(&tlsPort, "tls-port", 0, "the port to listen on for secure connections, defaults to a random value")
	pflag.StringVar(&tlsCertificate, "tls-certificate", "", "the certificate to use for secure connections")
	pflag.StringVar(&tlsCertificateKey, "tls-key", "", "the private key to use for secure conections")
	pflag.StringVar(&tlsCACertificate, "tls-ca", "", "the certificate authority file to be used with mutual tls auth")
	pflag.IntVar(&tlsListenLimit, "tls-listen-limit", 0, "limit the number of outstanding requests")
	pflag.DurationVar(&tlsKeepAlive, "tls-keep-alive", 3*time.Minute, "sets the TCP keep-alive timeouts on accepted connections. It prunes dead TCP connections ( e.g. closing laptop mid-download)")
	pflag.DurationVar(&tlsReadTimeout, "tls-read-timeout", 30*time.Second, "maximum duration before timing out read of the request")
	pflag.DurationVar(&tlsWriteTimeout, "tls-write-timeout", 30*time.Second, "maximum duration before timing out write of the response")
}

func (s *Server) hasScheme(scheme string) bool {
	for _, v := range s.EnabledListeners {
		if v == scheme {
			return true
		}
	}
	return false
}

// NewServer creates a new multikube server
func NewServer(h http.Handler) *Server {
	s := new(Server)

	s.EnabledListeners = enabledListeners
	s.CleanupTimeout = cleanupTimout
	s.MaxHeaderSize = maxHeaderSize
	s.SocketPath = socketPath
	s.Host = host
	s.Port = port
	s.ListenLimit = listenLimit
	s.KeepAlive = keepAlive
	s.ReadTimeout = readTimeout
	s.WriteTimeout = writeTimeout
	s.TLSHost = tlsHost
	s.TLSPort = tlsPort
	s.TLSCertificate = tlsCertificate
	s.TLSCertificateKey = tlsCertificateKey
	s.TLSCACertificate = tlsCACertificate
	s.TLSListenLimit = tlsListenLimit
	s.TLSKeepAlive = tlsKeepAlive
	s.TLSReadTimeout = tlsReadTimeout
	s.TLSWriteTimeout = tlsWriteTimeout
	s.shutdown = make(chan struct{})
	s.handler = h

	return s
}

// Listen configures server listeners
func (s *Server) Listen() error {
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

	var wg sync.WaitGroup

	if s.hasScheme(schemeUnix) {
		domainSocket := &graceful.Server{Server: new(http.Server)}
		domainSocket.MaxHeaderBytes = int(s.MaxHeaderSize)
		if int64(s.CleanupTimeout) > 0 {
			domainSocket.Timeout = s.CleanupTimeout
		}

		domainSocket.Handler = s.handler

		wg.Add(2)
		log.Printf("Serving multikube at unix://%s", s.SocketPath)
		go func(l net.Listener) {
			defer wg.Done()
			if err := domainSocket.Serve(l); err != nil {
				log.Fatalf("%v", err)
			}
			log.Printf("Stopped serving multikube at unix://%s", s.SocketPath)
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

		httpServer.Handler = s.handler

		wg.Add(2)
		log.Printf("Serving multikube at http://%s", s.httpServerL.Addr())
		go func(l net.Listener) {
			defer wg.Done()
			if err := httpServer.Serve(l); err != nil {
				log.Printf("%v", err)
			}
			log.Printf("Stopped serving multikube at http://%s", l.Addr())
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

		httpsServer.Handler = s.handler

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
		log.Printf("Serving multikube at https://%s", s.httpsServerL.Addr())
		go func(l net.Listener) {
			defer wg.Done()
			if err := httpsServer.Serve(l); err != nil {
				log.Fatalf("%v", err)
			}
			log.Printf("Stopped serving multikube at https://%s", l.Addr())
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
