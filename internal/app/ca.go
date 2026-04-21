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

	cav1 "github.com/amimof/multikube/api/ca/v1"
)

type CertificateAuthorityService struct {
	Repo     *repository.Repo[*cav1.CertificateAuthority]
	mu       sync.Mutex
	Exchange *events.Exchange
	Logger   logger.Logger
}

func (l *CertificateAuthorityService) Get(ctx context.Context, id keys.ID) (*cav1.CertificateAuthority, error) {
	ctx, span := tracer.Start(ctx, "ca.Get", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return l.Repo.Get(ctx, id)
}

func (l *CertificateAuthorityService) List(ctx context.Context, limit int32) ([]*cav1.CertificateAuthority, error) {
	ctx, span := tracer.Start(ctx, "ca.List")
	defer span.End()

	// Get cas from repo
	return l.Repo.List(ctx, limit)
}

func (l *CertificateAuthorityService) Create(ctx context.Context, ca *cav1.CertificateAuthority) (*cav1.CertificateAuthority, error) {
	ctx, span := tracer.Start(ctx, "ca.Create")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Ensure status field
	if err := EnsureCertInStatus(ca.GetConfig().GetCertificateData(), ca); err != nil {
		l.Logger.Error("error ensuring certificate status fields", "error", err, "name", ca.GetMeta().GetName())
	}

	if ca.GetConfig().Enabled == nil {
		ca.GetConfig().Enabled = new(true)
	}

	// Create ca in repo
	newCert, err := l.Repo.Create(ctx, ca)
	if err != nil {
		l.Logger.Error("error creating ca", "error", err, "name", newCert.GetMeta().GetName())
		return nil, err
	}

	// Publish event that ca is created
	err = l.Exchange.Forward(ctx, events.NewEvent(events.CertificateAuthorityCreate, ca))
	if err != nil {
		l.Logger.Error("error publishing ca create event", "error", err, "name", newCert.GetMeta().GetName())
		return nil, err
	}

	return newCert, nil
}

func (l *CertificateAuthorityService) Delete(ctx context.Context, id keys.ID) error {
	ctx, span := tracer.Start(ctx, "ca.Delete")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	ca, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	err = l.Repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = l.Exchange.Forward(ctx, events.NewEvent(events.CertificateAuthorityDelete, ca))
	if err != nil {
		l.Logger.Error("error publishing ca delete event", "error", err, "name", ca.GetMeta().GetName())
		return err
	}

	return nil
}

func (l *CertificateAuthorityService) Patch(ctx context.Context, id keys.ID, patch *cav1.CertificateAuthority) error {
	ctx, span := tracer.Start(ctx, "ca.Patch")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get existing ca from repo
	existing, err := l.Repo.Get(ctx, id)
	if err != nil {
		l.Logger.Error("error getting ca", "error", err, "name", patch.GetMeta().GetName())
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

	updated := maskedUpdate.(*cav1.CertificateAuthority)
	existing = protoutils.StrategicMerge(existing, updated)

	// Ensure status field
	if err := EnsureCertInStatus(existing.GetConfig().GetCertificateData(), existing); err != nil {
		l.Logger.Error("error ensuring certificate status fields", "error", err, "name", existing.GetMeta().GetName())
	}

	// Update the ca
	ca, err := l.Repo.Update(ctx, id, existing)
	if err != nil {
		l.Logger.Error("error updating ca", "error", err, "name", existing.GetMeta().GetName())
		return err
	}

	equal, err := protoutils.SpecEqual(existing.GetConfig(), ca.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if !equal {
		err = l.Exchange.Forward(ctx, events.NewEvent(events.CertificateAuthorityPatch, ca))
		if err != nil {
			l.Logger.Error("error publishing ca patch event", "error", err, "name", existing.GetMeta().GetName())
			return err
		}
	}

	return nil
}

func (l *CertificateAuthorityService) Update(ctx context.Context, id keys.ID, ca *cav1.CertificateAuthority) error {
	ctx, span := tracer.Start(ctx, "ca.Update")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get the existing ca before updating so we can compare specs
	existingCert, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Ensure status field
	if err := EnsureCertInStatus(ca.GetConfig().GetCertificateData(), ca); err != nil {
		l.Logger.Error("error ensuring certificate status fields", "error", err, "name", ca.GetMeta().GetName())
	}

	if ca.GetConfig().Enabled == nil {
		ca.GetConfig().Enabled = new(true)
	}

	// Update the ca
	updated, err := l.Repo.Update(ctx, id, ca)
	if err != nil {
		l.Logger.Error("error updating ca", "error", err, "name", ca.GetMeta().GetName())
		return err
	}

	equal, err := protoutils.SpecEqual(existingCert.GetConfig(), updated.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if !equal {
		l.Logger.Debug("ca was updated, emitting event to listeners", "event", "CertificateAuthorityUpdate", "name", updated.GetMeta().GetName())
		err = l.Exchange.Forward(ctx, events.NewEvent(events.CertificateAuthorityUpdate, updated))
		if err != nil {
			l.Logger.Error("error publishing ca update event", "error", err, "name", updated.GetMeta().GetName())
			return err
		}
	}

	return nil
}
