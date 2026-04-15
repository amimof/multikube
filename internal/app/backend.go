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

	backendv1 "github.com/amimof/multikube/api/backend/v1"
)

type BackendService struct {
	Repo     *repository.Repo[*backendv1.Backend]
	mu       sync.Mutex
	Exchange *events.Exchange
	Logger   logger.Logger
}

func applyMaskedUpdateBackend(dst, src *backendv1.BackendStatus, mask *fieldmaskpb.FieldMask) error {
	if mask == nil || len(mask.Paths) == 0 {
		return status.Error(codes.InvalidArgument, "update_mask is required")
	}
	for _, p := range mask.Paths {
		switch p {
		case "target_statuses":
			if src.TargetStatuses == nil {
				continue
			}
			if dst.TargetStatuses == nil {
				dst.TargetStatuses = make(map[string]*backendv1.TargetStatus, len(src.TargetStatuses))
			}
			for k, srcStatus := range src.TargetStatuses {
				if srcStatus == nil {
					continue
				}
				if existing, ok := dst.TargetStatuses[k]; ok && existing != nil {
					// Merge into the existing entry instead of replacing it.
					proto.Merge(existing, srcStatus)
				} else {
					// Clone so dst does not alias src memory.
					dst.TargetStatuses[k] = proto.Clone(srcStatus).(*backendv1.TargetStatus)
				}
			}

		default:
			return fmt.Errorf("unknown mask path %q", p)
		}
	}
	return nil
}

func (l *BackendService) Get(ctx context.Context, id keys.ID) (*backendv1.Backend, error) {
	ctx, span := tracer.Start(ctx, "volume.Get", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return l.Repo.Get(ctx, id)
}

func (l *BackendService) List(ctx context.Context, limit int32) ([]*backendv1.Backend, error) {
	ctx, span := tracer.Start(ctx, "volume.List")
	defer span.End()

	// Get volumes from repo
	return l.Repo.List(ctx, limit)
}

func (l *BackendService) Create(ctx context.Context, be *backendv1.Backend) (*backendv1.Backend, error) {
	ctx, span := tracer.Start(ctx, "volume.Create")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Default impersionation config
	if be.GetConfig().GetImpersonationConfig() == nil {
		be.GetConfig().ImpersonationConfig = &backendv1.ImpersonationConfig{
			Name:          "default",
			Enabled:       true,
			UsernameClaim: "sub",
			GroupsClaim:   "groups",
		}
	}

	// Create volume in repo
	newVolume, err := l.Repo.Create(ctx, be)
	if err != nil {
		l.Logger.Error("error creating volume", "error", err, "name", newVolume.GetMeta().GetName())
		return nil, err
	}

	// Publish event that volume is created
	err = l.Exchange.Forward(ctx, events.NewEvent(events.BackendCreate, be))
	if err != nil {
		l.Logger.Error("error publishing volume create event", "error", err, "name", newVolume.GetMeta().GetName())
		return nil, err
	}

	return newVolume, nil
}

// Delete publishes a delete request and the subscribers are responsible for deleting resources.
// Once they do, they will update there resource with the status Deleted
func (l *BackendService) Delete(ctx context.Context, id keys.ID) error {
	ctx, span := tracer.Start(ctx, "volume.Delete")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	volume, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	err = l.Repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = l.Exchange.Forward(ctx, events.NewEvent(events.BackendDelete, volume))
	if err != nil {
		l.Logger.Error("error publishing volume delete event", "error", err, "name", volume.GetMeta().GetName())
		return err
	}

	return nil
}

func (l *BackendService) Patch(ctx context.Context, id keys.ID, patch *backendv1.Backend) error {
	ctx, span := tracer.Start(ctx, "volume.Patch")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get existing volume from repo
	existing, err := l.Repo.Get(ctx, id)
	if err != nil {
		l.Logger.Error("error getting volume", "error", err, "name", patch.GetMeta().GetName())
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

	updated := maskedUpdate.(*backendv1.Backend)
	existing = protoutils.StrategicMerge(existing, updated)

	// Update the volume
	volume, err := l.Repo.Update(ctx, id, existing)
	if err != nil {
		l.Logger.Error("error updating volume", "error", err, "name", existing.GetMeta().GetName())
		return err
	}

	equal, err := protoutils.SpecEqual(existing.GetConfig(), volume.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if !equal {
		err = l.Exchange.Forward(ctx, events.NewEvent(events.BackendPatch, volume))
		if err != nil {
			l.Logger.Error("error publishing volume patch event", "error", err, "name", existing.GetMeta().GetName())
			return err
		}
	}

	return nil
}

func (l *BackendService) Update(ctx context.Context, id keys.ID, volume *backendv1.Backend) error {
	ctx, span := tracer.Start(ctx, "volume.Update")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get the existing volume before updating so we can compare specs
	existingVolume, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Update the volume
	updated, err := l.Repo.Update(ctx, id, volume)
	if err != nil {
		l.Logger.Error("error updating volume", "error", err, "name", volume.GetMeta().GetName())
		return err
	}

	equal, err := protoutils.SpecEqual(existingVolume.GetConfig(), updated.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if !equal {
		l.Logger.Debug("volume was updated, emitting event to listeners", "event", "VolumeUpdate", "name", updated.GetMeta().GetName())
		err = l.Exchange.Forward(ctx, events.NewEvent(events.BackendUpdate, updated))
		if err != nil {
			l.Logger.Error("error publishing volume update event", "error", err, "name", updated.GetMeta().GetName())
			return err
		}
	}

	return nil
}

// UpdateStatus implements [routesv1.RouteServieClient]
func (l *BackendService) UpdateStatus(ctx context.Context, id keys.ID, st *backendv1.BackendStatus, mask ...string) error {
	ctx, span := tracer.Start(ctx, "route.UpdateStatus")
	defer span.End()

	// Get the existing route before updating so we can compare specs
	existingBackend, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Apply mask safely
	base := &backendv1.BackendStatus{}
	if existingBackend.Status != nil {
		base = proto.Clone(existingBackend.Status).(*backendv1.BackendStatus)
	}
	if err := applyMaskedUpdateBackend(base, st, &fieldmaskpb.FieldMask{Paths: mask}); err != nil {
		return status.Errorf(codes.InvalidArgument, "bad mask: %v", err)
	}

	existingBackend.Status = base

	if _, err := l.Repo.Update(ctx, id, existingBackend); err != nil {
		return err
	}

	return nil
}
