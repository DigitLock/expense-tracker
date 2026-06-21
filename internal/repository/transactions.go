package repository

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

// deferRollback is a helper to safely rollback a transaction in defer
// It logs errors except for "transaction already committed" which is expected after Commit()
func deferRollback(ctx context.Context, tx pgx.Tx) {
	if err := tx.Rollback(ctx); err != nil {
		log.Printf("failed to rollback transaction: %v", err)
	}
}

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
	defer deferRollback(ctx, tx)

	// Set user ID for audit trail
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = '%s'", input.CreatedBy.String()))
	if err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to set audit user: %w", err)
	}

	amountBase := r.computeAmountBase(ctx, tx, input.Amount, input.Currency, input.TransactionDate)

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

// computeAmountBase converts amount to base currency (RSD) for cross-currency
// reporting. On a missing exchange rate it falls back to the amount 1:1 with a
// warning. amount_base does NOT affect account balance (native currency).
func (r *TransactionRepository) computeAmountBase(ctx context.Context, tx pgx.Tx, amount decimal.Decimal, currency string, date time.Time) decimal.Decimal {
	if currency == "RSD" {
		return amount
	}
	rate, err := r.getExchangeRate(ctx, tx, currency, "RSD", date)
	if err != nil {
		log.Printf("WARN: exchange rate %s→RSD unavailable for %s, using amount as amount_base fallback",
			currency, date.Format("2006-01-02"))
		return amount
	}
	return amount.Mul(rate)
}

// UpdateTransactionFullInput contains data for a full transaction update,
// including account_id and type (so a transaction can be moved between accounts
// or change type). amount_base is recomputed from amount/currency.
type UpdateTransactionFullInput struct {
	ID              uuid.UUID
	AccountID       uuid.UUID
	CategoryID      uuid.UUID
	Type            string
	Amount          decimal.Decimal
	Currency        string
	Description     string
	TransactionDate time.Time
	UpdatedBy       uuid.UUID
}

// UpdateFull updates all mutable fields of a transaction (incl. account_id and
// type). The balance trigger (migration 014) recalculates both the old and new
// account when account_id changes.
func (r *TransactionRepository) UpdateFull(ctx context.Context, input UpdateTransactionFullInput) (sqlc.Transaction, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer deferRollback(ctx, tx)

	// Audit trail.
	_, err = tx.Exec(ctx, fmt.Sprintf("SET LOCAL app.current_user_id = '%s'", input.UpdatedBy.String()))
	if err != nil {
		return sqlc.Transaction{}, fmt.Errorf("failed to set audit user: %w", err)
	}

	amountBase := r.computeAmountBase(ctx, tx, input.Amount, input.Currency, input.TransactionDate)

	qtx := sqlc.New(tx)

	var description pgtype.Text
	if input.Description != "" {
		description = pgtype.Text{String: input.Description, Valid: true}
	}

	result, err := qtx.UpdateTransactionFull(ctx, sqlc.UpdateTransactionFullParams{
		ID:              input.ID,
		AccountID:       input.AccountID,
		CategoryID:      input.CategoryID,
		Type:            input.Type,
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
	defer deferRollback(ctx, tx)

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

	// Transactions
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

	// Total count
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
