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
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "github.com/amimof/multikube/api/auth/v1"
	tokenv1 "github.com/amimof/multikube/api/token/v1"
)

type AuthService struct {
	Exchange *events.Exchange
	Logger   logger.Logger
	Users    UsersGetter
	Issuer   TokenManager
}

func (a *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*empty.Empty, error) {
	_, err := a.Issuer.VerifyAccessToken(ctx, req.GetAccessToken())
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
		return nil, status.Errorf(codes.PermissionDenied, "incorrect username/password")
	}

	if !u.GetConfig().GetEnabled() {
		return nil, status.Errorf(codes.PermissionDenied, "account disabled")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.GetConfig().GetPassword()), []byte(req.GetPassword())); err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "incorrect username/password")
	}

	tokenReq := &tokenv1.Token{
		Config: &tokenv1.TokenConfig{
			Subject: u.GetMeta().GetName(),
			Roles:   u.GetConfig().GetRoles(),
		},
	}

	token, err := a.Issuer.Issue(ctx, tokenReq)
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "cannot generate access token")
	}

	// Publish event that user logged in
	err = a.Exchange.Forward(ctx, events.NewEvent(events.AuthLogin, nil))
	if err != nil {
		a.Logger.Error("error publishing token issuance event", "error", err, "username", req.GetUsername())
		return nil, err
	}

	return &authv1.LoginResponse{AccessToken: token.AccessToken, RefreshToken: token.RefreshToken}, nil
}

func (a *AuthService) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	userClaims, err := a.Issuer.VerifyRefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "refresh token is invalid: %v", err)
	}

	userID, _ := userClaims["sub"].(string)

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
			Roles:   u.GetConfig().GetRoles(),
		},
	}

	token, err := a.Issuer.Issue(ctx, tokenReq)
	if err != nil {
		return nil, err
	}

	resp := &authv1.RefreshResponse{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}
	if !token.ExpiresAt.IsZero() {
		resp.ExpiresAt = timestamppb.New(token.ExpiresAt)
	}

	return resp, nil
}
