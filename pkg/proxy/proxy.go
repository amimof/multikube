package proxy

import (
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
	"strings"
	"time"
)

// MiddlewareFunc defines a function to process middleware.
type MiddlewareFunc func(http.Handler) http.Handler

type ctxKey string

var (
	contextKey = ctxKey("Context")
	subjectKey = ctxKey("Subject")
)

// Proxy implements an HTTP handler. It has a built-in transport with in-mem cache capabilities.
type Proxy struct {
	kubeConfig *api.Config
	transports map[string]http.RoundTripper
	middleware []MiddlewareFunc
}

// New creates a new Proxy instance
func New(c *api.Config) (*Proxy, error) {

	var transports = make(map[string]http.RoundTripper)

	for ctxKey := range c.Contexts {
		cluster := getClusterByContextName(c, ctxKey)
		auth := getAuthByContextName(c, ctxKey)
		tlsConfig, err := configureTLS(auth, cluster)
		if err != nil {
			return nil, err
		}
		transports[ctxKey] = &Transport{
			TLSClientConfig: tlsConfig,
			Cache:           cache.New(),
		}

	}

	return &Proxy{
		kubeConfig: c,
		transports: transports,
	}, nil
}

// WithHandler takes any http.Handler and returns it as a MiddlewareFunc so that it can be used in proxy
func WithHandler(next http.Handler) MiddlewareFunc {
	return func(inner http.Handler) http.Handler {
		return next
	}
}

// Apply chains all middlewares and resturns a MiddlewareFunc that can wrap an http.Handler
func (p *Proxy) Apply(middleware ...MiddlewareFunc) MiddlewareFunc {
	return func(final http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(p.middleware) - 1; i >= 0; i-- {
				last = p.middleware[i](last)
			}
			last.ServeHTTP(w, r)
		})
	}
}

// Use adds a middleware
func (p *Proxy) Use(middleware ...MiddlewareFunc) *Proxy {
	p.middleware = append(p.middleware, middleware...)
	return p
}

// Chain is a convenience function that chains all applied middleware and wraps proxy handler with it
func (p *Proxy) Chain() http.Handler {
	h := p.Apply(p.middleware...)
	return h(p)
}

// CacheTTL sets the TTL value of all transports to d
func (p *Proxy) CacheTTL(d time.Duration) {
	for key := range p.transports {
		if p.transports[key].(*Transport).Cache != nil {
			log.Printf("%s", key)
			p.transports[key].(*Transport).Cache.TTL = d
			cacheTTL.WithLabelValues(key).Set(d.Seconds())
		}
	}
}

// ServeHTTP routes the request to an apiserver. It determines, resolves an apiserver using
// data in the request itsel such as certificate data, authorization bearer tokens, http headers etc.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get the k8s context from the request
	ctx := ParseContextFromRequest(r, true)
	if ctx == "" {
		http.Error(w, "Not route: context not found", http.StatusBadGateway)
		return
	}

	// Get the subject from the request
	sub := ParseSubjectFromRequest(r)

	// // Get k8s cluster and authinfo from kubeconfig using the ctx name
	cluster := getClusterByContextName(p.kubeConfig, ctx)
	if cluster == nil {
		http.Error(w, fmt.Sprintf("no route: cluster not found for '%s'", ctx), http.StatusBadGateway)
		return
	}

	// Create an instance of golang reverse proxy and attach our own transport to it
	proxy := httputil.NewSingleHostReverseProxy(parseURL(cluster.Server))
	proxy.Transport = p.transports[ctx]

	// Add some headers to the client request
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Header.Set("Impersonate-User", sub)

	proxy.ServeHTTP(w, r)

}

// configureTLS composes a TLS configuration (tls.Config) from the provided Options parameter.
// This is useful when building HTTP requests (for example with the net/http package)
// and the TLS data is configured elsewhere.
func configureTLS(a *api.AuthInfo, c *api.Cluster) (*tls.Config, error) {

	tlsConfig := &tls.Config{
		InsecureSkipVerify: c.InsecureSkipTLSVerify,
	}

	// Load CA from file
	if c.CertificateAuthority != "" {
		caCert, err := ioutil.ReadFile(c.CertificateAuthority)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	// Load CA from block
	if c.CertificateAuthorityData != nil {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(c.CertificateAuthorityData)
		tlsConfig.RootCAs = caCertPool
	}

	// Load certs from file
	if a.ClientCertificate != "" && a.ClientKey != "" {
		cert, err := tls.LoadX509KeyPair(a.ClientCertificate, a.ClientKey)
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	// Load certs from block
	if a.ClientCertificateData != nil && a.ClientKeyData != nil {
		cert, err := tls.X509KeyPair(a.ClientCertificateData, a.ClientKeyData)
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

// ParseContextFromRequest tries to find the requested context name either by URL or HTTP header.
// Will return the value of 'Multikube-Context' HTTP header. Will return the first part
// of the URL path if no headers are set. Set replace to true if the URL path in provided request
// should be replaced with a path without context name.
func ParseContextFromRequest(req *http.Request, replace bool) string {
	val := req.Header.Get("Multikube-Context")
	if val != "" {
		return val
	}

	c, rem := getCtxFromURL(req.URL)
	if c != "" {
		val = c
	}

	if rem != "" && replace {
		req.URL.Path = rem
	}
	return val
}

// ParseSubjectFromRequest returns a string with the value of ContextKey key from
// the HTTP request Context (context.Context)
func ParseSubjectFromRequest(req *http.Request) string {
	if sub, ok := req.Context().Value(subjectKey).(string); ok {
		return sub
	}
	return ""
}

// getClusterByContextName returns an api.Cluster from the kubeconfig using context name.
// Returns a new empty Cluster object with non-nil maps if no cluster found in the kubeconfig.
func getClusterByContextName(kubeconfig *api.Config, n string) *api.Cluster {
	if ctx, ok1 := kubeconfig.Contexts[n]; ok1 {
		if clu, ok2 := kubeconfig.Clusters[ctx.Cluster]; ok2 {
			return clu
		}
	}
	return nil
}

// getAuthByContextName returns an api.AuthInfo from the kubeconfig using context name.
// Returns a new empty AuthInfo object with non-nil maps if no cluster found in the kubeconfig.
func getAuthByContextName(kubeconfig *api.Config, n string) *api.AuthInfo {
	if ctx, ok1 := kubeconfig.Contexts[n]; ok1 {
		if auth, ok2 := kubeconfig.AuthInfos[ctx.AuthInfo]; ok2 {
			return auth
		}
	}
	return nil
}

// getCtxFromURL reads path params from u and returns the kubeconfig context
// as well as the path params used for upstream communication
func getCtxFromURL(u *url.URL) (string, string) {
	val := ""
	rem := []string{}
	if vals := strings.Split(u.Path, "/"); len(vals) > 1 {
		val = vals[1]
		rem = vals[2:]
	}
	return val, fmt.Sprintf("/%s", strings.Join(rem, "/"))
}
