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

	policyv1 "github.com/amimof/multikube/api/policy/v1"
)

type PolicyService struct {
	Repo     *repository.Repo[*policyv1.Policy]
	mu       sync.Mutex
	Exchange *events.Exchange
	Logger   logger.Logger
}

func (l *PolicyService) Get(ctx context.Context, id keys.ID) (*policyv1.Policy, error) {
	ctx, span := tracer.Start(ctx, "policy.Get", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return l.Repo.Get(ctx, id)
}

func (l *PolicyService) List(ctx context.Context, limit int32) ([]*policyv1.Policy, error) {
	ctx, span := tracer.Start(ctx, "policy.List")
	defer span.End()

	return l.Repo.List(ctx, limit)
}

func (l *PolicyService) Create(ctx context.Context, policy *policyv1.Policy) (*policyv1.Policy, error) {
	ctx, span := tracer.Start(ctx, "policy.Create")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	newPolicy, err := l.Repo.Create(ctx, policy)
	if err != nil {
		l.Logger.Error("error creating policy", "error", err, "name", policy.GetMeta().GetName())
		return nil, err
	}

	err = l.Exchange.Forward(ctx, events.NewEvent(events.PolicyCreate, policy))
	if err != nil {
		l.Logger.Error("error publishing policy create event", "error", err, "name", newPolicy.GetMeta().GetName())
		return nil, err
	}

	return newPolicy, nil
}

func (l *PolicyService) Delete(ctx context.Context, id keys.ID) error {
	ctx, span := tracer.Start(ctx, "policy.Delete")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	policy, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	err = l.Repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = l.Exchange.Forward(ctx, events.NewEvent(events.PolicyDelete, policy))
	if err != nil {
		l.Logger.Error("error publishing policy delete event", "error", err, "name", policy.GetMeta().GetName())
		return err
	}

	return nil
}

func (l *PolicyService) Patch(ctx context.Context, id keys.ID, patch *policyv1.Policy) error {
	ctx, span := tracer.Start(ctx, "policy.Patch")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	existing, err := l.Repo.Get(ctx, id)
	if err != nil {
		l.Logger.Error("error getting policy", "error", err, "name", patch.GetMeta().GetName())
		return err
	}

	genFieldMask, err := protoutils.GenerateFieldMask(existing, patch)
	if err != nil {
		return err
	}

	maskedUpdate, err := protoutils.ApplyFieldMaskToNewMessage(patch, genFieldMask)
	if err != nil {
		return err
	}

	updated := maskedUpdate.(*policyv1.Policy)
	existing = protoutils.StrategicMerge(existing, updated)

	policy, err := l.Repo.Update(ctx, id, existing)
	if err != nil {
		l.Logger.Error("error updating policy", "error", err, "name", existing.GetMeta().GetName())
		return err
	}

	changed, err := protoutils.SpecEqual(existing.GetConfig(), policy.GetConfig())
	if err != nil {
		return err
	}

	if changed {
		err = l.Exchange.Forward(ctx, events.NewEvent(events.PolicyPatch, policy))
		if err != nil {
			l.Logger.Error("error publishing policy patch event", "error", err, "name", existing.GetMeta().GetName())
			return err
		}
	}

	return nil
}

func (l *PolicyService) Update(ctx context.Context, id keys.ID, policy *policyv1.Policy) error {
	ctx, span := tracer.Start(ctx, "policy.Update")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	existingPolicy, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	updated, err := l.Repo.Update(ctx, id, policy)
	if err != nil {
		l.Logger.Error("error updating policy", "error", err, "name", policy.GetMeta().GetName())
		return err
	}

	changed, err := protoutils.SpecEqual(existingPolicy.GetConfig(), updated.GetConfig())
	if err != nil {
		return err
	}

	if changed {
		l.Logger.Debug("policy was updated, emitting event to listeners", "event", "PolicyUpdate", "name", updated.GetMeta().GetName())
		err = l.Exchange.Forward(ctx, events.NewEvent(events.PolicyUpdate, updated))
		if err != nil {
			l.Logger.Error("error publishing policy update event", "error", err, "name", updated.GetMeta().GetName())
			return err
		}
	}

	return nil
}
