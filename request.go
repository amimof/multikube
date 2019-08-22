package multikube

import (
	"context"
	"io"
	"k8s.io/client-go/tools/clientcmd/api"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// Request is a simple type used to compose inidivudal requests to an HTTP server.
type Request struct {
	Transport    http.RoundTripper
	req          *http.Request
	url          *url.URL
	query        string
	apiVersion   string
	verb         string
	resourceType string
	resourceName string
	namespace    string
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
	ctx string
	sub string
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

// APIVer sets the api version to be used when building the URI for the request.
// Defaults to 'v1' if not set.
func (r *Request) APIVer(v string) *Request {
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

// Body sets the request body of the request being made.
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

// DoWithContext executes the request and returns an http.Response.
// DoWithContext expect a context to be provided.
func (r *Request) DoWithContext(ctx context.Context) (*http.Response, error) {

	// Use default transport if none provided
	if r.Transport == nil {
		r.Transport = http.DefaultTransport
	}

	// Return any error if any has been generated along the way before continuing
	if r.err != nil {
		return nil, r.err
	}

	u := r.URL().String()

	// Instantiate the http request for the roundtripper
	req, err := http.NewRequest(r.verb, u, r.body)
	if err != nil {
		return nil, err
	}
	r.req = req
	req.Header = r.headers

	// Make the call
	res, err := r.Transport.RoundTrip(r.req)
	if err != nil {
		return nil, err
	}

	return res, nil

}

// NewRequest will return a new Request object with the given URL
func NewRequest(u *url.URL) *Request {
	return &Request{url: u}
}
