package repository_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/testutil"

	"github.com/jackc/pgx/v5/pgxpool"
)

func countFamiliesUsers(t *testing.T, pool *pgxpool.Pool) (families, users int) {
	t.Helper()
	ctx := context.Background()
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM families").Scan(&families); err != nil {
		t.Fatalf("count families: %v", err)
	}
	if err := pool.QueryRow(ctx, "SELECT count(*) FROM users").Scan(&users); err != nil {
		t.Fatalf("count users: %v", err)
	}
	return families, users
}

func TestUserRepository_Register_HappyPath(t *testing.T) {
	pool := testutil.Pool(t)
	testutil.Truncate(t, pool)
	repos := repository.New(pool)
	ctx := context.Background()

	user, err := repos.Users.Register(ctx, repository.RegisterInput{
		Email:      "owner@example.com",
		Password:   "Secret123",
		Name:       "Test Owner",
		FamilyName: "Test Family",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}

	// Returned user is the family owner.
	if user.Role != "owner" {
		t.Errorf("user.Role = %q, want \"owner\"", user.Role)
	}

	// Password is bcrypt-hashed (cost 12), never stored as plaintext.
	if !strings.HasPrefix(user.PasswordHash, "$2a$12$") {
		t.Errorf("user.PasswordHash = %q, want bcrypt cost-12 prefix $2a$12$", user.PasswordHash)
	}
	if strings.Contains(user.PasswordHash, "Secret123") {
		t.Error("password stored in plaintext")
	}

	// Family was created with base_currency RSD.
	var baseCurrency string
	if err := pool.QueryRow(ctx,
		"SELECT base_currency FROM families WHERE id = $1", user.FamilyID,
	).Scan(&baseCurrency); err != nil {
		t.Fatalf("query family base_currency: %v", err)
	}
	if baseCurrency != "RSD" {
		t.Errorf("family.base_currency = %q, want \"RSD\"", baseCurrency)
	}

	// Default categories were seeded for the new family.
	var catCount int
	if err := pool.QueryRow(ctx,
		"SELECT count(*) FROM categories WHERE family_id = $1", user.FamilyID,
	).Scan(&catCount); err != nil {
		t.Fatalf("count categories: %v", err)
	}
	if catCount == 0 {
		t.Error("no default categories were seeded for the new family")
	}
}

func TestUserRepository_Register_DuplicateEmail(t *testing.T) {
	pool := testutil.Pool(t)
	testutil.Truncate(t, pool)
	repos := repository.New(pool)
	ctx := context.Background()

	const email = "dup@example.com"
	if _, err := repos.Users.Register(ctx, repository.RegisterInput{
		Email:      email,
		Password:   "Secret123",
		Name:       "First User",
		FamilyName: "First Family",
	}); err != nil {
		t.Fatalf("first Register: %v", err)
	}

	famBefore, usrBefore := countFamiliesUsers(t, pool)

	_, err := repos.Users.Register(ctx, repository.RegisterInput{
		Email:      email,
		Password:   "Other123",
		Name:       "Second User",
		FamilyName: "Second Family",
	})
	if !errors.Is(err, repository.ErrEmailExists) {
		t.Fatalf("second Register error = %v, want ErrEmailExists", err)
	}

	// The whole transaction must roll back: no orphan family or user.
	famAfter, usrAfter := countFamiliesUsers(t, pool)
	if famAfter != famBefore {
		t.Errorf("families count %d -> %d, want unchanged (rollback failed)", famBefore, famAfter)
	}
	if usrAfter != usrBefore {
		t.Errorf("users count %d -> %d, want unchanged (rollback failed)", usrBefore, usrAfter)
	}
}
