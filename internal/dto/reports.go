package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// --- Spending by Category Report ---

// SpendingByCategoryResponse - отчёт по расходам по категориям
type SpendingByCategoryResponse struct {
	ReportType         string             `json:"report_type"`
	Period             ReportPeriod       `json:"period"`
	Currency           string             `json:"currency"`
	TransactionType    string             `json:"transaction_type"`
	SpendingByCategory []CategorySpending `json:"spending_by_category"`
	TotalAmount        decimal.Decimal    `json:"total_amount"`
	TotalTransactions  int                `json:"total_transactions"`
	GeneratedAt        time.Time          `json:"generated_at"`
}

// ReportPeriod - период отчёта
type ReportPeriod struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// CategorySpending - расходы по одной категории
type CategorySpending struct {
	CategoryID            uuid.UUID       `json:"category_id"`
	CategoryName          string          `json:"category_name"`
	TotalAmount           decimal.Decimal `json:"total_amount"`
	TransactionCount      int             `json:"transaction_count"`
	Percentage            decimal.Decimal `json:"percentage"`
	AveragePerTransaction decimal.Decimal `json:"average_per_transaction"`
}

// --- Monthly Summary Report ---

// MonthlySummaryResponse - месячная сводка
type MonthlySummaryResponse struct {
	ReportType        string                     `json:"report_type"`
	Month             string                     `json:"month"`
	Currency          string                     `json:"currency"`
	Summary           MonthlySummary             `json:"summary"`
	IncomeBreakdown   map[string]decimal.Decimal `json:"income_breakdown"`
	ExpenseBreakdown  map[string]decimal.Decimal `json:"expense_breakdown"`
	AccountBalances   AccountBalances            `json:"account_balances"`
	TransactionCounts TransactionCounts          `json:"transaction_counts"`
	GeneratedAt       time.Time                  `json:"generated_at"`
}

// MonthlySummary - итоги месяца
type MonthlySummary struct {
	TotalIncome   decimal.Decimal `json:"total_income"`
	TotalExpenses decimal.Decimal `json:"total_expenses"`
	NetSavings    decimal.Decimal `json:"net_savings"`
	SavingsRate   decimal.Decimal `json:"savings_rate"`
}

// AccountBalances - балансы по счетам
type AccountBalances struct {
	Accounts map[string]decimal.Decimal `json:"accounts,omitempty"`
	Total    decimal.Decimal            `json:"total"`
}

// TransactionCounts - количество транзакций
type TransactionCounts struct {
	IncomeTransactions  int `json:"income_transactions"`
	ExpenseTransactions int `json:"expense_transactions"`
	TotalTransactions   int `json:"total_transactions"`
}
