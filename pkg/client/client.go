// Package client provides a client interface to interact with server APIs
package client

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/spf13/pflag"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"github.com/amimof/multikube/pkg/client/identity"
	"github.com/amimof/multikube/pkg/logger"

	backendv1 "github.com/amimof/multikube/pkg/client/backend/v1"
	cav1 "github.com/amimof/multikube/pkg/client/ca/v1"
	certificatev1 "github.com/amimof/multikube/pkg/client/certificate/v1"
	credentialv1 "github.com/amimof/multikube/pkg/client/credential/v1"
	healthv1 "github.com/amimof/multikube/pkg/client/health/v1"
	policyv1 "github.com/amimof/multikube/pkg/client/policy/v1"
	routev1 "github.com/amimof/multikube/pkg/client/route/v1"
	tokenv1 "github.com/amimof/multikube/pkg/client/token/v1"
)

var DefaultTLSConfig = &tls.Config{
	InsecureSkipVerify: false,
}

type NewClientOption func(c *ClientSet) error

func WithRouteClient(routev1Client routev1.ClientV1) NewClientOption {
	return func(c *ClientSet) error {
		c.routeV1Client = routev1Client
		return nil
	}
}

func WithBackendClient(backendv1Client backendv1.ClientV1) NewClientOption {
	return func(c *ClientSet) error {
		c.backendV1Client = backendv1Client
		return nil
	}
}

func WithCAClient(cav1Client cav1.ClientV1) NewClientOption {
	return func(c *ClientSet) error {
		c.caV1Client = cav1Client
		return nil
	}
}

func WithCredentialClient(credentialv1Client credentialv1.ClientV1) NewClientOption {
	return func(c *ClientSet) error {
		c.credentialV1Client = credentialv1Client
		return nil
	}
}

func WithCertificateClient(certificatev1Client certificatev1.ClientV1) NewClientOption {
	return func(c *ClientSet) error {
		c.certificateV1Client = certificatev1Client
		return nil
	}
}

func WithPolicyClient(policyv1Client policyv1.ClientV1) NewClientOption {
	return func(c *ClientSet) error {
		c.policyV1Client = policyv1Client
		return nil
	}
}

func WithIdentity(id *identity.AtomicIdentity) NewClientOption {
	return func(c *ClientSet) error {
		c.id = id
		return nil
	}
}

func WithGrpcDialOption(opts ...grpc.DialOption) NewClientOption {
	return func(c *ClientSet) error {
		c.grpcOpts = opts
		return nil
	}
}

func WithTLSConfig(t *tls.Config) NewClientOption {
	return func(c *ClientSet) error {
		c.tlsConfig = t
		return nil
	}
}

func WithLogger(l logger.Logger) NewClientOption {
	return func(c *ClientSet) error {
		c.logger = l
		return nil
	}
}

func WithTLSConfigFromFlags(f *pflag.FlagSet) NewClientOption {
	insecure, _ := f.GetBool("insecure")
	tlsCertificate, _ := f.GetString("tls-certificate")
	tlsCertificateKey, _ := f.GetString("tls-certificate-key")
	tlsCaCertificate, _ := f.GetString("tls-ca-certificate")
	return func(c *ClientSet) error {
		tlsConfig, err := getTLSConfig(tlsCertificate, tlsCertificateKey, tlsCaCertificate, insecure)
		if err != nil {
			return err
		}
		c.tlsConfig = tlsConfig
		return nil
	}
}

// WithTLSConfigFromCfg returns a NewClientOption using the provided client.Config.
// It runs Validate() on the config before returning. If passed in client config doesn't have
// tls configuration, then tls config is not set on the client.
func WithTLSConfigFromCfg(cfg *Config) NewClientOption {
	return func(c *ClientSet) error {
		if err := cfg.Validate(); err != nil {
			return err
		}

		current, err := cfg.CurrentServer()
		if err != nil {
			return err
		}

		if current.TLSConfig != nil {

			insecure := current.TLSConfig.Insecure
			tlsCertificate := current.TLSConfig.Certificate
			tlsCertificateKey := current.TLSConfig.Key
			tlsCaCertificate := current.TLSConfig.CA

			tlsConfig, err := getTLSConfig(tlsCertificate, tlsCertificateKey, tlsCaCertificate, insecure)
			if err != nil {
				return err
			}
			c.tlsConfig = tlsConfig
		}
		return nil
	}
}

func getTLSConfig(cert, key, ca string, insecure bool) (*tls.Config, error) {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: insecure,
	}

	if ca != "" {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM([]byte(ca)) {
			return nil, fmt.Errorf("error appending CA certitifacte to pool")
		}
		tlsConfig.RootCAs = certPool
	}

	// Add certificate pair to tls config
	if cert != "" && key != "" {
		certificate, err := tls.X509KeyPair([]byte(cert), []byte(key))
		if err != nil {
			return nil, fmt.Errorf("error loading x509 cert key pair: %v", err)
		}
		tlsConfig.Certificates = []tls.Certificate{certificate}
	}
	return tlsConfig, nil
}

type ClientSet struct {
	conn                *grpc.ClientConn
	healthV1Client      *healthv1.ClientV1
	backendV1Client     backendv1.ClientV1
	caV1Client          cav1.ClientV1
	certificateV1Client certificatev1.ClientV1
	credentialV1Client  credentialv1.ClientV1
	routeV1Client       routev1.ClientV1
	policyV1Client      policyv1.ClientV1
	tokenV1Client       tokenv1.ClientV1
	mu                  sync.Mutex
	grpcOpts            []grpc.DialOption
	tlsConfig           *tls.Config
	logger              logger.Logger
	id                  *identity.AtomicIdentity
}

func (c *ClientSet) BackendV1() backendv1.ClientV1 {
	return c.backendV1Client
}

func (c *ClientSet) CAV1() cav1.ClientV1 {
	return c.caV1Client
}

func (c *ClientSet) CertificateV1() certificatev1.ClientV1 {
	return c.certificateV1Client
}

func (c *ClientSet) CredentialV1() credentialv1.ClientV1 {
	return c.credentialV1Client
}

func (c *ClientSet) RouteV1() routev1.ClientV1 {
	return c.routeV1Client
}

func (c *ClientSet) PolicyV1() policyv1.ClientV1 {
	return c.policyV1Client
}

func (c *ClientSet) TokenV1() tokenv1.ClientV1 {
	return c.tokenV1Client
}

func (c *ClientSet) State() connectivity.State {
	return c.conn.GetState()
}

func (c *ClientSet) HealthV1() *healthv1.ClientV1 {
	return c.healthV1Client
}

func (c *ClientSet) Connect() {
	c.conn.Connect()
}

func (c *ClientSet) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	defer func() {
		if err := c.conn.Close(); err != nil {
			c.logger.Error("error closing connection", "error", err)
		}
	}()
	return nil
}

func New(server string, opts ...NewClientOption) (*ClientSet, error) {
	// Define connection backoff policy
	backoffConfig := backoff.Config{
		BaseDelay:  time.Second,       // Initial delay before retry
		Multiplier: 1.6,               // Multiplier for successive retries
		MaxDelay:   120 * time.Second, // Maximum delay
	}

	// Define keepalive parameters
	keepAliveParams := keepalive.ClientParameters{
		Time:                15 * time.Second, // Ping the server if no activity
		Timeout:             10 * time.Second, // Timeout for server response
		PermitWithoutStream: true,             // Ping even without active streams
	}

	// Default options
	defaultOpts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepAliveParams),
		grpc.WithConnectParams(
			grpc.ConnectParams{
				Backoff:           backoffConfig,
				MinConnectTimeout: 20 * time.Second,
			},
		),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	}

	id := &identity.AtomicIdentity{}
	id.Set(identity.ClientIdentity{Name: uuid.New().String(), UID: uuid.New().String()})

	// Default clientset
	c := &ClientSet{
		grpcOpts: defaultOpts,
		tlsConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
		logger: logger.ConsoleLogger{},
		id:     id,
	}

	// Allow passing in custom dial options
	for _, opt := range opts {
		err := opt(c)
		if err != nil {
			return nil, err
		}
	}

	// We always want TLS but TLSConfig might be changed by the user so that's why do this here
	c.grpcOpts = append(c.grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(c.tlsConfig)))

	// Add interceptors
	c.grpcOpts = append(c.grpcOpts, grpc.WithUnaryInterceptor(identity.IdentityUnaryInterceptor(c.id)), grpc.WithStreamInterceptor(identity.IdentityStreamInterceptor(c.id)))

	conn, err := grpc.NewClient(server, c.grpcOpts...)
	if err != nil {
		return nil, err
	}

	c.conn = conn
	if c.backendV1Client == nil {
		c.backendV1Client = backendv1.NewClientV1WithConn(conn)
	}
	if c.caV1Client == nil {
		c.caV1Client = cav1.NewClientV1WithConn(conn)
	}
	if c.certificateV1Client == nil {
		c.certificateV1Client = certificatev1.NewClientV1WithConn(conn)
	}
	if c.credentialV1Client == nil {
		c.credentialV1Client = credentialv1.NewClientV1WithConn(conn)
	}
	if c.routeV1Client == nil {
		c.routeV1Client = routev1.NewClientV1WithConn(conn)
	}
	if c.policyV1Client == nil {
		c.policyV1Client = policyv1.NewClientV1WithConn(conn)
	}
	if c.tokenV1Client == nil {
		c.tokenV1Client = tokenv1.NewClientV1WithConn(conn)
	}
	if c.healthV1Client == nil {
		c.healthV1Client = healthv1.NewClientV1WithConn(conn)
	}

	return c, nil
}
