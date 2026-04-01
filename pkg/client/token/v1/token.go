package v1

import (
	"context"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/amimof/multikube/pkg/errs"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
)

type CreateOption func(c *clientV1)

func WithClient(client tokenv1.TokenServiceClient) CreateOption {
	return func(c *clientV1) {
		c.Client = client
	}
}

type ClientV1 interface {
	IssueToken(context.Context, *tokenv1.Token) (*tokenv1.IssueResponse, error)
}

type clientV1 struct {
	Client tokenv1.TokenServiceClient
}

func (c *clientV1) IssueToken(ctx context.Context, ctr *tokenv1.Token) (*tokenv1.IssueResponse, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.token.Issue")
	defer span.End()

	resp, err := c.Client.Issue(ctx, &tokenv1.IssueRequest{Token: ctr})
	if err != nil {
		return nil, errs.ToStatus(err)
	}
	return resp, nil
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
		Client: tokenv1.NewTokenServiceClient(conn),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
