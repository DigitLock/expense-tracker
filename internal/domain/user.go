package domain

import "github.com/google/uuid"

// User is the transport-agnostic authenticated-user model.
type User struct {
	ID       string
	Email    string
	Name     string
	FamilyID uuid.UUID
}

// AuthResult is the result of a successful login.
type AuthResult struct {
	Token     string
	User      User
	ExpiresIn int // seconds
}
