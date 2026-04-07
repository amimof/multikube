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

	backendv1 "github.com/amimof/multikube/api/backend/v1"
)

const (
	BackendHealthHealthy   = "healthy"
	BackendHealthUnhealthy = "unhealthy"
)

type CreateOption func(c *clientV1)

func WithEmitLabels(l labels.Label) CreateOption {
	return func(c *clientV1) {
		c.emitLabels = l
	}
}

func WithClient(client backendv1.BackendServiceClient) CreateOption {
	return func(c *clientV1) {
		c.Client = client
	}
}

type ClientV1 interface {
	Create(context.Context, *backendv1.Backend, ...CreateOption) error
	Update(context.Context, string, *backendv1.Backend) error
	Patch(context.Context, string, *backendv1.Backend) error
	Get(context.Context, string) (*backendv1.Backend, error)
	Delete(context.Context, string) error
	List(context.Context, ...labels.Label) ([]*backendv1.Backend, error)
	UpdateStatus(context.Context, string, *backendv1.BackendStatus, ...string) error
}

type clientV1 struct {
	Client     backendv1.BackendServiceClient
	emitLabels labels.Label
}

func (c *clientV1) Create(ctx context.Context, ctr *backendv1.Backend, opts ...CreateOption) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.backend.Update")
	defer span.End()

	if ctr.Version == "" {
		ctr.Version = version.VersionBackend
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := c.Client.Create(ctx, &backendv1.CreateRequest{Backend: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Update(ctx context.Context, id string, ctr *backendv1.Backend) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.backend.Update")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	if ctr.Version == "" {
		ctr.Version = version.VersionBackend
	}

	_, err = c.Client.Update(ctx, &backendv1.UpdateRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Backend: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Patch(ctx context.Context, id string, ctr *backendv1.Backend) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.backend.Patch")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Patch(ctx, &backendv1.PatchRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Backend: ctr})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientV1) Get(ctx context.Context, id string) (*backendv1.Backend, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.backend.Get")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, &backendv1.GetRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
	if err != nil {
		return nil, err
	}
	return res.GetBackend(), nil
}

func (c *clientV1) List(ctx context.Context, l ...labels.Label) ([]*backendv1.Backend, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.backend.List")
	defer span.End()

	mergedLabels := util.MergeLabels(l...)
	res, err := c.Client.List(ctx, &backendv1.ListRequest{Selector: mergedLabels})
	if err != nil {
		return nil, err
	}
	return res.Backends, nil
}

func (c *clientV1) Delete(ctx context.Context, id string) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.backend.Delete")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Delete(ctx, &backendv1.DeleteRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientV1) UpdateStatus(ctx context.Context, id string, status *backendv1.BackendStatus, path ...string) error {
	// Construct field mask
	mask := &fieldmaskpb.FieldMask{
		Paths: path,
	}

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	req := &backendv1.UpdateStatusRequest{
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
		Client: backendv1.NewBackendServiceClient(conn),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
