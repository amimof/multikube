package multikube

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd/api"
	"net/http"
	"net/url"
	"path"
	"strings"
)

type Request struct {
	client       *http.Client
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
	TLSConfig    *tls.Config
}

type Options struct {
	*api.Cluster
	*api.AuthInfo
}

func (r *Request) Get() *Request {
	r.Method(http.MethodGet)
	return r
}

func (r *Request) Post() *Request {
	r.Method(http.MethodPost)
	return r
}

func (r *Request) Put() *Request {
	r.Method(http.MethodPut)
	return r
}

func (r *Request) Delete() *Request {
	r.Method(http.MethodDelete)
	return r
}

func (r *Request) Options() *Request {
	r.Method(http.MethodOptions)
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
	r.url.Path = p
	return r
}

func (r *Request) Query(q string) *Request {
	r.url.RawQuery = q
	return r
}

func (r *Request) Body(b io.Reader) *Request {
	r.body = b
	return r
}

func (r *Request) Headers(h http.Header) *Request {
	r.headers = h
	return r
}

func (r *Request) Impersonate(n string) *Request {
	r.impersonate = n
	return r
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

 	if r.impersonate != "" {
		 req.Header.Set("Impersonate-User", r.impersonate)
	 }

	res, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func NewRequest(options *Options) *Request {

	r := &Request{}
	tlsConfig := &tls.Config{
		InsecureSkipVerify: options.InsecureSkipTLSVerify,
		NextProtos:         []string{"http/1.1"},
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

	tr := &http.Transport{TLSClientConfig: tlsConfig}
	r.client = &http.Client{
		Transport: tr,
		Timeout:   0,
	}

	base, err := url.Parse(options.Server)
	if err != nil {
		r.err = newErr(err.Error())
		return r
	}
	r.url = base

	return r

}
