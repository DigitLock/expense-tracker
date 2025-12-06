package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// --- Requests ---

// CreateTransactionRequest - запрос на создание транзакции
type CreateTransactionRequest struct {
	Type        string          `json:"type" validate:"required,oneof=income expense"`
	Amount      decimal.Decimal `json:"amount" validate:"required"`
	Currency    string          `json:"currency" validate:"required,oneof=RSD EUR"`
	CategoryID  uuid.UUID       `json:"category_id" validate:"required"`
	AccountID   uuid.UUID       `json:"account_id" validate:"required"`
	Description string          `json:"description,omitempty" validate:"max=500"`
	Date        string          `json:"date" validate:"required"` // YYYY-MM-DD
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

// UpdateTransactionRequest - запрос на обновление транзакции (partial)
type UpdateTransactionRequest struct {
	Type        *string          `json:"type,omitempty" validate:"omitempty,oneof=income expense"`
	Amount      *decimal.Decimal `json:"amount,omitempty"`
	Currency    *string          `json:"currency,omitempty" validate:"omitempty,oneof=RSD EUR"`
	CategoryID  *uuid.UUID       `json:"category_id,omitempty"`
	AccountID   *uuid.UUID       `json:"account_id,omitempty"`
	Description *string          `json:"description,omitempty" validate:"omitempty,max=500"`
	Date        *string          `json:"date,omitempty"` // YYYY-MM-DD
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

// --- Responses ---

// TransactionResponse - транзакция в ответе API
type TransactionResponse struct {
	ID           uuid.UUID               `json:"id"`
	Type         string                  `json:"type"`
	Amount       decimal.Decimal         `json:"amount"`
	Currency     string                  `json:"currency"`
	AmountBase   decimal.Decimal         `json:"amount_base"`
	BaseCurrency string                  `json:"base_currency"`
	Category     TransactionCategoryInfo `json:"category"`
	Account      TransactionAccountInfo  `json:"account"`
	Description  *string                 `json:"description,omitempty"`
	Date         string                  `json:"date"`
	CreatedAt    time.Time               `json:"created_at"`
	CreatedBy    string                  `json:"created_by"`
}

// TransactionCategoryInfo - информация о категории в транзакции
type TransactionCategoryInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

// TransactionAccountInfo - информация об аккаунте в транзакции
type TransactionAccountInfo struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Type string    `json:"type"`
}

// TransactionListResponse - список транзакций с пагинацией
type TransactionListResponse struct {
	Transactions []TransactionResponse `json:"transactions"`
	Pagination   PaginationMeta        `json:"pagination"`
}
