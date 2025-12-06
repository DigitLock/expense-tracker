package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/DigitLock/expense-tracker/internal/auth"
	"github.com/DigitLock/expense-tracker/internal/dto"
)

type contextKey string

const (
	UserIDKey    contextKey = "user_id"
	FamilyIDKey  contextKey = "family_id"
	UserEmailKey contextKey = "user_email"
	UserNameKey  contextKey = "user_name"
)

func Auth(jwtService *auth.JWTService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get token from header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				unauthorizedResponse(w, "MISSING_TOKEN", "Missing authorization header")
				return
			}

			// Check Bearer prefix
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				unauthorizedResponse(w, "INVALID_TOKEN_FORMAT", "Invalid authorization header format")
				return
			}

			tokenString := parts[1]
			if tokenString == "" {
				unauthorizedResponse(w, "MISSING_TOKEN", "Missing token")
				return
			}

			// Validate JWT token
			claims, err := jwtService.ValidateToken(tokenString)
			if err != nil {
				if err == auth.ErrExpiredToken {
					unauthorizedResponse(w, "TOKEN_EXPIRED", "Token has expired")
					return
				}
				unauthorizedResponse(w, "INVALID_TOKEN", "Invalid token")
				return
			}

			// Add claims to context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, FamilyIDKey, claims.FamilyID)
			ctx = context.WithValue(ctx, UserEmailKey, claims.Email)
			ctx = context.WithValue(ctx, UserNameKey, claims.Name)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func unauthorizedResponse(w http.ResponseWriter, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(dto.NewErrorResponse(code, message, nil))
}

// Context helpers

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	userID, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return userID, ok
}

func GetFamilyID(ctx context.Context) (uuid.UUID, bool) {
	familyID, ok := ctx.Value(FamilyIDKey).(uuid.UUID)
	return familyID, ok
}

func GetUserEmail(ctx context.Context) (string, bool) {
	email, ok := ctx.Value(UserEmailKey).(string)
	return email, ok
}

func GetUserName(ctx context.Context) (string, bool) {
	name, ok := ctx.Value(UserNameKey).(string)
	return name, ok
}
