// Package transaction holds the application-layer service for the Transaction
// domain. All transaction business rules live here; REST and gRPC are thin
// adapters. Account balances are maintained solely by the DB trigger — the
// service never computes or writes current_balance.
package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

const dateLayout = "2006-01-02"

// Repo is the narrow set of transaction repository methods the service needs.
type Repo interface {
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Transaction, error)
	ListFiltered(ctx context.Context, filter repository.TransactionFilter) ([]sqlc.Transaction, int64, error)
	Create(ctx context.Context, input repository.CreateTransactionInput) (sqlc.Transaction, error)
	UpdateFull(ctx context.Context, input repository.UpdateTransactionFullInput) (sqlc.Transaction, error)
	Delete(ctx context.Context, id, deletedBy uuid.UUID) error
}

// AccountReader reads accounts (for currency/family checks).
type AccountReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Account, error)
}

// CategoryReader reads categories (for type/family checks).
type CategoryReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Category, error)
}

// Service implements Transaction use cases.
type Service struct {
	repo       Repo
	accounts   AccountReader
	categories CategoryReader
}

func New(repo Repo, accounts AccountReader, categories CategoryReader) *Service {
	return &Service{repo: repo, accounts: accounts, categories: categories}
}

// Input DTOs (transport-agnostic).

type CreateTransactionInput struct {
	Type        string
	Amount      decimal.Decimal
	Currency    string
	CategoryID  string
	AccountID   string
	Description string
	Date        string // YYYY-MM-DD
}

type UpdateTransactionInput struct {
	Type        *string
	Amount      *decimal.Decimal
	Currency    *string
	CategoryID  *string
	AccountID   *string
	Description *string
	Date        *string
}

type ListFilter struct {
	Type      *string
	AccountID *uuid.UUID
	Month     string // YYYY-MM or ""
	Page      int
	PerPage   int
}

type ListResult struct {
	Transactions []domain.Transaction
	Page         int
	PerPage      int
	Total        int
	TotalPages   int
}

func validType(t string) bool { return t == "income" || t == "expense" }

// Create validates cross-entity invariants and creates a transaction. The
// balance trigger updates the account balance.
func (s *Service) Create(ctx context.Context, familyID, userID uuid.UUID, in CreateTransactionInput) (domain.Transaction, error) {
	if !validType(in.Type) {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "invalid transaction type")
	}
	if !in.Amount.IsPositive() {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "amount must be positive")
	}
	date, err := time.Parse(dateLayout, in.Date)
	if err != nil {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "invalid date format, expected YYYY-MM-DD")
	}
	accountID, err := uuid.Parse(in.AccountID)
	if err != nil {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "invalid account_id")
	}
	categoryID, err := uuid.Parse(in.CategoryID)
	if err != nil {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "invalid category_id")
	}

	account, err := s.requireAccount(ctx, familyID, accountID)
	if err != nil {
		return domain.Transaction{}, err
	}
	category, err := s.requireCategory(ctx, familyID, categoryID)
	if err != nil {
		return domain.Transaction{}, err
	}
	if in.Currency != account.Currency {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "transaction currency must match account currency (%s)", account.Currency)
	}
	if category.Type != in.Type {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "category type does not match transaction type")
	}

	row, err := s.repo.Create(ctx, repository.CreateTransactionInput{
		FamilyID:        familyID,
		AccountID:       accountID,
		CategoryID:      categoryID,
		Type:            in.Type,
		Amount:          in.Amount,
		Currency:        in.Currency,
		Description:     in.Description,
		TransactionDate: date,
		CreatedBy:       userID,
	})
	if err != nil {
		return domain.Transaction{}, err
	}
	return toDomain(row), nil
}

// Update applies a partial update and re-checks all invariants against the
// resulting (merged) values. The balance trigger recalculates affected accounts
// (both old and new when account_id changes, per migration 014).
func (s *Service) Update(ctx context.Context, familyID, userID, id uuid.UUID, in UpdateTransactionInput) (domain.Transaction, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Transaction{}, domain.Errorf(domain.ErrNotFound, "transaction not found")
	}
	if existing.FamilyID != familyID {
		return domain.Transaction{}, domain.Errorf(domain.ErrPermissionDenied, "access denied")
	}

	// Merge: start from existing, apply provided fields.
	finalType := existing.Type
	if in.Type != nil {
		finalType = *in.Type
	}
	finalAmount := existing.Amount
	if in.Amount != nil {
		finalAmount = *in.Amount
	}
	finalCurrency := existing.Currency
	if in.Currency != nil {
		finalCurrency = *in.Currency
	}
	finalAccountID := existing.AccountID
	if in.AccountID != nil {
		parsed, err := uuid.Parse(*in.AccountID)
		if err != nil {
			return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "invalid account_id")
		}
		finalAccountID = parsed
	}
	finalCategoryID := existing.CategoryID
	if in.CategoryID != nil {
		parsed, err := uuid.Parse(*in.CategoryID)
		if err != nil {
			return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "invalid category_id")
		}
		finalCategoryID = parsed
	}
	finalDescription := ""
	if existing.Description.Valid {
		finalDescription = existing.Description.String
	}
	if in.Description != nil {
		finalDescription = *in.Description
	}
	finalDate := existing.TransactionDate.Time
	if in.Date != nil {
		parsed, err := time.Parse(dateLayout, *in.Date)
		if err != nil {
			return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "invalid date format, expected YYYY-MM-DD")
		}
		finalDate = parsed
	}

	// Re-check invariants against final values.
	if !validType(finalType) {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "invalid transaction type")
	}
	if !finalAmount.IsPositive() {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "amount must be positive")
	}
	account, err := s.requireAccount(ctx, familyID, finalAccountID)
	if err != nil {
		return domain.Transaction{}, err
	}
	category, err := s.requireCategory(ctx, familyID, finalCategoryID)
	if err != nil {
		return domain.Transaction{}, err
	}
	if finalCurrency != account.Currency {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "transaction currency must match account currency (%s)", account.Currency)
	}
	if category.Type != finalType {
		return domain.Transaction{}, domain.Errorf(domain.ErrInvalidArgument, "category type does not match transaction type")
	}

	row, err := s.repo.UpdateFull(ctx, repository.UpdateTransactionFullInput{
		ID:              id,
		AccountID:       finalAccountID,
		CategoryID:      finalCategoryID,
		Type:            finalType,
		Amount:          finalAmount,
		Currency:        finalCurrency,
		Description:     finalDescription,
		TransactionDate: finalDate,
		UpdatedBy:       userID,
	})
	if err != nil {
		return domain.Transaction{}, err
	}
	return toDomain(row), nil
}

// Delete soft-deletes a transaction; the balance trigger recalculates the account.
func (s *Service) Delete(ctx context.Context, familyID, userID, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Errorf(domain.ErrNotFound, "transaction not found")
	}
	if existing.FamilyID != familyID {
		return domain.Errorf(domain.ErrPermissionDenied, "access denied")
	}
	return s.repo.Delete(ctx, id, userID)
}

// List returns filtered, paginated transactions for the family.
func (s *Service) List(ctx context.Context, familyID uuid.UUID, f ListFilter) (ListResult, error) {
	page := f.Page
	if page < 1 {
		page = 1
	}
	perPage := f.PerPage
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}

	filter := repository.TransactionFilter{
		FamilyID:  familyID,
		Type:      f.Type,
		AccountID: f.AccountID,
		Limit:     int32(perPage),
		Offset:    int32((page - 1) * perPage),
	}
	if f.Month != "" {
		if start, err := time.Parse("2006-01", f.Month); err == nil {
			end := start.AddDate(0, 1, -1)
			filter.StartDate = &start
			filter.EndDate = &end
		}
	}

	rows, total, err := s.repo.ListFiltered(ctx, filter)
	if err != nil {
		return ListResult{}, err
	}

	txs := make([]domain.Transaction, len(rows))
	for i, r := range rows {
		txs[i] = toDomain(r)
	}
	return ListResult{
		Transactions: txs,
		Page:         page,
		PerPage:      perPage,
		Total:        int(total),
		TotalPages:   (int(total) + perPage - 1) / perPage,
	}, nil
}

// --- helpers ---

func (s *Service) requireAccount(ctx context.Context, familyID, id uuid.UUID) (sqlc.Account, error) {
	acc, err := s.accounts.GetByID(ctx, id)
	if err != nil {
		return sqlc.Account{}, domain.Errorf(domain.ErrNotFound, "account not found")
	}
	if acc.FamilyID != familyID {
		return sqlc.Account{}, domain.Errorf(domain.ErrPermissionDenied, "access denied")
	}
	return acc, nil
}

func (s *Service) requireCategory(ctx context.Context, familyID, id uuid.UUID) (sqlc.Category, error) {
	cat, err := s.categories.GetByID(ctx, id)
	if err != nil {
		return sqlc.Category{}, domain.Errorf(domain.ErrNotFound, "category not found")
	}
	if cat.FamilyID != familyID {
		return sqlc.Category{}, domain.Errorf(domain.ErrPermissionDenied, "access denied")
	}
	return cat, nil
}

func toDomain(t sqlc.Transaction) domain.Transaction {
	desc := ""
	if t.Description.Valid {
		desc = t.Description.String
	}
	return domain.Transaction{
		ID:           t.ID,
		Type:         t.Type,
		Amount:       t.Amount,
		Currency:     t.Currency,
		AmountBase:   t.AmountBase,
		BaseCurrency: "RSD",
		AccountID:    t.AccountID,
		CategoryID:   t.CategoryID,
		Description:  desc,
		Date:         t.TransactionDate.Time,
		CreatedBy:    t.CreatedBy,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}
