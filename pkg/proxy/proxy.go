package proxy

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/amimof/multikube/pkg/cache"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd/api"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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
}

// New creates a new Proxy instance
func New() *Proxy {
	return &Proxy{
		transports:        make(map[string]http.RoundTripper),
		OIDCUsernameClaim: "sub",
		OIDCPollInterval:  time.Second * 2,
		RS256PublicKey:    &rsa.PublicKey{},
		JWKS:              &JWKS{},
	}
}

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

	// Don't use transport cache if ttl is set to 0s
	resCache := cache.New()
	resCache.TTL = p.CacheTTL
	if p.CacheTTL.String() == "0s" {
		resCache = nil
	}

	// Create a transport that will be re-used for
	if p.transports[opts.ctx] == nil {
		// Setup TLS config
		tlsConfig, err := configureTLS(opts)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			panic(err)
		}
		p.transports[opts.ctx] = &Transport{
			TLSClientConfig: tlsConfig,
			Cache:           resCache,
		}
	}

	// Create an instance of golang reverse proxy and attach our own transport to it
	proxy := httputil.NewSingleHostReverseProxy(parseURL(opts.Server))
	proxy.Transport = p.transports[opts.ctx]

	// Add some headers to the client request
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Header.Set("Impersonate-User", opts.sub)
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", opts.AuthInfo.Token))

	proxy.ServeHTTP(w, r)

}

// GetJWKSFromURL fetches the keys of an OpenID Connect endpoint in a go routine. It polls the endpoint
// every n seconds. Returns a cancel function which can be called to stop polling and close the channel.
// The endpoint must support OpenID Connect discovery as per https://openid.net/specs/openid-connect-discovery-1_0.html
func (p *Proxy) GetJWKSFromURL() func() {

	// Make sure config has non-nil fields
	p.JWKS = &JWKS{
		Keys: []JSONWebKey{},
	}

	// Run a function in a go routine that continuously fetches from remote oidc provider
	quit := make(chan int)
	go func() {
		for {
			time.Sleep(p.OIDCPollInterval)
			select {
			case <-quit:
				close(quit)
				return
			default:
				// Make a request and fetch content of .well-known url (http://some-url/.well-known/openid-configuration)
				w, err := getWellKnown(p.OIDCIssuerURL, p.OIDCCa, p.OIDCInsecureSkipVerify)
				if err != nil {
					log.Printf("ERROR retrieving openid-configuration: %s", err)
					continue
				}
				// Get content of jwks_keys field
				j, err := getKeys(w.JwksURI, p.OIDCCa, p.OIDCInsecureSkipVerify)
				if err != nil {
					log.Printf("ERROR retrieving JWKS from provider: %s", err)
					continue
				}
				p.JWKS = j
			}
		}
	}()

	return func() {
		quit <- 1
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
