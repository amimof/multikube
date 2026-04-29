package app

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "github.com/amimof/multikube/api/auth/v1"
	tokenv1 "github.com/amimof/multikube/api/token/v1"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/golang-jwt/jwt"
)

type stubTokenManager struct {
	verifyErr error
}

func (s stubTokenManager) Issue(context.Context, *tokenv1.Token) (string, error) {
	return "", nil
}

func (s stubTokenManager) Revoke(context.Context, *tokenv1.Token) error {
	return nil
}

func (s stubTokenManager) Verify(context.Context, string) (*jwt.Token, error) {
	if s.verifyErr != nil {
		return nil, s.verifyErr
	}
	return &jwt.Token{Valid: true}, nil
}

func TestAuthServiceLogoutInvalidToken(t *testing.T) {
	svc := &AuthService{
		Exchange: events.NewExchange(),
		Logger:   &logger.DevNullLogger{},
		Issuser:  stubTokenManager{verifyErr: errors.New("bad token")},
	}

	_, err := svc.Logout(context.Background(), &authv1.LogoutRequest{AccessToken: "bad"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got %T", err)
	}
	if st.Code() != codes.Unauthenticated {
		t.Fatalf("code = %v, want %v", st.Code(), codes.Unauthenticated)
	}
}
