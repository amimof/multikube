package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/amimof/multikube/internal/app"
	"github.com/amimof/multikube/pkg/keys"

	credentialv1 "github.com/amimof/multikube/api/credential/v1"
)

var _ credentialv1.CredentialServiceServer = &CredentialService{}

type CredentialService struct {
	credentialv1.UnimplementedCredentialServiceServer
	app *app.CredentialService
}

func (n *CredentialService) Register(server *grpc.Server) {
	credentialv1.RegisterCredentialServiceServer(server, n)
}

func (n *CredentialService) Get(ctx context.Context, req *credentialv1.GetRequest) (*credentialv1.GetResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}
	credential, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}
	return &credentialv1.GetResponse{Credential: credential}, nil
}

func (n *CredentialService) Create(ctx context.Context, req *credentialv1.CreateRequest) (*credentialv1.CreateResponse, error) {
	credential, err := n.app.Create(ctx, req.GetCredential())
	if err != nil {
		return nil, toStatus(err)
	}
	return &credentialv1.CreateResponse{Credential: credential}, nil
}

func (n *CredentialService) Delete(ctx context.Context, req *credentialv1.DeleteRequest) (*emptypb.Empty, error) {
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

func (n *CredentialService) List(ctx context.Context, req *credentialv1.ListRequest) (*credentialv1.ListResponse, error) {
	credentials, err := n.app.List(ctx, req.GetLimit())
	if err != nil {
		return nil, toStatus(err)
	}
	return &credentialv1.ListResponse{Credentials: credentials}, nil
}

func (n *CredentialService) Update(ctx context.Context, req *credentialv1.UpdateRequest) (*credentialv1.UpdateResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Update(ctx, uid, req.GetCredential())
	if err != nil {
		return nil, toStatus(err)
	}

	credential, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &credentialv1.UpdateResponse{Credential: credential}, nil
}

func (n *CredentialService) Patch(ctx context.Context, req *credentialv1.PatchRequest) (*credentialv1.PatchResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Patch(ctx, uid, req.GetCredential())
	if err != nil {
		return nil, toStatus(err)
	}

	credential, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &credentialv1.PatchResponse{Credential: credential}, nil
}

func NewCredentialService(app *app.CredentialService) *CredentialService {
	return &CredentialService{app: app}
}
