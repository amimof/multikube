package client

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type AccessTokenProvider interface {
	GetAccessToken(ctx context.Context) (string, bool)
}

type ConfigAccessTokenProvider struct {
	Config *Config
}

func (p ConfigAccessTokenProvider) GetAccessToken(context.Context) (string, bool) {
	if p.Config == nil {
		return "", false
	}

	server, err := p.Config.CurrentServer()
	if err != nil || server == nil || server.Session == nil || server.Session.AccessToken == "" {
		return "", false
	}

	return server.Session.AccessToken, true
}

func AccessTokenUnaryInterceptor(p AccessTokenProvider) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		ctx = withAccessToken(ctx, p)
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func AccessTokenStreamInterceptor(p AccessTokenProvider) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		ctx = withAccessToken(ctx, p)
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func withAccessToken(ctx context.Context, p AccessTokenProvider) context.Context {
	if p == nil {
		return ctx
	}

	token, ok := p.GetAccessToken(ctx)
	if !ok || token == "" {
		return ctx
	}

	md, _ := metadata.FromOutgoingContext(ctx)
	md = md.Copy()
	md.Set("authorization", "Bearer "+token)
	return metadata.NewOutgoingContext(ctx, md)
}
