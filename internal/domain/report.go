package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Period is the inclusive date range of a report (YYYY-MM-DD strings).
type Period struct {
	StartDate string
	EndDate   string
}

// CategorySpending is one category's slice of a spending report. Money is
// decimal; percentages/averages are decimal too.
type CategorySpending struct {
	CategoryID            uuid.UUID
	CategoryName          string
	TotalAmount           decimal.Decimal
	TransactionCount      int
	Percentage            decimal.Decimal
	AveragePerTransaction decimal.Decimal
}

// SpendingReport is the transport-agnostic spending-by-category report.
// CurrencyNote is set only on a currency fallback (REST surfaces it; gRPC ignores).
type SpendingReport struct {
	ReportType        string
	Period            Period
	Currency          string
	TransactionType   string
	Categories        []CategorySpending
	TotalAmount       decimal.Decimal
	TotalTransactions int
	GeneratedAt       time.Time
	CurrencyNote      string
}

// MonthlySummary is the transport-agnostic monthly summary (TDD §5.5 shape).
// Income/Expenses/NetBalance are in Currency (converted from RSD amount_base).
type MonthlySummary struct {
	Month        string
	Currency     string
	Income       decimal.Decimal
	Expenses     decimal.Decimal
	NetBalance   decimal.Decimal
	IncomeCount  int
	ExpenseCount int
	GeneratedAt  time.Time
}
