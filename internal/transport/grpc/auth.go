package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/amimof/multikube/internal/app"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	authv1 "github.com/amimof/multikube/api/auth/v1"
)

var _ authv1.AuthServiceServer = &AuthService{}

type AuthService struct {
	authv1.UnimplementedAuthServiceServer
	app *app.AuthService
}

func (n *AuthService) Register(server *grpc.Server) {
	authv1.RegisterAuthServiceServer(server, n)
}

func (n *AuthService) RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return authv1.RegisterAuthServiceHandler(ctx, mux, conn)
}

func (n *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*emptypb.Empty, error) {
	return n.app.Logout(ctx, req)
}

func (n *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return n.app.Login(ctx, req)
}

func (n *AuthService) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	return n.app.Refresh(ctx, req)
}

func NewAuthService(app *app.AuthService) *AuthService {
	return &AuthService{app: app}
}
