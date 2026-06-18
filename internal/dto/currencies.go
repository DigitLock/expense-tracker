package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// Currency DTOs -->

// Exchange Rates

// ExchangeRatesResponse represents current exchange rates
type ExchangeRatesResponse struct {
	BaseCurrency string                     `json:"base_currency" example:"RSD"`
	Rates        map[string]decimal.Decimal `json:"rates" example:"RSD:1,EUR:0.008511"`
	LastUpdated  time.Time                  `json:"last_updated" example:"2025-12-06T18:30:00Z"`
	Source       string                     `json:"source" example:"api"`
}

// ExchangeRateSyncResponse summarizes a forced exchange-rate sync.
type ExchangeRateSyncResponse struct {
	SyncedPairs int       `json:"synced_pairs" example:"2"`
	Source      string    `json:"source" example:"fawazahmed0"`
	FetchedAt   time.Time `json:"fetched_at" example:"2025-12-06T18:30:00Z"`
}

// Currency Conversion

// ConvertCurrencyResponse represents currency conversion result
type ConvertCurrencyResponse struct {
	OriginalAmount   decimal.Decimal `json:"original_amount" example:"100.00"`
	OriginalCurrency string          `json:"original_currency" example:"EUR"`
	ConvertedAmount  decimal.Decimal `json:"converted_amount" example:"11750.00"`
	TargetCurrency   string          `json:"target_currency" example:"RSD"`
	ExchangeRate     decimal.Decimal `json:"exchange_rate" example:"117.50"`
	ConversionDate   time.Time       `json:"conversion_date" example:"2025-12-06T18:30:00Z"`
}
