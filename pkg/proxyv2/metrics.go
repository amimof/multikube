package proxy

import (
	"context"
	"encoding/json"
	"errors"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/nakabonne/tstorage"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	seriesStep      = time.Minute
	seriesPoints    = 120
	seriesIndexFile = "series-index.json"

	kindCounter   = "counter"
	kindHistogram = "histogram"
	kindGauge     = "gauge"
)

type MetricLabel struct {
	Name  string
	Value string
}

type MetricBucket struct {
	Start time.Time
	Value float64
	Count int64
	Sum   float64
}

type MetricSeries struct {
	Metric  string
	Kind    string
	Labels  []MetricLabel
	Buckets []MetricBucket
}

type seriesRegistry struct {
	mu      sync.RWMutex
	entries map[string]MetricSeries
}

func newSeriesRegistry() *seriesRegistry {
	return &seriesRegistry{entries: make(map[string]MetricSeries)}
}

func (r *seriesRegistry) Register(metricName, kind string, labels []tstorage.Label) {
	entry := MetricSeries{
		Metric: metricName,
		Kind:   kind,
		Labels: cloneMetricLabels(labels),
	}

	r.mu.Lock()
	r.entries[seriesRegistryKey(metricName, kind, labels)] = entry
	r.mu.Unlock()
}

func (r *seriesRegistry) RegisterSeries(entry MetricSeries) {
	r.mu.Lock()
	r.entries[seriesRegistryKey(entry.Metric, entry.Kind, metricLabelsToStorage(entry.Labels))] = MetricSeries{
		Metric: entry.Metric,
		Kind:   entry.Kind,
		Labels: append([]MetricLabel(nil), entry.Labels...),
	}
	r.mu.Unlock()
}

func (r *seriesRegistry) List() []MetricSeries {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]MetricSeries, 0, len(r.entries))
	for _, entry := range r.entries {
		entry.Buckets = nil
		out = append(out, entry)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Metric != out[j].Metric {
			return out[i].Metric < out[j].Metric
		}
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return labelsKey(metricLabelsToStorage(out[i].Labels)) < labelsKey(metricLabelsToStorage(out[j].Labels))
	})
	return out
}

func seriesRegistryKey(metricName, kind string, labels []tstorage.Label) string {
	return metricName + "|" + kind + "|" + labelsKey(labels)
}

func cloneMetricLabels(labels []tstorage.Label) []MetricLabel {
	if len(labels) == 0 {
		return nil
	}
	out := make([]MetricLabel, len(labels))
	for i, label := range labels {
		out[i] = MetricLabel{Name: label.Name, Value: label.Value}
	}
	return out
}

func metricLabelsToStorage(labels []MetricLabel) []tstorage.Label {
	if len(labels) == 0 {
		return nil
	}
	out := make([]tstorage.Label, len(labels))
	for i, label := range labels {
		out[i] = tstorage.Label{Name: label.Name, Value: label.Value}
	}
	return out
}

func labelsKey(labels []tstorage.Label) string {
	if len(labels) == 0 {
		return ""
	}
	normalized := cloneStorageLabels(labels)
	var b strings.Builder
	for i, label := range normalized {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(label.Name)
		b.WriteByte('=')
		b.WriteString(label.Value)
	}
	return b.String()
}

func cloneStorageLabels(labels []tstorage.Label) []tstorage.Label {
	if len(labels) == 0 {
		return nil
	}
	out := make([]tstorage.Label, 0, len(labels))
	for _, label := range labels {
		if label.Name == "" || label.Value == "" {
			continue
		}
		out = append(out, tstorage.Label{Name: label.Name, Value: label.Value})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Name != out[j].Name {
			return out[i].Name < out[j].Name
		}
		return out[i].Value < out[j].Value
	})
	return out
}

func normalizeAttributes(attrs []attribute.KeyValue) ([]attribute.KeyValue, []tstorage.Label) {
	if len(attrs) == 0 {
		return nil, nil
	}
	normalized := make([]attribute.KeyValue, 0, len(attrs))
	labels := make([]tstorage.Label, 0, len(attrs))
	for _, attr := range attrs {
		if !attr.Valid() {
			continue
		}
		value := attr.Value.Emit()
		if value == "" {
			continue
		}
		normalized = append(normalized, attr)
		labels = append(labels, tstorage.Label{Name: string(attr.Key), Value: value})
	}
	labels = cloneStorageLabels(labels)
	sort.Slice(normalized, func(i, j int) bool {
		if normalized[i].Key != normalized[j].Key {
			return normalized[i].Key < normalized[j].Key
		}
		return normalized[i].Value.Emit() < normalized[j].Value.Emit()
	})
	return normalized, labels
}

type metricsStorage struct {
	storage  tstorage.Storage
	registry *seriesRegistry
	index    *seriesIndex
}

func newMetricsStorage(storage tstorage.Storage, dataPath string) (*metricsStorage, error) {
	if storage == nil {
		var err error
		storage, err = tstorage.NewStorage()
		if err != nil {
			return nil, err
		}
	}
	registry := newSeriesRegistry()
	index := newSeriesIndex(dataPath)
	if err := index.LoadInto(registry); err != nil {
		return nil, err
	}
	return &metricsStorage{storage: storage, registry: registry, index: index}, nil
}

func (s *metricsStorage) Insert(metricName, kind string, labels []tstorage.Label, value float64, timestamp int64) {
	if s == nil || s.storage == nil {
		return
	}
	labels = cloneStorageLabels(labels)
	s.registry.Register(metricName, kind, labels)
	if s.index != nil {
		_ = s.index.Register(MetricSeries{Metric: metricName, Kind: kind, Labels: cloneMetricLabels(labels)})
	}
	_ = s.storage.InsertRows([]tstorage.Row{{
		Metric: metricName,
		Labels: labels,
		DataPoint: tstorage.DataPoint{
			Timestamp: timestamp,
			Value:     value,
		},
	}})
}

func (s *metricsStorage) Series(step time.Duration, count int) ([]MetricSeries, error) {
	if s == nil || s.storage == nil {
		return nil, nil
	}
	if step <= 0 {
		step = seriesStep
	}
	if count <= 0 {
		count = seriesPoints
	}

	series := s.registry.List()
	out := make([]MetricSeries, 0, len(series))
	for _, entry := range series {
		labels := metricLabelsToStorage(entry.Labels)
		buckets, err := s.loadBuckets(entry.Metric, entry.Kind, labels, step, count)
		if err != nil {
			return nil, err
		}
		entry.Buckets = buckets
		out = append(out, entry)
	}
	return out, nil
}

type seriesIndex struct {
	path string
	mu   sync.Mutex
	seen map[string]struct{}
}

func newSeriesIndex(dataPath string) *seriesIndex {
	if dataPath == "" {
		return nil
	}
	return &seriesIndex{
		path: filepath.Join(dataPath, seriesIndexFile),
		seen: make(map[string]struct{}),
	}
}

func (i *seriesIndex) LoadInto(registry *seriesRegistry) error {
	if i == nil {
		return nil
	}
	data, err := os.ReadFile(i.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var entries []MetricSeries
	if err := json.Unmarshal(data, &entries); err != nil {
		return err
	}
	for _, entry := range entries {
		registry.RegisterSeries(entry)
		i.seen[seriesRegistryKey(entry.Metric, entry.Kind, metricLabelsToStorage(entry.Labels))] = struct{}{}
	}
	return nil
}

func (i *seriesIndex) Register(entry MetricSeries) error {
	if i == nil {
		return nil
	}
	key := seriesRegistryKey(entry.Metric, entry.Kind, metricLabelsToStorage(entry.Labels))
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.seen[key]; ok {
		return nil
	}
	i.seen[key] = struct{}{}
	entries, err := i.entries()
	if err != nil {
		return err
	}
	entries = append(entries, MetricSeries{
		Metric: entry.Metric,
		Kind:   entry.Kind,
		Labels: append([]MetricLabel(nil), entry.Labels...),
	})
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(i.path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(i.path, data, 0o644)
}

func (i *seriesIndex) entries() ([]MetricSeries, error) {
	data, err := os.ReadFile(i.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var entries []MetricSeries
	if len(data) == 0 {
		return nil, nil
	}
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *metricsStorage) loadBuckets(metricName, kind string, labels []tstorage.Label, step time.Duration, count int) ([]MetricBucket, error) {
	now := time.Now().UTC().Truncate(step)
	start := now.Add(-time.Duration(count-1) * step)
	end := now.Add(step)
	buckets := makeBuckets(start, step, count)

	switch kind {
	case kindCounter:
		points, err := s.selectPoints(metricName, labels, start, end)
		if err != nil {
			return nil, err
		}
		for _, point := range points {
			idx := bucketIndex(point.Timestamp, start, step, count)
			if idx < 0 {
				continue
			}
			buckets[idx].Value += point.Value
		}
	case kindHistogram:
		countPoints, err := s.selectPoints(metricName+".count", labels, start, end)
		if err != nil {
			return nil, err
		}
		sumPoints, err := s.selectPoints(metricName+".sum", labels, start, end)
		if err != nil {
			return nil, err
		}
		for _, point := range countPoints {
			idx := bucketIndex(point.Timestamp, start, step, count)
			if idx < 0 {
				continue
			}
			buckets[idx].Count += int64(math.Round(point.Value))
		}
		for _, point := range sumPoints {
			idx := bucketIndex(point.Timestamp, start, step, count)
			if idx < 0 {
				continue
			}
			buckets[idx].Sum += point.Value
		}
	case kindGauge:
		points, err := s.selectPoints(metricName, labels, start, end)
		if err != nil {
			return nil, err
		}
		seen := make([]bool, len(buckets))
		for _, point := range points {
			idx := bucketIndex(point.Timestamp, start, step, count)
			if idx < 0 {
				continue
			}
			if !seen[idx] || point.Value > buckets[idx].Value {
				buckets[idx].Value = point.Value
				seen[idx] = true
			}
		}
	default:
		return nil, nil
	}

	return buckets, nil
}

func (s *metricsStorage) selectPoints(metricName string, labels []tstorage.Label, start, end time.Time) ([]*tstorage.DataPoint, error) {
	points, err := s.storage.Select(metricName, labels, start.UnixNano(), end.UnixNano())
	if errors.Is(err, tstorage.ErrNoDataPoints) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return points, nil
}

func makeBuckets(start time.Time, step time.Duration, count int) []MetricBucket {
	buckets := make([]MetricBucket, count)
	for i := range buckets {
		buckets[i].Start = start.Add(time.Duration(i) * step)
	}
	return buckets
}

func bucketIndex(timestamp int64, start time.Time, step time.Duration, count int) int {
	delta := timestamp - start.UnixNano()
	if delta < 0 {
		return -1
	}
	idx := int(delta / int64(step))
	if idx < 0 || idx >= count {
		return -1
	}
	return idx
}

type Int64Counter struct {
	Counter metric.Int64Counter
	storage *metricsStorage
	name    string
}

func (m *Int64Counter) Inc(ctx context.Context, v int64, attrs ...attribute.KeyValue) {
	normalized, labels := normalizeAttributes(attrs)
	m.Counter.Add(ctx, v, metric.WithAttributes(normalized...))
	m.storage.Insert(m.name, kindCounter, labels, float64(v), time.Now().UTC().UnixNano())
}

type Float64Histogram struct {
	Histogram metric.Float64Histogram
	storage   *metricsStorage
	name      string
}

func (m *Float64Histogram) Record(ctx context.Context, v float64, attrs ...attribute.KeyValue) {
	normalized, labels := normalizeAttributes(attrs)
	m.Histogram.Record(ctx, v, metric.WithAttributes(normalized...))
	now := time.Now().UTC().UnixNano()
	m.storage.Insert(m.name+".count", kindCounter, labels, 1, now)
	m.storage.Insert(m.name+".sum", kindCounter, labels, v, now)
	m.storage.registry.Register(m.name, kindHistogram, labels)
}

type Int64Histogram struct {
	Histogram metric.Int64Histogram
	storage   *metricsStorage
	name      string
}

func (m *Int64Histogram) Record(ctx context.Context, v int64, attrs ...attribute.KeyValue) {
	normalized, labels := normalizeAttributes(attrs)
	m.Histogram.Record(ctx, v, metric.WithAttributes(normalized...))
	now := time.Now().UTC().UnixNano()
	m.storage.Insert(m.name+".count", kindCounter, labels, 1, now)
	m.storage.Insert(m.name+".sum", kindCounter, labels, float64(v), now)
	m.storage.registry.Register(m.name, kindHistogram, labels)
}

type Int64UpDownCounter struct {
	Counter metric.Int64UpDownCounter
	storage *metricsStorage
	name    string

	mu      sync.Mutex
	current map[string]int64
}

func (m *Int64UpDownCounter) Add(ctx context.Context, v int64, attrs ...attribute.KeyValue) {
	normalized, labels := normalizeAttributes(attrs)
	m.Counter.Add(ctx, v, metric.WithAttributes(normalized...))

	key := labelsKey(labels)
	m.mu.Lock()
	if m.current == nil {
		m.current = make(map[string]int64)
	}
	m.current[key] += v
	current := m.current[key]
	m.mu.Unlock()

	m.storage.Insert(m.name, kindGauge, labels, float64(current), time.Now().UTC().UnixNano())
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

	storage *metricsStorage
}

func (m *ProxyMetrics) Series() ([]MetricSeries, error) {
	if m == nil || m.storage == nil {
		return nil, nil
	}
	return m.storage.Series(seriesStep, seriesPoints)
}

// InitMetrics creates a fully-initialised ProxyMetrics from the given OTel
// meter. Each metric is backed by both an OTel instrument and a labeled time
// series in tstorage for the web UI.
func InitMetrics(meter metric.Meter, storage tstorage.Storage, dataPath string) (*ProxyMetrics, error) {
	metricStore, err := newMetricsStorage(storage, dataPath)
	if err != nil {
		return nil, err
	}

	m := &ProxyMetrics{storage: metricStore}
	bindCounter := func(name string) Int64Counter {
		return Int64Counter{name: name, storage: metricStore}
	}
	bindFloatHistogram := func(name string) Float64Histogram {
		return Float64Histogram{name: name, storage: metricStore}
	}
	bindIntHistogram := func(name string) Int64Histogram {
		return Int64Histogram{name: name, storage: metricStore}
	}
	bindGauge := func(name string) Int64UpDownCounter {
		return Int64UpDownCounter{name: name, storage: metricStore}
	}

	m.RequestsTotal = bindCounter("proxy.http.requests.total")
	m.RequestDuration = bindFloatHistogram("proxy.http.request.duration.seconds")
	m.ActiveRequests = bindGauge("proxy.http.active.requests")
	m.RequestSizeBytes = bindIntHistogram("proxy.http.request.size.bytes")
	m.ResponseSizeBytes = bindIntHistogram("proxy.http.response.size.bytes")

	m.BackendRequestsTotal = bindCounter("proxy.backend.requests.total")
	m.BackendRequestDuration = bindFloatHistogram("proxy.backend.request.duration.seconds")
	m.BackendActiveRequests = bindGauge("proxy.backend.active.requests")

	m.AuthRequestsTotal = bindCounter("proxy.auth.requests.total")
	m.PolicyEvaluationsTotal = bindCounter("proxy.policy.evaluations.total")

	m.RouteMatchesTotal = bindCounter("proxy.route.matches.total")
	m.RouteNoMatchTotal = bindCounter("proxy.route.no_match.total")

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
