package client

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestConfigAccessTokenProviderGetAccessToken(t *testing.T) {
	provider := ConfigAccessTokenProvider{Config: &Config{
		Current: "prod",
		Servers: []*Server{{
			Name:    "prod",
			Address: "example.com:443",
			Session: &Session{AccessToken: "secret-token"},
		}},
	}}

	token, ok := provider.GetAccessToken(context.Background())
	if !ok {
		t.Fatal("expected access token to be available")
	}
	if token != "secret-token" {
		t.Fatalf("token = %q, want %q", token, "secret-token")
	}
}

func TestAccessTokenUnaryInterceptorAddsAuthorizationMetadata(t *testing.T) {
	provider := ConfigAccessTokenProvider{Config: &Config{
		Current: "prod",
		Servers: []*Server{{
			Name:    "prod",
			Address: "example.com:443",
			Session: &Session{AccessToken: "secret-token"},
		}},
	}}

	interceptor := AccessTokenUnaryInterceptor(provider)
	err := interceptor(context.Background(), "/auth.v1.AuthService/Login", nil, nil, nil, func(ctx context.Context, _ string, _ any, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("expected outgoing metadata")
		}
		values := md.Get("authorization")
		if len(values) != 1 {
			t.Fatalf("authorization values = %#v, want one value", values)
		}
		if values[0] != "Bearer secret-token" {
			t.Fatalf("authorization = %q, want %q", values[0], "Bearer secret-token")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
}

func TestAccessTokenUnaryInterceptorSkipsMissingToken(t *testing.T) {
	provider := ConfigAccessTokenProvider{Config: &Config{
		Current: "prod",
		Servers: []*Server{{
			Name:    "prod",
			Address: "example.com:443",
		}},
	}}

	interceptor := AccessTokenUnaryInterceptor(provider)
	err := interceptor(context.Background(), "/backend.v1.BackendService/List", nil, nil, nil, func(ctx context.Context, _ string, _ any, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if ok && len(md.Get("authorization")) > 0 {
			t.Fatalf("unexpected authorization metadata: %#v", md.Get("authorization"))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
}
