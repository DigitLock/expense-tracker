package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

// TransactionRepository handles transaction data operations
type TransactionRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

// NewTransactionRepository creates a new TransactionRepository
func NewTransactionRepository(queries *sqlc.Queries, pool *pgxpool.Pool) *TransactionRepository {
	return &TransactionRepository{
		queries: queries,
		pool:    pool,
	}
}

// GetByID retrieves a transaction by ID
func (r *TransactionRepository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Transaction, error) {
	return r.queries.GetTransaction(ctx, id)
}

// ListByFamily retrieves all transactions in a family
func (r *TransactionRepository) ListByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Transaction, error) {
	return r.queries.ListTransactionsByFamily(ctx, familyID)
}

// ListByAccount retrieves transactions for an account
func (r *TransactionRepository) ListByAccount(ctx context.Context, accountID uuid.UUID) ([]sqlc.Transaction, error) {
	return r.queries.ListTransactionsByAccount(ctx, accountID)
}

// ListByCategory retrieves transactions for a category
func (r *TransactionRepository) ListByCategory(ctx context.Context, categoryID uuid.UUID) ([]sqlc.Transaction, error) {
	return r.queries.ListTransactionsByCategory(ctx, categoryID)
}

// ListByDateRange retrieves transactions within a date range
func (r *TransactionRepository) ListByDateRange(ctx context.Context, familyID uuid.UUID, startDate, endDate time.Time) ([]sqlc.Transaction, error) {
	return r.queries.ListTransactionsByDateRange(ctx, sqlc.ListTransactionsByDateRangeParams{
		FamilyID:          familyID,
		TransactionDate:   pgtype.Date{Time: startDate, Valid: true},
		TransactionDate_2: pgtype.Date{Time: endDate, Valid: true},
	})
}

// ListPaginated retrieves transactions with pagination
func (r *TransactionRepository) ListPaginated(ctx context.Context, familyID uuid.UUID, limit, offset int32) ([]sqlc.Transaction, error) {
	return r.queries.ListTransactionsPaginated(ctx, sqlc.ListTransactionsPaginatedParams{
		FamilyID: familyID,
		Limit:    limit,
		Offset:   offset,
	})
}

// CreateTransactionInput contains data for creating a new transaction
type CreateTransactionInput struct {
	FamilyID        uuid.UUID
	AccountID       uuid.UUID
	CategoryID      uuid.UUID
	Type            string // income, expense
	Amount          decimal.Decimal
	Currency        string // RSD, EUR
	Description     string
	TransactionDate time.Time
	CreatedBy       uuid.UUID
}

// Create creates a new transaction with automatic amount_base calculation
func (r *TransactionRepository) Create(ctx context.Context, input CreateTransactionInput) (sqlc.Transaction, error) {
	// Start a transaction to set session variable for audit
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set user ID for audit trail
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = '%s'", input.CreatedBy.String()))
	if err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to set audit user: %w", err)
	}

	// Calculate amount_base (convert to base currency if needed)
	amountBase := input.Amount
	if input.Currency != "RSD" {
		// Get exchange rate
		rate, err := r.getExchangeRate(ctx, tx, input.Currency, "RSD", input.TransactionDate)
		if err != nil {
			return sqlc.Transaction{}, fmt.Errorf("failed to get exchange rate: %w", err)
		}
		amountBase = input.Amount.Mul(rate)
	}

	// Create transaction using queries with tx
	qtx := sqlc.New(tx)

	var description pgtype.Text
	if input.Description != "" {
		description = pgtype.Text{String: input.Description, Valid: true}
	}

	result, err := qtx.CreateTransaction(ctx, sqlc.CreateTransactionParams{
		ID:              uuid.New(),
		FamilyID:        input.FamilyID,
		AccountID:       input.AccountID,
		CategoryID:      input.CategoryID,
		Type:            input.Type,
		Amount:          input.Amount,
		Currency:        input.Currency,
		AmountBase:      amountBase,
		Description:     description,
		TransactionDate: pgtype.Date{Time: input.TransactionDate, Valid: true},
		CreatedBy:       input.CreatedBy,
	})
	if err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to create transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to commit: %w", err)
	}

	return result, nil
}

// getExchangeRate retrieves exchange rate for a given date
func (r *TransactionRepository) getExchangeRate(ctx context.Context, tx pgx.Tx, fromCurrency, toCurrency string, date time.Time) (decimal.Decimal, error) {
	qtx := sqlc.New(tx)

	rate, err := qtx.GetLatestExchangeRate(ctx, sqlc.GetLatestExchangeRateParams{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Date:         pgtype.Date{Time: date, Valid: true},
	})
	if err != nil {
		return decimal.Zero, err
	}

	return rate.Rate, nil
}

// UpdateTransactionInput contains data for updating a transaction
type UpdateTransactionInput struct {
	ID              uuid.UUID
	CategoryID      uuid.UUID
	Amount          decimal.Decimal
	Currency        string
	Description     string
	TransactionDate time.Time
	UpdatedBy       uuid.UUID
}

// Update updates a transaction
func (r *TransactionRepository) Update(ctx context.Context, input UpdateTransactionInput) (sqlc.Transaction, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set user ID for audit trail
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = '%s'", input.UpdatedBy.String()))
	if err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to set audit user: %w", err)
	}

	// Calculate amount_base
	amountBase := input.Amount
	if input.Currency != "RSD" {
		rate, err := r.getExchangeRate(ctx, tx, input.Currency, "RSD", input.TransactionDate)
		if err != nil {
			return sqlc.Transaction{}, fmt.Errorf("failed to get exchange rate: %w", err)
		}
		amountBase = input.Amount.Mul(rate)
	}

	qtx := sqlc.New(tx)

	var description pgtype.Text
	if input.Description != "" {
		description = pgtype.Text{String: input.Description, Valid: true}
	}

	result, err := qtx.UpdateTransaction(ctx, sqlc.UpdateTransactionParams{
		ID:              input.ID,
		CategoryID:      input.CategoryID,
		Amount:          input.Amount,
		Currency:        input.Currency,
		AmountBase:      amountBase,
		Description:     description,
		TransactionDate: pgtype.Date{Time: input.TransactionDate, Valid: true},
	})
	if err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to update transaction: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to commit: %w", err)
	}

	return result, nil
}

// Delete soft-deletes a transaction
func (r *TransactionRepository) Delete(ctx context.Context, id uuid.UUID, deletedBy uuid.UUID) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Set user ID for audit trail
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = '%s'", deletedBy.String()))
	if err != nil {
		return fmt.Errorf("failed to set audit user: %w", err)
	}

	qtx := sqlc.New(tx)
	if err := qtx.DeleteTransaction(ctx, id); err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}

	return tx.Commit(ctx)
}

// GetSummaryByType retrieves transaction summary grouped by type
func (r *TransactionRepository) GetSummaryByType(ctx context.Context, familyID uuid.UUID, startDate, endDate time.Time) ([]sqlc.GetTransactionsSummaryByTypeRow, error) {
	return r.queries.GetTransactionsSummaryByType(ctx, sqlc.GetTransactionsSummaryByTypeParams{
		FamilyID:          familyID,
		TransactionDate:   pgtype.Date{Time: startDate, Valid: true},
		TransactionDate_2: pgtype.Date{Time: endDate, Valid: true},
	})
}

// GetSummaryByCategory retrieves transaction summary grouped by category
func (r *TransactionRepository) GetSummaryByCategory(ctx context.Context, familyID uuid.UUID, transactionType string, startDate, endDate time.Time) ([]sqlc.GetTransactionsSummaryByCategoryRow, error) {
	return r.queries.GetTransactionsSummaryByCategory(ctx, sqlc.GetTransactionsSummaryByCategoryParams{
		FamilyID:          familyID,
		Type:              transactionType,
		TransactionDate:   pgtype.Date{Time: startDate, Valid: true},
		TransactionDate_2: pgtype.Date{Time: endDate, Valid: true},
	})
}

// TransactionFilter contains filter options for listing transactions
type TransactionFilter struct {
	FamilyID  uuid.UUID
	Type      *string // income, expense, or nil for all
	AccountID *uuid.UUID
	StartDate *time.Time
	EndDate   *time.Time
	Limit     int32
	Offset    int32
}

// ListFiltered retrieves transactions with filters and pagination
func (r *TransactionRepository) ListFiltered(ctx context.Context, filter TransactionFilter) ([]sqlc.Transaction, int64, error) {
	// Build params - empty string/zero UUID means "no filter" in SQL
	typeFilter := ""
	if filter.Type != nil {
		typeFilter = *filter.Type
	}

	var accountFilter uuid.UUID // zero UUID
	if filter.AccountID != nil {
		accountFilter = *filter.AccountID
	}

	var startDate pgtype.Date
	if filter.StartDate != nil {
		startDate = pgtype.Date{Time: *filter.StartDate, Valid: true}
	}

	var endDate pgtype.Date
	if filter.EndDate != nil {
		endDate = pgtype.Date{Time: *filter.EndDate, Valid: true}
	}

	// Get transactions
	transactions, err := r.queries.ListTransactionsFiltered(ctx, sqlc.ListTransactionsFilteredParams{
		FamilyID: filter.FamilyID,
		Column2:  typeFilter,
		Column3:  accountFilter,
		Column4:  startDate,
		Column5:  endDate,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	})
	if err != nil {
		return nil, 0, err
	}

	// Get total count
	total, err := r.queries.CountTransactionsFiltered(ctx, sqlc.CountTransactionsFilteredParams{
		FamilyID: filter.FamilyID,
		Column2:  typeFilter,
		Column3:  accountFilter,
		Column4:  startDate,
		Column5:  endDate,
	})
	if err != nil {
		return nil, 0, err
	}

	return transactions, total, nil
}

// GetByIDIncludingInactive retrieves a transaction by ID (even if inactive)
func (r *TransactionRepository) GetByIDIncludingInactive(ctx context.Context, id uuid.UUID) (sqlc.Transaction, error) {
	return r.queries.GetTransactionIncludingInactive(ctx, id)
}
