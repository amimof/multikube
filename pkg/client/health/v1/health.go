package v1

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type ClientV1 struct {
	client grpc_health_v1.HealthClient
}

func (c *ClientV1) Check(ctx context.Context) (*grpc_health_v1.HealthCheckResponse, error) {
	return c.client.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
}

func (c *ClientV1) List(ctx context.Context) (*grpc_health_v1.HealthListResponse, error) {
	return c.client.List(ctx, &grpc_health_v1.HealthListRequest{})
}

func (c *ClientV1) Watch(ctx context.Context) error {
	// return c.client.Watch(ctx, &grpc_health_v1.HealthCheckRequest{})
	return nil
}

func NewClientV1WithConn(conn *grpc.ClientConn) *ClientV1 {
	c := &ClientV1{
		client: grpc_health_v1.NewHealthClient(conn),
	}

	return c
}
