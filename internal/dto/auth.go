package dto

import "github.com/google/uuid"

// Requests -->

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// Responses <--

type LoginResponse struct {
	Token     string   `json:"token"`
	User      UserInfo `json:"user"`
	ExpiresIn int      `json:"expires_in"` // seconds
}

type UserInfo struct {
	ID       uuid.UUID `json:"id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	FamilyID uuid.UUID `json:"family_id"`
}
