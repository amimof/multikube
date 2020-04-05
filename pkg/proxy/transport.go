package proxy

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"github.com/amimof/multikube/pkg/cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"net/http/httputil"
	"time"
)

var backendHistogram = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "multikube_backend_request_duration_seconds",
		Help:    "A histogram of request latencies to backends",
		Buckets: prometheus.DefBuckets,
	},
	[]string{},
)

var backendCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "multikube_backend_requests_total",
		Help: "A counter for requests to backends.",
	},
	[]string{"code", "method"},
)

var backendGauge = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "multikube_backend_live_requests",
	Help: "A gauge of live requests currently in flight to backends",
})

// Transport is an implementation of RoundTripper and extension of http.Transport with the
// addition of a Cache.
type Transport struct {
	Cache           *cache.Cache
	TLSClientConfig *tls.Config
	transport       *http.Transport
}

func init() {
	prometheus.MustRegister(backendHistogram, backendCounter, backendGauge)
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

	// Either return a response from the cache or from a real request
	item := t.Cache.Get(req.URL.String())
	if item != nil && req.Method == http.MethodGet {

		// Cache hit!
		res, err = t.readResponse(req)
		if err != nil {
			return nil, err
		}

		// Set header to inform about the cached response
		res.Header.Set("Multikube-Cache", item.Age().String())

	} else {

		// Cache miss!
		res, err = roundTripper.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		if isCacheable(res.Request) {

			// Careful! DumpResponse will drain our original response and replace it with a new one
			resBytes, err := httputil.DumpResponse(res, true)
			if err != nil {
				return nil, err
			}

			//Cache the response if it's cacheable.
			if req.Method == http.MethodGet && (res.StatusCode == http.StatusOK || res.StatusCode == http.StatusNotModified) {
				t.Cache.Set(req.URL.String(), resBytes)
			}
		}

	}

	return res, nil

}

func (t *Transport) readResponse(req *http.Request) (*http.Response, error) {
	item := t.Cache.Get(req.URL.String())
	if item.Value == nil {
		return nil, nil
	}
	b := bytes.NewBuffer(item.Value)
	return http.ReadResponse(bufio.NewReader(b), req)
}

// isCacheable determines if an http request is eligible for caching
// by looking for watch and follow query parameters in the URL. This is very
// Kubernetes-specific and needs a better implementation. But will do for now.
func isCacheable(r *http.Request) bool {
	q := r.URL.Query()
	if q.Get("watch") == "true" || q.Get("follow") == "true" {
		return false
	}
	return true
}
