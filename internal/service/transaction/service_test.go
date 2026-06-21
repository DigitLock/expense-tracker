package transaction_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/transaction"
	"github.com/DigitLock/expense-tracker/internal/testutil"
)

type fixture struct {
	svc      *transaction.Service
	repos    *repository.Repositories
	pool     *pgxpool.Pool
	familyID uuid.UUID
	userID   uuid.UUID
}

// setup seeds a family, a user, two RSD accounts and an expense category.
func setup(t *testing.T) (fixture, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	pool := testutil.Pool(t) // skips if TEST_DATABASE_URL unset
	testutil.Truncate(t, pool)
	ctx := context.Background()

	famID := uuid.New()
	if _, err := pool.Exec(ctx, "INSERT INTO families (id, name) VALUES ($1,$2)", famID, "Test Family"); err != nil {
		t.Fatalf("seed family: %v", err)
	}
	userID := uuid.New()
	if _, err := pool.Exec(ctx,
		"INSERT INTO users (id, family_id, email, password_hash, name, role) VALUES ($1,$2,$3,$4,$5,$6)",
		userID, famID, "u@example.com", "x", "U", "owner"); err != nil {
		t.Fatalf("seed user: %v", err)
	}

	repos := repository.New(pool)
	accA, err := repos.Accounts.Create(ctx, repository.CreateAccountInput{
		FamilyID: famID, Name: "A", Type: "cash", Currency: "RSD", InitialBalance: decimal.NewFromInt(1000),
	})
	if err != nil {
		t.Fatalf("seed account A: %v", err)
	}
	accB, err := repos.Accounts.Create(ctx, repository.CreateAccountInput{
		FamilyID: famID, Name: "B", Type: "cash", Currency: "RSD", InitialBalance: decimal.NewFromInt(0),
	})
	if err != nil {
		t.Fatalf("seed account B: %v", err)
	}
	cat, err := repos.Categories.Create(ctx, repository.CreateCategoryInput{
		FamilyID: famID, Name: "Groceries", Type: "expense",
	})
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}

	svc := transaction.New(repos.Transactions, repos.Accounts, repos.Categories)
	return fixture{svc: svc, repos: repos, pool: pool, familyID: famID, userID: userID}, accA.ID, accB.ID, cat.ID
}

func balance(t *testing.T, f fixture, accountID uuid.UUID) decimal.Decimal {
	t.Helper()
	bal, _, err := f.repos.Accounts.GetBalance(context.Background(), accountID)
	if err != nil {
		t.Fatalf("get balance: %v", err)
	}
	return bal
}

func TestCreate_ValidUpdatesBalance(t *testing.T) {
	f, accA, _, cat := setup(t)
	ctx := context.Background()

	_, err := f.svc.Create(ctx, f.familyID, f.userID, transaction.CreateTransactionInput{
		Type: "expense", Amount: decimal.NewFromInt(250), Currency: "RSD",
		CategoryID: cat.String(), AccountID: accA.String(), Date: "2026-03-10",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	// Balance recalculated by trigger: 1000 - 250 = 750.
	if got := balance(t, f, accA); !got.Equal(decimal.NewFromInt(750)) {
		t.Errorf("account A balance = %s, want 750", got)
	}
}

func TestCreate_CategoryTypeMismatch(t *testing.T) {
	f, accA, _, cat := setup(t)
	// cat is expense; transaction type income -> mismatch.
	_, err := f.svc.Create(context.Background(), f.familyID, f.userID, transaction.CreateTransactionInput{
		Type: "income", Amount: decimal.NewFromInt(10), Currency: "RSD",
		CategoryID: cat.String(), AccountID: accA.String(), Date: "2026-03-10",
	})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestCreate_NonExistentAccount(t *testing.T) {
	f, _, _, cat := setup(t)
	_, err := f.svc.Create(context.Background(), f.familyID, f.userID, transaction.CreateTransactionInput{
		Type: "expense", Amount: decimal.NewFromInt(10), Currency: "RSD",
		CategoryID: cat.String(), AccountID: uuid.New().String(), Date: "2026-03-10",
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestCreate_CurrencyMismatch(t *testing.T) {
	f, accA, _, cat := setup(t)
	// account A is RSD; transaction EUR -> mismatch.
	_, err := f.svc.Create(context.Background(), f.familyID, f.userID, transaction.CreateTransactionInput{
		Type: "expense", Amount: decimal.NewFromInt(10), Currency: "EUR",
		CategoryID: cat.String(), AccountID: accA.String(), Date: "2026-03-10",
	})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestUpdate_AmountRecalculatesBalance(t *testing.T) {
	f, accA, _, cat := setup(t)
	ctx := context.Background()
	tx, err := f.svc.Create(ctx, f.familyID, f.userID, transaction.CreateTransactionInput{
		Type: "expense", Amount: decimal.NewFromInt(250), Currency: "RSD",
		CategoryID: cat.String(), AccountID: accA.String(), Date: "2026-03-10",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	newAmount := decimal.NewFromInt(400)
	if _, err := f.svc.Update(ctx, f.familyID, f.userID, tx.ID, transaction.UpdateTransactionInput{Amount: &newAmount}); err != nil {
		t.Fatalf("update: %v", err)
	}
	// 1000 - 400 = 600.
	if got := balance(t, f, accA); !got.Equal(decimal.NewFromInt(600)) {
		t.Errorf("account A balance = %s, want 600", got)
	}
}

func TestUpdate_AccountChangeRecalculatesBothAccounts(t *testing.T) {
	f, accA, accB, cat := setup(t)
	ctx := context.Background()
	tx, err := f.svc.Create(ctx, f.familyID, f.userID, transaction.CreateTransactionInput{
		Type: "expense", Amount: decimal.NewFromInt(250), Currency: "RSD",
		CategoryID: cat.String(), AccountID: accA.String(), Date: "2026-03-10",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// Before move: A = 750, B = 0.
	accBID := accB.String()
	if _, err := f.svc.Update(ctx, f.familyID, f.userID, tx.ID, transaction.UpdateTransactionInput{AccountID: &accBID}); err != nil {
		t.Fatalf("update account change: %v", err)
	}
	// After move: A healed back to 1000, B = 0 - 250 = -250 (migration 014: both recalculated).
	if got := balance(t, f, accA); !got.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("old account A balance = %s, want 1000 (restored)", got)
	}
	if got := balance(t, f, accB); !got.Equal(decimal.NewFromInt(-250)) {
		t.Errorf("new account B balance = %s, want -250", got)
	}
}

func TestDelete_RecalculatesBalance(t *testing.T) {
	f, accA, _, cat := setup(t)
	ctx := context.Background()
	tx, err := f.svc.Create(ctx, f.familyID, f.userID, transaction.CreateTransactionInput{
		Type: "expense", Amount: decimal.NewFromInt(250), Currency: "RSD",
		CategoryID: cat.String(), AccountID: accA.String(), Date: "2026-03-10",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := f.svc.Delete(ctx, f.familyID, f.userID, tx.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	// Soft-deleted tx excluded: A back to 1000.
	if got := balance(t, f, accA); !got.Equal(decimal.NewFromInt(1000)) {
		t.Errorf("account A balance = %s, want 1000", got)
	}
}

func TestUpdate_WrongFamily(t *testing.T) {
	f, accA, _, cat := setup(t)
	ctx := context.Background()
	tx, err := f.svc.Create(ctx, f.familyID, f.userID, transaction.CreateTransactionInput{
		Type: "expense", Amount: decimal.NewFromInt(250), Currency: "RSD",
		CategoryID: cat.String(), AccountID: accA.String(), Date: "2026-03-10",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	newAmount := decimal.NewFromInt(1)
	_, err = f.svc.Update(ctx, uuid.New(), f.userID, tx.ID, transaction.UpdateTransactionInput{Amount: &newAmount})
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("err = %v, want ErrPermissionDenied", err)
	}
}
