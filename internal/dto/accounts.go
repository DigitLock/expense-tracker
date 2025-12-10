package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Account Requests -->

// CreateAccountRequest represents data needed to create a new account
type CreateAccountRequest struct {
	Name           string          `json:"name" validate:"required,min=1,max=100" example:"Cash Wallet RSD"`
	Type           string          `json:"type" validate:"required,oneof=cash checking savings" example:"cash"`
	Currency       string          `json:"currency" validate:"required,oneof=RSD EUR" example:"RSD"`
	InitialBalance decimal.Decimal `json:"initial_balance" validate:"required" example:"5000.00"`
	Description    *string         `json:"description,omitempty"`
}

// UpdateAccountRequest represents data for updating an account (partial update)
type UpdateAccountRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100" example:"Updated Account Name"`
	Description *string `json:"description,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty" example:"true"`
}

// Account Responses <--

// AccountResponse represents account information
type AccountResponse struct {
	ID             uuid.UUID       `json:"id" example:"550e8400-e29b-41d4-a716-446655440010"`
	Name           string          `json:"name" example:"Cash Wallet RSD"`
	Type           string          `json:"type" example:"cash"`
	Currency       string          `json:"currency" example:"RSD"`
	Description    *string         `json:"description,omitempty"`
	InitialBalance decimal.Decimal `json:"initial_balance" example:"5000.00"`
	CurrentBalance decimal.Decimal `json:"current_balance" example:"4850.50"`
	IsActive       bool            `json:"is_active" example:"true"`
	CreatedAt      time.Time       `json:"created_at" example:"2025-11-01T10:00:00Z"`
	UpdatedAt      time.Time       `json:"updated_at" example:"2025-12-06T18:30:00Z"`
}

// AccountListResponse represents a list of accounts
type AccountListResponse struct {
	Accounts []AccountResponse `json:"accounts"`
}

// AccountBalanceResponse represents account balance information
type AccountBalanceResponse struct {
	AccountID      uuid.UUID       `json:"account_id" example:"550e8400-e29b-41d4-a716-446655440010"`
	AccountName    string          `json:"account_name" example:"Cash Wallet RSD"`
	Currency       string          `json:"currency" example:"RSD"`
	CurrentBalance decimal.Decimal `json:"current_balance" example:"4850.50"`
	BalanceDate    time.Time       `json:"balance_date" example:"2025-12-06T18:30:00Z"`
}
