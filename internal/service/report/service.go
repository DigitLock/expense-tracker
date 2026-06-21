// Package report holds the read-only application-layer service for the Report
// domain. It reuses the existing transaction aggregation (same repo functions
// as REST) so gRPC and REST report values are identical.
package report

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/domain"
)

const dateLayout = "2006-01-02"

// TxAggregator is the existing transaction aggregation reused for reports.
type TxAggregator interface {
	GetSummaryByCategory(ctx context.Context, familyID uuid.UUID, txType string, start, end time.Time) ([]sqlc.GetTransactionsSummaryByCategoryRow, error)
	GetSummaryByType(ctx context.Context, familyID uuid.UUID, start, end time.Time) ([]sqlc.GetTransactionsSummaryByTypeRow, error)
}

// CategoryReader resolves category names.
type CategoryReader interface {
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Category, error)
}

// RateReader reads the latest exchange rate (for currency conversion).
type RateReader interface {
	GetLatestRate(ctx context.Context, fromCurrency, toCurrency string, date time.Time) (sqlc.ExchangeRate, error)
}

type Service struct {
	tx         TxAggregator
	categories CategoryReader
	rates      RateReader
}

func New(tx TxAggregator, categories CategoryReader, rates RateReader) *Service {
	return &Service{tx: tx, categories: categories, rates: rates}
}

// SpendingParams are the (already string) query inputs for spending-by-category.
type SpendingParams struct {
	StartDate string // YYYY-MM-DD, optional
	EndDate   string // YYYY-MM-DD, optional
	Currency  string // optional, default RSD
	Type      string // income|expense, default expense
}

// Conversion converts RSD amount_base sums into the resolved report currency.
type Conversion struct {
	Currency string
	Rate     decimal.Decimal
	Convert  bool
	Note     string
}

// Apply converts an RSD amount into the resolved currency (target = rsd / rate).
func (c Conversion) Apply(rsd decimal.Decimal) decimal.Decimal {
	if !c.Convert {
		return rsd
	}
	return rsd.Div(c.Rate).Round(2)
}

// ResolveCurrency resolves the requested currency to a conversion. Unknown
// non-RSD currencies with no available rate fall back to RSD with a note.
// Exported so the REST adapter converts its extra fields identically.
func ResolveCurrency(ctx context.Context, rates RateReader, requested string) Conversion {
	switch requested {
	case "EUR", "USD":
	default:
		return Conversion{Currency: "RSD"}
	}
	rate, err := rates.GetLatestRate(ctx, requested, "RSD", time.Now())
	if err != nil || rate.Rate.IsZero() {
		return Conversion{Currency: "RSD", Note: "Requested " + requested + ", but no exchange rate available; showing RSD"}
	}
	return Conversion{Currency: requested, Rate: rate.Rate, Convert: true}
}

// SpendingByCategory builds the spending-by-category report, applying the same
// defaults and currency conversion as REST.
func (s *Service) SpendingByCategory(ctx context.Context, familyID uuid.UUID, p SpendingParams) (domain.SpendingReport, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	end := now

	if p.StartDate != "" {
		parsed, err := time.Parse(dateLayout, p.StartDate)
		if err != nil {
			return domain.SpendingReport{}, domain.Errorf(domain.ErrInvalidArgument, "invalid date format, expected YYYY-MM-DD")
		}
		start = parsed
	}
	if p.EndDate != "" {
		parsed, err := time.Parse(dateLayout, p.EndDate)
		if err != nil {
			return domain.SpendingReport{}, domain.Errorf(domain.ErrInvalidArgument, "invalid date format, expected YYYY-MM-DD")
		}
		end = parsed
	}
	if start.After(end) {
		return domain.SpendingReport{}, domain.Errorf(domain.ErrInvalidArgument, "start_date must be before end_date")
	}

	txType, err := resolveType(p.Type)
	if err != nil {
		return domain.SpendingReport{}, err
	}
	if err := validateCurrency(p.Currency); err != nil {
		return domain.SpendingReport{}, err
	}
	conv := ResolveCurrency(ctx, s.rates, p.Currency)

	summaries, err := s.tx.GetSummaryByCategory(ctx, familyID, txType, start, end)
	if err != nil {
		return domain.SpendingReport{}, err
	}

	var totalAmount decimal.Decimal
	var totalTransactions int
	for _, sm := range summaries {
		totalAmount = totalAmount.Add(sm.Total)
		totalTransactions += int(sm.Count)
	}

	categories := make([]domain.CategorySpending, len(summaries))
	for i, sm := range summaries {
		name := "Unknown"
		if cat, err := s.categories.GetByID(ctx, sm.CategoryID); err == nil {
			name = cat.Name
		}
		percentage := decimal.Zero
		if !totalAmount.IsZero() {
			percentage = sm.Total.Div(totalAmount).Mul(decimal.NewFromInt(100)).Round(1)
		}
		avg := decimal.Zero
		if sm.Count > 0 {
			avg = conv.Apply(sm.Total.Div(decimal.NewFromInt(sm.Count)).Round(2))
		}
		categories[i] = domain.CategorySpending{
			CategoryID:            sm.CategoryID,
			CategoryName:          name,
			TotalAmount:           conv.Apply(sm.Total),
			TransactionCount:      int(sm.Count),
			Percentage:            percentage,
			AveragePerTransaction: avg,
		}
	}

	return domain.SpendingReport{
		ReportType:        "spending_by_category",
		Period:            domain.Period{StartDate: start.Format(dateLayout), EndDate: end.Format(dateLayout)},
		Currency:          conv.Currency,
		TransactionType:   txType,
		Categories:        categories,
		TotalAmount:       conv.Apply(totalAmount),
		TotalTransactions: totalTransactions,
		GeneratedAt:       time.Now().UTC(),
		CurrencyNote:      conv.Note,
	}, nil
}

// MonthlySummary builds the core monthly summary (income/expenses/net/counts).
func (s *Service) MonthlySummary(ctx context.Context, familyID uuid.UUID, month, currency string) (domain.MonthlySummary, error) {
	now := time.Now()
	year, mon := now.Year(), now.Month()
	if month != "" {
		parsed, err := time.Parse("2006-01", month)
		if err != nil {
			return domain.MonthlySummary{}, domain.Errorf(domain.ErrInvalidArgument, "invalid month format, expected YYYY-MM")
		}
		year, mon = parsed.Year(), parsed.Month()
	}
	if err := validateCurrency(currency); err != nil {
		return domain.MonthlySummary{}, err
	}
	conv := ResolveCurrency(ctx, s.rates, currency)

	start := time.Date(year, mon, 1, 0, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 1, -1)

	typeSummaries, err := s.tx.GetSummaryByType(ctx, familyID, start, end)
	if err != nil {
		return domain.MonthlySummary{}, err
	}

	var income, expenses decimal.Decimal
	var incomeCount, expenseCount int
	for _, ts := range typeSummaries {
		switch ts.Type {
		case "income":
			income = ts.Total
			incomeCount = int(ts.Count)
		case "expense":
			expenses = ts.Total
			expenseCount = int(ts.Count)
		}
	}

	return domain.MonthlySummary{
		Month:        start.Format("2006-01"),
		Currency:     conv.Currency,
		Income:       conv.Apply(income),
		Expenses:     conv.Apply(expenses),
		NetBalance:   conv.Apply(income.Sub(expenses)),
		IncomeCount:  incomeCount,
		ExpenseCount: expenseCount,
		GeneratedAt:  time.Now().UTC(),
	}, nil
}

func resolveType(t string) (string, error) {
	switch t {
	case "":
		return "expense", nil // default
	case "income", "expense":
		return t, nil
	default:
		return "", domain.Errorf(domain.ErrInvalidArgument, "invalid transaction type")
	}
}

func validateCurrency(c string) error {
	switch c {
	case "", "RSD", "EUR", "USD":
		return nil
	default:
		return domain.Errorf(domain.ErrInvalidArgument, "invalid currency")
	}
}
