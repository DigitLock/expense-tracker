// Package domain holds cross-layer domain primitives shared by REST and gRPC.
package domain

import (
	"errors"
	"fmt"
)

// Sentinel domain errors. Handlers wrap these with context using
// fmt.Errorf("%w: ...", ErrXxx); callers detect them with errors.Is. The gRPC
// layer maps them to status codes (see internal/grpc/errmap.go). UNAUTHENTICATED
// is intentionally absent — it is handled by the gRPC auth interceptor.
var (
	// ErrNotFound: the requested resource does not exist (or not in this family).
	ErrNotFound = errors.New("not found")
	// ErrAlreadyExists: a uniqueness constraint would be violated.
	ErrAlreadyExists = errors.New("already exists")
	// ErrPermissionDenied: the resource exists but belongs to another family.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrFailedPrecondition: the resource is in a state that forbids the action
	// (e.g. already inactive, has active dependents).
	ErrFailedPrecondition = errors.New("failed precondition")
	// ErrInvalidArgument: the request is malformed or fails validation.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrUnauthenticated: credentials are missing or invalid (maps to 401 / Unauthenticated).
	ErrUnauthenticated = errors.New("unauthenticated")
)

// domainError carries a caller-facing message while remaining errors.Is-matchable
// to its sentinel category. This lets the service return a precise message
// (e.g. "name is required") that the gRPC mapper surfaces verbatim, while
// classification (InvalidArgument/NotFound/...) stays driven by the sentinel.
type domainError struct {
	sentinel error
	msg      string
}

func (e *domainError) Error() string { return e.msg }
func (e *domainError) Unwrap() error { return e.sentinel }

// Errorf builds a domain error categorized by sentinel with a formatted message.
// errors.Is(Errorf(ErrNotFound, ...), ErrNotFound) is true.
func Errorf(sentinel error, format string, args ...any) error {
	return &domainError{sentinel: sentinel, msg: fmt.Sprintf(format, args...)}
}
