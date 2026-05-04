package client

import (
	"context"
	"testing"
	"time"

	authv1 "github.com/amimof/multikube/api/auth/v1"
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
