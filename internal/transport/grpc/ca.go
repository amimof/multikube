package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/amimof/multikube/internal/app"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	cav1 "github.com/amimof/multikube/api/ca/v1"
)

var _ cav1.CertificateAuthorityServiceServer = &CertificateAuthorityService{}

type CertificateAuthorityService struct {
	cav1.UnimplementedCertificateAuthorityServiceServer
	app *app.CertificateAuthorityService
}

func (n *CertificateAuthorityService) Register(server *grpc.Server) {
	cav1.RegisterCertificateAuthorityServiceServer(server, n)
}

func (n *CertificateAuthorityService) RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return cav1.RegisterCertificateAuthorityServiceHandler(ctx, mux, conn)
}

func (n *CertificateAuthorityService) Get(ctx context.Context, req *cav1.GetRequest) (*cav1.GetResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}
	ca, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}
	return &cav1.GetResponse{CertificateAuthority: ca}, nil
}

func (n *CertificateAuthorityService) Create(ctx context.Context, req *cav1.CreateRequest) (*cav1.CreateResponse, error) {
	ca, err := n.app.Create(ctx, req.GetCertificateAuthority())
	if err != nil {
		return nil, toStatus(err)
	}
	return &cav1.CreateResponse{CertificateAuthority: ca}, nil
}

func (n *CertificateAuthorityService) Delete(ctx context.Context, req *cav1.DeleteRequest) (*emptypb.Empty, error) {
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

func (n *CertificateAuthorityService) List(ctx context.Context, req *cav1.ListRequest) (*cav1.ListResponse, error) {
	cas, err := n.app.List(ctx, req.GetLimit())
	if err != nil {
		return nil, toStatus(err)
	}
	return &cav1.ListResponse{CertificateAuthoritys: cas}, nil
}

func (n *CertificateAuthorityService) Update(ctx context.Context, req *cav1.UpdateRequest) (*cav1.UpdateResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Update(ctx, uid, req.GetCertificateAuthority())
	if err != nil {
		return nil, toStatus(err)
	}

	ca, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &cav1.UpdateResponse{CertificateAuthority: ca}, nil
}

func (n *CertificateAuthorityService) Patch(ctx context.Context, req *cav1.PatchRequest) (*cav1.PatchResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Patch(ctx, uid, req.GetCertificateAuthority())
	if err != nil {
		return nil, toStatus(err)
	}

	ca, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &cav1.PatchResponse{CertificateAuthority: ca}, nil
}

func NewCertificateAuthorityService(app *app.CertificateAuthorityService) *CertificateAuthorityService {
	return &CertificateAuthorityService{app: app}
}
