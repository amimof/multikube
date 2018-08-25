package multikube

import (
	"net/http"
	"net/http/httputil"
	"crypto/tls"
	"golang.org/x/net/http2"
	"bytes"
	"bufio"
)

// Tansport is an implementation of RoundTripper and extension of http.Transport with the 
// addition of Cache.
type Transport struct {
	Cache *Cache
	TLSClientConfig *tls.Config
	transport *http.Transport
}

func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {

	if t.transport == nil {
		t.transport = &http.Transport{
			TLSClientConfig: t.TLSClientConfig,
		}
		http2.ConfigureTransport(t.transport)
	}

	if t.Cache == nil {
		t.Cache = NewCache()
	}

	//var res *http.Response

	// Either return a response from the cache or from a real request
	item := t.Cache.Get(req.URL.String())
	if item.Value != nil {

		// Cache hit!
		res, err := t.readResponse(req)
		if err != nil {
			return nil, err
		}
		
		return res, nil

	} else {

		// Cache miss!
		res, err := t.transport.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// Cache the response 
		respBytes, err := httputil.DumpResponse(res, true)
		if err != nil {
			return nil, err
		}
		t.Cache.Set(req.URL.String(), respBytes)

		return res, nil

	}
	

}

func (t *Transport) readResponse(req *http.Request) (*http.Response, error) {

	item := t.Cache.Get(req.URL.String())
	if item.Value == nil {
		return nil, nil
	}

	b := bytes.NewBuffer(item.Value)

	return http.ReadResponse(bufio.NewReader(b), req)

}