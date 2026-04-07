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

	cav1 "github.com/amimof/multikube/api/ca/v1"
)

const (
	CertificateAuthorityHealthHealthy   = "healthy"
	CertificateAuthorityHealthUnhealthy = "unhealthy"
)

type CreateOption func(c *clientV1)

func WithEmitLabels(l labels.Label) CreateOption {
	return func(c *clientV1) {
		c.emitLabels = l
	}
}

func WithClient(client cav1.CertificateAuthorityServiceClient) CreateOption {
	return func(c *clientV1) {
		c.Client = client
	}
}

type ClientV1 interface {
	Create(context.Context, *cav1.CertificateAuthority, ...CreateOption) error
	Update(context.Context, string, *cav1.CertificateAuthority) error
	Patch(context.Context, string, *cav1.CertificateAuthority) error
	Get(context.Context, string) (*cav1.CertificateAuthority, error)
	Delete(context.Context, string) error
	List(context.Context, ...labels.Label) ([]*cav1.CertificateAuthority, error)
}

type clientV1 struct {
	Client     cav1.CertificateAuthorityServiceClient
	emitLabels labels.Label
}

func (c *clientV1) Create(ctx context.Context, ctr *cav1.CertificateAuthority, opts ...CreateOption) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.ca.Update")
	defer span.End()

	if ctr.Version == "" {
		ctr.Version = version.VersionCertificateAuthority
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := c.Client.Create(ctx, &cav1.CreateRequest{CertificateAuthority: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Update(ctx context.Context, id string, ctr *cav1.CertificateAuthority) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.ca.Update")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	if ctr.Version == "" {
		ctr.Version = version.VersionCertificateAuthority
	}

	_, err = c.Client.Update(ctx, &cav1.UpdateRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), CertificateAuthority: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Patch(ctx context.Context, id string, ctr *cav1.CertificateAuthority) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.ca.Patch")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Patch(ctx, &cav1.PatchRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), CertificateAuthority: ctr})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientV1) Get(ctx context.Context, id string) (*cav1.CertificateAuthority, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.ca.Get")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, &cav1.GetRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
	if err != nil {
		return nil, err
	}
	return res.GetCertificateAuthority(), nil
}

func (c *clientV1) List(ctx context.Context, l ...labels.Label) ([]*cav1.CertificateAuthority, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.ca.List")
	defer span.End()

	mergedLabels := util.MergeLabels(l...)
	res, err := c.Client.List(ctx, &cav1.ListRequest{Selector: mergedLabels})
	if err != nil {
		return nil, err
	}
	return res.CertificateAuthoritys, nil
}

func (c *clientV1) Delete(ctx context.Context, id string) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.ca.Delete")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Delete(ctx, &cav1.DeleteRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
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
		Client: cav1.NewCertificateAuthorityServiceClient(conn),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
