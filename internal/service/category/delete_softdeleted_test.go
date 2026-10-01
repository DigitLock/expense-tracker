package category_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/category"
)

// BE-001: only ACTIVE transactions block category deletion; soft-deleted
// transactions are history and do not.

// seedTx creates a transaction for catID on a fresh account and returns its ID.
func seedTx(t *testing.T, f fixture, catID uuid.UUID) uuid.UUID {
	t.Helper()
	ctx := context.Background()
	acc, err := f.repos.Accounts.Create(ctx, repository.CreateAccountInput{
		FamilyID: f.familyID, Name: "Acc " + uuid.NewString(), Type: "cash", Currency: "RSD", InitialBalance: decimal.NewFromInt(100),
	})
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}
	tx, err := f.repos.Transactions.Create(ctx, repository.CreateTransactionInput{
		FamilyID: f.familyID, AccountID: acc.ID, CategoryID: catID, Type: "expense",
		Amount: decimal.NewFromInt(10), Currency: "RSD", TransactionDate: mustDate(), CreatedBy: f.userID,
	})
	if err != nil {
		t.Fatalf("seed transaction: %v", err)
	}
	return tx.ID
}

func softDeleteTx(t *testing.T, f fixture, txID uuid.UUID) {
	t.Helper()
	if err := f.repos.Transactions.Delete(context.Background(), txID, f.userID); err != nil {
		t.Fatalf("soft-delete transaction: %v", err)
	}
}

func TestDelete_AllTransactionsSoftDeleted(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	cat, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	softDeleteTx(t, f, seedTx(t, f, cat.ID))
	softDeleteTx(t, f, seedTx(t, f, cat.ID))

	if err := f.svc.Delete(ctx, f.familyID, cat.ID); err != nil {
		t.Fatalf("delete with only soft-deleted tx: err = %v, want nil", err)
	}
	got, err := f.repos.Categories.GetByIDIncludingInactive(ctx, cat.ID)
	if err != nil {
		t.Fatalf("get category: %v", err)
	}
	if got.IsActive {
		t.Error("category should be inactive after delete")
	}
}

func TestDelete_ActiveAndSoftDeletedTransactions(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	cat, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	seedTx(t, f, cat.ID)
	softDeleteTx(t, f, seedTx(t, f, cat.ID))

	err = f.svc.Delete(ctx, f.familyID, cat.ID)
	if !errors.Is(err, domain.ErrFailedPrecondition) {
		t.Fatalf("delete with an active tx: err = %v, want ErrFailedPrecondition", err)
	}
	if err.Error() != "cannot delete category with existing transactions" {
		t.Errorf("error text = %q, want existing text", err.Error())
	}
}

func TestDelete_SoftDeletedTransactions_ThenRestore(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	cat, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	txID := seedTx(t, f, cat.ID)
	softDeleteTx(t, f, txID)

	if err := f.svc.Delete(ctx, f.familyID, cat.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	restored, err := f.repos.Categories.Restore(ctx, cat.ID)
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !restored.IsActive {
		t.Error("restored category should be active")
	}

	// The old transaction stays soft-deleted.
	tx, err := f.repos.Transactions.GetByIDIncludingInactive(ctx, txID)
	if err != nil {
		t.Fatalf("get transaction: %v", err)
	}
	if tx.IsActive {
		t.Error("old transaction should stay inactive after category restore")
	}
	if tx.CategoryID != cat.ID {
		t.Errorf("transaction category = %s, want %s", tx.CategoryID, cat.ID)
	}
}
