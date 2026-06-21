package account_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/account"
	"github.com/DigitLock/expense-tracker/internal/testutil"
)

func setup(t *testing.T) (*account.Service, uuid.UUID, *pgxpool.Pool) {
	t.Helper()
	pool := testutil.Pool(t) // skips if TEST_DATABASE_URL unset
	testutil.Truncate(t, pool)

	famID := uuid.New()
	if _, err := pool.Exec(context.Background(),
		"INSERT INTO families (id, name) VALUES ($1, $2)", famID, "Test Family"); err != nil {
		t.Fatalf("seed family: %v", err)
	}
	return account.New(repository.New(pool).Accounts), famID, pool
}

func ptr[T any](v T) *T { return &v }

func TestCreate_Valid(t *testing.T) {
	svc, famID, _ := setup(t)

	want := decimal.RequireFromString("500.00")
	acc, err := svc.Create(context.Background(), famID, account.CreateAccountInput{
		Name: "Cash EUR", Type: "cash", Currency: "EUR", InitialBalance: want, Description: "wallet",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if acc.Name != "Cash EUR" || acc.Type != "cash" || acc.Currency != "EUR" {
		t.Errorf("unexpected account: %+v", acc)
	}
	// CreateAccount sets current_balance = initial_balance.
	if !acc.InitialBalance.Equal(want) || !acc.CurrentBalance.Equal(want) {
		t.Errorf("balance: initial=%s current=%s, want 500/500", acc.InitialBalance, acc.CurrentBalance)
	}
	if !acc.IsActive {
		t.Error("new account should be active")
	}
}

func TestCreate_Validation(t *testing.T) {
	svc, famID, _ := setup(t)
	ctx := context.Background()

	cases := []struct {
		name string
		in   account.CreateAccountInput
	}{
		{"missing name", account.CreateAccountInput{Type: "cash", Currency: "RSD"}},
		{"invalid type", account.CreateAccountInput{Name: "X", Type: "crypto", Currency: "RSD"}},
		{"invalid currency", account.CreateAccountInput{Name: "X", Type: "cash", Currency: "GBP"}},
		{"negative balance", account.CreateAccountInput{Name: "X", Type: "cash", Currency: "RSD", InitialBalance: decimal.NewFromInt(-1)}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := svc.Create(ctx, famID, tc.in)
			if !errors.Is(err, domain.ErrInvalidArgument) {
				t.Errorf("err = %v, want ErrInvalidArgument", err)
			}
		})
	}
}

func TestCreate_DuplicateName(t *testing.T) {
	svc, famID, _ := setup(t)
	ctx := context.Background()
	in := account.CreateAccountInput{Name: "Cash", Type: "cash", Currency: "RSD"}

	if _, err := svc.Create(ctx, famID, in); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := svc.Create(ctx, famID, in)
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Fatalf("duplicate create err = %v, want ErrAlreadyExists", err)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc, famID, _ := setup(t)
	_, err := svc.Update(context.Background(), famID, uuid.New(), account.UpdateAccountInput{Name: ptr("New")})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdate_WrongFamily(t *testing.T) {
	svc, famID, _ := setup(t)
	ctx := context.Background()
	acc, err := svc.Create(ctx, famID, account.CreateAccountInput{Name: "Cash", Type: "cash", Currency: "RSD"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	otherFamily := uuid.New()
	_, err = svc.Update(ctx, otherFamily, acc.ID, account.UpdateAccountInput{Name: ptr("Hacked")})
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("err = %v, want ErrPermissionDenied", err)
	}
}

func TestUpdate_PartialKeepsOtherFields(t *testing.T) {
	svc, famID, _ := setup(t)
	ctx := context.Background()
	acc, err := svc.Create(ctx, famID, account.CreateAccountInput{
		Name: "Cash", Type: "cash", Currency: "RSD", Description: "orig",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := svc.Update(ctx, famID, acc.ID, account.UpdateAccountInput{Name: ptr("Renamed")})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Renamed" {
		t.Errorf("name = %q, want Renamed", updated.Name)
	}
	// Type/currency/description untouched by a name-only update.
	if updated.Type != "cash" || updated.Currency != "RSD" || updated.Description != "orig" {
		t.Errorf("partial update changed other fields: %+v", updated)
	}
}

func TestDelete_AlreadyInactive(t *testing.T) {
	svc, famID, _ := setup(t)
	ctx := context.Background()
	acc, err := svc.Create(ctx, famID, account.CreateAccountInput{Name: "Cash", Type: "cash", Currency: "RSD"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := svc.Delete(ctx, famID, acc.ID); err != nil {
		t.Fatalf("first delete: %v", err)
	}
	err = svc.Delete(ctx, famID, acc.ID)
	if !errors.Is(err, domain.ErrFailedPrecondition) {
		t.Errorf("second delete err = %v, want ErrFailedPrecondition", err)
	}
}
