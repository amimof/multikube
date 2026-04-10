package grpc

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/amimof/multikube/internal/app"
	"github.com/amimof/multikube/pkg/keys"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	certv1 "github.com/amimof/multikube/api/certificate/v1"
)

var _ certv1.CertificateServiceServer = &CertificateService{}

type CertificateService struct {
	certv1.UnimplementedCertificateServiceServer
	app *app.CertificateService
}

func (n *CertificateService) Register(server *grpc.Server) {
	certv1.RegisterCertificateServiceServer(server, n)
}

func (n *CertificateService) RegisterHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return certv1.RegisterCertificateServiceHandler(ctx, mux, conn)
}

func (n *CertificateService) Get(ctx context.Context, req *certv1.GetRequest) (*certv1.GetResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}
	cert, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}
	return &certv1.GetResponse{Certificate: cert}, nil
}

func (n *CertificateService) Create(ctx context.Context, req *certv1.CreateRequest) (*certv1.CreateResponse, error) {
	cert, err := n.app.Create(ctx, req.GetCertificate())
	if err != nil {
		return nil, toStatus(err)
	}
	return &certv1.CreateResponse{Certificate: cert}, nil
}

func (n *CertificateService) Delete(ctx context.Context, req *certv1.DeleteRequest) (*emptypb.Empty, error) {
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

func (n *CertificateService) List(ctx context.Context, req *certv1.ListRequest) (*certv1.ListResponse, error) {
	certs, err := n.app.List(ctx, req.GetLimit())
	if err != nil {
		return nil, toStatus(err)
	}
	return &certv1.ListResponse{Certificates: certs}, nil
}

func (n *CertificateService) Update(ctx context.Context, req *certv1.UpdateRequest) (*certv1.UpdateResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Update(ctx, uid, req.GetCertificate())
	if err != nil {
		return nil, toStatus(err)
	}

	cert, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &certv1.UpdateResponse{Certificate: cert}, nil
}

func (n *CertificateService) Patch(ctx context.Context, req *certv1.PatchRequest) (*certv1.PatchResponse, error) {
	uid, err := keys.FromUIDOrName(req.GetUid(), req.GetName())
	if err != nil {
		return nil, toStatus(err)
	}

	err = n.app.Patch(ctx, uid, req.GetCertificate())
	if err != nil {
		return nil, toStatus(err)
	}

	cert, err := n.app.Get(ctx, uid)
	if err != nil {
		return nil, toStatus(err)
	}

	return &certv1.PatchResponse{Certificate: cert}, nil
}

func NewCertificateService(app *app.CertificateService) *CertificateService {
	return &CertificateService{app: app}
}
