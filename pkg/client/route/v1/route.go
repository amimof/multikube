package v1

import (
	"context"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/fieldmaskpb"

	"github.com/amimof/multikube/pkg/client/version"
	"github.com/amimof/multikube/pkg/errs"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/labels"
	"github.com/amimof/multikube/pkg/util"

	routev1 "github.com/amimof/multikube/api/route/v1"
)

const (
	RouteHealthHealthy   = "healthy"
	RouteHealthUnhealthy = "unhealthy"
)

type CreateOption func(c *clientV1)

func WithEmitLabels(l labels.Label) CreateOption {
	return func(c *clientV1) {
		c.emitLabels = l
	}
}

func WithClient(client routev1.RouteServiceClient) CreateOption {
	return func(c *clientV1) {
		c.Client = client
	}
}

type ClientV1 interface {
	Create(context.Context, *routev1.Route, ...CreateOption) error
	Update(context.Context, string, *routev1.Route) error
	Patch(context.Context, string, *routev1.Route) error
	Get(context.Context, string) (*routev1.Route, error)
	Delete(context.Context, string) error
	List(context.Context, ...labels.Label) ([]*routev1.Route, error)
	UpdateStatus(context.Context, string, *routev1.RouteStatus, ...string) error
}

type clientV1 struct {
	Client     routev1.RouteServiceClient
	emitLabels labels.Label
}

func (c *clientV1) Create(ctx context.Context, ctr *routev1.Route, opts ...CreateOption) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.task.Update")
	defer span.End()

	if ctr.Version == "" {
		ctr.Version = version.VersionRoute
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := c.Client.Create(ctx, &routev1.CreateRequest{Route: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Update(ctx context.Context, id string, ctr *routev1.Route) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.task.Update")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Update(ctx, &routev1.UpdateRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Route: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Patch(ctx context.Context, id string, ctr *routev1.Route) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.task.Patch")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Patch(ctx, &routev1.PatchRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Route: ctr})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientV1) Get(ctx context.Context, id string) (*routev1.Route, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.task.Get")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, &routev1.GetRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
	if err != nil {
		return nil, err
	}
	return res.GetRoute(), nil
}

func (c *clientV1) List(ctx context.Context, l ...labels.Label) ([]*routev1.Route, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.task.List")
	defer span.End()

	mergedLabels := util.MergeLabels(l...)
	res, err := c.Client.List(ctx, &routev1.ListRequest{Selector: mergedLabels})
	if err != nil {
		return nil, err
	}
	return res.Routes, nil
}

func (c *clientV1) Delete(ctx context.Context, id string) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.task.Delete")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Delete(ctx, &routev1.DeleteRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientV1) UpdateStatus(ctx context.Context, id string, status *routev1.RouteStatus, path ...string) error {
	// Construct field mask
	mask := &fieldmaskpb.FieldMask{
		Paths: path,
	}

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	req := &routev1.UpdateStatusRequest{
		Name:       uid.NameStr(),
		Uid:        uid.UUIDStr(),
		UpdateMask: mask,
		Status:     status,
	}

	_, err = c.Client.UpdateStatus(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func NewClientV1(opts ...CreateOption) ClientV1 {
	c := &clientV1{}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func NewClientV1WithConn(conn *grpc.ClientConn, opts ...CreateOption) ClientV1 {
	c := &clientV1{
		Client: routev1.NewRouteServiceClient(conn),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
