// Package auth holds the application-layer service for the Auth domain. It sits
// on top of the existing JWTService (token generation/validation is NOT
// duplicated here) and the user repository. Roles are not placed in the token
// (deferred to v0.8.0).
package auth

import (
	"context"
	"time"

	"github.com/google/uuid"

	jwtauth "github.com/DigitLock/expense-tracker/internal/auth"
	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/domain"
)

// Authenticator verifies credentials (email lookup + bcrypt). *repository.UserRepository satisfies it.
type Authenticator interface {
	Authenticate(ctx context.Context, email, password string) (sqlc.User, error)
}

// Tokens is the existing JWT service (token generation/validation reused, not duplicated).
type Tokens interface {
	GenerateToken(userID, familyID uuid.UUID, email, name string) (string, time.Time, error)
	ValidateToken(token string) (*jwtauth.Claims, error)
	GetExpirationSeconds() int
}

type Service struct {
	users  Authenticator
	tokens Tokens
}

func New(users Authenticator, tokens Tokens) *Service {
	return &Service{users: users, tokens: tokens}
}

// Login verifies credentials and issues a JWT. Empty fields → ErrInvalidArgument;
// bad credentials → ErrUnauthenticated. The role is intentionally not in the token.
func (s *Service) Login(ctx context.Context, email, password string) (domain.AuthResult, error) {
	if email == "" || password == "" {
		return domain.AuthResult{}, domain.Errorf(domain.ErrInvalidArgument, "email and password are required")
	}

	user, err := s.users.Authenticate(ctx, email, password)
	if err != nil {
		return domain.AuthResult{}, domain.Errorf(domain.ErrUnauthenticated, "invalid email or password")
	}

	token, _, err := s.tokens.GenerateToken(user.ID, user.FamilyID, user.Email, user.Name)
	if err != nil {
		return domain.AuthResult{}, err // internal
	}

	return domain.AuthResult{
		Token: token,
		User: domain.User{
			ID:       user.ID.String(),
			Email:    user.Email,
			Name:     user.Name,
			FamilyID: user.FamilyID,
		},
		ExpiresIn: s.tokens.GetExpirationSeconds(),
	}, nil
}

// ValidateToken validates a JWT and returns the user context. An invalid or
// expired token is reported as valid=false (not an error). Empty token →
// ErrInvalidArgument.
func (s *Service) ValidateToken(_ context.Context, token string) (bool, domain.User, error) {
	if token == "" {
		return false, domain.User{}, domain.Errorf(domain.ErrInvalidArgument, "token is required")
	}

	claims, err := s.tokens.ValidateToken(token)
	if err != nil {
		return false, domain.User{}, nil // invalid/expired — not an error, just not valid
	}

	return true, domain.User{
		ID:       claims.UserID.String(),
		Email:    claims.Email,
		Name:     claims.Name,
		FamilyID: claims.FamilyID,
	}, nil
}
