package dto

import "github.com/google/uuid"

// Requests -->

// LoginRequest represents user login credentials
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"demo@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"Demo123!"`
}

// RegisterRequest represents self-registration data. Creates a new family
// with the user as its owner. base_currency is fixed to RSD (not accepted here).
type RegisterRequest struct {
	Email      string `json:"email" validate:"required,email" example:"new@example.com"`
	Password   string `json:"password" validate:"required,min=8" example:"Secret123"`
	Name       string `json:"name" validate:"required,min=1,max=100" example:"Igor Kudinov"`
	FamilyName string `json:"family_name" validate:"omitempty,min=1,max=100" example:"Kudinov Family"`
}

// Responses <--

// LoginResponse represents successful login response with JWT token
type LoginResponse struct {
	Token     string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User      UserInfo `json:"user"`
	ExpiresIn int      `json:"expires_in" example:"86400"` // seconds
}

// UserInfo represents authenticated user information
type UserInfo struct {
	ID       uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Email    string    `json:"email" example:"demo@example.com"`
	Name     string    `json:"name" example:"Demo User"`
	FamilyID uuid.UUID `json:"family_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role     string    `json:"role,omitempty" example:"owner"`
}
