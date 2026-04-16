package app

import (
	"context"

	"github.com/amimof/multikube/internal/infra"
	"github.com/amimof/multikube/pkg/audit"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/logger"

	auditv1 "github.com/amimof/multikube/api/audit/v1"
)

type AuditService struct {
	Logger   logger.Logger
	Exchange *events.Exchange
	Manager  infra.SessionManager
	Store    audit.Store
}

// AuditLog streams audit entries
func (n *AuditService) AuditLog(ctx context.Context, req *auditv1.AuditLogRequest, stream auditv1.AuditService_AuditLogServer) error {
	events, err := n.Store.Subscribe(ctx, 0, true)
	if err != nil {
		return err
	}

	// Send log entry for each line that comes in from the line channel. After x amount of time anf if no lines are
	// received, exit out. This is a blocking operation.
	for {
		select {
		case <-ctx.Done():
			n.Logger.Debug("log stream cancelled")
			return ctx.Err()
		case line, ok := <-events:
			if !ok {
				n.Logger.Debug("log stream completed")
				return nil
			}

			// Send the line as log entry to the server
			if err := stream.Send(&auditv1.AuditLogResponse{
				SessionId: req.GetSessionId(),
				Entry:     line,
			}); err != nil {
				n.Logger.Error(
					"error pushing log entry",
					"error", err,
				)
				return err
			}
		}
	}
}
