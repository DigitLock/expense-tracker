package account_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/account"
)

// BE-001: only ACTIVE transactions block account deletion; soft-deleted
// transactions are history and do not.

type txSeeder struct {
	repos    *repository.Repositories
	familyID uuid.UUID
	userID   uuid.UUID
	catID    uuid.UUID
}

// newTxSeeder seeds a user and an expense category in famID so transactions
// can be created against accounts in that family.
func newTxSeeder(t *testing.T, pool *pgxpool.Pool, famID uuid.UUID) txSeeder {
	t.Helper()
	ctx := context.Background()
	userID := uuid.New()
	if _, err := pool.Exec(ctx,
		"INSERT INTO users (id, family_id, email, password_hash, name, role) VALUES ($1,$2,$3,$4,$5,$6)",
		userID, famID, "u@example.com", "x", "U", "owner"); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	repos := repository.New(pool)
	cat, err := repos.Categories.Create(ctx, repository.CreateCategoryInput{FamilyID: famID, Name: "Food", Type: "expense"})
	if err != nil {
		t.Fatalf("seed category: %v", err)
	}
	return txSeeder{repos: repos, familyID: famID, userID: userID, catID: cat.ID}
}

func (s txSeeder) create(t *testing.T, accID uuid.UUID) uuid.UUID {
	t.Helper()
	tx, err := s.repos.Transactions.Create(context.Background(), repository.CreateTransactionInput{
		FamilyID: s.familyID, AccountID: accID, CategoryID: s.catID, Type: "expense",
		Amount: decimal.NewFromInt(10), Currency: "RSD",
		TransactionDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), CreatedBy: s.userID,
	})
	if err != nil {
		t.Fatalf("seed transaction: %v", err)
	}
	return tx.ID
}

func (s txSeeder) softDelete(t *testing.T, txID uuid.UUID) {
	t.Helper()
	if err := s.repos.Transactions.Delete(context.Background(), txID, s.userID); err != nil {
		t.Fatalf("soft-delete transaction: %v", err)
	}
}

func TestDelete_AllTransactionsSoftDeleted(t *testing.T) {
	svc, famID, pool := setup(t)
	ctx := context.Background()
	seed := newTxSeeder(t, pool, famID)
	acc, err := svc.Create(ctx, famID, account.CreateAccountInput{Name: "Cash", Type: "cash", Currency: "RSD"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	seed.softDelete(t, seed.create(t, acc.ID))
	seed.softDelete(t, seed.create(t, acc.ID))

	if err := svc.Delete(ctx, famID, acc.ID); err != nil {
		t.Fatalf("delete with only soft-deleted tx: err = %v, want nil", err)
	}
	got, err := seed.repos.Accounts.GetByIDIncludingInactive(ctx, acc.ID)
	if err != nil {
		t.Fatalf("get account: %v", err)
	}
	if got.IsActive {
		t.Error("account should be inactive after delete")
	}
}

func TestDelete_ActiveAndSoftDeletedTransactions(t *testing.T) {
	svc, famID, pool := setup(t)
	ctx := context.Background()
	seed := newTxSeeder(t, pool, famID)
	acc, err := svc.Create(ctx, famID, account.CreateAccountInput{Name: "Cash", Type: "cash", Currency: "RSD"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	seed.create(t, acc.ID)
	seed.softDelete(t, seed.create(t, acc.ID))

	err = svc.Delete(ctx, famID, acc.ID)
	if !errors.Is(err, domain.ErrFailedPrecondition) {
		t.Fatalf("delete with an active tx: err = %v, want ErrFailedPrecondition", err)
	}
	if err.Error() != "cannot delete account with existing transactions" {
		t.Errorf("error text = %q, want existing text", err.Error())
	}
}
