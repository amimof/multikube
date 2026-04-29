package app

import (
	"context"
	"testing"
	"time"

	eventv1 "github.com/amimof/multikube/api/event/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	userv1 "github.com/amimof/multikube/api/user/v1"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/amimof/multikube/pkg/repository"
	repobadger "github.com/amimof/multikube/pkg/repository/badger"
	badgerdb "github.com/dgraph-io/badger/v4"
	"golang.org/x/crypto/bcrypt"
)

func TestUserServicePatch_EmitsUserPatchEvent(t *testing.T) {
	db, err := badgerdb.Open(badgerdb.DefaultOptions(t.TempDir()).WithLogger(nil))
	if err != nil {
		t.Fatalf("open badger: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewUserRepo[*userv1.User](repobadger.New(db))
	exchange := events.NewExchange(events.WithExchangeLogger(logger.NilLogger{}))
	service := &UserService{
		Repo:     repo,
		Exchange: exchange,
		Logger:   logger.NilLogger{},
	}

	created, err := service.Create(context.Background(), &userv1.User{
		Version: "multikube.io/user",
		Meta:    &metav1.Meta{Name: "alice"},
		Config: &userv1.UserConfig{
			Email:    "alice@example.com",
			Password: "secret123",
			Roles:    []string{"viewer"},
		},
	})
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	patchEvents := exchange.Subscribe(context.Background(), eventv1.Event_EVENT_USER_PATCH)
	id, err := keys.Name(created.GetMeta().GetName())
	if err != nil {
		t.Fatalf("user key: %v", err)
	}

	err = service.Patch(context.Background(), id, &userv1.User{
		Meta: &metav1.Meta{Name: "alice"},
		Config: &userv1.UserConfig{
			Email: "alice+patched@example.com",
		},
	})
	if err != nil {
		t.Fatalf("patch user: %v", err)
	}

	select {
	case ev := <-patchEvents:
		if ev.GetEvent() != eventv1.Event_EVENT_USER_PATCH {
			t.Fatalf("event = %v, want %v", ev.GetEvent(), eventv1.Event_EVENT_USER_PATCH)
		}
	case <-time.After(time.Second):
		t.Fatal("expected user patch event")
	}

	stored, err := repo.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get patched user: %v", err)
	}
	if got := stored.GetConfig().GetEmail(); got != "alice+patched@example.com" {
		t.Fatalf("email = %q, want %q", got, "alice+patched@example.com")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(stored.GetConfig().GetPassword()), []byte("secret123")); err != nil {
		t.Fatalf("password hash no longer matches original password: %v", err)
	}
}
