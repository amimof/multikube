package multikube

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd/api"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

const (
	SubjectUndefined string = "No route: subject undefined"
	ContextUndefined string = "No route: context undefined"
	ContextNotFound  string = "No route: context not found"
)

type Proxy struct {
	CertChain  *x509.Certificate
	Config     *api.Config
	mw         http.Handler
	transports map[string]http.RoundTripper
	tlsconfigs map[string]*tls.Config
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

// NewProxy crerates a new Proxy and initialises router and configuration
func NewProxy() *Proxy {
	return &Proxy{
		transports: make(map[string]http.RoundTripper),
		tlsconfigs: make(map[string]*tls.Config),
	}
}

func NewProxyFrom(c *api.Config) *Proxy {
	p := NewProxy()
	p.Config = c
	return p
}

// Use chains all middlewares and applies a context to the request flow
func (p *Proxy) Use(mw ...Middleware) Middleware {
	return func(final http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			ctx := context.WithValue(r.Context(), "signer", p.CertChain)
			last.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (p *Proxy) getCluster(n string) *api.Cluster {
	for k, v := range p.Config.Clusters {
		if k == n {
			return v
		}
	}
	return nil
}

func (p *Proxy) getAuthInfo(n string) *api.AuthInfo {
	for k, v := range p.Config.AuthInfos {
		if k == n {
			return v
		}
	}
	return nil
}

func (p *Proxy) getContext(n string) *api.Context {
	for k, v := range p.Config.Contexts {
		if k == n {
			return v
		}
	}
	return nil
}

func (p *Proxy) getOptions(n string) *Options {
	ctx := p.getContext(n)
	if ctx == nil {
		return nil
	}
	authInfo := p.getAuthInfo(ctx.AuthInfo)
	if authInfo == nil {
		return nil
	}
	cluster := p.getCluster(ctx.Cluster)
	if cluster == nil {
		return nil
	}
	return &Options{
		cluster,
		authInfo,
	}
}

// ServeHTTP routes the request to an apiserver. It determines, resolves an apiserver using
// data in the request itsel such as certificate data, authorization bearer tokens, http headers etc.
func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// Make sure Subject is set
	sub, ok := r.Context().Value("Subject").(string)
	if !ok || sub == "" {
		http.Error(w, SubjectUndefined, http.StatusInternalServerError)
		return
	}

	// Make sure Context is set
	ctx, ok := r.Context().Value("Context").(string)
	if !ok || ctx == "" {
		http.Error(w, ContextUndefined, http.StatusInternalServerError)
		return
	}

	// Get a kubeconfig context
	opts := p.getOptions(ctx)
	if opts == nil {
		http.Error(w, ContextNotFound, http.StatusInternalServerError)
		return
	}

	// Tunnel the connection if server sends Upgrade
	if r.Header.Get("Upgrade") != "" {
		p.transports[ctx].(*Transport).TLSClientConfig.NextProtos = []string{"http/1.1"}
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
	req.Header("Impersonate-User", sub)
	req.Header("Authorization", fmt.Sprintf("Bearer %s", opts.Token))

	// Setup TLS config
	tlsConfig, err := configureTLS(opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p.tlsconfigs[ctx] = tlsConfig

	// Remember the transport created by the restclient so that we can re-use the connection
	if p.transports[ctx] == nil {
		p.transports[ctx] = &Transport{
			TLSClientConfig: tlsConfig,
		}
	}

	// Assign our transport to the request
	req.Transport = p.transports[ctx]

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
		w.Write(b)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
	}

}

// tunnel hijacks the client request, creates a pipe between client and backend server
// and starts streaming data between the two connections.
func (p *Proxy) tunnel(w http.ResponseWriter, r *http.Request) {

	// Make sure Context is set
	ctx, ok := r.Context().Value("Context").(string)
	if !ok || ctx == "" {
		http.Error(w, ContextUndefined, http.StatusInternalServerError)
		return
	}

	// Get a kubeconfig context
	opts := p.getOptions(ctx)
	if opts == nil {
		http.Error(w, ContextNotFound, http.StatusInternalServerError)
		return
	}

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	u, err := url.Parse(opts.Server)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	dst_conn, err := tls.Dial("tcp", u.Host, p.tlsconfigs[ctx])
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	dst_conn.Write(dump)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	src_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	go transfer(dst_conn, src_conn)
	go transfer(src_conn, dst_conn)

}

// transfer reads the data from src into a buffer before it writes it into dst
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

// configureTLS composes a TLS configuration from the provided Options parameter.
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
