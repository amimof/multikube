package grpc

import (
	"io"
	"net"
	"time"

	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/amimof/multikube/pkg/repository"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

func init() {
	grpclog.SetLoggerV2(grpclog.NewLoggerV2(io.Discard, io.Discard, io.Discard))
}

type GrpcService interface {
	Register(*grpc.Server)
}

type Server struct {
	grpcOpts   []grpc.ServerOption
	grpcServer *grpc.Server
	logger     logger.Logger
	exchange   *events.Exchange
	db         repository.DB
}

type NewServerOption func(*Server)

func WithLogger(lgr logger.Logger) NewServerOption {
	return func(s *Server) {
		s.logger = lgr
	}
}

func WithExchange(e *events.Exchange) NewServerOption {
	return func(s *Server) {
		s.exchange = e
	}
}

func WithDB(d repository.DB) NewServerOption {
	return func(s *Server) {
		s.db = d
	}
}

func WithGrpcOption(opts ...grpc.ServerOption) NewServerOption {
	return func(s *Server) {
		s.grpcOpts = append(s.grpcOpts, opts...)
	}
}

func NewServer(opts ...NewServerOption) (*Server, error) {
	grpcOpts := []grpc.ServerOption{
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             15 * time.Second, // Min time between pings
			PermitWithoutStream: true,
		}),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time:    15 * time.Second,
			Timeout: 10 * time.Second,
		}),
	}

	// Setup server
	server := &Server{
		grpcOpts: grpcOpts,
		logger:   logger.ConsoleLogger{},
	}

	// Apply options
	for _, opt := range opts {
		opt(server)
	}

	server.grpcServer = grpc.NewServer(server.grpcOpts...)
	server.grpcServer.RegisterService(&grpc_health_v1.Health_ServiceDesc, health.NewServer())

	reflection.Register(server.grpcServer)

	return server, nil
}

func (s *Server) Serve(lis net.Listener) error {
	return s.grpcServer.Serve(lis)
}

func (s *Server) Shutdown() {
	s.grpcServer.GracefulStop()
}

func (s *Server) ForceShutdown() {
	s.grpcServer.Stop()
}

func (s *Server) RegisterService(svcs ...GrpcService) {
	for _, svc := range svcs {
		svc.Register(s.grpcServer)
	}
}
