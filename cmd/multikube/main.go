package main

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/spf13/pflag"
	"gitlab.com/amimof/multikube"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	"log"
	"os"
	"time"
)

var (
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

	tlsHost              string
	tlsPort              int
	tlsListenLimit       int
	tlsKeepAlive         time.Duration
	tlsReadTimeout       time.Duration
	tlsWriteTimeout      time.Duration
	tlsCertificate       string
	tlsCertificateKey    string
	tlsCACertificate     string
	tlsSignerCertificate string

	kubeconfigPath string
)

func init() {
	pflag.StringVar(&socketPath, "socket-path", "/var/run/multikube.sock", "the unix socket to listen on")
	pflag.StringVar(&host, "host", "localhost", "the IP to listen on")
	pflag.StringVar(&tlsHost, "tls-host", "localhost", "the IP to listen on")
	pflag.StringVar(&tlsCertificate, "tls-certificate", "", "the certificate to use for secure connections")
	pflag.StringVar(&tlsCertificateKey, "tls-key", "", "the private key to use for secure conections")
	pflag.StringVar(&tlsCACertificate, "tls-ca", "", "the certificate authority file to be used with mutual tls auth")
	pflag.StringVar(&tlsSignerCertificate, "tls-signer-certificate", "", "the certificate to use when verifying client certificates and JWT token signature")
	pflag.StringVar(&kubeconfigPath, "kubeconfig", "~/.kube/config", "absolute path to a kubeconfig file")
	pflag.StringSliceVar(&enabledListeners, "scheme", []string{}, "the listeners to enable, this can be repeated and defaults to the schemes in the swagger spec")

	pflag.IntVar(&port, "port", 443, "the port to listen on for insecure connections, defaults to 443")
	pflag.IntVar(&listenLimit, "listen-limit", 0, "limit the number of outstanding requests")
	pflag.IntVar(&tlsPort, "tls-port", 0, "the port to listen on for secure connections, defaults to a random value")
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

	pflag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage:\n")
		fmt.Fprint(os.Stderr, "  multikube-server [OPTIONS]\n\n")

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

	// Read provided kubeconfig file
	c, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	// Create the proxy
	p := multikube.NewProxyFrom(c)

	// Read provided signer cert file
	if tlsSignerCertificate != "" {
		signer, err := ioutil.ReadFile(tlsSignerCertificate)
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

	// Listen and serve!
	err = s.Serve()
	if err != nil {
		log.Fatal(err)
	}

}
