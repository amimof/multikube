package app

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/trace"
	"golang.org/x/crypto/bcrypt"

	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/amimof/multikube/pkg/protoutils"
	"github.com/amimof/multikube/pkg/repository"

	userv1 "github.com/amimof/multikube/api/user/v1"
)

type UserService struct {
	Repo     *repository.Repo[*userv1.User]
	mu       sync.Mutex
	Exchange *events.Exchange
	Logger   logger.Logger
}

func (l *UserService) Get(ctx context.Context, id keys.ID) (*userv1.User, error) {
	ctx, span := tracer.Start(ctx, "user.Get", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	return l.Repo.Get(ctx, id)
}

func (l *UserService) List(ctx context.Context, limit int32) ([]*userv1.User, error) {
	ctx, span := tracer.Start(ctx, "user.List")
	defer span.End()

	// Get users from repo
	return l.Repo.List(ctx, limit)
}

func (l *UserService) Create(ctx context.Context, user *userv1.User) (*userv1.User, error) {
	ctx, span := tracer.Start(ctx, "user.Create")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	if user.GetConfig().Enabled == nil {
		user.GetConfig().Enabled = new(true)
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.GetConfig().GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user.GetConfig().Password = string(hashedPass)

	// Create user in repo
	newUser, err := l.Repo.Create(ctx, user)
	if err != nil {
		l.Logger.Error("error creating user", "error", err, "name", newUser.GetMeta().GetName())
		return nil, err
	}

	// Publish event that user is created
	err = l.Exchange.Forward(ctx, events.NewEvent(events.UserCreate, user))
	if err != nil {
		l.Logger.Error("error publishing user create event", "error", err, "name", newUser.GetMeta().GetName())
		return nil, err
	}

	return newUser, nil
}

// Delete publishes a delete request and the subscribers are responsible for deleting resources.
// Once they do, they will update there resource with the status Deleted
func (l *UserService) Delete(ctx context.Context, id keys.ID) error {
	ctx, span := tracer.Start(ctx, "user.Delete")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	user, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	err = l.Repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	err = l.Exchange.Forward(ctx, events.NewEvent(events.UserDelete, user))
	if err != nil {
		l.Logger.Error("error publishing user delete event", "error", err, "name", user.GetMeta().GetName())
		return err
	}

	return nil
}

func (l *UserService) Patch(ctx context.Context, id keys.ID, patch *userv1.User) error {
	ctx, span := tracer.Start(ctx, "user.Patch")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get existing user from repo
	existing, err := l.Repo.Get(ctx, id)
	if err != nil {
		l.Logger.Error("error getting user", "error", err, "name", patch.GetMeta().GetName())
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

	updated := maskedUpdate.(*userv1.User)
	candidate := protoutils.StrategicMerge(existing, updated)

	if patch.GetConfig() == nil || patch.GetConfig().GetPassword() == "" {
		candidate.GetConfig().Password = existing.GetConfig().GetPassword()
	} else {
		hashedPass, err := bcrypt.GenerateFromPassword([]byte(candidate.GetConfig().GetPassword()), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		candidate.GetConfig().Password = string(hashedPass)
	}

	equal, err := protoutils.SpecEqual(existing.GetConfig(), candidate.GetConfig())
	if err != nil {
		return err
	}

	// Update the user
	user, err := l.Repo.Update(ctx, id, candidate)
	if err != nil {
		l.Logger.Error("error updating user", "error", err, "name", candidate.GetMeta().GetName())
		return err
	}

	// Only publish if spec is updated
	if !equal {
		err = l.Exchange.Forward(ctx, events.NewEvent(events.UserPatch, user))
		if err != nil {
			l.Logger.Error("error publishing user patch event", "error", err, "name", candidate.GetMeta().GetName())
			return err
		}
	}

	return nil
}

func (l *UserService) Update(ctx context.Context, id keys.ID, user *userv1.User) error {
	ctx, span := tracer.Start(ctx, "user.Update")
	defer span.End()

	l.mu.Lock()
	defer l.mu.Unlock()

	// Get the existing user before updating so we can compare specs
	existingUser, err := l.Repo.Get(ctx, id)
	if err != nil {
		return err
	}

	if user.GetConfig().Enabled == nil {
		user.GetConfig().Enabled = new(true)
	}

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(user.GetConfig().GetPassword()), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.GetConfig().Password = string(hashedPass)

	// Update the user
	updated, err := l.Repo.Update(ctx, id, user)
	if err != nil {
		l.Logger.Error("error updating user", "error", err, "name", user.GetMeta().GetName())
		return err
	}

	equal, err := protoutils.SpecEqual(existingUser.GetConfig(), updated.GetConfig())
	if err != nil {
		return err
	}

	// Only publish if spec is updated
	if !equal {
		l.Logger.Debug("user was updated, emitting event to listeners", "event", "UserUpdate", "name", updated.GetMeta().GetName())
		err = l.Exchange.Forward(ctx, events.NewEvent(events.UserUpdate, updated))
		if err != nil {
			l.Logger.Error("error publishing user update event", "error", err, "name", updated.GetMeta().GetName())
			return err
		}
	}

	return nil
}
