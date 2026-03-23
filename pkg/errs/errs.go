package errs

import (
	"errors"
	"os"

	"github.com/amimof/multikube/pkg/repository"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrLeaseHeld = errors.New("lease: already held by another holder")

func ToStatus(err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok {
		// already a status (maybe from downstream grpc call)
		return st.Err()
	}

	switch {
	case IsNotFound(err):
		return status.Error(codes.NotFound, err.Error())
	case IsConflict(err):
		return status.Error(codes.AlreadyExists, err.Error())
	case IsPermissionDenied(err):
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

func IsPermissionDenied(err error) bool {
	if err == nil {
		return false
	}

	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.PermissionDenied {
			return true
		}
	}

	// switch {
	// case errors.Is(err, domain.ErrInvalidToken):
	// 	return true
	// }

	return false
}

func IsConflict(err error) bool {
	if err == nil {
		return false
	}

	// grpc errors
	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.AlreadyExists {
			return true
		}
	}

	switch {
	case errors.Is(err, ErrLeaseHeld):
		return true
	// case errors.Is(err, domain.ErrLeaseHeld):
	// 	return true
	case errors.Is(err, repository.ErrIdxExists):
		return true
	}

	return false
}

func IsNotFound(err error) bool {
	var b bool

	// grpc errors
	if st, ok := status.FromError(err); ok {
		if st.Code() == codes.NotFound {
			b = true
		}
	}

	// containerd errors
	// if errdefs.IsNotFound(err) {
	// 	b = true
	// }

	// repo errors
	if errors.Is(err, repository.ErrNotFound) {
		b = true
	}

	// os errors
	if errors.Is(err, os.ErrNotExist) {
		b = true
	}

	return b
}

func IgnoreNotFound(err error) error {
	if IsNotFound(err) {
		return nil
	}
	return err
}
