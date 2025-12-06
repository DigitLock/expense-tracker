package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// Exchange Rates

type ExchangeRatesResponse struct {
	BaseCurrency string                     `json:"base_currency"`
	Rates        map[string]decimal.Decimal `json:"rates"`
	LastUpdated  time.Time                  `json:"last_updated"`
	Source       string                     `json:"source"`
}

// Currency Conversion

type ConvertCurrencyResponse struct {
	OriginalAmount   decimal.Decimal `json:"original_amount"`
	OriginalCurrency string          `json:"original_currency"`
	ConvertedAmount  decimal.Decimal `json:"converted_amount"`
	TargetCurrency   string          `json:"target_currency"`
	ExchangeRate     decimal.Decimal `json:"exchange_rate"`
	ConversionDate   time.Time       `json:"conversion_date"`
}
