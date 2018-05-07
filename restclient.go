package multikube

import (
	"net/http"
	"net/url"
	"io"
	"io/ioutil"
	"crypto/tls"
	"crypto/x509"
	"path"
	"strings"
)

type Request struct {
	client *http.Client
	baseURL *url.URL
	verb string
	resourceType string
	resourceName string
	namespace string
	headers *http.Header
	params *url.Values
	body *io.Reader
	data []byte
}


func NewRequest(cl *Cluster) (*Request, error) {

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(cl.Cert, cl.Key)
	if err != nil {
		return nil, err
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(cl.CA)
	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs: caCertPool,
	}

	tlsConfig.BuildNameToCertificate()
	tr := &http.Transport{ TLSClientConfig: tlsConfig }
	client := &http.Client{ Transport: tr }

	base, err := url.Parse(cl.Hostname)
	if err != nil {
		return nil, err
	}

	return &Request{
		client: client,
		baseURL: base,
	}, nil

}

func (r *Request) Get() *Request {
	r.verb = "GET"
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

func (r *Request) SetHeader(key string, values ...string) *Request {
	if r.headers == nil {
		r.headers = &http.Header{}
	}
	r.headers.Del(key)
	for _, value := range values {
		r.headers.Add(key, value)
	}
	return r
}

func (r *Request) Do() (*Request, error) {
	url := r.URL().String()
	req, err := http.NewRequest(r.verb, url, nil)
	if err != nil {
		return nil, err
	}

	res, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	r.data = body

	return r, nil

}

func (r *Request) Data() []byte {
	return r.data
}

// URL returns the current working URL.
func (r *Request) URL() *url.URL {
	
	p := "/api/v1/"
	
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
	// TODO: Include query params in request
	//
	// query := url.Values{}
	// for key, values := range r.params {
	// 	for _, value := range values {
	// 		query.Add(key, value)
	// 	}
	// }

	return r.baseURL
}
