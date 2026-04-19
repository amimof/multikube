package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/amimof/multikube/internal/app"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	metricsv1 "github.com/amimof/multikube/api/metrics/v1"
	// nodesv1 "github.com/amimof/multikube/api/services/nodes/v1"
	// "github.com/amimof/multikube/api/types/v1"
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
	return &metricsv1.GetResponse{}, nil
}

func NewMetricsService(app *app.MetricsService) *MetricsService {
	return &MetricsService{app: app}
}
