package version

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	backendv1 "github.com/amimof/multikube/api/backend/v1"
)

var apiVersionByFullName = map[protoreflect.FullName]string{
	"backend.v1.Backend": "backend/v1",
}

var VersionBackend = Version((&backendv1.Backend{}))

func Version(m proto.Message) string {
	return apiVersionByFullName[m.ProtoReflect().Descriptor().FullName()]
}
