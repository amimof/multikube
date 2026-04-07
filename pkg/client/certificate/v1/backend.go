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

	certv1 "github.com/amimof/multikube/api/certificate/v1"
)

const (
	CertificateHealthHealthy   = "healthy"
	CertificateHealthUnhealthy = "unhealthy"
)

type CreateOption func(c *clientV1)

func WithEmitLabels(l labels.Label) CreateOption {
	return func(c *clientV1) {
		c.emitLabels = l
	}
}

func WithClient(client certv1.CertificateServiceClient) CreateOption {
	return func(c *clientV1) {
		c.Client = client
	}
}

type ClientV1 interface {
	Create(context.Context, *certv1.Certificate, ...CreateOption) error
	Update(context.Context, string, *certv1.Certificate) error
	Patch(context.Context, string, *certv1.Certificate) error
	Get(context.Context, string) (*certv1.Certificate, error)
	Delete(context.Context, string) error
	List(context.Context, ...labels.Label) ([]*certv1.Certificate, error)
}

type clientV1 struct {
	Client     certv1.CertificateServiceClient
	emitLabels labels.Label
}

func (c *clientV1) Create(ctx context.Context, ctr *certv1.Certificate, opts ...CreateOption) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.certificate.Update")
	defer span.End()

	if ctr.Version == "" {
		ctr.Version = version.VersionCertificate
	}

	for _, opt := range opts {
		opt(c)
	}

	_, err := c.Client.Create(ctx, &certv1.CreateRequest{Certificate: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Update(ctx context.Context, id string, ctr *certv1.Certificate) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.certificate.Update")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	if ctr.Version == "" {
		ctr.Version = version.VersionCertificate
	}

	_, err = c.Client.Update(ctx, &certv1.UpdateRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Certificate: ctr})
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Patch(ctx context.Context, id string, ctr *certv1.Certificate) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.certificate.Patch")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Patch(ctx, &certv1.PatchRequest{Uid: uid.UUIDStr(), Name: uid.NameStr(), Certificate: ctr})
	if err != nil {
		return err
	}
	return nil
}

func (c *clientV1) Get(ctx context.Context, id string) (*certv1.Certificate, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.certificate.Get")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return nil, err
	}

	res, err := c.Client.Get(ctx, &certv1.GetRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
	if err != nil {
		return nil, err
	}
	return res.GetCertificate(), nil
}

func (c *clientV1) List(ctx context.Context, l ...labels.Label) ([]*certv1.Certificate, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.certificate.List")
	defer span.End()

	mergedLabels := util.MergeLabels(l...)
	res, err := c.Client.List(ctx, &certv1.ListRequest{Selector: mergedLabels})
	if err != nil {
		return nil, err
	}
	return res.Certificates, nil
}

func (c *clientV1) Delete(ctx context.Context, id string) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.certificate.Delete")
	defer span.End()

	uid, err := keys.ParseStr(id)
	if err != nil {
		return err
	}

	_, err = c.Client.Delete(ctx, &certv1.DeleteRequest{Uid: uid.UUIDStr(), Name: uid.NameStr()})
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
		Client: certv1.NewCertificateServiceClient(conn),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
