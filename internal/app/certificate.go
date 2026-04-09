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

	certv1 "github.com/amimof/multikube/api/certificate/v1"
)

type CertificateService struct {
	Repo     *repository.Repo[*certv1.Certificate]
	mu       sync.Mutex
	Exchange *events.Exchange
	Logger   logger.Logger
}

func (l *CertificateService) Get(ctx context.Context, id keys.ID) (*certv1.Certificate, error) {
	ctx, span := tracer.Start(ctx, "certificate.Get", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return l.Repo.Get(ctx, id)
}

func (l *CertificateService) List(ctx context.Context, limit int32) ([]*certv1.Certificate, error) {
	ctx, span := tracer.Start(ctx, "certificate.List")
	defer span.End()

	// Get certificates from repo
	return l.Repo.List(ctx, limit)
}

func (l *CertificateService) Create(ctx context.Context, certificate *certv1.Certificate) (*certv1.Certificate, error) {
	ctx, span := tracer.Start(ctx, "certificate.Create")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Create certificate in repo
	newCertificate, err := l.Repo.Create(ctx, certificate)
	if err != nil {
		l.Logger.Error("error creating certificate", "error", err, "name", newCertificate.GetMeta().GetName())
		return nil, err
	}

	// Publish event that certificate is created
	err = l.Exchange.Forward(ctx, events.NewEvent(events.CertificateCreate, certificate))
	if err != nil {
		l.Logger.Error("error publishing certificate create event", "error", err, "name", newCertificate.GetMeta().GetName())
		return nil, err
	}

	return newCertificate, nil
}

func (l *CertificateService) Delete(ctx context.Context, id keys.ID) error {
	ctx, span := tracer.Start(ctx, "certificate.Delete")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	certificate, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	err = l.Repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = l.Exchange.Forward(ctx, events.NewEvent(events.CertificateDelete, certificate))
	if err != nil {
		l.Logger.Error("error publishing certificate delete event", "error", err, "name", certificate.GetMeta().GetName())
		return err
	}

	return nil
}

func (l *CertificateService) Patch(ctx context.Context, id keys.ID, patch *certv1.Certificate) error {
	ctx, span := tracer.Start(ctx, "certificate.Patch")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get existing certificate from repo
	existing, err := l.Repo.Get(ctx, id)
	if err != nil {
		l.Logger.Error("error getting certificate", "error", err, "name", patch.GetMeta().GetName())
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

	updated := maskedUpdate.(*certv1.Certificate)
	existing = protoutils.StrategicMerge(existing, updated)

	// Update the certificate
	certificate, err := l.Repo.Update(ctx, id, existing)
	if err != nil {
		l.Logger.Error("error updating certificate", "error", err, "name", existing.GetMeta().GetName())
		return err
	}

	equal, err := protoutils.SpecEqual(existing.GetConfig(), certificate.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if !equal {
		err = l.Exchange.Forward(ctx, events.NewEvent(events.CertificatePatch, certificate))
		if err != nil {
			l.Logger.Error("error publishing certificate patch event", "error", err, "name", existing.GetMeta().GetName())
			return err
		}
	}

	return nil
}

func (l *CertificateService) Update(ctx context.Context, id keys.ID, certificate *certv1.Certificate) error {
	ctx, span := tracer.Start(ctx, "certificate.Update")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get the existing certificate before updating so we can compare specs
	existingCertificate, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	// Update the certificate
	updated, err := l.Repo.Update(ctx, id, certificate)
	if err != nil {
		l.Logger.Error("error updating certificate", "error", err, "name", certificate.GetMeta().GetName())
		return err
	}

	equal, err := protoutils.SpecEqual(existingCertificate.GetConfig(), updated.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if !equal {
		l.Logger.Debug("certificate was updated, emitting event to listeners", "event", "CertificateUpdate", "name", updated.GetMeta().GetName())
		err = l.Exchange.Forward(ctx, events.NewEvent(events.CertificateUpdate, updated))
		if err != nil {
			l.Logger.Error("error publishing certificate update event", "error", err, "name", updated.GetMeta().GetName())
			return err
		}
	}

	return nil
}
