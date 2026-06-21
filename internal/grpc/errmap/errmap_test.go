package errmap

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"

	"github.com/DigitLock/expense-tracker/internal/domain"
)

func TestToStatus(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want codes.Code
	}{
		{"invalid_argument", domain.ErrInvalidArgument, codes.InvalidArgument},
		{"not_found", domain.ErrNotFound, codes.NotFound},
		{"already_exists", domain.ErrAlreadyExists, codes.AlreadyExists},
		{"permission_denied", domain.ErrPermissionDenied, codes.PermissionDenied},
		{"failed_precondition", domain.ErrFailedPrecondition, codes.FailedPrecondition},
		{"unauthenticated", domain.ErrUnauthenticated, codes.Unauthenticated},
		{"wrapped_not_found", fmt.Errorf("account %s: %w", "abc", domain.ErrNotFound), codes.NotFound},
		{"unknown_internal", errors.New("boom"), codes.Internal},
		{"nil_ok", nil, codes.OK},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			st := ToStatus(tc.err)
			if st.Code() != tc.want {
				t.Errorf("ToStatus(%v).Code() = %v, want %v", tc.err, st.Code(), tc.want)
			}
		})
	}
}

func TestToStatus_MessageIsErrorText(t *testing.T) {
	err := fmt.Errorf("account %q: %w", "Cash EUR", domain.ErrAlreadyExists)
	st := ToStatus(err)
	if st.Message() != err.Error() {
		t.Errorf("message = %q, want %q", st.Message(), err.Error())
	}
}
