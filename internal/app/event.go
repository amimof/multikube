package app

import (
	"context"

	"github.com/amimof/multikube/internal/infra"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/logger"

	eventv1 "github.com/amimof/multikube/api/event/v1"
)

type EventService struct {
	Logger   logger.Logger
	Exchange *events.Exchange
	Manager  infra.SessionManager
}

// Publish implements events.EventServiceClient.
func (n *EventService) Publish(ctx context.Context, ev *eventv1.Envelope) (*eventv1.Envelope, error) {
	ctx, span := tracer.Start(ctx, "event.Publish")
	defer span.End()

	err := n.Exchange.Publish(ctx, ev)
	if err != nil {
		return nil, err
	}

	return ev, nil
}

// Subscribe implements events.EventServiceClient.
func (n *EventService) Subscribe(ctx context.Context, in infra.NodeConnectInput) (infra.Session, error) {
	return n.Manager.Connect(ctx, in)
}
