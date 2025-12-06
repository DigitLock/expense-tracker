package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Requests -->

type CreateAccountRequest struct {
	Name           string          `json:"name" validate:"required,min=1,max=100"`
	Type           string          `json:"type" validate:"required,oneof=cash checking savings"`
	Currency       string          `json:"currency" validate:"required,oneof=RSD EUR"`
	InitialBalance decimal.Decimal `json:"initial_balance"`
}

// UpdateAccountRequest - запрос на обновление счёта (partial)
type UpdateAccountRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// --- Responses ---

// AccountResponse - счёт в ответе API
type AccountResponse struct {
	ID             uuid.UUID       `json:"id"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	Currency       string          `json:"currency"`
	InitialBalance decimal.Decimal `json:"initial_balance"`
	CurrentBalance decimal.Decimal `json:"current_balance"`
	IsActive       bool            `json:"is_active"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// AccountListResponse - список счетов
type AccountListResponse struct {
	Accounts []AccountResponse `json:"accounts"`
}

// AccountBalanceResponse - ответ на запрос баланса
type AccountBalanceResponse struct {
	AccountID           uuid.UUID       `json:"account_id"`
	AccountName         string          `json:"account_name"`
	Currency            string          `json:"currency"`
	CurrentBalance      decimal.Decimal `json:"current_balance"`
	BalanceDate         time.Time       `json:"balance_date"`
	LastTransactionDate *time.Time      `json:"last_transaction_date,omitempty"`
}
