package report_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/report"
	"github.com/DigitLock/expense-tracker/internal/testutil"
)

// seed builds a family with one account, an expense + income category, and
// three transactions in 2026-03: expenses 100 & 200 (Food), income 1000 (Salary).
func seed(t *testing.T) (*report.Service, uuid.UUID) {
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
	acc, err := repos.Accounts.Create(ctx, repository.CreateAccountInput{
		FamilyID: famID, Name: "A", Type: "cash", Currency: "RSD", InitialBalance: decimal.NewFromInt(10000),
	})
	if err != nil {
		t.Fatalf("seed account: %v", err)
	}
	food, _ := repos.Categories.Create(ctx, repository.CreateCategoryInput{FamilyID: famID, Name: "Food", Type: "expense"})
	salary, _ := repos.Categories.Create(ctx, repository.CreateCategoryInput{FamilyID: famID, Name: "Salary", Type: "income"})

	d := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	mk := func(cat uuid.UUID, typ string, amt int64) {
		if _, err := repos.Transactions.Create(ctx, repository.CreateTransactionInput{
			FamilyID: famID, AccountID: acc.ID, CategoryID: cat, Type: typ,
			Amount: decimal.NewFromInt(amt), Currency: "RSD", TransactionDate: d, CreatedBy: userID,
		}); err != nil {
			t.Fatalf("seed tx: %v", err)
		}
	}
	mk(food.ID, "expense", 100)
	mk(food.ID, "expense", 200)
	mk(salary.ID, "income", 1000)

	return report.New(repos.Transactions, repos.Categories, repos.ExchangeRates), famID
}

func TestSpendingByCategory_CustomPeriod(t *testing.T) {
	svc, famID := seed(t)
	rep, err := svc.SpendingByCategory(context.Background(), famID, report.SpendingParams{
		StartDate: "2026-03-01", EndDate: "2026-03-31",
	})
	if err != nil {
		t.Fatalf("SpendingByCategory: %v", err)
	}
	if rep.TransactionType != "expense" || rep.Currency != "RSD" {
		t.Errorf("type=%s currency=%s, want expense/RSD", rep.TransactionType, rep.Currency)
	}
	if len(rep.Categories) != 1 || rep.Categories[0].CategoryName != "Food" {
		t.Fatalf("categories = %+v, want 1 Food", rep.Categories)
	}
	if !rep.TotalAmount.Equal(decimal.NewFromInt(300)) || rep.TotalTransactions != 2 {
		t.Errorf("total=%s txns=%d, want 300/2", rep.TotalAmount, rep.TotalTransactions)
	}
	c := rep.Categories[0]
	if !c.TotalAmount.Equal(decimal.NewFromInt(300)) || c.TransactionCount != 2 || !c.Percentage.Equal(decimal.NewFromInt(100)) {
		t.Errorf("category = %+v, want total 300 count 2 pct 100", c)
	}
}

func TestSpendingByCategory_InvalidDate(t *testing.T) {
	svc, famID := seed(t)
	_, err := svc.SpendingByCategory(context.Background(), famID, report.SpendingParams{StartDate: "03-2026"})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestSpendingByCategory_StartAfterEnd(t *testing.T) {
	svc, famID := seed(t)
	_, err := svc.SpendingByCategory(context.Background(), famID, report.SpendingParams{
		StartDate: "2026-03-31", EndDate: "2026-03-01",
	})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("err = %v, want ErrInvalidArgument", err)
	}
}

func TestMonthlySummary_Values(t *testing.T) {
	svc, famID := seed(t)
	sum, err := svc.MonthlySummary(context.Background(), famID, "2026-03", "")
	if err != nil {
		t.Fatalf("MonthlySummary: %v", err)
	}
	if !sum.Income.Equal(decimal.NewFromInt(1000)) || !sum.Expenses.Equal(decimal.NewFromInt(300)) {
		t.Errorf("income=%s expenses=%s, want 1000/300", sum.Income, sum.Expenses)
	}
	if !sum.NetBalance.Equal(decimal.NewFromInt(700)) {
		t.Errorf("net=%s, want 700", sum.NetBalance)
	}
	if sum.IncomeCount != 1 || sum.ExpenseCount != 2 {
		t.Errorf("counts income=%d expense=%d, want 1/2", sum.IncomeCount, sum.ExpenseCount)
	}
}

func TestMonthlySummary_EmptyMonthZeros(t *testing.T) {
	svc, famID := seed(t)
	sum, err := svc.MonthlySummary(context.Background(), famID, "2020-01", "")
	if err != nil {
		t.Fatalf("MonthlySummary: %v", err)
	}
	if !sum.Income.IsZero() || !sum.Expenses.IsZero() || !sum.NetBalance.IsZero() || sum.IncomeCount != 0 || sum.ExpenseCount != 0 {
		t.Errorf("empty month not zero: %+v", sum)
	}
}

func TestMonthlySummary_InvalidMonth(t *testing.T) {
	svc, famID := seed(t)
	_, err := svc.MonthlySummary(context.Background(), famID, "2026/03", "")
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Errorf("err = %v, want ErrInvalidArgument", err)
	}
}
