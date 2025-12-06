package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/DigitLock/expense-tracker/internal/auth"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtService *auth.JWTService
	validate   *validator.Validate
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtService *auth.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
		validate:   validator.New(),
	}
}

// Login godoc
// @Summary User login
// @Description Authenticates user and returns JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login credentials"
// @Success 200 {object} dto.SuccessResponse{data=dto.LoginResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		validationErrors := formatValidationErrors(err)
		writeValidationError(w, validationErrors)
		return
	}

	user, err := h.userRepo.Authenticate(r.Context(), req.Email, req.Password)
	if err != nil {
		// Don't reveal whether email exists or password is wrong
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Invalid email or password")
		return
	}

	if !user.IsActive {
		writeError(w, http.StatusUnauthorized, "USER_INACTIVE", "User account is inactive")
		return
	}

	// Generate JWT token
	token, _, err := h.jwtService.GenerateToken(user.ID, user.FamilyID, user.Email, user.Name)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", "Failed to generate token")
		return
	}

	response := dto.LoginResponse{
		Token: token,
		User: dto.UserInfo{
			ID:       user.ID,
			Email:    user.Email,
			Name:     user.Name,
			FamilyID: user.FamilyID,
		},
		ExpiresIn: h.jwtService.GetExpirationSeconds(),
	}

	writeSuccess(w, http.StatusOK, response)
}

// Helper functions
func writeSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.NewSuccessResponse(data))
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.NewErrorResponse(code, message, nil))
}

func writeValidationError(w http.ResponseWriter, details []dto.ValidationError) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(dto.NewErrorResponse("VALIDATION_ERROR", "Invalid input", details))
}

func formatValidationErrors(err error) []dto.ValidationError {
	var errors []dto.ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, dto.ValidationError{
				Field:   e.Field(),
				Message: formatValidationMessage(e),
			})
		}
	}

	return errors
}

func formatValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		return "Value is too short"
	case "max":
		return "Value is too long"
	default:
		return "Invalid value"
	}
}
