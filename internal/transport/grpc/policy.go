package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/amimof/multikube/internal/app"
	"github.com/amimof/multikube/pkg/keys"

	policyv1 "github.com/amimof/multikube/api/policy/v1"
)

var _ policyv1.PolicyServiceServer = &PolicyService{}

type PolicyService struct {
	policyv1.UnimplementedPolicyServiceServer
	app *app.PolicyService
}

func (n *PolicyService) Register(server *grpc.Server) {
	policyv1.RegisterPolicyServiceServer(server, n)
}

func (n *PolicyService) Get(ctx context.Context, req *policyv1.GetRequest) (*policyv1.GetResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}
	policy, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}
	return &policyv1.GetResponse{Policy: policy}, nil
}

func (n *PolicyService) Create(ctx context.Context, req *policyv1.CreateRequest) (*policyv1.CreateResponse, error) {
	policy, err := n.app.Create(ctx, req.GetPolicy())
	if err != nil {
		return nil, toStatus(err)
	}
	return &policyv1.CreateResponse{Policy: policy}, nil
}

func (n *PolicyService) Delete(ctx context.Context, req *policyv1.DeleteRequest) (*emptypb.Empty, error) {
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

func (n *PolicyService) List(ctx context.Context, req *policyv1.ListRequest) (*policyv1.ListResponse, error) {
	policies, err := n.app.List(ctx, req.GetLimit())
	if err != nil {
		return nil, toStatus(err)
	}
	return &policyv1.ListResponse{Policys: policies}, nil
}

func (n *PolicyService) Update(ctx context.Context, req *policyv1.UpdateRequest) (*policyv1.UpdateResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Update(ctx, uid, req.GetPolicy())
	if err != nil {
		return nil, toStatus(err)
	}

	policy, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &policyv1.UpdateResponse{Policy: policy}, nil
}

func (n *PolicyService) Patch(ctx context.Context, req *policyv1.PatchRequest) (*policyv1.PatchResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Patch(ctx, uid, req.GetPolicy())
	if err != nil {
		return nil, toStatus(err)
	}

	policy, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &policyv1.PatchResponse{Policy: policy}, nil
}

func NewPolicyService(app *app.PolicyService) *PolicyService {
	return &PolicyService{app: app}
}
