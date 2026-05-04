package app

import (
	"context"
	"testing"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
	"github.com/amimof/multikube/internal/infra"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/golang-jwt/jwt"
)

type stubTokenIssuer struct {
	resp infra.IssueResponse
	err  error
}

func (s *stubTokenIssuer) Issue(context.Context, *tokenv1.Token) (infra.IssueResponse, error) {
	return s.resp, s.err
}

func (s *stubTokenIssuer) Revoke(context.Context, *tokenv1.Token) error {
	return nil
}

func (s *stubTokenIssuer) VerifyAccessToken(context.Context, string) (jwt.MapClaims, error) {
	return nil, nil
}

func (s *stubTokenIssuer) VerifyRefreshToken(context.Context, string) (jwt.MapClaims, error) {
	return nil, nil
}

func TestTokenServiceIssueOmitsExpiresAtForNonExpiringToken(t *testing.T) {
	svc := &TokenService{
		Exchange: events.NewExchange(),
		Logger:   &logger.DevNullLogger{},
		Issuer: &stubTokenIssuer{resp: infra.IssueResponse{
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			KeyID:        "key-id",
			TokenType:    "access",
		}},
	}

	resp, err := svc.IssueToken(context.Background(), &tokenv1.Token{Config: &tokenv1.TokenConfig{Subject: "alice"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExpiresAt != nil {
		t.Fatalf("expires at = %v, want nil", resp.ExpiresAt)
	}
}
