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

	policyv1 "github.com/amimof/multikube/api/policy/v1"
)

type CreateOption func(c *clientV1)

func WithClient(client policyv1.PolicyServiceClient) CreateOption {
	return func(c *clientV1) {
		c.Client = client
	}
}

type ClientV1 interface {
	Create(context.Context, *policyv1.Policy, ...CreateOption) error
	Update(context.Context, string, *policyv1.Policy) error
	Patch(context.Context, string, *policyv1.Policy) error
	Get(context.Context, string) (*policyv1.Policy, error)
	Delete(context.Context, string) error
	List(context.Context, ...labels.Label) ([]*policyv1.Policy, error)
}

type clientV1 struct {
	Client policyv1.PolicyServiceClient
}

func (c *clientV1) Create(ctx context.Context, policy *policyv1.Policy, opts ...CreateOption) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.policy.Create")
	defer span.End()

	if policy.Version == "" {
		policy.Version = version.VersionPolicy
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := c.Client.Create(ctx, &policyv1.CreateRequest{Policy: policy})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Update(ctx context.Context, id string, policy *policyv1.Policy) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.policy.Update")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Update(ctx, &policyv1.UpdateRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Policy: policy})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Patch(ctx context.Context, id string, policy *policyv1.Policy) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.policy.Patch")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Patch(ctx, &policyv1.PatchRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Policy: policy})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientV1) Get(ctx context.Context, id string) (*policyv1.Policy, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.policy.Get")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, &policyv1.GetRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
	if err != nil {
		return nil, err
	}
	return res.GetPolicy(), nil
}

func (c *clientV1) List(ctx context.Context, l ...labels.Label) ([]*policyv1.Policy, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.policy.List")
	defer span.End()

	mergedLabels := util.MergeLabels(l...)
	res, err := c.Client.List(ctx, &policyv1.ListRequest{Selector: mergedLabels})
	if err != nil {
		return nil, err
	}
	return res.Policys, nil
}

func (c *clientV1) Delete(ctx context.Context, id string) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.policy.Delete")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Delete(ctx, &policyv1.DeleteRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
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
		Client: policyv1.NewPolicyServiceClient(conn),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
