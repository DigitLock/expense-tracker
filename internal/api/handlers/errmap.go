package handlers

import (
	"errors"
	"net/http"

	"github.com/DigitLock/expense-tracker/internal/domain"
)

// writeDomainError maps a domain sentinel error to the project's standard error
// envelope (dto.NewErrorResponse via writeError), preserving the existing
// {success,error:{code,message}} JSON shape. The message is the domain error's
// own text. Unknown errors map to 500 INTERNAL_ERROR.
func writeDomainError(w http.ResponseWriter, err error) {
	status, code := domainErrorToHTTP(err)
	writeError(w, status, code, err.Error())
}

func domainErrorToHTTP(err error) (int, string) {
	switch {
	case errors.Is(err, domain.ErrInvalidArgument):
		return http.StatusBadRequest, "VALIDATION_ERROR"
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND"
	case errors.Is(err, domain.ErrAlreadyExists):
		return http.StatusConflict, "ALREADY_EXISTS"
	case errors.Is(err, domain.ErrPermissionDenied):
		return http.StatusForbidden, "FORBIDDEN"
	case errors.Is(err, domain.ErrFailedPrecondition):
		return http.StatusConflict, "FAILED_PRECONDITION"
	case errors.Is(err, domain.ErrUnauthenticated):
		return http.StatusUnauthorized, "UNAUTHENTICATED"
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR"
	}
}
