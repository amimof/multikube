package app

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/trace"

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

// func applyMaskedUpdateVolume(dst, src *backendv1.Status, mask *fieldmaskpb.FieldMask) error {
// 	if mask == nil || len(mask.Paths) == 0 {
// 		return status.Error(codes.InvalidArgument, "update_mask is required")
// 	}
//
// 	for _, p := range mask.Paths {
// 		switch p {
// 		case "controllers":
// 			if src.Controllers == nil {
// 				continue
// 			}
// 			dst.Controllers = src.Controllers
// 		default:
// 			return fmt.Errorf("unknown mask path %q", p)
// 		}
// 	}
//
// 	return nil
// }

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

func (l *BackendService) Create(ctx context.Context, volume *backendv1.Backend) (*backendv1.Backend, error) {
	ctx, span := tracer.Start(ctx, "volume.Create")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Create volume in repo
	newVolume, err := l.Repo.Create(ctx, volume)
	if err != nil {
		l.Logger.Error("error creating volume", "error", err, "name", newVolume.GetMeta().GetName())
		return nil, err
	}

	// Publish event that volume is created
	err = l.Exchange.Forward(ctx, events.NewEvent(events.BackendCreate, volume))
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

	changed, err := protoutils.SpecEqual(existing.GetConfig(), volume.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if changed {
		err = l.Exchange.Forward(ctx, events.NewEvent(events.BackendPatch, volume))
		if err != nil {
			l.Logger.Error("error publishing volume patch event", "error", err, "name", existing.GetMeta().GetName())
			return err
		}
	}

	return nil
}

// UpdateStatus implements volumes.VolumeServieClient.
// func (l *VolumeService) UpdateStatus(ctx context.Context, id keys.ID, st *backendv1.Status, mask ...string) error {
// 	l.mu.Lock()
// 	defer l.mu.Unlock()
//
// 	ctx, span := tracer.Start(ctx, "volume.UpdateStatus")
// 	defer span.End()
//
// 	// Get the existing volume before updating so we can compare specs
// 	existingVolume, err := l.Repo.Get(ctx, id)
// 	if err != nil {
// 		return err
// 	}
//
// 	// Apply mask safely
// 	base := proto.Clone(existingVolume.Status).(*backendv1.Status)
// 	if err := applyMaskedUpdateVolume(base, st, &fieldmaskpb.FieldMask{Paths: mask}); err != nil {
// 		return status.Errorf(codes.InvalidArgument, "bad mask: %v", err)
// 	}
//
// 	existingVolume.Status = base
//
// 	if _, err := l.Repo.Update(ctx, id, existingVolume); err != nil {
// 		return err
// 	}
//
// 	return nil
// }

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

	changed, err := protoutils.SpecEqual(existingVolume.GetConfig(), updated.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if changed {
		l.Logger.Debug("volume was updated, emitting event to listeners", "event", "VolumeUpdate", "name", updated.GetMeta().GetName())
		err = l.Exchange.Forward(ctx, events.NewEvent(events.BackendUpdate, updated))
		if err != nil {
			l.Logger.Error("error publishing volume update event", "error", err, "name", updated.GetMeta().GetName())
			return err
		}
	}

	return nil
}

// func (l *VolumeService) Condition(ctx context.Context, id keys.ID, req *typesv1.ConditionRequest) error {
// 	st := &backendv1.Status{
// 		Conditions: req.GetConditions(),
// 	}
//
// 	err := l.UpdateStatus(ctx, id, st, "conditions")
// 	if err != nil {
// 		return err
// 	}
//
// 	err = l.Exchange.Publish(ctx, events.NewEvent(events.ConditionReported, req))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }
