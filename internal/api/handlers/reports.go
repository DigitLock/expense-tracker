package handlers

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/api/middleware"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

type ReportHandler struct {
	transactionRepo *repository.TransactionRepository
	accountRepo     *repository.AccountRepository
	categoryRepo    *repository.CategoryRepository
}

func NewReportHandler(
	transactionRepo *repository.TransactionRepository,
	accountRepo *repository.AccountRepository,
	categoryRepo *repository.CategoryRepository,
) *ReportHandler {
	return &ReportHandler{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
	}
}

// SpendingByCategory godoc
// @Summary Spending by category report
// @Description Returns spending breakdown by category for a date range
// @Tags reports
// @Produce json
// @Security BearerAuth
// @Param start_date query string false "Start date (YYYY-MM-DD), default: first day of current month"
// @Param end_date query string false "End date (YYYY-MM-DD), default: today"
// @Param type query string false "Transaction type: income or expense (default: expense)"
// @Success 200 {object} dto.SuccessResponse{data=dto.SpendingByCategoryResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/reports/spending-by-category [get]
func (h *ReportHandler) SpendingByCategory(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	// Parse parameters
	now := time.Now()
	startDate := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := now

	if sd := r.URL.Query().Get("start_date"); sd != "" {
		if parsed, err := time.Parse("2006-01-02", sd); err == nil {
			startDate = parsed
		}
	}

	if ed := r.URL.Query().Get("end_date"); ed != "" {
		if parsed, err := time.Parse("2006-01-02", ed); err == nil {
			endDate = parsed
		}
	}

	transactionType := r.URL.Query().Get("type")
	if transactionType != "income" && transactionType != "expense" {
		transactionType = "expense"
	}

	// Get summary by category
	summaries, err := h.transactionRepo.GetSummaryByCategory(r.Context(), familyID, transactionType, startDate, endDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to generate report")
		return
	}

	// Calculate totals and build response
	var totalAmount decimal.Decimal
	var totalTransactions int

	for _, s := range summaries {
		totalAmount = totalAmount.Add(s.Total)
		totalTransactions += int(s.Count)
	}

	// Build category spending list
	categorySpending := make([]dto.CategorySpending, len(summaries))
	for i, s := range summaries {
		// Get category name
		categoryName := "Unknown"
		if cat, err := h.categoryRepo.GetByID(r.Context(), s.CategoryID); err == nil {
			categoryName = cat.Name
		}

		percentage := decimal.Zero
		if !totalAmount.IsZero() {
			percentage = s.Total.Div(totalAmount).Mul(decimal.NewFromInt(100)).Round(1)
		}

		avgPerTransaction := decimal.Zero
		if s.Count > 0 {
			avgPerTransaction = s.Total.Div(decimal.NewFromInt(s.Count)).Round(2)
		}

		categorySpending[i] = dto.CategorySpending{
			CategoryID:            s.CategoryID,
			CategoryName:          categoryName,
			TotalAmount:           s.Total,
			TransactionCount:      int(s.Count),
			Percentage:            percentage,
			AveragePerTransaction: avgPerTransaction,
		}
	}

	response := dto.SpendingByCategoryResponse{
		ReportType: "spending_by_category",
		Period: dto.ReportPeriod{
			StartDate: startDate.Format("2006-01-02"),
			EndDate:   endDate.Format("2006-01-02"),
		},
		Currency:           "RSD",
		TransactionType:    transactionType,
		SpendingByCategory: categorySpending,
		TotalAmount:        totalAmount,
		TotalTransactions:  totalTransactions,
		GeneratedAt:        time.Now().UTC(),
	}

	writeSuccess(w, http.StatusOK, response)
}

// MonthlySummary godoc
// @Summary Monthly summary report
// @Description Returns financial summary for a specific month
// @Tags reports
// @Produce json
// @Security BearerAuth
// @Param month query string false "Month (YYYY-MM), default: current month"
// @Success 200 {object} dto.SuccessResponse{data=dto.MonthlySummaryResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/reports/monthly-summary [get]
func (h *ReportHandler) MonthlySummary(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	// Parse month parameter
	now := time.Now()
	year, month := now.Year(), now.Month()

	if m := r.URL.Query().Get("month"); m != "" {
		if parsed, err := time.Parse("2006-01", m); err == nil {
			year, month = parsed.Year(), parsed.Month()
		}
	}

	startDate := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1) // Last day of month

	// Get summary by type (income/expense totals)
	typeSummaries, err := h.transactionRepo.GetSummaryByType(r.Context(), familyID, startDate, endDate)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to generate report")
		return
	}

	// Extract income and expense totals
	var totalIncome, totalExpenses decimal.Decimal
	var incomeCount, expenseCount int

	for _, s := range typeSummaries {
		if s.Type == "income" {
			totalIncome = s.Total
			incomeCount = int(s.Count)
		} else if s.Type == "expense" {
			totalExpenses = s.Total
			expenseCount = int(s.Count)
		}
	}

	netSavings := totalIncome.Sub(totalExpenses)
	savingsRate := decimal.Zero
	if !totalIncome.IsZero() {
		savingsRate = netSavings.Div(totalIncome).Mul(decimal.NewFromInt(100)).Round(1)
	}

	// Get income breakdown by category
	incomeSummaries, _ := h.transactionRepo.GetSummaryByCategory(r.Context(), familyID, "income", startDate, endDate)
	incomeBreakdown := make(map[string]decimal.Decimal)
	for _, s := range incomeSummaries {
		if cat, err := h.categoryRepo.GetByID(r.Context(), s.CategoryID); err == nil {
			incomeBreakdown[cat.Name] = s.Total
		}
	}

	// Get expense breakdown by category
	expenseSummaries, _ := h.transactionRepo.GetSummaryByCategory(r.Context(), familyID, "expense", startDate, endDate)
	expenseBreakdown := make(map[string]decimal.Decimal)
	for _, s := range expenseSummaries {
		if cat, err := h.categoryRepo.GetByID(r.Context(), s.CategoryID); err == nil {
			expenseBreakdown[cat.Name] = s.Total
		}
	}

	// Get account balances
	accounts, _ := h.accountRepo.ListByFamily(r.Context(), familyID)
	accountBalances := make(map[string]decimal.Decimal)
	var totalBalance decimal.Decimal
	for _, acc := range accounts {
		accountBalances[acc.Name] = acc.CurrentBalance
		totalBalance = totalBalance.Add(acc.CurrentBalance)
	}

	response := dto.MonthlySummaryResponse{
		ReportType: "monthly_summary",
		Month:      startDate.Format("2006-01"),
		Currency:   "RSD",
		Summary: dto.MonthlySummary{
			TotalIncome:   totalIncome,
			TotalExpenses: totalExpenses,
			NetSavings:    netSavings,
			SavingsRate:   savingsRate,
		},
		IncomeBreakdown:  incomeBreakdown,
		ExpenseBreakdown: expenseBreakdown,
		AccountBalances: dto.AccountBalances{
			Accounts: accountBalances,
			Total:    totalBalance,
		},
		TransactionCounts: dto.TransactionCounts{
			IncomeTransactions:  incomeCount,
			ExpenseTransactions: expenseCount,
			TotalTransactions:   incomeCount + expenseCount,
		},
		GeneratedAt: time.Now().UTC(),
	}

	writeSuccess(w, http.StatusOK, response)
}
