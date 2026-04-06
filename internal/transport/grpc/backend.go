package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/amimof/multikube/internal/app"
	"github.com/amimof/multikube/pkg/keys"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
)

var _ backendv1.BackendServiceServer = &BackendService{}

type BackendService struct {
	backendv1.UnimplementedBackendServiceServer
	app *app.BackendService
}

func (n *BackendService) Register(server *grpc.Server) {
	backendv1.RegisterBackendServiceServer(server, n)
}

func (n *BackendService) Get(ctx context.Context, req *backendv1.GetRequest) (*backendv1.GetResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}
	node, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}
	return &backendv1.GetResponse{Backend: node}, nil
}

func (n *BackendService) Create(ctx context.Context, req *backendv1.CreateRequest) (*backendv1.CreateResponse, error) {
	node, err := n.app.Create(ctx, req.GetBackend())
	if err != nil {
		return nil, toStatus(err)
	}
	return &backendv1.CreateResponse{Backend: node}, nil
}

func (n *BackendService) Delete(ctx context.Context, req *backendv1.DeleteRequest) (*emptypb.Empty, error) {
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

func (n *BackendService) List(ctx context.Context, req *backendv1.ListRequest) (*backendv1.ListResponse, error) {
	nodes, err := n.app.List(ctx, req.GetLimit())
	if err != nil {
		return nil, toStatus(err)
	}
	return &backendv1.ListResponse{Backends: nodes}, nil
}

func (n *BackendService) Update(ctx context.Context, req *backendv1.UpdateRequest) (*backendv1.UpdateResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Update(ctx, uid, req.GetBackend())
	if err != nil {
		return nil, toStatus(err)
	}

	node, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &backendv1.UpdateResponse{Backend: node}, nil
}

func (n *BackendService) Patch(ctx context.Context, req *backendv1.PatchRequest) (*backendv1.PatchResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Patch(ctx, uid, req.GetBackend())
	if err != nil {
		return nil, toStatus(err)
	}

	node, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &backendv1.PatchResponse{Backend: node}, nil
}

func (n *BackendService) UpdateStatus(ctx context.Context, req *backendv1.UpdateStatusRequest) (*emptypb.Empty, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.UpdateStatus(ctx, uid, req.GetStatus(), req.GetUpdateMask().GetPaths()...)
	if err != nil {
		return nil, toStatus(err)
	}

	return &emptypb.Empty{}, nil
}

func NewBackendService(app *app.BackendService) *BackendService {
	return &BackendService{app: app}
}
