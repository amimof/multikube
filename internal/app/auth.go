package app

import (
	"context"

	"github.com/amimof/multikube/pkg/events"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/amimof/multikube/pkg/logger"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "github.com/amimof/multikube/api/auth/v1"
	tokenv1 "github.com/amimof/multikube/api/token/v1"
)

type AuthService struct {
	// Repo     *repository.Repo[*authv1.Auth]
	Exchange *events.Exchange
	Logger   logger.Logger
	Users    UsersGetter
	// Manager  JWTManager
	Issuser TokenManager
}

func (a *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*empty.Empty, error) {
	_, err := a.Issuser.VerifyAccessToken(ctx, req.GetAccessToken())
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid access token")
	}

	// Publish event that user logged out
	err = a.Exchange.Forward(ctx, events.NewEvent(events.AuthLogout, nil))
	if err != nil {
		a.Logger.Error("error publishing token issuance event", "error", err)
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (a *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	userID, err := keys.Name(req.GetUsername())
	if err != nil {
		return nil, err
	}

	u, err := a.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if !u.GetConfig().GetEnabled() {
		return nil, status.Errorf(codes.PermissionDenied, "account disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.GetConfig().GetPassword()), []byte(req.GetPassword())); err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "incorrect username/password %s", err.Error())
	}

	tokenReq := &tokenv1.Token{
		Config: &tokenv1.TokenConfig{
			Subject: u.GetMeta().GetName(),
		},
	}

	accessToken, refreshToken, err := a.Issuser.Issue(ctx, tokenReq)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "cannot generate access token")
	}

	// Publish event that user logged in
	err = a.Exchange.Forward(ctx, events.NewEvent(events.AuthLogin, nil))
	if err != nil {
		a.Logger.Error("error publishing token issuance event", "error", err, "username", req.GetUsername())
		return nil, err
	}

	return &authv1.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (a *AuthService) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	userID, err := a.Issuser.VerifyRefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "refresh token is invalid: %v", err)
	}

	uid, err := keys.ParseStr(userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "error parsing userid: %v", err)
	}

	u, err := a.Users.Get(ctx, uid)
	if err != nil {
		return nil, err
	}
	if !u.GetConfig().GetEnabled() {
		return nil, status.Errorf(codes.PermissionDenied, "account disabled")
	}

	tokenReq := &tokenv1.Token{
		Config: &tokenv1.TokenConfig{
			Subject: u.GetMeta().GetName(),
		},
	}

	token, _, err := a.Issuser.Issue(ctx, tokenReq)
	if err != nil {
		return nil, err
	}

	return &authv1.RefreshResponse{
		AccessToken: token,
	}, nil
}
