// Package errmap maps domain sentinel errors to gRPC statuses. It is a leaf
// package (imported by gRPC handlers) so it does not create an import cycle
// with the gRPC server package.
package errmap

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/DigitLock/expense-tracker/internal/domain"
)

// ToStatus maps a domain sentinel error to a *grpc status (TDD §3). It uses
// errors.Is, so wrapped errors (fmt.Errorf("%w: ...", domain.ErrXxx)) map
// correctly. The status message is the error's full text. Unknown errors map to
// Internal so internal details are reported under a generic code.
//
// This helper is self-contained: it is NOT yet wired into repositories, REST,
// or handlers (that happens in later stages).
func ToStatus(err error) *status.Status {
	switch {
	case err == nil:
		return status.New(codes.OK, "")
	case errors.Is(err, domain.ErrInvalidArgument):
		return status.New(codes.InvalidArgument, err.Error())
	case errors.Is(err, domain.ErrNotFound):
		return status.New(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrAlreadyExists):
		return status.New(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrPermissionDenied):
		return status.New(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrFailedPrecondition):
		return status.New(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrUnauthenticated):
		return status.New(codes.Unauthenticated, err.Error())
	default:
		return status.New(codes.Internal, err.Error())
	}
}
