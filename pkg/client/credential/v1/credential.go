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

	credentialv1 "github.com/amimof/multikube/api/credential/v1"
)

type CreateOption func(c *clientV1)

func WithEmitLabels(l labels.Label) CreateOption {
	return func(c *clientV1) {
		c.emitLabels = l
	}
}

func WithClient(client credentialv1.CredentialServiceClient) CreateOption {
	return func(c *clientV1) {
		c.Client = client
	}
}

type ClientV1 interface {
	Create(context.Context, *credentialv1.Credential, ...CreateOption) error
	Update(context.Context, string, *credentialv1.Credential) error
	Patch(context.Context, string, *credentialv1.Credential) error
	Get(context.Context, string) (*credentialv1.Credential, error)
	Delete(context.Context, string) error
	List(context.Context, ...labels.Label) ([]*credentialv1.Credential, error)
}

type clientV1 struct {
	Client     credentialv1.CredentialServiceClient
	emitLabels labels.Label
}

func (c *clientV1) Create(ctx context.Context, ctr *credentialv1.Credential, opts ...CreateOption) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.credential.Update")
	defer span.End()

	if ctr.Version == "" {
		ctr.Version = version.VersionCredential
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := c.Client.Create(ctx, &credentialv1.CreateRequest{Credential: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Update(ctx context.Context, id string, ctr *credentialv1.Credential) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.credential.Update")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	if ctr.Version == "" {
		ctr.Version = version.VersionCredential
	}

	_, err = c.Client.Update(ctx, &credentialv1.UpdateRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Credential: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Patch(ctx context.Context, id string, ctr *credentialv1.Credential) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.credential.Patch")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Patch(ctx, &credentialv1.PatchRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Credential: ctr})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientV1) Get(ctx context.Context, id string) (*credentialv1.Credential, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.credential.Get")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, &credentialv1.GetRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
	if err != nil {
		return nil, err
	}
	return res.GetCredential(), nil
}

func (c *clientV1) List(ctx context.Context, l ...labels.Label) ([]*credentialv1.Credential, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.credential.List")
	defer span.End()

	mergedLabels := util.MergeLabels(l...)
	res, err := c.Client.List(ctx, &credentialv1.ListRequest{Selector: mergedLabels})
	if err != nil {
		return nil, err
	}
	return res.Credentials, nil
}

func (c *clientV1) Delete(ctx context.Context, id string) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.credential.Delete")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Delete(ctx, &credentialv1.DeleteRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
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
		Client: credentialv1.NewCredentialServiceClient(conn),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
