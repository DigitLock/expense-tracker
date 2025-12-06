package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

// AccountRepository handles account data operations
type AccountRepository struct {
	queries *sqlc.Queries
}

// NewAccountRepository creates a new AccountRepository
func NewAccountRepository(queries *sqlc.Queries) *AccountRepository {
	return &AccountRepository{queries: queries}
}

// GetByID retrieves an active account by ID
func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Account, error) {
	return r.queries.GetAccount(ctx, id)
}

// GetByIDIncludingInactive retrieves an account by ID (even if inactive)
func (r *AccountRepository) GetByIDIncludingInactive(ctx context.Context, id uuid.UUID) (sqlc.Account, error) {
	return r.queries.GetAccountIncludingInactive(ctx, id)
}

// ListByFamily retrieves all active accounts in a family
func (r *AccountRepository) ListByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Account, error) {
	return r.queries.ListAccountsByFamily(ctx, familyID)
}

// ListAllByFamily retrieves all accounts in a family (including inactive)
func (r *AccountRepository) ListAllByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Account, error) {
	return r.queries.ListAllAccountsByFamily(ctx, familyID)
}

// ListByType retrieves accounts by type within a family
func (r *AccountRepository) ListByType(ctx context.Context, familyID uuid.UUID, accountType string) ([]sqlc.Account, error) {
	return r.queries.ListAccountsByType(ctx, sqlc.ListAccountsByTypeParams{
		FamilyID: familyID,
		Type:     accountType,
	})
}

// CreateAccountInput contains data for creating a new account
type CreateAccountInput struct {
	FamilyID       uuid.UUID
	Name           string
	Type           string // cash, checking, savings
	Currency       string // RSD, EUR
	InitialBalance decimal.Decimal
}

// Create creates a new account
func (r *AccountRepository) Create(ctx context.Context, input CreateAccountInput) (sqlc.Account, error) {
	return r.queries.CreateAccount(ctx, sqlc.CreateAccountParams{
		ID:             uuid.New(),
		FamilyID:       input.FamilyID,
		Name:           input.Name,
		Type:           input.Type,
		Currency:       input.Currency,
		InitialBalance: input.InitialBalance,
	})
}

// UpdateAccountInput contains data for updating an account (partial update)
type UpdateAccountInput struct {
	ID       uuid.UUID
	Name     *string
	IsActive *bool
}

// Update updates account details (partial update)
func (r *AccountRepository) Update(ctx context.Context, input UpdateAccountInput) (sqlc.Account, error) {
	// First get current account (including inactive to allow reactivation)
	current, err := r.queries.GetAccountIncludingInactive(ctx, input.ID)
	if err != nil {
		return sqlc.Account{}, err
	}

	// Apply updates (keep current values if not provided)
	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}

	isActive := current.IsActive
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	// Update with merged values
	return r.queries.UpdateAccount(ctx, sqlc.UpdateAccountParams{
		ID:       input.ID,
		Name:     name,
		IsActive: isActive,
	})
}

// Delete soft-deletes an account
func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteAccount(ctx, id)
}

// GetBalance retrieves current balance and currency
func (r *AccountRepository) GetBalance(ctx context.Context, id uuid.UUID) (decimal.Decimal, string, error) {
	result, err := r.queries.GetAccountBalance(ctx, id)
	if err != nil {
		return decimal.Zero, "", err
	}
	return result.CurrentBalance, result.Currency, nil
}

// GetTotalBalanceByFamily retrieves total balance across all accounts
func (r *AccountRepository) GetTotalBalanceByFamily(ctx context.Context, familyID uuid.UUID) (decimal.Decimal, int64, error) {
	result, err := r.queries.GetTotalBalanceByFamily(ctx, familyID)
	if err != nil {
		return decimal.Zero, 0, err
	}
	return result.TotalBalance, result.AccountCount, nil
}
