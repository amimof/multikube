package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"github.com/amimof/multikube/pkg/cache"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

// Transport is an implementation of RoundTripper and extension of http.Transport with the
// addition of a Cache.
type Transport struct {
	Cache           *cache.Cache
	TLSClientConfig *tls.Config
	transport       *http.Transport
}

// RoundTrip implements http.Transport
func (t *Transport) RoundTrip(req *http.Request) (res *http.Response, err error) {

	// Use default transport with http2 if not told otherwise
	if t.transport == nil {
		t.transport = &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			TLSClientConfig:       t.TLSClientConfig,
		}
	}

	// Wrap our RoundTripper with Prometheus middleware.
	roundTripper := promhttp.InstrumentRoundTripperCounter(backendCounter,
		promhttp.InstrumentRoundTripperInFlight(backendGauge,
			promhttp.InstrumentRoundTripperDuration(backendHistogram, t.transport),
		),
	)

	// If no cache exists then carry out the request as usual
	if t.Cache == nil {
		res, err = roundTripper.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	// Cache the response
	item := t.Cache.Get(req.URL.String())
	if item != nil {
		res, err = t.readResponse(req)
		if err != nil {
			return nil, err
		}
		res.Header.Set("Multikube-Cache-Age", item.Age().String())
	} else {
		res, err = roundTripper.RoundTrip(req)
		if err != nil {
			return nil, err
		}
	}

	// Cache any response
	_, err = t.cacheResponse(res)
	if err != nil {
		return nil, err
	}

	return res, nil

}

// cacheResponse tries to commit a http.Response to the transport cache.
// Careful! cacheResponse makes use of http.DumpResponse which will drain the original response and replace it with a new one
func (t *Transport) cacheResponse(res *http.Response) (bool, error) {
	if t.Cache == nil {
		return false, nil
	}
	// Don't cache if method is not GET
	if res.Request.Method != http.MethodGet {
		return false, nil
	}
	// Don't cache if response code isn't 200 (OK) or 304 (NotModified)
	if !(res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotModified) {
		return false, nil
	}
	// Don't cache if certain url params are present (kubernetes streams)
	q := res.Request.URL.Query()
	if q.Get("watch") == "true" || q.Get("follow") == "true" {
		return false, nil
	}
	b, err := httputil.DumpResponse(res, true)
	if err != nil {
		return false, err
	}
	t.Cache.Set(res.Request.URL.String(), b)
	return true, nil
}

func (t *Transport) readResponse(req *http.Request) (*http.Response, error) {
	item := t.Cache.Get(req.URL.String())
	if item.Value == nil {
		return nil, nil
	}
	b := bytes.NewBuffer(item.Value)
	return http.ReadResponse(bufio.NewReader(b), req)
}
