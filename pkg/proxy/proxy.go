package proxy

import (
	"bufio"
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/amimof/multikube/pkg/cache"
	"golang.org/x/net/http/httpproxy"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd/api"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

const (
	contextNotFound string = "No route: context not found"
)

// Proxy implements an HTTP handler. It has a built-in transport with in-mem cache capabilities.
type Proxy struct {
	KubeConfig             *api.Config
	OIDCIssuerURL          string
	OIDCUsernameClaim      string
	OIDCPollInterval       time.Duration
	OIDCInsecureSkipVerify bool
	OIDCCa                 *x509.Certificate
	RS256PublicKey         *rsa.PublicKey
	CacheTTL               time.Duration
	JWKS                   *JWKS
	mw                     http.Handler
	transports             map[string]http.RoundTripper
	tlsconfigs             map[string]*tls.Config
}

// New creates a new Proxy instance
func New() *Proxy {
	return &Proxy{
		transports:        make(map[string]http.RoundTripper),
		tlsconfigs:        make(map[string]*tls.Config),
		OIDCUsernameClaim: "sub",
		OIDCPollInterval:  time.Second * 2,
		RS256PublicKey:    &rsa.PublicKey{},
		JWKS:              &JWKS{},
	}
}

// NewProxyFrom creates an instance of Proxy
// func NewProxyFrom(kc *api.Config) *Proxy {
// 	p := NewProxy()
// 	p.KubeConfig = kc
// 	p.Config = &Config{
// 		OIDCIssuerURL:     "",
// 		OIDCPollInterval:  time.Second * 2,
// 		OIDCUsernameClaim: "sub",
// 		RS256PublicKey:    &rsa.PublicKey{},
// 		JWKS:              &JWKS{},
// 	}
// 	return p
// }

// Use chains all middlewares and applies a context to the request flow
func (p *Proxy) Use(mw ...Middleware) Middleware {
	return func(c *Proxy, final http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](p, last)
			}
			last.ServeHTTP(w, r)
		})
	}
}

// ServeHTTP routes the request to an apiserver. It determines, resolves an apiserver using
// data in the request itsel such as certificate data, authorization bearer tokens, http headers etc.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get a kubeconfig context
	opts := optsFromCtx(r.Context(), p.KubeConfig)
	if opts == nil {
		http.Error(w, contextNotFound, http.StatusInternalServerError)
		return
	}

	// Setup TLS config
	tlsConfig, err := configureTLS(opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		panic(err)
	}

	if p.tlsconfigs[opts.ctx] == nil {
		p.tlsconfigs[opts.ctx] = tlsConfig
	}

	// Tunnel the connection if server sends Upgrade
	if r.Header.Get("Upgrade") != "" {
		p.transports[opts.ctx].(*Transport).TLSClientConfig.NextProtos = []string{"http/1.1"}
		p.tunnel(w, r)
		return
	}

	// Build the request and execute the call to the backend apiserver
	req :=
		NewRequest(parseURL(opts.Server)).
			Method(r.Method).
			Body(r.Body).
			Path(r.URL.Path).
			Query(r.URL.RawQuery).
			Headers(r.Header)

	// Set the Impersonate header
	req.Header("Impersonate-User", opts.sub)
	req.Header("Authorization", fmt.Sprintf("Bearer %s", opts.Token))

	// Don't use transport cache if ttl is set to 0s
	resCache := cache.New()
	resCache.TTL = p.CacheTTL
	if p.CacheTTL.String() == "0s" {
		resCache = nil
	}

	// Remember the transport created by the restclient so that we can re-use the connection
	if p.transports[opts.ctx] == nil {
		p.transports[opts.ctx] = &Transport{
			TLSClientConfig: tlsConfig,
			Cache:           resCache,
		}
	}

	// Assign our transport to the request
	req.Transport = p.transports[opts.ctx]

	// Execute!
	res, err := req.Do()

	// Catch any unexpected errors
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Body.Close()

	// Copy all response headers
	copyHeader(w.Header(), res.Header)
	w.WriteHeader(res.StatusCode)

	// Read body into buffer before writing to response and wait until client cancels
	buf := make([]byte, 4096)
	for {
		n, err := res.Body.Read(buf)
		if n == 0 && err != nil {
			break
		}
		b := buf[:n]
		_, err = w.Write(b)
		if err != nil {
			break
		}
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

}

// tunnel makes use of the two methods directTunnel() and proxiedTunnel(). It chooses one of these
// depending on the http proxy environment variables configured in the runtime environment.
func (p *Proxy) tunnel(w http.ResponseWriter, r *http.Request) {
	if httpproxy.FromEnvironment().HTTPSProxy != "" {
		p.proxiedTunnel(w, r)
	} else {
		p.directTunnel(w, r)
	}
}

// proxiedTunnel establishes a connection to an http proxy, configured in environment variables,
// and starts streaming data between a client and the backend server through the http proxy.
// Makes use of http/1.1 CONNECT and reads the HTTPS_PROXY environment variable, ignoring HTTP_PROXY
// since connections from multikube to a kubernetes API will always be HTTPS.
func (p *Proxy) proxiedTunnel(w http.ResponseWriter, r *http.Request) {

	opts := optsFromCtx(r.Context(), p.KubeConfig)
	if opts == nil {
		http.Error(w, contextNotFound, http.StatusInternalServerError)
		return
	}

	u, err := url.Parse(opts.Server)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	pconn, err := net.Dial("tcp", getHostAndPort(httpproxy.FromEnvironment().HTTPSProxy))
	if err != nil {
		panic(err)
	}

	connectedReq := &http.Request{
		Method: "CONNECT",
		URL:    &url.URL{Opaque: u.String()},
		Host:   u.Host,
		Header: nil,
	}
	connectedReq.Write(pconn)

	br := bufio.NewReader(pconn)
	resp, err := http.ReadResponse(br, connectedReq)
	if err != nil {
		pconn.Close()
		panic(err)
	}

	if resp.StatusCode != 200 {
		pconn.Close()
		fmt.Println(resp.StatusCode)
		return
	}

	p.tlsconfigs[opts.ctx].InsecureSkipVerify = true

	dstConn := tls.Client(pconn, p.tlsconfigs[opts.ctx])
	err = dstConn.Handshake()
	if err != nil {
		panic(err)
	}

	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", opts.AuthInfo.Token))
	r.Header.Set("Impersonate-User", opts.sub)

	err = stream(dstConn, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		panic(err)
	}

}

// directTunnel establishes a direct connection between multikube and a the kubernetes API
// and starts streaming data between the two connections on behalf of the client. directTunnel will bypass
// http proxy environment variables. For proxied connections, use proxiedTunnel().
func (p *Proxy) directTunnel(w http.ResponseWriter, r *http.Request) {

	opts := optsFromCtx(r.Context(), p.KubeConfig)
	if opts == nil {
		http.Error(w, contextNotFound, http.StatusInternalServerError)
		return
	}

	u, err := url.Parse(opts.Server)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	dstConn, err := tls.Dial("tcp", u.Host, p.tlsconfigs[opts.ctx])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		panic(err)
	}

	err = stream(dstConn, w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		panic(err)
	}

}

// stream hijacks the client connection and copies the data from the destination connection (conn)
// to the hijacked client connection.
func stream(conn net.Conn, w http.ResponseWriter, r *http.Request) error {

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		return err
	}

	conn.Write(dump)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		return newErr("Hijacking not supported")
	}

	srcConn, _, err := hijacker.Hijack()
	if err != nil {
		return err
	}

	go transfer(conn, srcConn)
	go transfer(srcConn, conn)

	return nil
}

// transfer reads data on src and copies it to dst. Data read from src is first copied into
// a buffer before it's written to dst.
func transfer(src, dst net.Conn) {
	buff := make([]byte, 65535)

	defer src.Close()
	defer dst.Close()

	for {
		n, err := src.Read(buff)
		if err != nil {
			break
		}
		b := buff[:n]
		_, err = dst.Write(b)
		if err != nil {
			break
		}
	}

}

// configureTLS composes a TLS configuration (tls.Config) from the provided Options parameter.
// This is useful when building HTTP requests (for example with the net/http package)
// and the TLS data is configured elsewhere.
func configureTLS(options *Options) (*tls.Config, error) {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: options.InsecureSkipTLSVerify,
	}

	// Load CA from file
	if options.CertificateAuthority != "" {
		caCert, err := ioutil.ReadFile(options.CertificateAuthority)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	// Load CA from block
	if options.CertificateAuthorityData != nil {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(options.CertificateAuthorityData)
		tlsConfig.RootCAs = caCertPool
	}

	// Load certs from file
	if options.ClientCertificate != "" && options.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(options.ClientCertificate, options.ClientKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	// Load certs from block
	if options.ClientCertificateData != nil && options.ClientKeyData != nil {
		cert, err := tls.X509KeyPair(options.ClientCertificateData, options.ClientKeyData)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	return tlsConfig, nil
}

// parseURL is a helper function that tries to parse a string and return an url.URL.
// Will return nil if errors occur.
func parseURL(str string) *url.URL {
	u, err := url.Parse(str)
	if err != nil {
		return nil
	}
	return u
}

// copyHeader adds all headers from dst to src
func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// getAuthInfo returns an api.AuthInfo in map authinfos identified by it's key n
func getAuthInfo(authinfos map[string]*api.AuthInfo, n string) *api.AuthInfo {
	for k, v := range authinfos {
		if k == n {
			return v
		}
	}
	return nil
}

// getContext returns an api.Context in map contexts identified by it's key n
func getContext(contexts map[string]*api.Context, n string) *api.Context {
	for k, v := range contexts {
		if k == n {
			return v
		}
	}
	return nil
}

// getCluster returns an api.Cluster in map clusters identified by it's key n
func getCluster(clusters map[string]*api.Cluster, n string) *api.Cluster {
	for k, v := range clusters {
		if k == n {
			return v
		}
	}
	return nil
}

// getOptions returns a new pointer to an Options instance which is constructed
// from an api.Config object and name n. Returns a nil value if unable to find an Options that matches.
func getOptions(config *api.Config, n string) *Options {
	ctx := getContext(config.Contexts, n)
	if ctx == nil {
		return nil
	}
	authInfo := getAuthInfo(config.AuthInfos, ctx.AuthInfo)
	if authInfo == nil {
		return nil
	}
	cluster := getCluster(config.Clusters, ctx.Cluster)
	if cluster == nil {
		return nil
	}
	return &Options{
		cluster,
		authInfo,
		n,
		"",
	}
}

// optsFromCtx returns a pointer to an Options instance defined by context and api.Config. Returns
// a nil value if unable to find an Options matching values in context and the given config.
func optsFromCtx(ctx context.Context, config *api.Config) *Options {

	// Make sure Subject is set
	sub, ok := ctx.Value(subjectKey).(string)
	if !ok || sub == "" {
		return nil
	}

	// Make sure Context is set
	cont, ok := ctx.Value(contextKey).(string)
	if !ok || cont == "" {
		return nil
	}

	// Get a kubeconfig context
	opts := getOptions(config, cont)
	if opts == nil {
		return nil
	}

	opts.ctx = cont
	opts.sub = sub

	return opts

}

// getHostAndPort takes a string, and returns host and port in the format host:port.
// If uri is a url then the scheme is removed. For example https://amimof.com becomes amimof.com:80
func getHostAndPort(uri string) string {

	// Strip http scheme
	schemeStrip := strings.Replace(uri, "http://", "", 1)
	// Strip https scheme
	schemeStrip = strings.Replace(schemeStrip, "https://", "", 1)

	// Append default port 80 if missing
	if !strings.Contains(schemeStrip, ":") {
		schemeStrip = fmt.Sprintf("%s:%s", schemeStrip, "80")
	}

	return schemeStrip

}

func newErrf(s string, f ...interface{}) error {
	return fmt.Errorf(s, f...)
}

func newErr(s string) error {
	return errors.New(s)
}
