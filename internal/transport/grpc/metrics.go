package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/amimof/multikube/internal/app"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	metricsv1 "github.com/amimof/multikube/api/metrics/v1"
	proxy "github.com/amimof/multikube/pkg/proxyv2"
)

type MetricsService struct {
	metricsv1.UnimplementedMetricsServiceServer
	app *app.MetricsService
}

func (n *MetricsService) Register(server *grpc.Server) {
	server.RegisterService(&metricsv1.MetricsService_ServiceDesc, n)
}

func (n *MetricsService) RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return metricsv1.RegisterMetricsServiceHandler(ctx, mux, conn)
}

func (s *MetricsService) Get(ctx context.Context, req *metricsv1.GetRequest) (*metricsv1.GetResponse, error) {
	m := s.app.Metrics

	return &metricsv1.GetResponse{
		// Request metrics
		RequestsTotal:     counterMetricProto(&m.RequestsTotal),
		RequestDuration:   histogramMetricProto(&m.RequestDuration),
		ActiveRequests:    gaugeMetricProto(&m.ActiveRequests),
		RequestSizeBytes:  int64HistogramMetricProto(&m.RequestSizeBytes),
		ResponseSizeBytes: int64HistogramMetricProto(&m.ResponseSizeBytes),

		// Backend metrics
		BackendRequestsTotal:   counterMetricProto(&m.BackendRequestsTotal),
		BackendRequestDuration: histogramMetricProto(&m.BackendRequestDuration),
		BackendActiveRequests:  gaugeMetricProto(&m.BackendActiveRequests),

		// Auth metrics
		AuthRequestsTotal:      counterMetricProto(&m.AuthRequestsTotal),
		PolicyEvaluationsTotal: counterMetricProto(&m.PolicyEvaluationsTotal),

		// Route metrics
		RouteMatchesTotal: counterMetricProto(&m.RouteMatchesTotal),
		RouteNoMatchTotal: counterMetricProto(&m.RouteNoMatchTotal),
	}, nil
}

func NewMetricsService(app *app.MetricsService) *MetricsService {
	return &MetricsService{app: app}
}

func counterMetricProto(c *proxy.Int64Counter) *metricsv1.CounterMetric {
	buckets := c.SnapshotBuckets()
	series := make([]*metricsv1.Int64Series, len(buckets))
	for i, b := range buckets {
		series[i] = &metricsv1.Int64Series{
			Start: timestamppb.New(b.Start),
			Value: b.Value,
		}
	}
	return &metricsv1.CounterMetric{
		Total:   c.Load(),
		Buckets: series,
	}
}

func histogramMetricProto(h *proxy.Float64Histogram) *metricsv1.HistogramMetric {
	buckets := h.SnapshotBuckets()
	series := make([]*metricsv1.Float64Series, len(buckets))
	for i, b := range buckets {
		series[i] = &metricsv1.Float64Series{
			Start: timestamppb.New(b.Start),
			Count: b.Count,
			Sum:   b.Sum,
		}
	}
	return &metricsv1.HistogramMetric{
		Buckets: series,
	}
}

func int64HistogramMetricProto(h *proxy.Int64Histogram) *metricsv1.Int64HistogramMetric {
	buckets := h.SnapshotBuckets()
	series := make([]*metricsv1.Int64HistogramSeries, len(buckets))
	for i, b := range buckets {
		series[i] = &metricsv1.Int64HistogramSeries{
			Start: timestamppb.New(b.Start),
			Count: b.Count,
			Sum:   b.Sum,
		}
	}
	return &metricsv1.Int64HistogramMetric{
		Buckets: series,
	}
}

func gaugeMetricProto(g *proxy.Int64UpDownCounter) *metricsv1.GaugeMetric {
	buckets := g.SnapshotBuckets()
	series := make([]*metricsv1.GaugeSeries, len(buckets))
	for i, b := range buckets {
		series[i] = &metricsv1.GaugeSeries{
			Start: timestamppb.New(b.Start),
			Max:   b.Max,
		}
	}
	return &metricsv1.GaugeMetric{
		Buckets: series,
	}
}
