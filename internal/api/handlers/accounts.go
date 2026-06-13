package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/DigitLock/expense-tracker/internal/api/middleware"
	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

type AccountHandler struct {
	accountRepo *repository.AccountRepository
	validate    *validator.Validate
}

func NewAccountHandler(accountRepo *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{
		accountRepo: accountRepo,
		validate:    validator.New(),
	}
}

// stringToPgText converts *string to pgtype.Text
func stringToPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{
		String: *s,
		Valid:  true,
	}
}

// List godoc
// @Summary      List accounts
// @Description  Get all accounts for the authenticated user's family with optional filter for inactive accounts
// @Tags         Accounts
// @Produce      json
// @Security     BearerAuth
// @Param        include_inactive query bool false "Include inactive (deleted) accounts" default(false)
// @Success      200 {object} dto.SuccessResponse{data=dto.AccountListResponse} "List of accounts"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /accounts [get]
func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"

	var dbAccounts []sqlc.Account
	var err error

	if includeInactive {
		dbAccounts, err = h.accountRepo.ListAllByFamily(r.Context(), familyID)
	} else {
		dbAccounts, err = h.accountRepo.ListByFamily(r.Context(), familyID)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch accounts")
		return
	}

	accounts := mapAccounts(dbAccounts)
	writeSuccess(w, http.StatusOK, dto.AccountListResponse{Accounts: accounts})
}

// Create godoc
// @Summary      Create account
// @Description  Create a new financial account (cash, checking, or savings) for the authenticated user's family
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateAccountRequest true "Account creation data"
// @Success      201 {object} dto.SuccessResponse{data=dto.AccountResponse} "Account created successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid request body or validation error"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /accounts [post]
func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	var req dto.CreateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, formatValidationErrors(err))
		return
	}

	// Business validation
	if req.InitialBalance.IsNegative() {
		writeValidationError(w, []dto.ValidationError{
			{Field: "initial_balance", Message: "Initial balance cannot be negative"},
		})
		return
	}

	account, err := h.accountRepo.Create(r.Context(), repository.CreateAccountInput{
		FamilyID:       familyID,
		Name:           req.Name,
		Type:           req.Type,
		Currency:       req.Currency,
		InitialBalance: req.InitialBalance,
		Description:    stringToPgText(req.Description),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create account")
		return
	}

	writeSuccess(w, http.StatusCreated, mapAccount(account))
}

// Get godoc
// @Summary      Get account by ID
// @Description  Retrieve detailed information about a specific account
// @Tags         Accounts
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Account ID (UUID)" format(uuid)
// @Success      200 {object} dto.SuccessResponse{data=dto.AccountResponse} "Account details"
// @Failure      400 {object} dto.ErrorResponse "Invalid account ID format"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Account not found or does not belong to user's family"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /accounts/{id} [get]
func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid account ID format")
		return
	}

	account, err := h.accountRepo.GetByID(r.Context(), accountID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}

	// Account belongs to user's family
	if account.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}

	writeSuccess(w, http.StatusOK, mapAccount(account))
}

// Update godoc
// @Summary      Update account
// @Description  Update account information (partial update - only provided fields will be updated)
// @Tags         Accounts
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Account ID (UUID)" format(uuid)
// @Param        request body dto.UpdateAccountRequest true "Account update data (partial)"
// @Success      200 {object} dto.SuccessResponse{data=dto.AccountResponse} "Account updated successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid request body or validation error"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Account not found or does not belong to user's family"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /accounts/{id} [patch]
func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid account ID format")
		return
	}

	// Account exists and belongs to family
	existing, err := h.accountRepo.GetByIDIncludingInactive(r.Context(), accountID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}
	if existing.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}

	var req dto.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, formatValidationErrors(err))
		return
	}

	// Updated input
	input := repository.UpdateAccountInput{
		ID:          accountID,
		FamilyID:    familyID,
		Description: stringToPgText(req.Description),
	}
	if req.Name != nil {
		input.Name = req.Name
	}
	if req.IsActive != nil {
		input.IsActive = req.IsActive
	}

	account, err := h.accountRepo.Update(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update account")
		return
	}

	writeSuccess(w, http.StatusOK, mapAccount(account))
}

// Delete godoc
// @Summary      Delete account
// @Description  Soft delete an account (marks as inactive, preserves transaction history)
// @Tags         Accounts
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Account ID (UUID)" format(uuid)
// @Success      200 {object} dto.MessageResponse "Account deleted successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid account ID format"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Account not found or does not belong to user's family"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /accounts/{id} [delete]
func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid account ID format")
		return
	}

	// Account exists and belongs to family
	existing, err := h.accountRepo.GetByID(r.Context(), accountID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}
	if existing.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}

	// Block deletion if the account has ANY transactions (active or inactive).
	// Prevents orphaned transaction records and accidental data loss.
	hasTx, err := h.accountRepo.HasTransactions(r.Context(), accountID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to check account transactions")
		return
	}
	if hasTx {
		writeError(w, http.StatusBadRequest, "ACCOUNT_HAS_TRANSACTIONS",
			"Cannot delete account with existing transactions. Delete or move transactions first.")
		return
	}

	if err := h.accountRepo.Delete(r.Context(), accountID); err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete account")
		return
	}

	writeMessage(w, http.StatusOK, "Account deleted successfully")
}

// GetBalance godoc
// @Summary      Get account balance
// @Description  Retrieve current balance and related information for a specific account
// @Tags         Accounts
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Account ID (UUID)" format(uuid)
// @Success      200 {object} dto.SuccessResponse{data=dto.AccountBalanceResponse} "Account balance information"
// @Failure      400 {object} dto.ErrorResponse "Invalid account ID format"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Account not found or does not belong to user's family"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /accounts/{id}/balance [get]
func (h *AccountHandler) GetBalance(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	accountID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid account ID format")
		return
	}

	account, err := h.accountRepo.GetByID(r.Context(), accountID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}

	if account.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}

	response := dto.AccountBalanceResponse{
		AccountID:      account.ID,
		AccountName:    account.Name,
		Currency:       account.Currency,
		CurrentBalance: account.CurrentBalance,
		BalanceDate:    account.UpdatedAt,
	}

	writeSuccess(w, http.StatusOK, response)
}

// Helper functions

func writeMessage(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(dto.NewMessageResponse(message)); err != nil {
		_ = err
	}
}

func mapAccount(a sqlc.Account) dto.AccountResponse {
	var description *string
	if a.Description.Valid {
		description = &a.Description.String
	}

	return dto.AccountResponse{
		ID:             a.ID,
		Name:           a.Name,
		Type:           a.Type,
		Currency:       a.Currency,
		InitialBalance: a.InitialBalance,
		CurrentBalance: a.CurrentBalance,
		Description:    description,
		IsActive:       a.IsActive,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

func mapAccounts(accounts []sqlc.Account) []dto.AccountResponse {
	result := make([]dto.AccountResponse, len(accounts))
	for i, a := range accounts {
		result[i] = mapAccount(a)
	}
	return result
}
