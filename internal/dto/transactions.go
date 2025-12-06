package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Transaction Requests -->

// CreateTransactionRequest represents data needed to create a new transaction
type CreateTransactionRequest struct {
	Type        string          `json:"type" validate:"required,oneof=income expense" example:"expense"`
	Amount      decimal.Decimal `json:"amount" validate:"required" example:"1500.00"`
	Currency    string          `json:"currency" validate:"required,oneof=RSD EUR" example:"RSD"`
	CategoryID  uuid.UUID       `json:"category_id" validate:"required" example:"e3b5cae7-cf1d-4c22-a0fe-4014590a7a85"`
	AccountID   uuid.UUID       `json:"account_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440010"`
	Description string          `json:"description,omitempty" validate:"max=500" example:"Weekly groceries at Maxi"`
	Date        string          `json:"date" validate:"required" example:"2025-12-06"` // YYYY-MM-DD
}

// ValidateBusiness performs business logic validation
func (r *CreateTransactionRequest) ValidateBusiness() []ValidationError {
	var errors []ValidationError

	// Amount must be positive
	if r.Amount.LessThanOrEqual(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   "amount",
			Message: "Amount must be positive",
		})
	}

	// Validate date format and not in future
	date, err := time.Parse("2006-01-02", r.Date)
	if err != nil {
		errors = append(errors, ValidationError{
			Field:   "date",
			Message: "Invalid date format, use YYYY-MM-DD",
		})
	} else if date.After(time.Now()) {
		errors = append(errors, ValidationError{
			Field:   "date",
			Message: "Transaction date cannot be in the future",
		})
	}

	return errors
}

// UpdateTransactionRequest represents data for updating a transaction (partial update)
type UpdateTransactionRequest struct {
	Type        *string          `json:"type,omitempty" validate:"omitempty,oneof=income expense" example:"expense"`
	Amount      *decimal.Decimal `json:"amount,omitempty" example:"1600.00"`
	Currency    *string          `json:"currency,omitempty" validate:"omitempty,oneof=RSD EUR" example:"RSD"`
	CategoryID  *uuid.UUID       `json:"category_id,omitempty" example:"e3b5cae7-cf1d-4c22-a0fe-4014590a7a85"`
	AccountID   *uuid.UUID       `json:"account_id,omitempty" example:"550e8400-e29b-41d4-a716-446655440010"`
	Description *string          `json:"description,omitempty" validate:"omitempty,max=500" example:"Updated description"`
	Date        *string          `json:"date,omitempty" example:"2025-12-05"` // YYYY-MM-DD
}

// ValidateBusiness performs business logic validation
func (r *UpdateTransactionRequest) ValidateBusiness() []ValidationError {
	var errors []ValidationError

	// Amount must be positive if provided
	if r.Amount != nil && r.Amount.LessThanOrEqual(decimal.Zero) {
		errors = append(errors, ValidationError{
			Field:   "amount",
			Message: "Amount must be positive",
		})
	}

	// Validate date if provided
	if r.Date != nil {
		date, err := time.Parse("2006-01-02", *r.Date)
		if err != nil {
			errors = append(errors, ValidationError{
				Field:   "date",
				Message: "Invalid date format, use YYYY-MM-DD",
			})
		} else if date.After(time.Now()) {
			errors = append(errors, ValidationError{
				Field:   "date",
				Message: "Transaction date cannot be in the future",
			})
		}
	}

	return errors
}

// Transaction Responses <--

// TransactionResponse represents detailed transaction information
type TransactionResponse struct {
	ID           uuid.UUID               `json:"id" example:"c776462c-5df9-4dc5-b69b-90d55547eda5"`
	Type         string                  `json:"type" example:"expense"`
	Amount       decimal.Decimal         `json:"amount" example:"1500.00"`
	Currency     string                  `json:"currency" example:"RSD"`
	AmountBase   decimal.Decimal         `json:"amount_base" example:"1500.00"`
	BaseCurrency string                  `json:"base_currency" example:"RSD"`
	Category     TransactionCategoryInfo `json:"category"`
	Account      TransactionAccountInfo  `json:"account"`
	Description  *string                 `json:"description,omitempty" example:"Weekly groceries at Maxi"`
	Date         string                  `json:"date" example:"2025-12-06"`
	CreatedAt    time.Time               `json:"created_at" example:"2025-12-06T14:02:29Z"`
	CreatedBy    string                  `json:"created_by" example:"Demo User"`
}

// TransactionCategoryInfo represents category information in transaction response
type TransactionCategoryInfo struct {
	ID   uuid.UUID `json:"id" example:"e3b5cae7-cf1d-4c22-a0fe-4014590a7a85"`
	Name string    `json:"name" example:"Groceries"`
	Type string    `json:"type" example:"expense"`
}

// TransactionAccountInfo represents account information in transaction response
type TransactionAccountInfo struct {
	ID   uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440010"`
	Name string    `json:"name" example:"Cash Wallet RSD"`
	Type string    `json:"type" example:"cash"`
}

// TransactionListResponse represents a paginated list of transactions
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Pagination   PaginationMeta        `json:"pagination"`
}
