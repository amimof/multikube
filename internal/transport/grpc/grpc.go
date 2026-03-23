package grpc

import (
	"github.com/amimof/multikube/pkg/errs"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toStatus(err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		// already a status (maybe from downstream grpc call)
		return st.Err()
	}

	switch {
	case errs.IsNotFound(err):
		return status.Error(codes.NotFound, err.Error())
	case errs.IsConflict(err):
		return status.Error(codes.AlreadyExists, err.Error())
	case errs.IsPermissionDenied(err):
		return status.Error(codes.PermissionDenied, err.Error())
	// case errs.IsValidation(err):
	// 	return status.Error(codes.InvalidArgument, err.Error())
	// case errors.Is(err, context.Canceled):
	// 	return status.Error(codes.Canceled, "request canceled")
	// case errors.Is(err, context.DeadlineExceeded):
	// 	return status.Error(codes.DeadlineExceeded, "deadline exceeded")
	default:
		// avoid leaking internal error strings to clients
		return status.Errorf(codes.Internal, "internal error: %v", err)
	}
}
