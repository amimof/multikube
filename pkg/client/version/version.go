package version

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
	cav1 "github.com/amimof/multikube/api/ca/v1"
	certv1 "github.com/amimof/multikube/api/certificate/v1"
	routev1 "github.com/amimof/multikube/api/route/v1"
)

var apiVersionByFullName = map[protoreflect.FullName]string{
	"backend.v1.Backend":         "backend/v1",
	"ca.v1.CertificateAuthority": "ca/v1",
	"certificate.v1.Certificate": "certificate/v1",
	"route.v1.Route":             "route/v1",
}

var (
	VersionBackend              = Version((&backendv1.Backend{}))
	VersionCertificateAuthority = Version((&cav1.CertificateAuthority{}))
	VersionCertificate          = Version((&certv1.Certificate{}))
	VersionRoute                = Version((&routev1.Route{}))
)

func Version(m proto.Message) string {
	return apiVersionByFullName[m.ProtoReflect().Descriptor().FullName()]
}
