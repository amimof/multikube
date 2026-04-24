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
	series, err := s.app.MetricsSeries(ctx)
	if err != nil {
		return nil, err
	}

	response := &metricsv1.GetResponse{
		Series: make([]*metricsv1.MetricSeries, 0, len(series)),
	}
	for _, metricSeries := range series {
		response.Series = append(response.Series, metricSeriesProto(metricSeries))
	}
	return response, nil
}

func NewMetricsService(app *app.MetricsService) *MetricsService {
	return &MetricsService{app: app}
}

func metricSeriesProto(series proxy.MetricSeries) *metricsv1.MetricSeries {
	labels := make([]*metricsv1.Label, len(series.Labels))
	for i, label := range series.Labels {
		labels[i] = &metricsv1.Label{Name: label.Name, Value: label.Value}
	}

	buckets := make([]*metricsv1.MetricBucket, len(series.Buckets))
	for i, bucket := range series.Buckets {
		buckets[i] = &metricsv1.MetricBucket{
			Start: timestamppb.New(bucket.Start),
			Value: bucket.Value,
			Count: bucket.Count,
			Sum:   bucket.Sum,
		}
	}

	return &metricsv1.MetricSeries{
		Metric:  series.Metric,
		Kind:    series.Kind,
		Labels:  labels,
		Buckets: buckets,
	}
}
