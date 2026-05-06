package client

import (
	"context"
	"errors"
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
	source := NewRefreshTokenSource(client, "", "refresh-token", time.Time{}, nil)

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
	source := NewRefreshTokenSource(nil, "cached-token", "refresh-token", time.Now().Add(time.Hour), nil)
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
	source := NewRefreshTokenSource(client, "cached-token", "refresh-token", time.Now().Add(-time.Minute), nil)
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
	source := NewRefreshTokenSource(client, "expired-token", "refresh-token", time.Now().Add(-time.Minute), nil)
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

func TestRefreshTokenSourceCallsCallbackOnRefresh(t *testing.T) {
	client := &stubRefreshAuthClient{resp: &authv1.RefreshResponse{AccessToken: "new-access-token", RefreshToken: "new-refresh-token"}}
	callbackCalls := 0
	source := NewRefreshTokenSource(client, "expired-token", "refresh-token", time.Now().Add(-time.Minute), func(tok *Token) error {
		callbackCalls++
		if tok.AccessToken != "new-access-token" {
			t.Fatalf("callback access token = %q, want %q", tok.AccessToken, "new-access-token")
		}
		if tok.RefreshToken != "new-refresh-token" {
			t.Fatalf("callback refresh token = %q, want %q", tok.RefreshToken, "new-refresh-token")
		}
		return nil
	})

	_, err := source.Token(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callbackCalls != 1 {
		t.Fatalf("callback calls = %d, want 1", callbackCalls)
	}

	token, ok := source.GetAccessToken(context.Background())
	if !ok || token != "new-access-token" {
		t.Fatalf("GetAccessToken = (%q, %v), want (new-access-token, true)", token, ok)
	}
}

func TestRefreshTokenSourceReturnsCallbackErrorAndBlocksTokenReuse(t *testing.T) {
	client := &stubRefreshAuthClient{resp: &authv1.RefreshResponse{AccessToken: "new-access-token", RefreshToken: "new-refresh-token"}}
	callbackErr := errors.New("persist failed")
	callbackCalls := 0
	source := NewRefreshTokenSource(client, "expired-token", "refresh-token", time.Now().Add(-time.Minute), func(*Token) error {
		callbackCalls++
		return callbackErr
	})

	_, err := source.Token(context.Background())
	if !errors.Is(err, callbackErr) {
		t.Fatalf("Token error = %v, want %v", err, callbackErr)
	}
	if callbackCalls != 1 {
		t.Fatalf("callback calls = %d, want 1", callbackCalls)
	}

	if token, ok := source.GetAccessToken(context.Background()); ok || token != "" {
		t.Fatalf("GetAccessToken = (%q, %v), want empty false", token, ok)
	}

	_, err = source.Token(context.Background())
	if !errors.Is(err, callbackErr) {
		t.Fatalf("second Token error = %v, want %v", err, callbackErr)
	}
	if callbackCalls != 2 {
		t.Fatalf("callback calls = %d, want 2", callbackCalls)
	}
	if client.calls != 1 {
		t.Fatalf("refresh calls = %d, want 1", client.calls)
	}
}

func TestWithTokenRefreshCallbackUpdatesExistingTokenSource(t *testing.T) {
	clientSet := &ClientSet{}
	if err := WithCredentialSet("access-token", "refresh-token")(clientSet); err != nil {
		t.Fatalf("WithCredentialSet error: %v", err)
	}

	callbackCalled := false
	callback := func(*Token) error {
		callbackCalled = true
		return nil
	}
	if err := WithTokenRefreshCallback(callback)(clientSet); err != nil {
		t.Fatalf("WithTokenRefreshCallback error: %v", err)
	}

	if clientSet.tokenSource == nil {
		t.Fatal("expected tokenSource to be initialized")
	}
	if clientSet.tokenSource.callback == nil {
		t.Fatal("expected tokenSource callback to be updated")
	}
	if err := clientSet.tokenSource.callback(&Token{}); err != nil {
		t.Fatalf("callback returned error: %v", err)
	}
	if !callbackCalled {
		t.Fatal("expected callback to be invoked")
	}
}
