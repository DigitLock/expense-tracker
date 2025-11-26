package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

// ExchangeRateRepository handles exchange rate data operations
type ExchangeRateRepository struct {
	queries *sqlc.Queries
}

// NewExchangeRateRepository creates a new ExchangeRateRepository
func NewExchangeRateRepository(queries *sqlc.Queries) *ExchangeRateRepository {
	return &ExchangeRateRepository{queries: queries}
}

// GetRate retrieves exchange rate for specific date
func (r *ExchangeRateRepository) GetRate(ctx context.Context, fromCurrency, toCurrency string, date time.Time) (sqlc.ExchangeRate, error) {
	return r.queries.GetExchangeRate(ctx, sqlc.GetExchangeRateParams{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Date:         pgtype.Date{Time: date, Valid: true},
	})
}

// GetLatestRate retrieves latest exchange rate up to given date
func (r *ExchangeRateRepository) GetLatestRate(ctx context.Context, fromCurrency, toCurrency string, date time.Time) (sqlc.ExchangeRate, error) {
	return r.queries.GetLatestExchangeRate(ctx, sqlc.GetLatestExchangeRateParams{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Date:         pgtype.Date{Time: date, Valid: true},
	})
}

// ListByDate retrieves all exchange rates for a specific date
func (r *ExchangeRateRepository) ListByDate(ctx context.Context, date time.Time) ([]sqlc.ExchangeRate, error) {
	return r.queries.ListExchangeRatesByDate(ctx, pgtype.Date{Time: date, Valid: true})
}

// ListHistory retrieves exchange rate history for a currency pair
func (r *ExchangeRateRepository) ListHistory(ctx context.Context, fromCurrency, toCurrency string, startDate, endDate time.Time) ([]sqlc.ExchangeRate, error) {
	return r.queries.ListExchangeRatesHistory(ctx, sqlc.ListExchangeRatesHistoryParams{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Date:         pgtype.Date{Time: startDate, Valid: true},
		Date_2:       pgtype.Date{Time: endDate, Valid: true},
	})
}

// CreateRateInput contains data for creating a new exchange rate
type CreateRateInput struct {
	FromCurrency string
	ToCurrency   string
	Rate         decimal.Decimal
	Date         time.Time
}

// Create creates a new exchange rate
func (r *ExchangeRateRepository) Create(ctx context.Context, input CreateRateInput) (sqlc.ExchangeRate, error) {
	return r.queries.CreateExchangeRate(ctx, sqlc.CreateExchangeRateParams{
		ID:           uuid.New(),
		FromCurrency: input.FromCurrency,
		ToCurrency:   input.ToCurrency,
		Rate:         input.Rate,
		Date:         pgtype.Date{Time: input.Date, Valid: true},
	})
}

// Upsert creates or updates an exchange rate
func (r *ExchangeRateRepository) Upsert(ctx context.Context, input CreateRateInput) (sqlc.ExchangeRate, error) {
	return r.queries.UpsertExchangeRate(ctx, sqlc.UpsertExchangeRateParams{
		ID:           uuid.New(),
		FromCurrency: input.FromCurrency,
		ToCurrency:   input.ToCurrency,
		Rate:         input.Rate,
		Date:         pgtype.Date{Time: input.Date, Valid: true},
	})
}
