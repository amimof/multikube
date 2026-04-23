package app

import (
	"context"
	"sync"
	"testing"
	"time"

	eventv1 "github.com/amimof/multikube/api/event/v1"
	metav1 "github.com/amimof/multikube/api/meta/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/amimof/multikube/pkg/repository"
	repobadger "github.com/amimof/multikube/pkg/repository/badger"
	badgerdb "github.com/dgraph-io/badger/v4"
)

func TestRouteServicePatch_DoesNotPanicOrDeadlock(t *testing.T) {
	db, err := badgerdb.Open(badgerdb.DefaultOptions(t.TempDir()).WithLogger(nil))
	if err != nil {
		t.Fatalf("open badger: %v", err)
	}
	defer func() {
		_ = db.Close()
	}()

	repo := repository.NewRouteRepo[*routev1.Route](repobadger.New(db))
	exchange := events.NewExchange(events.WithExchangeLogger(logger.NilLogger{}))
	service := &RouteService{
		Repo:     repo,
		Exchange: exchange,
		Logger:   logger.NilLogger{},
	}

	created, err := service.Create(context.Background(), &routev1.Route{
		Meta: &metav1.Meta{Name: "route"},
		Config: &routev1.RouteConfig{
			BackendRef: "be-a",
			Match:      &routev1.Match{Path: "/a"},
		},
	})
	if err != nil {
		t.Fatalf("create route: %v", err)
	}

	patchEvents := exchange.Subscribe(context.Background(), eventv1.Event_EVENT_ROUTE_PATCH)
	id, err := keys.Name(created.GetMeta().GetName())
	if err != nil {
		t.Fatalf("route key: %v", err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	panicCh := make(chan any, 1)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				panicCh <- r
			}
		}()
		errCh <- service.Patch(context.Background(), id, &routev1.Route{
			Meta: &metav1.Meta{Name: "route"},
			Config: &routev1.RouteConfig{
				BackendRef: "be-b",
				Match:      &routev1.Match{Path: "/b"},
			},
		})
	}()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("patch route timed out")
	}

	select {
	case p := <-panicCh:
		t.Fatalf("patch route panicked: %v", p)
	default:
	}

	if err := <-errCh; err != nil {
		t.Fatalf("patch route: %v", err)
	}

	select {
	case ev := <-patchEvents:
		t.Fatalf("unexpected patch event: %v", ev.GetEvent())
	case <-time.After(time.Second):
		// Current Patch implementation compares the merged object against the
		// stored update result, so it does not currently emit RoutePatch.
	}

	stored, err := repo.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("get patched route: %v", err)
	}
	if got := stored.GetConfig().GetBackendRef(); got != "be-b" {
		t.Fatalf("backend_ref = %q, want %q", got, "be-b")
	}
	if got := stored.GetConfig().GetMatch().GetPath(); got != "/b" {
		t.Fatalf("path = %q, want %q", got, "/b")
	}
}
