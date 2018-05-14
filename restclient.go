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
	"encoding/json"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
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
	interf interface{}
	err error
}


func NewRequest(config *ClusterConfig) *Request {

	r := &Request{}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(config.Cert, config.Key)
	if err != nil {
		r.err = newErr(err.Error())
		return r
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(config.CA)
	if err != nil {
		r.err = newErr(err.Error())
		return r
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs: caCertPool,
	}

	tlsConfig.BuildNameToCertificate()
	tr := &http.Transport{ TLSClientConfig: tlsConfig }
	r.client = &http.Client{ Transport: tr }

	base, err := url.Parse(config.Hostname)
	if err != nil {
		r.err = newErr(err.Error())
		return r
	}
	r.baseURL = base

	return r

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

	// Lets try to see if response is a failure
	status := &v1.Status{}
	err = json.Unmarshal(r.Data(), status)
	if err != nil {
		return nil, err
	}
	err = handleResponse(status)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(r.Data(), r.interf)
	if err != nil {
		return nil, err
	}

	// Return any error if any has been generated along the way
	if r.err != nil {
		return nil, r.err
	}

	return r, nil

}

func (r *Request) Into(obj interface{}) *Request {
	r.interf = obj
	return r
}

func (r *Request) Data() []byte {
	return r.data
}

// URL returns the current working URL.
func (r *Request) URL() *url.URL {
	
	if r.baseURL.Path != "" {
		r.baseURL.Path = ""
	}
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
