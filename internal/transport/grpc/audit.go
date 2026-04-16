package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc"

	"github.com/amimof/multikube/internal/app"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	auditv1 "github.com/amimof/multikube/api/audit/v1"
	// nodesv1 "github.com/amimof/multikube/api/services/nodes/v1"
	// "github.com/amimof/multikube/api/types/v1"
)

var ErrClientExists = errors.New("client already exists")

type NewServiceOption func(s *AuditService)

type AuditService struct {
	auditv1.UnimplementedAuditServiceServer
	app *app.AuditService
}

func (n *AuditService) Register(server *grpc.Server) {
	server.RegisterService(&auditv1.AuditService_ServiceDesc, n)
}

func (n *AuditService) RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return auditv1.RegisterAuditServiceHandler(ctx, mux, conn)
}

func (s *AuditService) AuditLog(req *auditv1.AuditLogRequest, stream auditv1.AuditService_AuditLogServer) error {
	return s.app.AuditLog(stream.Context(), req, stream)
}

func NewAuditService(app *app.AuditService) *AuditService {
	return &AuditService{app: app}
}
