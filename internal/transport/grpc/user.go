package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/amimof/multikube/internal/app"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	userv1 "github.com/amimof/multikube/api/user/v1"
)

var _ userv1.UserServiceServer = &UserService{}

type UserService struct {
	userv1.UnimplementedUserServiceServer
	app *app.UserService
}

func (n *UserService) Register(server *grpc.Server) {
	userv1.RegisterUserServiceServer(server, n)
}

func (n *UserService) RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return userv1.RegisterUserServiceHandler(ctx, mux, conn)
}

func (n *UserService) Get(ctx context.Context, req *userv1.GetRequest) (*userv1.GetResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}
	user, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}
	user.GetConfig().Password = ""
	return &userv1.GetResponse{User: user}, nil
}

func (n *UserService) Create(ctx context.Context, req *userv1.CreateRequest) (*userv1.CreateResponse, error) {
	user, err := n.app.Create(ctx, req.GetUser())
	if err != nil {
		return nil, toStatus(err)
	}
	user.GetConfig().Password = ""
	return &userv1.CreateResponse{User: user}, nil
}

func (n *UserService) Delete(ctx context.Context, req *userv1.DeleteRequest) (*emptypb.Empty, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Delete(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func (n *UserService) List(ctx context.Context, req *userv1.ListRequest) (*userv1.ListResponse, error) {
	users, err := n.app.List(ctx, req.GetLimit())
	if err != nil {
		return nil, toStatus(err)
	}
	for _, u := range users {
		u.GetConfig().Password = ""
	}
	return &userv1.ListResponse{Users: users}, nil
}

func (n *UserService) Update(ctx context.Context, req *userv1.UpdateRequest) (*userv1.UpdateResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Update(ctx, uid, req.GetUser())
	if err != nil {
		return nil, toStatus(err)
	}

	user, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}
	user.GetConfig().Password = ""

	return &userv1.UpdateResponse{User: user}, nil
}

func (n *UserService) Patch(ctx context.Context, req *userv1.PatchRequest) (*userv1.PatchResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Patch(ctx, uid, req.GetUser())
	if err != nil {
		return nil, toStatus(err)
	}

	user, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}
	user.GetConfig().Password = ""

	return &userv1.PatchResponse{User: user}, nil
}

func NewUserService(app *app.UserService) *UserService {
	return &UserService{app: app}
}
