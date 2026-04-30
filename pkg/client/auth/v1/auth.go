package v1

import (
	"context"

	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/amimof/multikube/pkg/errs"

	authv1 "github.com/amimof/multikube/api/auth/v1"
)

type CreateOption func(c *clientV1)

func WithClient(client authv1.AuthServiceClient) CreateOption {
	return func(c *clientV1) {
		c.Client = client
	}
}

type ClientV1 interface {
	Login(context.Context, *authv1.LoginRequest) (*authv1.LoginResponse, error)
	Logout(context.Context, *authv1.LogoutRequest) error
	Refresh(context.Context, *authv1.RefreshRequest) (*authv1.RefreshResponse, error)
}

type clientV1 struct {
	Client authv1.AuthServiceClient
}

func (c *clientV1) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.auth.Login")
	defer span.End()

	resp, err := c.Client.Login(ctx, req)
	if err != nil {
		return nil, errs.ToStatus(err)
	}
	return resp, nil
}

func (c *clientV1) Logout(ctx context.Context, req *authv1.LogoutRequest) error {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.auth.Logout")
	defer span.End()

	_, err := c.Client.Logout(ctx, req)
	if err != nil {
		return errs.ToStatus(err)
	}
	return nil
}

func (c *clientV1) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	tracer := otel.Tracer("client-v1")
	ctx, span := tracer.Start(ctx, "client.auth.Refresh")
	defer span.End()

	resp, err := c.Client.Refresh(ctx, req)
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
		Client: authv1.NewAuthServiceClient(conn),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}
