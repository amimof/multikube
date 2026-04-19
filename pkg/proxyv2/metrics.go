package proxy

import (
	"go.opentelemetry.io/otel/metric"
)

// ProxyMetrics holds all OpenTelemetry instruments for the proxy.
type ProxyMetrics struct {
	// Request metrics
	RequestsTotal     metric.Int64Counter
	RequestDuration   metric.Float64Histogram
	ActiveRequests    metric.Int64UpDownCounter
	RequestSizeBytes  metric.Int64Histogram
	ResponseSizeBytes metric.Int64Histogram

	// Backend metrics
	BackendRequestsTotal   metric.Int64Counter
	BackendRequestDuration metric.Float64Histogram
	BackendActiveRequests  metric.Int64UpDownCounter

	// Auth metrics
	AuthRequestsTotal      metric.Int64Counter
	PolicyEvaluationsTotal metric.Int64Counter

	// Route matching metrics
	RouteMatchesTotal metric.Int64Counter
	RouteNoMatchTotal metric.Int64Counter
}

func initMetrics(meter metric.Meter) (*ProxyMetrics, error) {
	m := &ProxyMetrics{}
	var err error

	// --- Request metrics ---

	if m.RequestsTotal, err = meter.Int64Counter(
		"proxy.http.requests.total",
		metric.WithDescription("Total number of HTTP requests processed"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	if m.RequestDuration, err = meter.Float64Histogram(
		"proxy.http.request.duration.seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}

	if m.ActiveRequests, err = meter.Int64UpDownCounter(
		"proxy.http.active.requests",
		metric.WithDescription("Number of in-flight HTTP requests"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	if m.RequestSizeBytes, err = meter.Int64Histogram(
		"proxy.http.request.size.bytes",
		metric.WithDescription("Size of HTTP request bodies in bytes"),
		metric.WithUnit("By"),
	); err != nil {
		return nil, err
	}

	if m.ResponseSizeBytes, err = meter.Int64Histogram(
		"proxy.http.response.size.bytes",
		metric.WithDescription("Size of HTTP response bodies in bytes"),
		metric.WithUnit("By"),
	); err != nil {
		return nil, err
	}

	// --- Backend metrics ---

	if m.BackendRequestsTotal, err = meter.Int64Counter(
		"proxy.backend.requests.total",
		metric.WithDescription("Total number of requests forwarded to backends"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	if m.BackendRequestDuration, err = meter.Float64Histogram(
		"proxy.backend.request.duration.seconds",
		metric.WithDescription("Duration of upstream backend requests in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}

	if m.BackendActiveRequests, err = meter.Int64UpDownCounter(
		"proxy.backend.active.requests",
		metric.WithDescription("Number of in-flight requests to backends"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	// --- Auth metrics ---

	if m.AuthRequestsTotal, err = meter.Int64Counter(
		"proxy.auth.requests.total",
		metric.WithDescription("Total JWT authentication attempts by result"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	if m.PolicyEvaluationsTotal, err = meter.Int64Counter(
		"proxy.policy.evaluations.total",
		metric.WithDescription("Total policy evaluation outcomes"),
		metric.WithUnit("{evaluation}"),
	); err != nil {
		return nil, err
	}

	// --- Route matching metrics ---

	if m.RouteMatchesTotal, err = meter.Int64Counter(
		"proxy.route.matches.total",
		metric.WithDescription("Total route matches by match kind"),
		metric.WithUnit("{match}"),
	); err != nil {
		return nil, err
	}

	if m.RouteNoMatchTotal, err = meter.Int64Counter(
		"proxy.route.no_match.total",
		metric.WithDescription("Total requests that matched no route"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	return m, nil
}
