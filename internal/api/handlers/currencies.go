package handlers

import (
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

type CurrencyHandler struct {
	exchangeRateRepo *repository.ExchangeRateRepository
}

func NewCurrencyHandler(exchangeRateRepo *repository.ExchangeRateRepository) *CurrencyHandler {
	return &CurrencyHandler{
		exchangeRateRepo: exchangeRateRepo,
	}
}

// GetRates godoc
// @Summary Get exchange rates
// @Description Returns current exchange rates
// @Tags currencies
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.ExchangeRatesResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/currencies/rates [get]
func (h *CurrencyHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	now := time.Now()

	// Get latest EUR to RSD rate
	eurToRsdRate, err := h.exchangeRateRepo.GetLatestRate(r.Context(), "EUR", "RSD", now)

	var eurToRsd decimal.Decimal
	var lastUpdated time.Time
	var source string

	if err != nil {
		// If no rate found, use a default
		eurToRsd = decimal.NewFromFloat(117.5)
		lastUpdated = now
		source = "fallback"
	} else {
		eurToRsd = eurToRsdRate.Rate
		lastUpdated = eurToRsdRate.CreatedAt
		source = eurToRsdRate.Source
	}

	// Calculate RSD to EUR (inverse)
	rsdToEur := decimal.NewFromInt(1).Div(eurToRsd).Round(6)

	rates := map[string]decimal.Decimal{
		"RSD": decimal.NewFromInt(1),
		"EUR": rsdToEur,
	}

	response := dto.ExchangeRatesResponse{
		BaseCurrency: "RSD",
		Rates:        rates,
		LastUpdated:  lastUpdated,
		Source:       source,
	}

	writeSuccess(w, http.StatusOK, response)
}

// Convert godoc
// @Summary Convert currency
// @Description Converts an amount from one currency to another
// @Tags currencies
// @Produce json
// @Security BearerAuth
// @Param amount query number true "Amount to convert"
// @Param from query string true "Source currency (RSD or EUR)"
// @Param to query string true "Target currency (RSD or EUR)"
// @Success 200 {object} dto.SuccessResponse{data=dto.ConvertCurrencyResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/currencies/convert [get]
func (h *CurrencyHandler) Convert(w http.ResponseWriter, r *http.Request) {
	// Parse parameters
	amountStr := r.URL.Query().Get("amount")
	fromCurrency := r.URL.Query().Get("from")
	toCurrency := r.URL.Query().Get("to")

	// Validate amount
	if amountStr == "" {
		writeValidationError(w, []dto.ValidationError{
			{Field: "amount", Message: "Amount is required"},
		})
		return
	}

	amount, err := decimal.NewFromString(amountStr)
	if err != nil {
		writeValidationError(w, []dto.ValidationError{
			{Field: "amount", Message: "Invalid amount format"},
		})
		return
	}

	// Validate currencies
	validCurrencies := map[string]bool{"RSD": true, "EUR": true}
	if !validCurrencies[fromCurrency] {
		writeValidationError(w, []dto.ValidationError{
			{Field: "from", Message: "Currency must be RSD or EUR"},
		})
		return
	}
	if !validCurrencies[toCurrency] {
		writeValidationError(w, []dto.ValidationError{
			{Field: "to", Message: "Currency must be RSD or EUR"},
		})
		return
	}

	// Same currency - no conversion needed
	if fromCurrency == toCurrency {
		response := dto.ConvertCurrencyResponse{
			OriginalAmount:   amount,
			OriginalCurrency: fromCurrency,
			ConvertedAmount:  amount,
			TargetCurrency:   toCurrency,
			ExchangeRate:     decimal.NewFromInt(1),
			ConversionDate:   time.Now().UTC(),
		}
		writeSuccess(w, http.StatusOK, response)
		return
	}

	now := time.Now()
	var rate decimal.Decimal
	var convertedAmount decimal.Decimal

	if fromCurrency == "EUR" && toCurrency == "RSD" {
		// EUR to RSD
		eurToRsdRate, err := h.exchangeRateRepo.GetLatestRate(r.Context(), "EUR", "RSD", now)
		if err != nil {
			rate = decimal.NewFromFloat(117.5) // fallback
		} else {
			rate = eurToRsdRate.Rate
		}
		convertedAmount = amount.Mul(rate).Round(2)
	} else {
		// RSD to EUR
		eurToRsdRate, err := h.exchangeRateRepo.GetLatestRate(r.Context(), "EUR", "RSD", now)
		var eurToRsd decimal.Decimal
		if err != nil {
			eurToRsd = decimal.NewFromFloat(117.5) // fallback
		} else {
			eurToRsd = eurToRsdRate.Rate
		}
		rate = decimal.NewFromInt(1).Div(eurToRsd).Round(6)
		convertedAmount = amount.Mul(rate).Round(2)
	}

	response := dto.ConvertCurrencyResponse{
		OriginalAmount:   amount,
		OriginalCurrency: fromCurrency,
		ConvertedAmount:  convertedAmount,
		TargetCurrency:   toCurrency,
		ExchangeRate:     rate,
		ConversionDate:   time.Now().UTC(),
	}

	writeSuccess(w, http.StatusOK, response)
}
