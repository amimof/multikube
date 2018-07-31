package multikube

import (
	"io"
	"path"
	"strings"
	"net/url"
	"net/http"
	"io/ioutil"
	"crypto/tls"
	"crypto/x509"
)

type Request struct {
	client       *http.Client
	baseURL      *url.URL
	path         string
	query			 	 string
	apiVersion   string
	verb         string
	resourceType string
	resourceName string
	namespace    string
	headers      http.Header
	params       *url.Values
	body         io.Reader
	data         []byte
	interf       interface{}
	err          error
	resp			 	 *http.Response
	tlsConfig		 *tls.Config
}

type Options interface {
	Hostname() string
	CA() string
	Cert() string
	Key() string
	Insecure() bool
}

func (r *Request) TLSConfig() *tls.Config {
	return r.tlsConfig
}

func (r *Request) Get() *Request {
	r.Method("GET")
	return r
}

func (r *Request) Post() *Request {
	r.Method("POST")
	return r
}

func (r *Request) Put() *Request {
	r.Method("PUT")
	return r
}

func (r *Request) Delete() *Request {
	r.Method("DELETE")
	return r
}

func (r *Request) Options() *Request {
	r.Method("OPTIONS")
	return r
}

func (r *Request) Method(m string) *Request {
	r.verb = m
	return r
}

func (r *Request) Resource(res string) *Request {
	r.resourceType = res
	return r
}

func (r *Request) Name(n string) *Request {
	r.resourceName = n
	return r
}

func (r *Request) Namespace(ns string) *Request {
	r.namespace = ns
	return r
}

func (r *Request) ApiVer(v string) *Request {
	r.apiVersion = v
	return r
}

func (r *Request) Into(obj interface{}) *Request {
	r.interf = obj
	return r
}

func (r *Request) Path(p string) *Request {
	r.path = p
	return r
}

func (r *Request) Query(q string) *Request {
	r.query = q
	return r
}

func (r *Request) Body(b io.Reader) *Request {
	r.body = b
	return r
}

func (r *Request) Data() []byte {
	return r.data
}

func (r *Request) Headers(h http.Header) *Request {
	r.headers = h
	return r
}

func (r *Request) Response() *http.Response {
	return r.resp
}

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

// URL returns the current working URL.
func (r *Request) URL() *url.URL {

	if r.baseURL == nil {
		r.baseURL = &url.URL{}
	}

	if r.baseURL.Path != "" {
		r.baseURL.Path = ""
	}

	p := "/api/v1/"
	if r.path != "" {
		p = r.path
	}

	if r.query != "" {
		r.baseURL.RawQuery = r.query
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

	r.baseURL.Path = path.Join(r.baseURL.Path, p)

	return r.baseURL
}

// Doo executes the request and returns an http.Response. The caller is responible of closing the Body
func (r *Request) Do() (*http.Response, error) {

	// Return any error if any has been generated along the way before continuing
	if r.err != nil {
		return nil, r.err
	}
	
	u := r.URL().String()

	req, err := http.NewRequest(r.verb, u, r.body)
	if err != nil {
		return nil, err
	}
	req.Close = false
	req.Header = r.headers

	res, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	r.resp = res

	return res, nil

}

func NewRequest(options Options) *Request {

	r := &Request{}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: options.Insecure(),
		NextProtos: []string{"http/1.1"},
	}

	if options.CA() != "" {

		// Load CA certificate
		caCert, err := ioutil.ReadFile(options.CA())
		if err != nil {
			r.err = newErr(err.Error())
			return r
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool

	}

	if options.Cert() != "" && options.Key() != "" {
		//Load client certificate
		cert, err := tls.LoadX509KeyPair(options.Cert(), options.Key())
		if err != nil {
			r.err = newErr(err.Error())
			return r
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
		tlsConfig.BuildNameToCertificate()
	}

	r.tlsConfig = tlsConfig

	tr := &http.Transport{TLSClientConfig: tlsConfig}
	r.client = &http.Client{
		Transport: tr,
		Timeout: 0,
	}

	base, err := url.Parse(options.Hostname())
	if err != nil {
		r.err = newErr(err.Error())
		return r
	}
	r.baseURL = base

	return r

}