package multikube

import (
	"log"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd/api"
	"net/http"
	"net/url"
	"path"
	"strings"
	"context"
	"golang.org/x/net/http2"
)

// Request is a simple type used to compose inidivudal requests to an HTTP server.
type Request struct {
	Opts         *Options
	Transport		 *http.Transport
	TLSConfig    *tls.Config
	req					 *http.Request
	url          *url.URL
	query        string
	apiVersion   string
	verb         string
	resourceType string
	resourceName string
	namespace    string
	impersonate  string
	headers      http.Header
	body         io.Reader
	interf       interface{}
	err          error
}

// Options embeds Cluster and AuthInfo from https://godoc.org/k8s.io/client-go/tools/clientcmd/api
// so that fields and methods are easily accessible from one type.
type Options struct {
	*api.Cluster
	*api.AuthInfo
}

// Get method sets the method on a request to GET. Get will invoke Method(http.MethodGet).
func (r *Request) Get() *Request {
	r.Method(http.MethodGet)
	return r
}

// Post method sets the method on a request to POST. Post will invoke Method(http.MethodPost).
func (r *Request) Post() *Request {
	r.Method(http.MethodPost)
	return r
}

// Put method sets the method on a request to PUT. Put will invoke Method(http.MethodPut).
func (r *Request) Put() *Request {
	r.Method(http.MethodPut)
	return r
}

// Delete method sets the method on a request to DELETE. Delete will invoke Method(http.MethodDelete).
func (r *Request) Delete() *Request {
	r.Method(http.MethodDelete)
	return r
}

// Options method sets the method on a request to OPTIONS. Options will invoke Method(http.MethodOptions),
func (r *Request) Options() *Request {
	r.Method(http.MethodOptions)
	return r
}

// Method methdo sets the method on a request.
func (r *Request) Method(m string) *Request {
	r.verb = m
	return r
}

// Resource sets the Kubernetes resource to be used when building the URI. For example
// setting the resource to 'Pod' will create an uri like /api/v1/namespaces/pods.
func (r *Request) Resource(res string) *Request {
	r.resourceType = res
	return r
}

// Name sets the name of the Kubernetes resource to be used when building the URI. For example
// setting the name to 'app-pod-1' will create an uri like /api/v1/namespaces/pods/app-pod-1.
func (r *Request) Name(n string) *Request {
	r.resourceName = n
	return r
}

// Namespace sets the Kubernetes namespace to be used when building the URI. For example
// setting the namespace to 'default' will create an uri like /api/v1/namespaces/default.
func (r *Request) Namespace(ns string) *Request {
	r.namespace = ns
	return r
}

// ApiVer sets the api version to be used when building the URI for the request.
// Defaults to 'v1' if not set.
func (r *Request) ApiVer(v string) *Request {
	r.apiVersion = v
	return r
}

// Into sets the interface in which the returning data will be marshaled into.
func (r *Request) Into(obj interface{}) *Request {
	r.interf = obj
	return r
}

// Path sets the raw URI path later used by the request.
func (r *Request) Path(p string) *Request {
	r.url.Path = p
	return r
}

// Query sets the raw query path to be used when performing the request
func (r *Request) Query(q string) *Request {
	r.url.RawQuery = q
	return r
}

// Body sets the request body of the request beeing made.
func (r *Request) Body(b io.Reader) *Request {
	r.body = b
	return r
}

// Headers overrides the entire headers field of the http request.
// Use Header() method to set individual headers.
func (r *Request) Headers(h http.Header) *Request {
	r.headers = h
	return r
}

// Impersonat sets the Impersonate-User HTTP header for the request
func (r *Request) Impersonate(n string) *Request {
	r.impersonate = n
	return r
}

// Header sets one header and replacing any headers with equal key
func (r *Request) Header(key string, values ...string) *Request {
	if r.headers == nil {
		r.headers = http.Header{}
	}
	r.headers.Del(key)
	for _, value := range values {
		r.headers.Add(key, value)
	}
	return r
}

// URL composes a complete URL and return an url.URL then used by the request
func (r *Request) URL() *url.URL {

	p := "/api/v1/"
	if r.url.Path != "" {
		p = r.url.Path
	}

	// Set Api version only if not v1
	if len(r.apiVersion) > 0 && r.apiVersion != "v1" {
		p = path.Join("/apis", r.apiVersion)
	}

	// Is this resource namespaced?
	if len(r.namespace) > 0 {
		p = path.Join(p, "namespaces", r.namespace)
	}

	// Append resource scope
	if len(r.resourceType) != 0 {
		p = path.Join(p, strings.ToLower(r.resourceType))
	}

	// Append resource name scope
	if len(r.resourceName) != 0 {
		p = path.Join(p, r.resourceName)
	}

	r.url.Path = p

	return r.url
}

// Do executes the request and returns an http.Response.
// The caller is responible of closing the Body.
func (r *Request) Do() (*http.Response, error) {
	res, err := r.DoWithContext(context.Background())
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *Request) DoWithContext(ctx context.Context) (*http.Response, error) {

	// Return any error if any has been generated along the way before continuing
	if r.err != nil {
		return nil, r.err
	}

	u := r.URL().String()

	req, err := http.NewRequest(r.verb, u, r.body)
	if err != nil {
		return nil, err
	}
	r.req = req
	req.Header = r.headers

	// Set any headers that we might want
	if r.impersonate != "" {
		req.Header.Set("Impersonate-User", r.impersonate)
	}
	if r.Opts.Token != "" {
		r.headers.Set("Authorization", fmt.Sprintf("Bearer %s", r.Opts.Token))
	}

	// Make the call
	res, err := r.Transport.RoundTrip(r.req)
	if err != nil {
		log.Printf("Err: %s", err.Error())
		return nil, err
	}

	log.Printf("<- %s %s %s %s %s %s", req.Method, req.URL.Path, req.URL.RawQuery, req.RemoteAddr, res.Proto, res.Status)

	return res, nil

}

// NewRequest builds an http request to be used for execution and returns a Request type.
// and expects an Option interface. NewRequest creates an http.Client for each individual request
// but you may access the Client field on the Request type returned by NewRequest in order to
// override the defaults.
func NewRequest(options *Options) *Request {

	r := &Request{Opts: options}

	base, err := url.Parse(options.Server)
	if err != nil {
		r.err = err
		return r
	}
	r.url = base

	// Use already defined transport
	if r.Transport != nil {
		return r
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: options.InsecureSkipTLSVerify,
		NextProtos: []string{"h2", "http/1.1"},
	}

	// Load CA from file
	if options.CertificateAuthority != "" {
		caCert, err := ioutil.ReadFile(options.CertificateAuthority)
		if err != nil {
			r.err = newErr(err.Error())
			return r
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
			r.err = newErr(err.Error())
			return r
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	// Load certs from block
	if options.ClientCertificateData != nil && options.ClientKeyData != nil {
		cert, err := tls.X509KeyPair(options.ClientCertificateData, options.ClientKeyData)
		if err != nil {
			r.err = newErr(err.Error())
			return r
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	r.TLSConfig = tlsConfig
	r.Transport = &http.Transport{TLSClientConfig: tlsConfig}
	http2.ConfigureTransport(r.Transport)

	return r

}
