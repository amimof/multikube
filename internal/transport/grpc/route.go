package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/amimof/multikube/internal/app"
	"github.com/amimof/multikube/pkg/keys"

	routev1 "github.com/amimof/multikube/api/route/v1"
)

var _ routev1.RouteServiceServer = &RouteService{}

type RouteService struct {
	routev1.UnimplementedRouteServiceServer
	app *app.RouteService
}

func (n *RouteService) Register(server *grpc.Server) {
	routev1.RegisterRouteServiceServer(server, n)
}

func (n *RouteService) Get(ctx context.Context, req *routev1.GetRequest) (*routev1.GetResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}
	route, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}
	return &routev1.GetResponse{Route: route}, nil
}

func (n *RouteService) Create(ctx context.Context, req *routev1.CreateRequest) (*routev1.CreateResponse, error) {
	route, err := n.app.Create(ctx, req.GetRoute())
	if err != nil {
		return nil, toStatus(err)
	}
	return &routev1.CreateResponse{Route: route}, nil
}

func (n *RouteService) Delete(ctx context.Context, req *routev1.DeleteRequest) (*emptypb.Empty, error) {
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

func (n *RouteService) List(ctx context.Context, req *routev1.ListRequest) (*routev1.ListResponse, error) {
	routes, err := n.app.List(ctx, req.GetLimit())
	if err != nil {
		return nil, toStatus(err)
	}
	return &routev1.ListResponse{Routes: routes}, nil
}

func (n *RouteService) Update(ctx context.Context, req *routev1.UpdateRequest) (*routev1.UpdateResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Update(ctx, uid, req.GetRoute())
	if err != nil {
		return nil, toStatus(err)
	}

	route, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &routev1.UpdateResponse{Route: route}, nil
}

func (n *RouteService) Patch(ctx context.Context, req *routev1.PatchRequest) (*routev1.PatchResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Patch(ctx, uid, req.GetRoute())
	if err != nil {
		return nil, toStatus(err)
	}

	route, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &routev1.PatchResponse{Route: route}, nil
}

func (n *RouteService) UpdateStatus(ctx context.Context, req *routev1.UpdateStatusRequest) (*emptypb.Empty, error) {
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

func NewRouteService(app *app.RouteService) *RouteService {
	return &RouteService{app: app}
}
