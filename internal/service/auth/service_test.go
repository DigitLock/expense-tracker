package auth_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	jwtauth "github.com/DigitLock/expense-tracker/internal/auth"
	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/repository"
	authsvc "github.com/DigitLock/expense-tracker/internal/service/auth"
	"github.com/DigitLock/expense-tracker/internal/testutil"
)

const (
	testEmail    = "auth_svc@example.com"
	testPassword = "Secret123"
)

func setup(t *testing.T) *authsvc.Service {
	t.Helper()
	pool := testutil.Pool(t)
	testutil.Truncate(t, pool)

	repos := repository.New(pool)
	// Register seeds a family + owner user with a bcrypt-hashed password.
	if _, err := repos.Users.Register(context.Background(), repository.RegisterInput{
		Email: testEmail, Password: testPassword, Name: "Auth User", FamilyName: "Fam",
	}); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	jwt := jwtauth.NewJWTService("test-secret", 24)
	return authsvc.New(repos.Users, jwt)
}

func TestLogin_Valid(t *testing.T) {
	svc := setup(t)
	res, err := svc.Login(context.Background(), testEmail, testPassword)
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if res.Token == "" {
		t.Error("token is empty")
	}
	if res.User.Email != testEmail || res.User.ID == "" {
		t.Errorf("user = %+v, want email %s and non-empty id", res.User, testEmail)
	}
	if res.ExpiresIn != 24*3600 {
		t.Errorf("expires_in = %d, want 86400", res.ExpiresIn)
	}
}

func TestLogin_BadCredentials(t *testing.T) {
	svc := setup(t)
	_, err := svc.Login(context.Background(), testEmail, "wrongpassword")
	if !errors.Is(err, domain.ErrUnauthenticated) {
		t.Errorf("err = %v, want ErrUnauthenticated", err)
	}
}

func TestLogin_MissingField(t *testing.T) {
	svc := setup(t)
	if _, err := svc.Login(context.Background(), "", testPassword); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("empty email err = %v, want ErrInvalidArgument", err)
	}
	if _, err := svc.Login(context.Background(), testEmail, ""); !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("empty password err = %v, want ErrInvalidArgument", err)
	}
}

func TestValidateToken_Valid(t *testing.T) {
	svc := setup(t)
	res, err := svc.Login(context.Background(), testEmail, testPassword)
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	valid, user, err := svc.ValidateToken(context.Background(), res.Token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if !valid {
		t.Fatal("valid = false, want true")
	}
	if user.Email != testEmail {
		t.Errorf("user.Email = %s, want %s", user.Email, testEmail)
	}
	if _, perr := uuid.Parse(user.ID); perr != nil {
		t.Errorf("user.ID = %q not a uuid", user.ID)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	svc := setup(t)
	valid, _, err := svc.ValidateToken(context.Background(), "garbage.token.value")
	if err != nil {
		t.Errorf("err = %v, want nil (invalid token is not an error)", err)
	}
	if valid {
		t.Error("valid = true, want false for garbage token")
	}
}

func TestValidateToken_Empty(t *testing.T) {
	svc := setup(t)
	_, _, err := svc.ValidateToken(context.Background(), "")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("empty token err = %v, want ErrInvalidArgument", err)
	}
}
