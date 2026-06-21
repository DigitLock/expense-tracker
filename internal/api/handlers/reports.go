package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/api/middleware"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/report"
)

// ReportHandler is a thin REST adapter over the report query service. The core
// report values come from the service (guaranteeing REST↔gRPC parity); the
// monthly response's richer fields (breakdowns, account balances, savings rate)
// are assembled here from the same repos + shared currency conversion, so the
// REST response shape is unchanged.
type ReportHandler struct {
	svc              *report.Service
	transactionRepo  *repository.TransactionRepository
	accountRepo      *repository.AccountRepository
	categoryRepo     *repository.CategoryRepository
	exchangeRateRepo *repository.ExchangeRateRepository
}

func NewReportHandler(
	transactionRepo *repository.TransactionRepository,
	accountRepo *repository.AccountRepository,
	categoryRepo *repository.CategoryRepository,
	exchangeRateRepo *repository.ExchangeRateRepository,
) *ReportHandler {
	return &ReportHandler{
		svc:              report.New(transactionRepo, categoryRepo, exchangeRateRepo),
		transactionRepo:  transactionRepo,
		accountRepo:      accountRepo,
		categoryRepo:     categoryRepo,
		exchangeRateRepo: exchangeRateRepo,
	}
}

// SpendingByCategory godoc
// @Summary      Spending by category report
// @Tags         Reports
// @Produce      json
// @Security     BearerAuth
// @Param        start_date query string false "Start date YYYY-MM-DD"
// @Param        end_date query string false "End date YYYY-MM-DD"
// @Param        type query string false "income or expense" default(expense)
// @Success      200 {object} dto.SuccessResponse{data=dto.SpendingByCategoryResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Router       /reports/spending-by-category [get]
func (h *ReportHandler) SpendingByCategory(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	rep, err := h.svc.SpendingByCategory(r.Context(), familyID, report.SpendingParams{
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
		Currency:  r.URL.Query().Get("currency"),
		Type:      r.URL.Query().Get("type"),
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	categories := make([]dto.CategorySpending, len(rep.Categories))
	for i, c := range rep.Categories {
		categories[i] = dto.CategorySpending{
			CategoryID:            c.CategoryID,
			CategoryName:          c.CategoryName,
			TotalAmount:           c.TotalAmount,
			TransactionCount:      c.TransactionCount,
			Percentage:            c.Percentage,
			AveragePerTransaction: c.AveragePerTransaction,
		}
	}

	writeSuccess(w, http.StatusOK, dto.SpendingByCategoryResponse{
		ReportType:         rep.ReportType,
		Period:             dto.ReportPeriod{StartDate: rep.Period.StartDate, EndDate: rep.Period.EndDate},
		Currency:           rep.Currency,
		TransactionType:    rep.TransactionType,
		SpendingByCategory: categories,
		TotalAmount:        rep.TotalAmount,
		TotalTransactions:  rep.TotalTransactions,
		GeneratedAt:        rep.GeneratedAt,
		CurrencyNote:       rep.CurrencyNote,
	})
}

// MonthlySummary godoc
// @Summary      Monthly financial summary
// @Tags         Reports
// @Produce      json
// @Security     BearerAuth
// @Param        month query string false "Month YYYY-MM (defaults to current)"
// @Success      200 {object} dto.SuccessResponse{data=dto.MonthlySummaryResponse}
// @Failure      400 {object} dto.ErrorResponse
// @Failure      401 {object} dto.ErrorResponse
// @Router       /reports/monthly-summary [get]
func (h *ReportHandler) MonthlySummary(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	currency := r.URL.Query().Get("currency")
	core, err := h.svc.MonthlySummary(r.Context(), familyID, r.URL.Query().Get("month"), currency)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	// Period derived from the (validated) month for the breakdown queries.
	start, _ := time.Parse("2006-01", core.Month)
	end := start.AddDate(0, 1, -1)
	conv := report.ResolveCurrency(r.Context(), h.exchangeRateRepo, currency)

	savingsRate := decimal.Zero
	if !core.Income.IsZero() {
		savingsRate = core.NetBalance.Div(core.Income).Mul(decimal.NewFromInt(100)).Round(1)
	}

	writeSuccess(w, http.StatusOK, dto.MonthlySummaryResponse{
		ReportType: "monthly_summary",
		Month:      core.Month,
		Currency:   core.Currency,
		Summary: dto.MonthlySummary{
			TotalIncome:   core.Income,
			TotalExpenses: core.Expenses,
			NetSavings:    core.NetBalance,
			SavingsRate:   savingsRate,
		},
		IncomeBreakdown:  h.breakdown(r.Context(), familyID, "income", start, end, conv),
		ExpenseBreakdown: h.breakdown(r.Context(), familyID, "expense", start, end, conv),
		AccountBalances:  h.accountBalances(r.Context(), familyID),
		TransactionCounts: dto.TransactionCounts{
			IncomeTransactions:  core.IncomeCount,
			ExpenseTransactions: core.ExpenseCount,
			TotalTransactions:   core.IncomeCount + core.ExpenseCount,
		},
		GeneratedAt:  core.GeneratedAt,
		CurrencyNote: conv.Note,
	})
}

// breakdown builds a category-name → converted-amount map for a type/period.
func (h *ReportHandler) breakdown(ctx context.Context, familyID uuid.UUID, txType string, start, end time.Time, conv report.Conversion) map[string]decimal.Decimal {
	out := make(map[string]decimal.Decimal)
	summaries, err := h.transactionRepo.GetSummaryByCategory(ctx, familyID, txType, start, end)
	if err != nil {
		return out
	}
	for _, s := range summaries {
		if cat, err := h.categoryRepo.GetByID(ctx, s.CategoryID); err == nil {
			out[cat.Name] = conv.Apply(s.Total)
		}
	}
	return out
}

// accountBalances returns native (unconverted) current balances per account.
func (h *ReportHandler) accountBalances(ctx context.Context, familyID uuid.UUID) dto.AccountBalances {
	accounts, _ := h.accountRepo.ListByFamily(ctx, familyID)
	balances := make(map[string]decimal.Decimal)
	var total decimal.Decimal
	for _, acc := range accounts {
		balances[acc.Name] = acc.CurrentBalance
		total = total.Add(acc.CurrentBalance)
	}
	return dto.AccountBalances{Accounts: balances, Total: total}
}
