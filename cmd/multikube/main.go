package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"buf.build/go/protovalidate"
	"github.com/amimof/multikube/pkg/audit"
	"github.com/amimof/multikube/pkg/client"
	"github.com/amimof/multikube/pkg/compile"
	"github.com/amimof/multikube/pkg/controller"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/proxy"
	proxyv2 "github.com/amimof/multikube/pkg/proxyv2"
	"github.com/amimof/multikube/pkg/repository"
	"github.com/amimof/multikube/pkg/server"
	"github.com/dgraph-io/badger/v4"
	"github.com/golang-jwt/jwt"
	protovalidate_middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/protovalidate"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/amimof/multikube/internal/app"
	transport "github.com/amimof/multikube/internal/transport/grpc"
	badgerrepo "github.com/amimof/multikube/pkg/repository/badger"
)

var (
	// VERSION of the app. Is set when project is built and should never be set manually
	VERSION string
	// COMMIT is the Git commit currently used when compiling. Is set when project is built and should never be set manually
	COMMIT string
	// BRANCH is the Git branch currently used when compiling. Is set when project is built and should never be set manually
	BRANCH string
	// GOVERSION used to compile. Is set when project is built and should never be set manually
	GOVERSION string

	enabledListeners []string
	cleanupTimeout   time.Duration
	maxHeaderSize    uint64

	socketPath string

	serverAddress  string
	metricsAddress string
	proxyAddress   string

	listenLimit  int
	keepAlive    time.Duration
	readTimeout  time.Duration
	writeTimeout time.Duration

	oidcPollInterval       time.Duration
	oidcIssuerURL          string
	oidcUsernameClaim      string
	oidcCaFile             string
	oidcInsecureSkipVerify bool
	tlsListenLimit         int
	tlsKeepAlive           time.Duration
	tlsReadTimeout         time.Duration
	tlsWriteTimeout        time.Duration
	tlsCertificate         string
	tlsCertificateKey      string
	tlsCACertificate       string

	rs256PrivateKey string
	rs256PublicKey  string
	kubeconfigPath  string
	cacheTTL        time.Duration
	dataPath        string
	logLevel        string

	log *slog.Logger
)

func parseSlogLevel(lvl string) (slog.Level, error) {
	switch strings.ToLower(lvl) {
	case "error":
		return slog.LevelError, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "info":
		return slog.LevelInfo, nil
	case "debug":
		return slog.LevelDebug, nil
	}

	var l slog.Level
	return l, fmt.Errorf("not a valid log level: %q", lvl)
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Errorf("could not determine home directory: %w", err))
	}
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		panic(fmt.Errorf("could not determine cache directory: %w", err))
	}

	defaultStatePath := filepath.Join(home, ".local", "state", "multiUserCacheDirkube")
	defaultSocketPath := filepath.Join(cacheDir, "multikube.sock")

	pflag.StringVar(&socketPath, "socket-path", defaultSocketPath, "the unix socket to listen on")
	pflag.StringVar(&serverAddress, "server-address", "0.0.0.0:5743", "Address to listen the TCP server on")
	pflag.StringVar(&metricsAddress, "metrics-address", "0.0.0.0:8888", "Address to listen the metrics server on")
	pflag.StringVar(&proxyAddress, "proxy-address", "0.0.0.0:8443", "Address to listen the http proxy server on")
	pflag.StringVar(&tlsCertificate, "tls-certificate", "", "the certificate to use for secure connections")
	pflag.StringVar(&tlsCertificateKey, "tls-key", "", "the private key to use for secure conections")
	pflag.StringVar(&tlsCACertificate, "tls-ca", "", "the certificate authority file to be used with mutual tls auth")
	pflag.StringVar(&rs256PrivateKey, "rs256-private-key", "", "the RS256 private key used to sign JWT's")
	pflag.StringVar(&rs256PublicKey, "rs256-public-key", "", "the RS256 public key used to validate the signature of client JWT's")
	pflag.StringVar(&kubeconfigPath, "kubeconfig", "/etc/multikube/kubeconfig", "absolute path to a kubeconfig file")
	pflag.StringVar(&oidcIssuerURL, "oidc-issuer-url", "", "The URL of the OpenID issuer, only HTTPS scheme will be accepted. If set, it will be used to verify the OIDC JSON Web Token (JWT)")
	pflag.StringVar(&oidcUsernameClaim, "oidc-username-claim", "sub", " The OpenID claim to use as the user name. Note that claims other than the default is not guaranteed to be unique and immutable")
	pflag.StringVar(&oidcCaFile, "oidc-ca-file", "", "the certificate authority file to be used for verifyign the OpenID server")
	pflag.StringVar(&dataPath, "data-path", defaultStatePath, "Directory to store state")
	pflag.StringVar(&logLevel, "log-level", "info", "The level of verbosity of log output")
	pflag.StringSliceVar(&enabledListeners, "scheme", []string{"https"}, "the listeners to enable, this can be repeated and defaults to the schemes in the swagger spec")

	pflag.IntVar(&listenLimit, "listen-limit", 0, "limit the number of outstanding requests")
	pflag.IntVar(&tlsListenLimit, "tls-listen-limit", 0, "limit the number of outstanding requests")
	pflag.Uint64Var(&maxHeaderSize, "max-header-size", 1000000, "controls the maximum number of bytes the server will read parsing the request header's keys and values, including the request line. It does not limit the size of the request body")

	pflag.DurationVar(&cleanupTimeout, "cleanup-timeout", 10*time.Second, "grace period for which to wait before shutting down the server")
	pflag.DurationVar(&keepAlive, "keep-alive", 3*time.Minute, "sets the TCP keep-alive timeouts on accepted connections. It prunes dead TCP connections ( e.g. closing laptop mid-download)")
	pflag.DurationVar(&readTimeout, "read-timeout", 30*time.Second, "maximum duration before timing out read of the request")
	pflag.DurationVar(&writeTimeout, "write-timeout", 30*time.Second, "maximum duration before timing out write of the response")
	pflag.DurationVar(&tlsKeepAlive, "tls-keep-alive", 3*time.Minute, "sets the TCP keep-alive timeouts on accepted connections. It prunes dead TCP connections ( e.g. closing laptop mid-download)")
	pflag.DurationVar(&tlsReadTimeout, "tls-read-timeout", 30*time.Second, "maximum duration before timing out read of the request")
	pflag.DurationVar(&tlsWriteTimeout, "tls-write-timeout", 30*time.Second, "maximum duration before timing out write of the response")
	pflag.DurationVar(&oidcPollInterval, "oidc-poll-interval", 2*time.Second, "maximum duration between intervals in which the oidc issuer url (--oidc-issuer-url) is polled")
	pflag.DurationVar(&cacheTTL, "cache-ttl", 1*time.Second, "maximum duration before cached responses are invalidated. Set this value to 0s to disable the cache")

	pflag.BoolVar(&oidcInsecureSkipVerify, "oidc-insecure-skip-verify", false, "")

	// Create build_info metrics
	if err := prometheus.Register(prometheus.NewGaugeFunc(
		prometheus.GaugeOpts{
			Name: "multikube_build_info",
			Help: "A constant gauge with build info labels.",
			ConstLabels: prometheus.Labels{
				"branch":    BRANCH,
				"goversion": GOVERSION,
				"commit":    COMMIT,
				"version":   VERSION,
			},
		},
		func() float64 { return 1 },
	)); err != nil {
		log.Info("Unable to register 'multikube_build_info metric'", "error", err.Error())
	}
}

func main() {
	showver := pflag.Bool("version", false, "Print version")

	pflag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage:\n")
		fmt.Fprint(os.Stderr, "  multikube [OPTIONS]\n\n")

		title := "Kubernetes multi-cluster manager"
		fmt.Fprint(os.Stderr, title+"\n\n")
		desc := "Manages multiple Kubernetes clusters and provides a single API to clients"
		if desc != "" {
			fmt.Fprintf(os.Stderr, "%s\n\n", desc)
		}
		fmt.Fprintln(os.Stderr, pflag.CommandLine.FlagUsages())
	}

	// parse the CLI flags
	pflag.Parse()

	// Show version if requested
	if *showver {
		fmt.Printf("Version: %s\nCommit: %s\nBranch: %s\nGoVersion: %s\n", VERSION, COMMIT, BRANCH, GOVERSION)
		return
	}

	// Setup logging
	lvl, err := parseSlogLevel(logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing log level: %v", err)
		os.Exit(1)
	}
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: lvl, AddSource: true}))

	// Setup signal handlers
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)

	// Setup badgerdb and repo
	db, err := badger.Open(badger.DefaultOptions(path.Join(dataPath, "db")))
	if err != nil {
		log.Error("error opening badger database", "error", err)
		os.Exit(1)

	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("error closing database", "error", err)
		}
	}()

	// Setup repo
	repo := badgerrepo.New(db)

	// Setup event exchange bus
	exchange := events.NewExchange(events.WithExchangeLogger(log))

	// Setup RS256 key pairs
	rs256PrivKey, err := generateSigningKey()
	if err != nil {
		log.Error("error parsing rs256 key", "error", err)
		os.Exit(1)
	}

	// Setup grpc services
	backendService := transport.NewBackendService(&app.BackendService{
		Repo:     repository.NewBackendRepo(repo),
		Exchange: exchange,
		Logger:   log,
	})
	caService := transport.NewCertificateAuthorityService(&app.CertificateAuthorityService{
		Repo:     repository.NewCertificateAuthorityRepo(repo),
		Exchange: exchange,
		Logger:   log,
	})
	certService := transport.NewCertificateService(&app.CertificateService{
		Repo:     repository.NewCertificateRepo(repo),
		Exchange: exchange,
		Logger:   log,
	})
	routeService := transport.NewRouteService(&app.RouteService{
		Repo:     repository.NewRouteRepo(repo),
		Exchange: exchange,
		Logger:   log,
	})
	policyService := transport.NewPolicyService(&app.PolicyService{
		Repo:     repository.NewPolicyRepo(repo),
		Exchange: exchange,
		Logger:   log,
	})
	credentialService := transport.NewCredentialService(&app.CredentialService{
		Repo:     repository.NewCredentialRepo(repo),
		Exchange: exchange,
		Logger:   log,
	})
	tokenService := transport.NewTokenService(&app.TokenService{
		Exchange:     exchange,
		Logger:       log,
		Issuer:       "https://auth.multikube.io",
		Key:          rs256PrivKey,
		DefaultTTL:   time.Hour * 24,
		MaxTTL:       time.Hour * 72,
		DefaultAud:   []string{"multikube"},
		SigningKeyID: "key-2026-04",
	})

	validator, err := protovalidate.New()
	if err != nil {
		log.Error("Failed to create protovalidate validator", "error", err)
		os.Exit(1)
	}

	var serverOpts []transport.NewServerOption

	// Load in certificates either from flags or auto-generated
	cert, err := generateCertificates()
	if err != nil {
		log.Error("error loading x509 certificates", "error", err)
		os.Exit(1)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Enable mTLS for gRPC server, if CA cert provided
	if tlsCACertificate != "" {
		caCert, err := os.ReadFile(tlsCACertificate)
		if err != nil {
			log.Error("error reading CA certificate file", "error", err)
			os.Exit(1)
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			log.Error("error appending CA certificate to pool", "caCert", tlsCACertificate)
			os.Exit(1)
		}

		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert, // ← mTLS
			ClientCAs:    certPool,
		}
		log.Info("mutual TLS enabled for gRPC server")
	}

	creds := credentials.NewTLS(tlsConfig)
	serverOpts = append(serverOpts,
		transport.WithGrpcOption(grpc.Creds(creds),
			grpc.StatsHandler(otelgrpc.NewServerHandler()),
		),
	)

	serverOpts = append(serverOpts,
		// transport.WithGrpcOption(metricsOpts),
		transport.WithGrpcOption(grpc.UnaryInterceptor(protovalidate_middleware.UnaryServerInterceptor(validator))),
		transport.WithExchange(exchange),
		transport.WithLogger(log),
		transport.WithDB(repo),
	)
	errChan := make(chan error)

	// Setup server
	srv, err := transport.NewServer(serverOpts...)
	if err != nil {
		log.Error("error setting up gRPC server", "error", err)
		os.Exit(1)
	}

	// Register services to gRPC server
	srv.RegisterService(
		backendService,
		caService,
		certService,
		routeService,
		policyService,
		credentialService,
		tokenService,
	)

	// Context
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Only allow one of the flags rs256-public-key and oidc-issuer-url
	// TODO: Remove these from flags. They are now in the API instead
	if rs256PublicKey != "" && oidcIssuerURL != "" {
		log.Error("Only one of `--rs256-public-key` or `--oidc-issue-url` cat be set")
		os.Exit(1)
	}

	go serveTCP(serverAddress, srv, errChan)
	go serveUnix(srv, errChan)

	// Used by clientset and the gateway to connect internally
	socketAddr := fmt.Sprintf("unix:///%s", socketPath)

	// Setup a clientset for the controllers
	cs, err := client.New(socketAddr, client.WithLogger(log), client.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}), client.WithGrpcDialOption(grpc.WithAuthority("localhost")))
	if err != nil {
		log.Error("error creating clientset", "error", err.Error())
	}

	defer func() {
		if err := cs.Close(); err != nil {
			log.Error("error closing clientset connection", "error", err)
		}
	}()

	// Create the proxy
	p, err := proxy.New(nil)
	if err != nil {
		log.Error("error setting up proxy", "error", err)
		os.Exit(1)
	}
	p.CacheTTL(cacheTTL)

	p.Use(
		proxy.WithEmpty(),
		proxy.WithLogging(),
		proxy.WithJWT(),
		proxy.WithHeader(),
	)

	// Add JWK validation middleware if issuer url is provided on cmd line
	if oidcIssuerURL != "" {
		oidcConfig := proxy.OIDCConfig{
			OIDCIssuerURL:          oidcIssuerURL,
			OIDCPollInterval:       oidcPollInterval,
			OIDCUsernameClaim:      oidcUsernameClaim,
			OIDCInsecureSkipVerify: oidcInsecureSkipVerify,
			OIDCCa:                 readCert(oidcCaFile),
		}
		// middlewares = append(middlewares, proxy.WithOIDC(oidcConfig))
		p.Use(proxy.WithOIDC(oidcConfig))
	}

	// Setup controller
	runtimeStore := proxyv2.NewRuntimeStore()
	compiler := compile.NewCompiler()
	ctrl := controller.New(
		cs,
		controller.WithLogger(log),
		controller.WithExchange(exchange),
		controller.WithCompiler(compiler),
		controller.WithRuntime(runtimeStore),
	)
	go ctrl.Run(ctx)
	log.Info("started proxy Controller")

	// Setup Audit logging
	fileSink, err := audit.NewFileSink(filepath.Join(dataPath, "audit.json"))
	if err != nil {
		log.Error("error setting up file sink for audit logging", "error", err)
		os.Exit(1)
	}
	publisher := audit.NewAsyncPublisher(fileSink)
	publisher.Start()

	prox := proxyv2.NewProxy(
		runtimeStore,
		proxyv2.WithPublicKey(&rs256PrivKey.PublicKey),
		proxyv2.WithPublisher(publisher),
	)

	handler := proxyv2.AuditMiddleware(publisher)(prox)

	// Create the server
	s := &server.Server{
		EnabledListeners: enabledListeners,
		CleanupTimeout:   cleanupTimeout,
		MaxHeaderSize:    maxHeaderSize,
		SocketPath:       socketPath,
		ListenLimit:      listenLimit,
		KeepAlive:        keepAlive,
		ReadTimeout:      readTimeout,
		WriteTimeout:     writeTimeout,
		TLSAddress:       proxyAddress,
		TLSConfig:        tlsConfig,
		TLSListenLimit:   tlsListenLimit,
		TLSKeepAlive:     tlsKeepAlive,
		TLSReadTimeout:   tlsReadTimeout,
		TLSWriteTimeout:  tlsWriteTimeout,
		Logger:           log,
		// Handler:          p.Chain(),
		Handler: handler,
	}

	// Listen and serve!
	go serveProxyServer(s, errChan)

	// Metrics server
	ms := server.NewServer()
	ms.Address = metricsAddress
	ms.Name = "metrics"
	ms.Handler = promhttp.Handler()
	ms.Logger = log
	go func() {
		if err := ms.Serve(); err != nil {
			log.Error("error setting up metrics server", "error", err)
			os.Exit(1)
		}
	}()

	// Setup opentracing
	cfg := config.Configuration{
		ServiceName: "multikube",
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}

	//nolint:all
	tracer, closer, err := cfg.New("multikube", config.Logger(jaeger.StdLogger))
	if err != nil {
		log.Warn("error setting up tracer", "error", err)
	}

	opentracing.SetGlobalTracer(tracer)
	defer func() { _ = closer.Close() }()
	select {
	case <-exit:
		log.Info("received shutdown signal")
		cancel()
	case e := <-errChan:
		log.Error("fatal error in server component", "error", e)
		cancel()
	case <-ctx.Done():
		log.Info("context cancelled externally")
	}

	// Shut down server with force shutdown as fallback
	serverShutdownDone := make(chan struct{})
	go func() {
		srv.Shutdown()
		close(serverShutdownDone)
	}()

	select {
	case <-serverShutdownDone:
		log.Info("server shut down gracefully")
	case <-time.After(10 * time.Second):
		log.Warn("timeout exceeded, forcing shutdown")
		srv.ForceShutdown()
	}

	close(exit)
	close(errChan)
}

// Reads an x509 certificate from the filesystem and returns an instance of x509.Certiticate. Returns nil on errors
func readCert(p string) *x509.Certificate {
	signer, err := os.ReadFile(p)
	if err != nil {
		return nil
	}
	block, _ := pem.Decode(signer)
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil
	}
	return cert
}

// Reads a RSA public key file from the filesystem and parses it into an instance of rsa.PublicKey
// func readPublicKey(p string) *rsa.PublicKey {
// 	f, err := os.ReadFile(p)
// 	if err != nil {
// 		log.Error("error reading public keyl", "error", err)
// 		return nil
// 	}
// 	pubkey, err := crypto.ParseRSAPublicKeyFromPEM(f)
// 	if err != nil {
// 		log.Error("error parsing rsa public key from pem", "error", err)
// 		return nil
// 	}
// 	return pubkey
// }

func serveUnix(s *transport.Server, errChan chan error) {
	// Remove the socket file if it already exists
	if _, err := os.Stat(socketPath); err == nil {
		if err := os.RemoveAll(socketPath); err != nil {
			errChan <- fmt.Errorf("failed to remove existing Unix socket: %v", err)
			return
		}
	}

	// Create socket dir if doesn't exist
	dirPath := filepath.Dir(socketPath)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(dirPath, 0o755); err != nil {
			errChan <- fmt.Errorf("failed to create socket directory: %v", err)
			return
		}
	}
	unixListener, err := net.Listen("unix", socketPath)
	if err != nil {
		errChan <- fmt.Errorf("error setting up Unix socket listener: %v", err)
		return
	}

	log.Info("server listening", "socket", socketPath)
	if err := s.Serve(unixListener); err != nil {
		errChan <- fmt.Errorf("error serving server: %v", err)
		return
	}
}

func serveTCP(addr string, s *transport.Server, errChan chan error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		errChan <- fmt.Errorf("error setting up server listener: %v", err)
		return
	}
	log.Info("server listening", "address", addr)
	if err := s.Serve(l); err != nil {
		errChan <- fmt.Errorf("error serving server: %v", err)
		return
	}
}

func serveProxyServer(ps *server.Server, errChan chan error) {
	err := ps.Serve()
	if err != nil {
		errChan <- fmt.Errorf("error serving proxy: %v", err)
		return
	}
}

func generateCertificates() (tls.Certificate, error) {
	if tlsCertificate != "" && tlsCertificateKey != "" {
		return tls.LoadX509KeyPair(tlsCertificate, tlsCertificateKey)
	}

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
		DNSNames:              []string{"localhost", "multikube-node"},
		IPAddresses: []net.IP{
			net.IPv4(127, 0, 0, 1),
		},
		Subject: pkix.Name{
			CommonName:         "multikube-node",
			Country:            []string{"SE"},
			Province:           []string{"Halland"},
			Locality:           []string{"Varberg"},
			Organization:       []string{"multikube-node"},
			OrganizationalUnit: []string{"multikube"},
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

	log.Info("generated x509 key pair")

	return cert, nil
}

func generateSigningKey() (*ecdsa.PrivateKey, error) {
	b, err := os.ReadFile(rs256PrivateKey)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
			if err != nil {
				return nil, err
			}

			pemKey, err := encodePem(key)
			if err != nil {
				return nil, err
			}

			pemKeyB, err := pemKey.Bytes()
			if err != nil {
				return nil, err
			}

			err = os.WriteFile(path.Join(dataPath, "sign.pem"), pemKeyB, 0o755)
			if err != nil {
				return nil, err
			}

			_, err = generateVerifyKey(key)
			if err != nil {
				return nil, err
			}

			return pemKey, nil
		}
		return nil, err
	}

	return ecdsa.ParseRawPrivateKey(elliptic.P256(), b)
}

func generateVerifyKey(key *ecdsa.PrivateKey) (*ecdsa.PublicKey, error) {
	b, err := os.ReadFile(rs256PublicKey)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			pemPub, err := encodePubPem(&key.PublicKey)
			if err != nil {
				return nil, err
			}

			pemPubB, err := pemPub.Bytes()
			if err != nil {
				return nil, err
			}

			err = os.WriteFile(path.Join(dataPath, "verify.pem"), pemPubB, 0o755)
			if err != nil {
				return nil, err
			}

			return pemPub, nil
		}
		return nil, err
	}

	return ecdsa.ParseUncompressedPublicKey(elliptic.P256(), b)
}

func encodePem(key *ecdsa.PrivateKey) (*ecdsa.PrivateKey, error) {
	pkcs8Der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		panic(err)
	}
	pkcs8Pem := pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: pkcs8Der})
	return jwt.ParseECPrivateKeyFromPEM(pkcs8Pem)
}

func encodePubPem(key *ecdsa.PublicKey) (*ecdsa.PublicKey, error) {
	pkcs8Der, err := x509.MarshalPKIXPublicKey(key)
	if err != nil {
		panic(err)
	}
	pkcs8Pem := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pkcs8Der})
	return jwt.ParseECPublicKeyFromPEM(pkcs8Pem)
}
