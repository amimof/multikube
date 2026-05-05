package client

import (
	"context"
	"testing"
	"time"

	authv1 "github.com/amimof/multikube/api/auth/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type stubRefreshAuthClient struct {
	resp  *authv1.RefreshResponse
	calls int
}

func (s *stubRefreshAuthClient) Login(context.Context, *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return nil, nil
}

func (s *stubRefreshAuthClient) Logout(context.Context, *authv1.LogoutRequest) error {
	return nil
}

func (s *stubRefreshAuthClient) Refresh(context.Context, *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	s.calls++
	return s.resp, nil
}

func TestRefreshTokenSourceHandlesOmittedExpiresAt(t *testing.T) {
	client := &stubRefreshAuthClient{resp: &authv1.RefreshResponse{
		AccessToken:  "new-access-token",
		RefreshToken: "new-refresh-token",
	}}
	source := NewRefreshTokenSource(client, "", "refresh-token", time.Time{})

	tok, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.ExpiresAt != (time.Time{}) {
		t.Fatalf("expires at = %v, want zero", tok.ExpiresAt)
	}
	if client.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1", client.calls)
	}

	tok, err = source.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tok.AccessToken != "new-access-token" {
		t.Fatalf("access token = %q, want %q", tok.AccessToken, "new-access-token")
	}
	if client.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1 after cached token reuse", client.calls)
	}
}

func TestAccessTokenUnaryInterceptorUsesCachedRefreshSourceToken(t *testing.T) {
	source := NewRefreshTokenSource(nil, "cached-token", "refresh-token", time.Now().Add(time.Hour))
	interceptor := AccessTokenUnaryInterceptor(source)

	err := interceptor(context.Background(), "/backend.v1.BackendService/List", nil, nil, nil, func(ctx context.Context, _ string, _ any, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("expected outgoing metadata")
		}
		values := md.Get("authorization")
		if len(values) != 1 || values[0] != "Bearer cached-token" {
			t.Fatalf("authorization = %#v, want bearer cached-token", values)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
}

func TestRefreshUnaryInterceptorSkipsRefreshMethod(t *testing.T) {
	client := &stubRefreshAuthClient{resp: &authv1.RefreshResponse{AccessToken: "new-access-token", RefreshToken: "new-refresh-token"}}
	source := NewRefreshTokenSource(client, "cached-token", "refresh-token", time.Now().Add(-time.Minute))
	interceptor := RefreshUnaryInterceptor(source)

	calls := 0
	err := interceptor(context.Background(), refreshMethod, nil, nil, nil, func(ctx context.Context, method string, _ any, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		calls++
		if method != refreshMethod {
			t.Fatalf("method = %q, want %q", method, refreshMethod)
		}
		return status.Error(codes.Unauthenticated, "unauthenticated")
	})
	if status.Code(err) != codes.Unauthenticated {
		t.Fatalf("error code = %v, want unauthenticated", status.Code(err))
	}
	if calls != 1 {
		t.Fatalf("invoker calls = %d, want 1", calls)
	}
	if client.calls != 0 {
		t.Fatalf("refresh calls = %d, want 0", client.calls)
	}
}

func TestRefreshUnaryInterceptorRefreshesAndRetries(t *testing.T) {
	client := &stubRefreshAuthClient{resp: &authv1.RefreshResponse{AccessToken: "new-access-token", RefreshToken: "new-refresh-token"}}
	source := NewRefreshTokenSource(client, "expired-token", "refresh-token", time.Now().Add(-time.Minute))
	interceptor := RefreshUnaryInterceptor(source)

	calls := 0
	err := interceptor(context.Background(), "/backend.v1.BackendService/List", nil, nil, nil, func(ctx context.Context, _ string, _ any, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		calls++
		if calls == 1 {
			return status.Error(codes.Unauthenticated, "expired")
		}

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			t.Fatal("expected outgoing metadata on retry")
		}
		values := md.Get("authorization")
		if len(values) != 1 || values[0] != "Bearer new-access-token" {
			t.Fatalf("authorization = %#v, want bearer new-access-token", values)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("interceptor returned error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("invoker calls = %d, want 2", calls)
	}
	if client.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1", client.calls)
	}
}
