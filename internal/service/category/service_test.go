package category_test

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
	"github.com/DigitLock/expense-tracker/internal/service/category"
	"github.com/DigitLock/expense-tracker/internal/testutil"
)

type fixture struct {
	svc      *category.Service
	repos    *repository.Repositories
	pool     *pgxpool.Pool
	familyID uuid.UUID
	userID   uuid.UUID
}

func setup(t *testing.T) fixture {
	t.Helper()
	pool := testutil.Pool(t)
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
	return fixture{svc: category.New(repos.Categories), repos: repos, pool: pool, familyID: famID, userID: userID}
}

func strptr(s string) *string { return &s }

func mustDate() time.Time { return time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC) }

func TestCreate_WithValidParent(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	parent, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	if err != nil {
		t.Fatalf("create parent: %v", err)
	}
	pid := parent.ID.String()
	child, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Groceries", Type: "expense", ParentID: &pid})
	if err != nil {
		t.Fatalf("create child: %v", err)
	}
	if child.ParentID == nil || *child.ParentID != parent.ID {
		t.Errorf("child parent = %v, want %v", child.ParentID, parent.ID)
	}
}

func TestCreate_ParentTypeMismatch(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	parent, _ := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Salary", Type: "income"})
	pid := parent.ID.String()
	_, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "X", Type: "expense", ParentID: &pid})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestCreate_ThirdLevelRejected(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	parent, _ := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	ppid := parent.ID.String()
	child, _ := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Groceries", Type: "expense", ParentID: &ppid})
	cpid := child.ID.String()
	_, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Fruit", Type: "expense", ParentID: &cpid})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("3rd-level err = %v, want ErrInvalidArgument", err)
	}
}

func TestCreate_DuplicateActiveName(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	in := category.CreateCategoryInput{Name: "Food", Type: "expense"}
	if _, err := f.svc.Create(ctx, f.familyID, in); err != nil {
		t.Fatalf("first create: %v", err)
	}
	_, err := f.svc.Create(ctx, f.familyID, in)
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("dup err = %v, want ErrAlreadyExists", err)
	}
}

func TestUpdate_CyclicParent(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c, _ := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	self := c.ID.String()
	_, err := f.svc.Update(ctx, f.familyID, c.ID, category.UpdateCategoryInput{ParentID: &self})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("self-parent err = %v, want ErrInvalidArgument", err)
	}
}

func TestDelete_WithSubcategories(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	parent, _ := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	ppid := parent.ID.String()
	if _, err := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Groceries", Type: "expense", ParentID: &ppid}); err != nil {
		t.Fatalf("create child: %v", err)
	}
	err := f.svc.Delete(ctx, f.familyID, parent.ID)
	if !errors.Is(err, domain.ErrFailedPrecondition) {
		t.Errorf("delete-with-children err = %v, want ErrFailedPrecondition", err)
	}
}

func TestDelete_WithTransactions(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	cat, _ := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	acc, err := f.repos.Accounts.Create(ctx, repository.CreateAccountInput{
		FamilyID: f.familyID, Name: "A", Type: "cash", Currency: "RSD", InitialBalance: decimal.NewFromInt(100),
	})
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}
	if _, err := f.repos.Transactions.Create(ctx, repository.CreateTransactionInput{
		FamilyID: f.familyID, AccountID: acc.ID, CategoryID: cat.ID, Type: "expense",
		Amount: decimal.NewFromInt(10), Currency: "RSD", TransactionDate: mustDate(), CreatedBy: f.userID,
	}); err != nil {
		t.Fatalf("seed transaction: %v", err)
	}
	err = f.svc.Delete(ctx, f.familyID, cat.ID)
	if !errors.Is(err, domain.ErrFailedPrecondition) {
		t.Errorf("delete-with-tx err = %v, want ErrFailedPrecondition", err)
	}
}

func TestList_TypeAndIncludeInactive(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	inc, _ := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Salary", Type: "income"})
	exp, _ := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})

	incomeOnly, err := f.svc.List(ctx, f.familyID, "income", false)
	if err != nil {
		t.Fatalf("list income: %v", err)
	}
	if len(incomeOnly) != 1 || incomeOnly[0].Type != "income" {
		t.Errorf("income filter = %+v, want 1 income", incomeOnly)
	}

	// soft-delete the expense category, then list with include_inactive.
	if err := f.svc.Delete(ctx, f.familyID, exp.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	active, _ := f.svc.List(ctx, f.familyID, "", false)
	all, _ := f.svc.List(ctx, f.familyID, "", true)
	if len(all) <= len(active) {
		t.Errorf("include_inactive=%d should exceed active=%d", len(all), len(active))
	}
	_ = inc
}

func TestUpdate_WrongFamily(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	c, _ := f.svc.Create(ctx, f.familyID, category.CreateCategoryInput{Name: "Food", Type: "expense"})
	_, err := f.svc.Update(ctx, uuid.New(), c.ID, category.UpdateCategoryInput{Name: strptr("Hacked")})
	if !errors.Is(err, domain.ErrPermissionDenied) {
		t.Errorf("err = %v, want ErrPermissionDenied", err)
	}
}
