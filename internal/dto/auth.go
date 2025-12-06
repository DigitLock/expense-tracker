package dto

import "github.com/google/uuid"

// Requests -->

// LoginRequest represents user login credentials
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email" example:"demo@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"Demo123!"`
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
}
