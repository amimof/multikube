package proxy

import (
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/amimof/multikube/pkg/cache"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd/api"
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
	KubeConfig     *api.Config
	RS256PublicKey *rsa.PublicKey
	CacheTTL       time.Duration
	mw             http.Handler
	transports     map[string]http.RoundTripper
}

// New creates a new Proxy instance
func New() *Proxy {
	return &Proxy{
		transports: make(map[string]http.RoundTripper),
		KubeConfig: api.NewConfig(),
	}
}

// Use chains all middlewares and applies a context to the request flow
func (p *Proxy) Use(mw ...MiddlewareFunc) MiddlewareFunc {
	return func(final http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			last.ServeHTTP(w, r)
		})
	}
}

// ServeHTTP routes the request to an apiserver. It determines, resolves an apiserver using
// data in the request itsel such as certificate data, authorization bearer tokens, http headers etc.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Get the k8s context from the request
	ctx := ParseContextFromRequest(r)
	if ctx == "" {
		http.Error(w, "Not route: context not found", http.StatusBadGateway)
		return
	}

	// Get the subject from the request
	sub := ParseSubjectFromRequest(r)

	// Get k8s cluster and authinfo from kubeconfig using the ctx name
	cluster := getClusterByContextName(p.KubeConfig, ctx)
	if cluster == nil {
		http.Error(w, fmt.Sprintf("no route: cluster not found for '%s'", ctx), http.StatusBadGateway)
		return
	}
	auth := getAuthByContextName(p.KubeConfig, ctx)
	if auth == nil {
		http.Error(w, fmt.Sprintf("no route: authinfo not found for '%s'", ctx), http.StatusBadGateway)
		return
	}

	// Don't use transport cache if ttl is set to 0s
	resCache := cache.New()
	resCache.TTL = p.CacheTTL
	if p.CacheTTL.String() == "0s" {
		resCache = nil
	}

	// Create a transport that will be re-used for
	if p.transports[ctx] == nil {
		// Setup TLS config
		tlsConfig, err := configureTLS(auth, cluster)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			panic(err)
		}
		p.transports[ctx] = &Transport{
			TLSClientConfig: tlsConfig,
			Cache:           resCache,
		}
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
// of the URL path if no headers are set.
func ParseContextFromRequest(req *http.Request) string {
	val := req.Header.Get("Multikube-Context")
	if val != "" {
		return val
	}

	c, rem := getCtxFromURL(req.URL)
	if c != "" {
		val = c
		if rem != "" {
			req.URL.Path = rem
		}
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
