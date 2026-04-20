package proxy

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/metric"
)

var (
	interval = time.Second * 60
	points   = 120
)

// bucketPtr is the generic constraint: T is a struct type whose pointer *T
// implements start() and zero(). This lets advanceBuckets and makeBuckets
// work on []T (value slices) while calling pointer-receiver methods.
type bucketPtr[T any] interface {
	~*T
	start() time.Time
	zero(t time.Time)
}

// Int64Bucket stores a running sum for a single time interval (counters).
type Int64Bucket struct {
	Start time.Time
	Value int64
}

func (b *Int64Bucket) start() time.Time { return b.Start }
func (b *Int64Bucket) zero(t time.Time) { b.Start = t; b.Value = 0 }

// Float64Bucket stores observation count and sum for a single time interval
// (float64 histograms such as request duration).
type Float64Bucket struct {
	Start time.Time
	Count int64
	Sum   float64
}

func (b *Float64Bucket) start() time.Time { return b.Start }
func (b *Float64Bucket) zero(t time.Time) { b.Start = t; b.Count = 0; b.Sum = 0 }

// Int64HistogramBucket stores observation count and sum for a single time
// interval (int64 histograms such as request/response size).
type Int64HistogramBucket struct {
	Start time.Time
	Count int64
	Sum   int64
}

func (b *Int64HistogramBucket) start() time.Time { return b.Start }
func (b *Int64HistogramBucket) zero(t time.Time) { b.Start = t; b.Count = 0; b.Sum = 0 }

// Int64GaugeBucket stores the maximum gauge value observed during a single
// time interval (up-down counters such as active requests).
type Int64GaugeBucket struct {
	Start time.Time
	Max   int64
}

func (b *Int64GaugeBucket) start() time.Time { return b.Start }
func (b *Int64GaugeBucket) zero(t time.Time) { b.Start = t; b.Max = 0 }

// advanceBuckets shifts the rolling window forward so that the last bucket
// covers the interval containing now. Buckets that fall outside the window
// are zeroed.
func advanceBuckets[T any, P bucketPtr[T]](buckets []T, now time.Time) {
	last := len(buckets) - 1
	steps := int(now.Sub(P(&buckets[last]).start()) / interval)
	if steps <= 0 {
		return
	}
	if steps >= len(buckets) {
		for i := range buckets {
			P(&buckets[i]).zero(now.Add(time.Duration(i-len(buckets)+1) * interval))
		}
		return
	}
	copy(buckets, buckets[steps:])
	for i := len(buckets) - steps; i < len(buckets); i++ {
		P(&buckets[i]).zero(P(&buckets[i-1]).start().Add(interval))
	}
}

// makeBuckets pre-allocates a rolling window of n buckets ending at now.
func makeBuckets[T any, P bucketPtr[T]](n int, now time.Time) []T {
	s := make([]T, n)
	for i := range s {
		P(&s[i]).zero(now.Add(time.Duration(i-n+1) * interval))
	}
	return s
}

type Int64Counter struct {
	mu      sync.Mutex
	Counter metric.Int64Counter
	Current atomic.Int64
	Buckets []Int64Bucket
}

func (m *Int64Counter) Inc(ctx context.Context, v int64, attrs ...metric.AddOption) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Counter.Add(ctx, v, attrs...)
	m.Current.Add(v)

	now := time.Now().Truncate(interval)
	advanceBuckets[Int64Bucket, *Int64Bucket](m.Buckets, now)
	m.Buckets[len(m.Buckets)-1].Value += v
}

func (m *Int64Counter) Load() int64 {
	return m.Current.Load()
}

func (m *Int64Counter) SnapshotBuckets() []Int64Bucket {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().Truncate(interval)
	advanceBuckets[Int64Bucket, *Int64Bucket](m.Buckets, now)
	out := make([]Int64Bucket, len(m.Buckets))
	copy(out, m.Buckets)
	return out
}

type Float64Histogram struct {
	mu        sync.Mutex
	Histogram metric.Float64Histogram
	Buckets   []Float64Bucket
}

func (m *Float64Histogram) Record(ctx context.Context, v float64, attrs ...metric.RecordOption) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Histogram.Record(ctx, v, attrs...)

	now := time.Now().Truncate(interval)
	advanceBuckets[Float64Bucket, *Float64Bucket](m.Buckets, now)
	last := &m.Buckets[len(m.Buckets)-1]
	last.Count++
	last.Sum += v
}

func (m *Float64Histogram) SnapshotBuckets() []Float64Bucket {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().Truncate(interval)
	advanceBuckets[Float64Bucket, *Float64Bucket](m.Buckets, now)
	out := make([]Float64Bucket, len(m.Buckets))
	copy(out, m.Buckets)
	return out
}

type Int64Histogram struct {
	mu        sync.Mutex
	Histogram metric.Int64Histogram
	Buckets   []Int64HistogramBucket
}

func (m *Int64Histogram) Record(ctx context.Context, v int64, attrs ...metric.RecordOption) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Histogram.Record(ctx, v, attrs...)

	now := time.Now().Truncate(interval)
	advanceBuckets[Int64HistogramBucket, *Int64HistogramBucket](m.Buckets, now)
	last := &m.Buckets[len(m.Buckets)-1]
	last.Count++
	last.Sum += v
}

func (m *Int64Histogram) SnapshotBuckets() []Int64HistogramBucket {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().Truncate(interval)
	advanceBuckets[Int64HistogramBucket, *Int64HistogramBucket](m.Buckets, now)
	out := make([]Int64HistogramBucket, len(m.Buckets))
	copy(out, m.Buckets)
	return out
}

type Int64UpDownCounter struct {
	mu      sync.Mutex
	Counter metric.Int64UpDownCounter
	current int64
	Buckets []Int64GaugeBucket
}

func (m *Int64UpDownCounter) Add(ctx context.Context, v int64, attrs ...metric.AddOption) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Counter.Add(ctx, v, attrs...)
	m.current += v

	now := time.Now().Truncate(interval)
	advanceBuckets[Int64GaugeBucket, *Int64GaugeBucket](m.Buckets, now)
	last := &m.Buckets[len(m.Buckets)-1]
	if m.current > last.Max {
		last.Max = m.current
	}
}

func (m *Int64UpDownCounter) SnapshotBuckets() []Int64GaugeBucket {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now().Truncate(interval)
	advanceBuckets[Int64GaugeBucket, *Int64GaugeBucket](m.Buckets, now)
	out := make([]Int64GaugeBucket, len(m.Buckets))
	copy(out, m.Buckets)
	return out
}

// ProxyMetrics holds all OpenTelemetry instruments for the proxy.
type ProxyMetrics struct {
	// Request metrics
	RequestsTotal     Int64Counter
	RequestDuration   Float64Histogram
	ActiveRequests    Int64UpDownCounter
	RequestSizeBytes  Int64Histogram
	ResponseSizeBytes Int64Histogram

	// Backend metrics
	BackendRequestsTotal   Int64Counter
	BackendRequestDuration Float64Histogram
	BackendActiveRequests  Int64UpDownCounter

	// Auth metrics
	AuthRequestsTotal      Int64Counter
	PolicyEvaluationsTotal Int64Counter

	// Route matching metrics
	RouteMatchesTotal Int64Counter
	RouteNoMatchTotal Int64Counter
}

// InitMetrics creates a fully-initialised ProxyMetrics from the given OTel
// meter. Each metric is backed by both an OTel instrument (for Prometheus
// scraping) and a rolling in-memory time series (for the web UI).
func InitMetrics(meter metric.Meter) (*ProxyMetrics, error) {
	now := time.Now().Truncate(interval)

	m := &ProxyMetrics{
		RequestsTotal:          Int64Counter{Buckets: makeBuckets[Int64Bucket, *Int64Bucket](points, now)},
		BackendRequestsTotal:   Int64Counter{Buckets: makeBuckets[Int64Bucket, *Int64Bucket](points, now)},
		AuthRequestsTotal:      Int64Counter{Buckets: makeBuckets[Int64Bucket, *Int64Bucket](points, now)},
		PolicyEvaluationsTotal: Int64Counter{Buckets: makeBuckets[Int64Bucket, *Int64Bucket](points, now)},
		RouteMatchesTotal:      Int64Counter{Buckets: makeBuckets[Int64Bucket, *Int64Bucket](points, now)},
		RouteNoMatchTotal:      Int64Counter{Buckets: makeBuckets[Int64Bucket, *Int64Bucket](points, now)},

		RequestDuration:        Float64Histogram{Buckets: makeBuckets[Float64Bucket, *Float64Bucket](points, now)},
		BackendRequestDuration: Float64Histogram{Buckets: makeBuckets[Float64Bucket, *Float64Bucket](points, now)},

		RequestSizeBytes:  Int64Histogram{Buckets: makeBuckets[Int64HistogramBucket, *Int64HistogramBucket](points, now)},
		ResponseSizeBytes: Int64Histogram{Buckets: makeBuckets[Int64HistogramBucket, *Int64HistogramBucket](points, now)},

		ActiveRequests:        Int64UpDownCounter{Buckets: makeBuckets[Int64GaugeBucket, *Int64GaugeBucket](points, now)},
		BackendActiveRequests: Int64UpDownCounter{Buckets: makeBuckets[Int64GaugeBucket, *Int64GaugeBucket](points, now)},
	}

	var err error

	// Request metrics
	if m.RequestsTotal.Counter, err = meter.Int64Counter(
		"proxy.http.requests.total",
		metric.WithDescription("Total number of HTTP requests processed"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	if m.RequestDuration.Histogram, err = meter.Float64Histogram(
		"proxy.http.request.duration.seconds",
		metric.WithDescription("Duration of HTTP requests in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}

	if m.ActiveRequests.Counter, err = meter.Int64UpDownCounter(
		"proxy.http.active.requests",
		metric.WithDescription("Number of in-flight HTTP requests"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	if m.RequestSizeBytes.Histogram, err = meter.Int64Histogram(
		"proxy.http.request.size.bytes",
		metric.WithDescription("Size of HTTP request bodies in bytes"),
		metric.WithUnit("By"),
	); err != nil {
		return nil, err
	}

	if m.ResponseSizeBytes.Histogram, err = meter.Int64Histogram(
		"proxy.http.response.size.bytes",
		metric.WithDescription("Size of HTTP response bodies in bytes"),
		metric.WithUnit("By"),
	); err != nil {
		return nil, err
	}

	// Backend metrics
	if m.BackendRequestsTotal.Counter, err = meter.Int64Counter(
		"proxy.backend.requests.total",
		metric.WithDescription("Total number of requests forwarded to backends"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	if m.BackendRequestDuration.Histogram, err = meter.Float64Histogram(
		"proxy.backend.request.duration.seconds",
		metric.WithDescription("Duration of upstream backend requests in seconds"),
		metric.WithUnit("s"),
	); err != nil {
		return nil, err
	}

	if m.BackendActiveRequests.Counter, err = meter.Int64UpDownCounter(
		"proxy.backend.active.requests",
		metric.WithDescription("Number of in-flight requests to backends"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	// Auth metrics
	if m.AuthRequestsTotal.Counter, err = meter.Int64Counter(
		"proxy.auth.requests.total",
		metric.WithDescription("Total JWT authentication attempts by result"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	if m.PolicyEvaluationsTotal.Counter, err = meter.Int64Counter(
		"proxy.policy.evaluations.total",
		metric.WithDescription("Total policy evaluation outcomes"),
		metric.WithUnit("{evaluation}"),
	); err != nil {
		return nil, err
	}

	// Route matching metrics
	if m.RouteMatchesTotal.Counter, err = meter.Int64Counter(
		"proxy.route.matches.total",
		metric.WithDescription("Total route matches by match kind"),
		metric.WithUnit("{match}"),
	); err != nil {
		return nil, err
	}

	if m.RouteNoMatchTotal.Counter, err = meter.Int64Counter(
		"proxy.route.no_match.total",
		metric.WithDescription("Total requests that matched no route"),
		metric.WithUnit("{request}"),
	); err != nil {
		return nil, err
	}

	return m, nil
}
