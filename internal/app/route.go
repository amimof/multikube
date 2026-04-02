package app

import (
	"context"
	"fmt"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/amimof/multikube/pkg/protoutils"
	"github.com/amimof/multikube/pkg/repository"

	routev1 "github.com/amimof/multikube/api/route/v1"
)

type RouteService struct {
	Repo     *repository.Repo[*routev1.Route]
	mu       sync.Mutex
	Exchange *events.Exchange
	Logger   logger.Logger
}

func applyMaskedUpdateRoute(dst, src *routev1.RouteStatus, mask *fieldmaskpb.FieldMask) error {
	if mask == nil || len(mask.Paths) == 0 {
		return status.Error(codes.InvalidArgument, "update_mask is required")
	}
	for _, p := range mask.Paths {
		switch p {
		case "phase":
			if src.Phase == nil {
				continue
			}
			dst.Phase = src.Phase
		case "reason":
			if src.Reason == nil {
				continue
			}
			dst.Reason = src.Reason
		case "last_transition_time":
			if src.LastTransitionTime == nil {
				continue
			}
			dst.LastTransitionTime = src.LastTransitionTime
		default:
			return fmt.Errorf("unknown mask path %q", p)
		}
	}
	return nil
}

func (l *RouteService) Get(ctx context.Context, id keys.ID) (*routev1.Route, error) {
	ctx, span := tracer.Start(ctx, "route.Get", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return l.Repo.Get(ctx, id)
}

func (l *RouteService) List(ctx context.Context, limit int32) ([]*routev1.Route, error) {
	ctx, span := tracer.Start(ctx, "route.List")
	defer span.End()

	// Get routes from repo
	return l.Repo.List(ctx, limit)
}

func (l *RouteService) Create(ctx context.Context, route *routev1.Route) (*routev1.Route, error) {
	ctx, span := tracer.Start(ctx, "route.Create")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Create route in repo
	newRoute, err := l.Repo.Create(ctx, route)
	if err != nil {
		l.Logger.Error("error creating route", "error", err, "name", newRoute.GetMeta().GetName())
		return nil, err
	}

	// Publish event that route is created
	err = l.Exchange.Forward(ctx, events.NewEvent(events.RouteCreate, route))
	if err != nil {
		l.Logger.Error("error publishing route create event", "error", err, "name", newRoute.GetMeta().GetName())
		return nil, err
	}

	return newRoute, nil
}

// Delete publishes a delete request and the subscribers are responsible for deleting resources.
// Once they do, they will update there resource with the status Deleted
func (l *RouteService) Delete(ctx context.Context, id keys.ID) error {
	ctx, span := tracer.Start(ctx, "route.Delete")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	route, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	err = l.Repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = l.Exchange.Forward(ctx, events.NewEvent(events.RouteDelete, route))
	if err != nil {
		l.Logger.Error("error publishing route delete event", "error", err, "name", route.GetMeta().GetName())
		return err
	}

	return nil
}

func (l *RouteService) Patch(ctx context.Context, id keys.ID, patch *routev1.Route) error {
	ctx, span := tracer.Start(ctx, "route.Patch")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get existing route from repo
	existing, err := l.Repo.Get(ctx, id)
	if err != nil {
		l.Logger.Error("error getting route", "error", err, "name", patch.GetMeta().GetName())
		return err
	}

	// Generate field mask
	genFieldMask, err := protoutils.GenerateFieldMask(existing, patch)
	if err != nil {
		return err
	}

	// Handle partial update
	maskedUpdate, err := protoutils.ApplyFieldMaskToNewMessage(patch, genFieldMask)
	if err != nil {
		return err
	}

	updated := maskedUpdate.(*routev1.Route)
	existing = protoutils.StrategicMerge(existing, updated)

	// Update the route
	route, err := l.Repo.Update(ctx, id, existing)
	if err != nil {
		l.Logger.Error("error updating route", "error", err, "name", existing.GetMeta().GetName())
		return err
	}

	changed, err := protoutils.SpecEqual(existing.GetConfig(), route.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if changed {
		err = l.Exchange.Forward(ctx, events.NewEvent(events.RoutePatch, route))
		if err != nil {
			l.Logger.Error("error publishing route patch event", "error", err, "name", existing.GetMeta().GetName())
			return err
		}
	}

	return nil
}

func (l *RouteService) Update(ctx context.Context, id keys.ID, route *routev1.Route) error {
	ctx, span := tracer.Start(ctx, "route.Update")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get the existing route before updating so we can compare specs
	existingRoute, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Update the route
	updated, err := l.Repo.Update(ctx, id, route)
	if err != nil {
		l.Logger.Error("error updating route", "error", err, "name", route.GetMeta().GetName())
		return err
	}

	changed, err := protoutils.SpecEqual(existingRoute.GetConfig(), updated.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if changed {
		l.Logger.Debug("route was updated, emitting event to listeners", "event", "RouteUpdate", "name", updated.GetMeta().GetName())
		err = l.Exchange.Forward(ctx, events.NewEvent(events.RouteUpdate, updated))
		if err != nil {
			l.Logger.Error("error publishing route update event", "error", err, "name", updated.GetMeta().GetName())
			return err
		}
	}

	return nil
}

// UpdateStatus implements [routesv1.RouteServieClient]
func (l *RouteService) UpdateStatus(ctx context.Context, id keys.ID, st *routev1.RouteStatus, mask ...string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	ctx, span := tracer.Start(ctx, "route.UpdateStatus")
	defer span.End()

	// Get the existing route before updating so we can compare specs
	existingRoute, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Apply mask safely
	base := &routev1.RouteStatus{}
	if existingRoute.Status != nil {
		base = proto.Clone(existingRoute.Status).(*routev1.RouteStatus)
	}
	if err := applyMaskedUpdateRoute(base, st, &fieldmaskpb.FieldMask{Paths: mask}); err != nil {
		return status.Errorf(codes.InvalidArgument, "bad mask: %v", err)
	}

	existingRoute.Status = base

	if _, err := l.Repo.Update(ctx, id, existingRoute); err != nil {
		return err
	}

	return nil
}
