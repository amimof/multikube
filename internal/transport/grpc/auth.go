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
	app   *app.AuthService
	token *app.TokenService
}

func (n *AuthService) Register(server *grpc.Server) {
	authv1.RegisterAuthServiceServer(server, n)
}

func (n *AuthService) RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return authv1.RegisterAuthServiceHandler(ctx, mux, conn)
}

func (a *AuthService) Logout(ctx context.Context, req *authv1.LogoutRequest) (*emptypb.Empty, error) {
	return a.app.Logout(ctx, req)
}

func (a *AuthService) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	return a.app.Login(ctx, req)
}

func NewAuthService(app *app.AuthService) *AuthService {
	return &AuthService{app: app}
}
