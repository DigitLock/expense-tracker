package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Report DTOs -->

// Spending by Category Report

// SpendingByCategoryResponse represents spending analysis grouped by category
type SpendingByCategoryResponse struct {
	ReportType         string             `json:"report_type" example:"spending_by_category"`
	Period             ReportPeriod       `json:"period"`
	Currency           string             `json:"currency" example:"RSD"`
	TransactionType    string             `json:"transaction_type" example:"expense"`
	SpendingByCategory []CategorySpending `json:"spending_by_category"`
	TotalAmount        decimal.Decimal    `json:"total_amount" example:"58975.00"`
	TotalTransactions  int                `json:"total_transactions" example:"15"`
	GeneratedAt        time.Time          `json:"generated_at" example:"2025-12-06T18:30:00Z"`
	// CurrencyNote is set only when a non-RSD currency was requested but no
	// exchange rate was available, so amounts fall back to RSD.
	CurrencyNote string `json:"currency_note,omitempty" example:"Requested EUR, but no exchange rate available; showing RSD"`
}

// ReportPeriod represents the time period for a report
type ReportPeriod struct {
	StartDate string `json:"start_date" example:"2025-11-01"`
	EndDate   string `json:"end_date" example:"2025-11-30"`
}

// CategorySpending represents spending details for a single category
type CategorySpending struct {
	CategoryID            uuid.UUID       `json:"category_id" example:"e3b5cae7-cf1d-4c22-a0fe-4014590a7a85"`
	CategoryName          string          `json:"category_name" example:"Groceries"`
	TotalAmount           decimal.Decimal `json:"total_amount" example:"11500.00"`
	TransactionCount      int             `json:"transaction_count" example:"8"`
	Percentage            decimal.Decimal `json:"percentage" example:"19.5"`
	AveragePerTransaction decimal.Decimal `json:"average_per_transaction" example:"1437.50"`
}

// Monthly Summary Report

// MonthlySummaryResponse represents comprehensive monthly financial summary
type MonthlySummaryResponse struct {
	ReportType        string                     `json:"report_type" example:"monthly_summary"`
	Month             string                     `json:"month" example:"2025-11"`
	Currency          string                     `json:"currency" example:"RSD"`
	Summary           MonthlySummary             `json:"summary"`
	IncomeBreakdown   map[string]decimal.Decimal `json:"income_breakdown" example:"Freelance:25000,Salary:50000"`
	ExpenseBreakdown  map[string]decimal.Decimal `json:"expense_breakdown" example:"Groceries:11500,Utilities:8500"`
	AccountBalances   AccountBalances            `json:"account_balances"`
	TransactionCounts TransactionCounts          `json:"transaction_counts"`
	GeneratedAt       time.Time                  `json:"generated_at" example:"2025-12-06T18:30:00Z"`
	// CurrencyNote is set only when a non-RSD currency was requested but no
	// exchange rate was available, so amounts fall back to RSD.
	CurrencyNote string `json:"currency_note,omitempty" example:"Requested EUR, but no exchange rate available; showing RSD"`
}

// MonthlySummary represents monthly financial summary metrics
type MonthlySummary struct {
	TotalIncome   decimal.Decimal `json:"total_income" example:"75000.00"`
	TotalExpenses decimal.Decimal `json:"total_expenses" example:"58975.00"`
	NetSavings    decimal.Decimal `json:"net_savings" example:"16025.00"`
	SavingsRate   decimal.Decimal `json:"savings_rate" example:"21.4"`
}

// AccountBalances represents current account balances
type AccountBalances struct {
	Accounts map[string]decimal.Decimal `json:"accounts,omitempty" example:"Cash Wallet RSD:2000,Bank Card RSD:206700"`
	Total    decimal.Decimal            `json:"total" example:"208700.00"`
}

// TransactionCounts represents transaction count statistics
type TransactionCounts struct {
	IncomeTransactions  int `json:"income_transactions" example:"5"`
	ExpenseTransactions int `json:"expense_transactions" example:"15"`
	TotalTransactions   int `json:"total_transactions" example:"20"`
}
