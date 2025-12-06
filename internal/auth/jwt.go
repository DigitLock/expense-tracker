package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID   uuid.UUID `json:"user_id"`
	FamilyID uuid.UUID `json:"family_id"`
	Email    string    `json:"email"`
	Name     string    `json:"name"`
	jwt.RegisteredClaims
}

type JWTService struct {
	secret          []byte
	expirationHours int
}

func NewJWTService(secret string, expirationHours int) *JWTService {
	return &JWTService{
		secret:          []byte(secret),
		expirationHours: expirationHours,
	}
}

func (s *JWTService) GenerateToken(userID, familyID uuid.UUID, email, name string) (string, time.Time, error) {
	expiresAt := time.Now().Add(time.Duration(s.expirationHours) * time.Hour)

	claims := &Claims{
		UserID:   userID,
		FamilyID: familyID,
		Email:    email,
		Name:     name,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "expense-tracker",
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return s.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func (s *JWTService) GetExpirationSeconds() int {
	return s.expirationHours * 3600
}
