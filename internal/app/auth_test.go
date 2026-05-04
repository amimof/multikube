package app

import (
	"context"
	"errors"
	"testing"
	"time"

	authv1 "github.com/amimof/multikube/api/auth/v1"
	meta "github.com/amimof/multikube/api/meta/v1"
	tokenv1 "github.com/amimof/multikube/api/token/v1"
	userv1 "github.com/amimof/multikube/api/user/v1"
	"github.com/amimof/multikube/internal/infra"
	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/golang-jwt/jwt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type stubTokenManager struct {
	accessSubject  string
	refreshSubject string
	accessErr      error
	refreshErr     error
	issuedAccess   string
	issuedRefresh  string
	issueErr       error
	issueCalls     int
	issuedExpires  time.Time
}

func (s *stubTokenManager) Issue(context.Context, *tokenv1.Token) (infra.IssueResponse, error) {
	s.issueCalls++
	if s.issueErr != nil {
		return infra.IssueResponse{}, s.issueErr
	}
	return infra.IssueResponse{
		AccessToken:  s.issuedAccess,
		RefreshToken: s.issuedRefresh,
		ExpiresAt:    s.issuedExpires,
	}, nil
}

func (s *stubTokenManager) Revoke(context.Context, *tokenv1.Token) error {
	return nil
}

func (s *stubTokenManager) VerifyAccessToken(context.Context, string) (jwt.MapClaims, error) {
	if s.accessErr != nil {
		return nil, s.accessErr
	}
	return jwt.MapClaims{"sub": s.accessSubject}, nil
}

func (s *stubTokenManager) VerifyRefreshToken(context.Context, string) (jwt.MapClaims, error) {
	if s.refreshErr != nil {
		return nil, s.refreshErr
	}
	return jwt.MapClaims{"sub": s.refreshSubject}, nil
}

type stubUsersGetter struct {
	user *userv1.User
	err  error
}

func (s stubUsersGetter) List(context.Context, int32) ([]*userv1.User, error) {
	return nil, nil
}

func (s stubUsersGetter) Get(context.Context, keys.ID) (*userv1.User, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.user, nil
}

func TestAuthServiceLogoutInvalidToken(t *testing.T) {
	issuer := &stubTokenManager{accessErr: errors.New("bad token")}
	svc := &AuthService{
		Exchange: events.NewExchange(),
		Logger:   &logger.DevNullLogger{},
		Issuer:   issuer,
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

func TestAuthServiceRefreshSuccess(t *testing.T) {
	issuer := &stubTokenManager{
		refreshSubject: "alice",
		issuedAccess:   "new-access-token",
		issuedRefresh:  "new-refresh-token",
	}
	svc := &AuthService{
		Exchange: events.NewExchange(),
		Logger:   &logger.DevNullLogger{},
		Users: stubUsersGetter{user: &userv1.User{
			Meta: &meta.Meta{Name: "alice"},
			Config: &userv1.UserConfig{
				Enabled: boolPointer(true),
			},
		}},
		Issuer: issuer,
	}

	resp, err := svc.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "refresh-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetAccessToken() != "new-access-token" {
		t.Fatalf("access token = %q, want %q", resp.GetAccessToken(), "new-access-token")
	}
	if issuer.issueCalls != 1 {
		t.Fatalf("issue calls = %d, want 1", issuer.issueCalls)
	}
}

func TestAuthServiceRefreshOmitsExpiresAtForNonExpiringToken(t *testing.T) {
	issuer := &stubTokenManager{
		refreshSubject: "alice",
		issuedAccess:   "new-access-token",
		issuedRefresh:  "new-refresh-token",
	}
	svc := &AuthService{
		Exchange: events.NewExchange(),
		Logger:   &logger.DevNullLogger{},
		Users: stubUsersGetter{user: &userv1.User{
			Meta: &meta.Meta{Name: "alice"},
			Config: &userv1.UserConfig{
				Enabled: boolPointer(true),
			},
		}},
		Issuer: issuer,
	}

	resp, err := svc.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "refresh-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ExpiresAt != nil {
		t.Fatalf("expires at = %v, want nil", resp.ExpiresAt)
	}
}

func TestAuthServiceRefreshInvalidToken(t *testing.T) {
	issuer := &stubTokenManager{refreshErr: errors.New("bad refresh token")}
	svc := &AuthService{
		Exchange: events.NewExchange(),
		Logger:   &logger.DevNullLogger{},
		Issuer:   issuer,
	}

	_, err := svc.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "bad"})
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

func TestAuthServiceRefreshDisabledUser(t *testing.T) {
	issuer := &stubTokenManager{refreshSubject: "alice"}
	svc := &AuthService{
		Exchange: events.NewExchange(),
		Logger:   &logger.DevNullLogger{},
		Users: stubUsersGetter{user: &userv1.User{
			Meta: &meta.Meta{Name: "alice"},
			Config: &userv1.UserConfig{
				Enabled: boolPointer(false),
			},
		}},
		Issuer: issuer,
	}

	_, err := svc.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "refresh-token"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got %T", err)
	}
	if st.Code() != codes.PermissionDenied {
		t.Fatalf("code = %v, want %v", st.Code(), codes.PermissionDenied)
	}
	if issuer.issueCalls != 0 {
		t.Fatalf("issue calls = %d, want 0", issuer.issueCalls)
	}
}

func TestAuthServiceRefreshInvalidSubject(t *testing.T) {
	issuer := &stubTokenManager{refreshSubject: ""}
	svc := &AuthService{
		Exchange: events.NewExchange(),
		Logger:   &logger.DevNullLogger{},
		Issuer:   issuer,
	}

	_, err := svc.Refresh(context.Background(), &authv1.RefreshRequest{RefreshToken: "refresh-token"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("expected grpc status error, got %T", err)
	}
	if st.Code() != codes.Internal {
		t.Fatalf("code = %v, want %v", st.Code(), codes.Internal)
	}
}

func boolPointer(v bool) *bool {
	return &v
}
