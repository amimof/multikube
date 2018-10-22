package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/pflag"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/config"
	"gitlab.com/amimof/multikube"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"time"
)

var (
	VERSION   string
	COMMIT    string
	BRANCH    string
	GOVERSION string

	enabledListeners []string
	cleanupTimeout   time.Duration
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

	metricsHost				string
	metricsPort				int

	rs256PublicKey string

	kubeconfigPath string
)

func init() {
	pflag.StringVar(&socketPath, "socket-path", "/var/run/multikube.sock", "the unix socket to listen on")
	pflag.StringVar(&host, "host", "localhost", "The host address on which to listen for the --port port")
	pflag.StringVar(&tlsHost, "tls-host", "localhost", "The host address on which to listen for the --tls-port port")
	pflag.StringVar(&tlsCertificate, "tls-certificate", "", "the certificate to use for secure connections")
	pflag.StringVar(&tlsCertificateKey, "tls-key", "", "the private key to use for secure conections")
	pflag.StringVar(&tlsCACertificate, "tls-ca", "", "the certificate authority file to be used with mutual tls auth")
	pflag.StringVar(&rs256PublicKey, "rs256-public-key", "", "the RS256 public key used to validate the signature of client JWT's")
	pflag.StringVar(&kubeconfigPath, "kubeconfig", "/etc/multikube/kubeconfig", "absolute path to a kubeconfig file")
	pflag.StringVar(&metricsHost, "metrics-host", "localhost", "The host address on which to listen for the --metrics-port port")
	pflag.StringSliceVar(&enabledListeners, "scheme", []string{"https"}, "the listeners to enable, this can be repeated and defaults to the schemes in the swagger spec")

	pflag.IntVar(&port, "port", 8080, "the port to listen on for insecure connections, defaults to 8080")
	pflag.IntVar(&tlsPort, "tls-port", 8443, "the port to listen on for secure connections, defaults to 8443")
	pflag.IntVar(&metricsPort, "metrics-port", 8888, "the port to listen on for Prometheus metrics, defaults to 8888")
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
			fmt.Fprintf(os.Stderr, desc+"\n\n")
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

	// rs256-public-key is required
	if rs256PublicKey == "" {
		log.Fatalf("the required flag `--rs256-public-key` was not specified")
	}

	// Read provided kubeconfig file
	c, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create the proxy
	p := multikube.NewProxyFrom(c)

	// Read provided signer cert file
	if rs256PublicKey != "" {
		signer, err := ioutil.ReadFile(rs256PublicKey)
		if err != nil {
			log.Fatal(err)
		}
		block, _ := pem.Decode(signer)
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			log.Fatal(err)
		}
		p.CertChain = cert
	}

	// Setup middlewares
	m := p.Use(
		multikube.WithEmpty,
		multikube.WithLogging,
		multikube.WithMetrics,
		multikube.WithValidate,
	)

	// Create the server
	s := &multikube.Server{
		EnabledListeners:  enabledListeners,
		CleanupTimeout:    cleanupTimeout,
		MaxHeaderSize:     maxHeaderSize,
		SocketPath:        socketPath,
		Host:              host,
		Port:              port,
		ListenLimit:       listenLimit,
		KeepAlive:         keepAlive,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		TLSHost:           tlsHost,
		TLSPort:           tlsPort,
		TLSCertificate:    tlsCertificate,
		TLSCertificateKey: tlsCertificateKey,
		TLSCACertificate:  tlsCACertificate,
		TLSListenLimit:    tlsListenLimit,
		TLSKeepAlive:      tlsKeepAlive,
		TLSReadTimeout:    tlsReadTimeout,
		TLSWriteTimeout:   tlsWriteTimeout,
		Handler:           m(p),
	}

	// Metrics server
	ms := multikube.NewServer()
	ms.Port = metricsPort
	ms.Host = metricsHost

	// Setup opentracing
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            true,
			BufferFlushInterval: 1 * time.Second,
		},
	}
	tracer, closer, err := cfg.New("multikube", config.Logger(jaeger.StdLogger))
	opentracing.SetGlobalTracer(tracer)
	defer closer.Close()

	ms.Handler = promhttp.Handler()
	go ms.Serve()

	// Listen and serve!
	err = s.Serve()
	if err != nil {
		log.Fatal(err)
	}

}
