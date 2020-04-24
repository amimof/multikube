package proxy

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (

	// http
	httpDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "multikube_http_duration_seconds",
		Help: "A histogram of http request durations.",
	},
		[]string{"context", "method", "protocol"},
	)
	httpRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "multikube_http_requests_total",
		Help: "A counter for total http requests.",
	},
		[]string{"context", "method", "protocol", "code"},
	)
	httpRequestsCached = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "multikube_http_requests_cached_total",
		Help: "A counter for total cached http requests.",
	},
		[]string{"context", "method", "protocol", "code"},
	)

	// Backend
	backendHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "multikube_backend_request_duration_seconds",
		Help:    "A histogram of request latencies to backends",
		Buckets: prometheus.DefBuckets,
	},
		[]string{},
	)
	backendCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "multikube_backend_requests_total",
		Help: "A counter for requests to backends.",
	},
		[]string{},
	)
	backendGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "multikube_backend_live_requests",
		Help: "A gauge of live requests currently in flight to backends",
	})

	// OIDC
	oidcIssuerUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "multikube_oidc_provider_up",
		Help: "",
	},
		[]string{"context", "issuer"},
	)
	oidcReqsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "multikube_oidc_requests_total",
		Help: "A counter for total http requests.",
	},
		[]string{"context"},
	)
	oidcReqsAuthorized = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "multikube_oidc_requests_authorized_total",
		Help: "A counter for successfully authorized requests.",
	},
		[]string{"context"},
	)
	oidcReqsUnauthorized = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "multikube_oidc_requests_unauthorized_total",
		Help: "A counter for unauthorized requests.",
	},
		[]string{"context"},
	)

	// RS256
	rs256ReqsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "multikube_rs256_requests_total",
		Help: "A counter for total http requests.",
	},
		[]string{"context"},
	)
	rs256ReqsAuthorized = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "multikube_rs256_requests_authorized_total",
		Help: "A counter for successfully authorized requests.",
	},
		[]string{"context"},
	)
	rs256ReqsUnauthorized = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "multikube_rs256_requests_unauthorized_total",
		Help: "A counter for unauthorized requests.",
	},
		[]string{"context"},
	)

	// Cache
	cacheLen = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "multikube_cache_items_total",
		Help: "A gauge for the total amount of cached items.",
	},
		[]string{"context"},
	)
	cacheTTL = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "multikube_cache_ttl_seconds",
		Help: "A gauge for the cache TTL in seconds.",
	},
		[]string{"context"},
	)
)

func init() {
	prometheus.MustRegister(
		httpDuration,
		httpRequests,
		httpRequestsCached,
		backendHistogram,
		backendCounter,
		backendGauge,
		oidcIssuerUp,
		oidcReqsTotal,
		oidcReqsAuthorized,
		oidcReqsUnauthorized,
		rs256ReqsTotal,
		rs256ReqsAuthorized,
		rs256ReqsUnauthorized,
		cacheTTL,
	)
}
