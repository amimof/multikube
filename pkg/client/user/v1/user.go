package v1

import (
	"context"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/amimof/multikube/pkg/client/version"
	"github.com/amimof/multikube/pkg/errs"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/labels"
	"github.com/amimof/multikube/pkg/util"

	userv1 "github.com/amimof/multikube/api/user/v1"
)

type CreateOption func(c *clientV1)

func WithEmitLabels(l labels.Label) CreateOption {
	return func(c *clientV1) {
		c.emitLabels = l
	}
}

func WithClient(client userv1.UserServiceClient) CreateOption {
	return func(c *clientV1) {
		c.Client = client
	}
}

type ClientV1 interface {
	Create(context.Context, *userv1.User, ...CreateOption) error
	Update(context.Context, string, *userv1.User) error
	Patch(context.Context, string, *userv1.User) error
	Get(context.Context, string) (*userv1.User, error)
	Delete(context.Context, string) error
	List(context.Context, ...labels.Label) ([]*userv1.User, error)
}

type clientV1 struct {
	Client     userv1.UserServiceClient
	emitLabels labels.Label
}

func (c *clientV1) Create(ctx context.Context, ctr *userv1.User, opts ...CreateOption) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.user.Update")
	defer span.End()

	if ctr.Version == "" {
		ctr.Version = version.VersionUser
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := c.Client.Create(ctx, &userv1.CreateRequest{User: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Update(ctx context.Context, id string, ctr *userv1.User) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.user.Update")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	if ctr.Version == "" {
		ctr.Version = version.VersionUser
	}

	_, err = c.Client.Update(ctx, &userv1.UpdateRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), User: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Patch(ctx context.Context, id string, ctr *userv1.User) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.user.Patch")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Patch(ctx, &userv1.PatchRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), User: ctr})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientV1) Get(ctx context.Context, id string) (*userv1.User, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.user.Get")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, &userv1.GetRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
	if err != nil {
		return nil, err
	}
	return res.GetUser(), nil
}

func (c *clientV1) List(ctx context.Context, l ...labels.Label) ([]*userv1.User, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.user.List")
	defer span.End()

	mergedLabels := util.MergeLabels(l...)
	res, err := c.Client.List(ctx, &userv1.ListRequest{Selector: mergedLabels})
	if err != nil {
		return nil, err
	}
	return res.Users, nil
}

func (c *clientV1) Delete(ctx context.Context, id string) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.user.Delete")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Delete(ctx, &userv1.DeleteRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
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
		Client: userv1.NewUserServiceClient(conn),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
