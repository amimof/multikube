package identity

import (
	"context"
	"sync/atomic"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type ClientIdentity struct {
	Name string
	UID  string
}

type IdentityProvider interface {
	Get(ctx context.Context) (ClientIdentity, bool)
}

type AtomicIdentity struct {
	v atomic.Value // stores NodeIdentity
}

func (a *AtomicIdentity) Set(id ClientIdentity) { a.v.Store(id) }

func (a *AtomicIdentity) Get(ctx context.Context) (ClientIdentity, bool) {
	val := a.v.Load()
	if val == nil {
		return ClientIdentity{}, false
	}
	return val.(ClientIdentity), true
}

func IdentityUnaryInterceptor(p IdentityProvider) grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req any,
		reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		if id, ok := p.Get(ctx); ok && id.UID != "" {
			md, _ := metadata.FromOutgoingContext(ctx)
			md = md.Copy()
			md.Set("x-voiyd-node-uid", id.UID)
			md.Set("x-voiyd-node-name", id.Name)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

func IdentityStreamInterceptor(p IdentityProvider) grpc.StreamClientInterceptor {
	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		if id, ok := p.Get(ctx); ok && id.UID != "" {
			md, _ := metadata.FromOutgoingContext(ctx)
			md = md.Copy()
			md.Set("x-voiyd-node-uid", id.UID)
			md.Set("x-voiyd-node-name", id.Name)
			ctx = metadata.NewOutgoingContext(ctx, md)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
