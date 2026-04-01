package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/amimof/multikube/internal/app"
	"github.com/amimof/multikube/pkg/keys"

	tokenv1 "github.com/amimof/multikube/api/token/v1"
)

var _ tokenv1.TokenServiceServer = &TokenService{}

type TokenService struct {
	tokenv1.UnimplementedTokenServiceServer
	app *app.TokenService
}

func (n *TokenService) Register(server *grpc.Server) {
	tokenv1.RegisterTokenServiceServer(server, n)
}

func (n *TokenService) Issue(ctx context.Context, req *tokenv1.IssueRequest) (*tokenv1.IssueResponse, error) {
	tokenResp, err := n.app.IssueToken(ctx, req.GetToken().GetConfig())
	if err != nil {
		return nil, toStatus(err)
	}
	return tokenResp, nil
}

func (n *TokenService) Revoke(ctx context.Context, req *tokenv1.RevokeRequest) (*emptypb.Empty, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Revoke(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func NewTokenService(app *app.TokenService) *TokenService {
	return &TokenService{app: app}
}
