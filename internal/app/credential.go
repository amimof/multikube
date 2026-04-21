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

	credentialv1 "github.com/amimof/multikube/api/credential/v1"
)

type CredentialService struct {
	Repo     *repository.Repo[*credentialv1.Credential]
	mu       sync.Mutex
	Exchange *events.Exchange
	Logger   logger.Logger
}

func (l *CredentialService) Get(ctx context.Context, id keys.ID) (*credentialv1.Credential, error) {
	ctx, span := tracer.Start(ctx, "credential.Get", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return l.Repo.Get(ctx, id)
}

func (l *CredentialService) List(ctx context.Context, limit int32) ([]*credentialv1.Credential, error) {
	ctx, span := tracer.Start(ctx, "credential.List")
	defer span.End()

	// Get credentials from repo
	return l.Repo.List(ctx, limit)
}

func (l *CredentialService) Create(ctx context.Context, credential *credentialv1.Credential) (*credentialv1.Credential, error) {
	ctx, span := tracer.Start(ctx, "credential.Create")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	if credential.GetConfig().Enabled == nil {
		credential.GetConfig().Enabled = new(true)
	}

	// Create credential in repo
	newCredential, err := l.Repo.Create(ctx, credential)
	if err != nil {
		l.Logger.Error("error creating credential", "error", err, "name", newCredential.GetMeta().GetName())
		return nil, err
	}

	// Publish event that credential is created
	err = l.Exchange.Forward(ctx, events.NewEvent(events.CredentialCreate, credential))
	if err != nil {
		l.Logger.Error("error publishing credential create event", "error", err, "name", newCredential.GetMeta().GetName())
		return nil, err
	}

	return newCredential, nil
}

// Delete publishes a delete request and the subscribers are responsible for deleting resources.
// Once they do, they will update there resource with the status Deleted
func (l *CredentialService) Delete(ctx context.Context, id keys.ID) error {
	ctx, span := tracer.Start(ctx, "credential.Delete")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	credential, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	err = l.Repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = l.Exchange.Forward(ctx, events.NewEvent(events.CredentialDelete, credential))
	if err != nil {
		l.Logger.Error("error publishing credential delete event", "error", err, "name", credential.GetMeta().GetName())
		return err
	}

	return nil
}

func (l *CredentialService) Patch(ctx context.Context, id keys.ID, patch *credentialv1.Credential) error {
	ctx, span := tracer.Start(ctx, "credential.Patch")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get existing credential from repo
	existing, err := l.Repo.Get(ctx, id)
	if err != nil {
		l.Logger.Error("error getting credential", "error", err, "name", patch.GetMeta().GetName())
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

	updated := maskedUpdate.(*credentialv1.Credential)
	existing = protoutils.StrategicMerge(existing, updated)

	// Update the credential
	credential, err := l.Repo.Update(ctx, id, existing)
	if err != nil {
		l.Logger.Error("error updating credential", "error", err, "name", existing.GetMeta().GetName())
		return err
	}

	equal, err := protoutils.SpecEqual(existing.GetConfig(), credential.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if !equal {
		err = l.Exchange.Forward(ctx, events.NewEvent(events.CredentialPatch, credential))
		if err != nil {
			l.Logger.Error("error publishing credential patch event", "error", err, "name", existing.GetMeta().GetName())
			return err
		}
	}

	return nil
}

func (l *CredentialService) Update(ctx context.Context, id keys.ID, credential *credentialv1.Credential) error {
	ctx, span := tracer.Start(ctx, "credential.Update")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get the existing credential before updating so we can compare specs
	existingCredential, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if credential.GetConfig().Enabled == nil {
		credential.GetConfig().Enabled = new(true)
	}

	// Update the credential
	updated, err := l.Repo.Update(ctx, id, credential)
	if err != nil {
		l.Logger.Error("error updating credential", "error", err, "name", credential.GetMeta().GetName())
		return err
	}

	equal, err := protoutils.SpecEqual(existingCredential.GetConfig(), updated.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if !equal {
		l.Logger.Debug("credential was updated, emitting event to listeners", "event", "CredentialUpdate", "name", updated.GetMeta().GetName())
		err = l.Exchange.Forward(ctx, events.NewEvent(events.CredentialUpdate, updated))
		if err != nil {
			l.Logger.Error("error publishing credential update event", "error", err, "name", updated.GetMeta().GetName())
			return err
		}
	}

	return nil
}
