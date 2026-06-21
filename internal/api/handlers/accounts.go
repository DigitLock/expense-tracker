package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DigitLock/expense-tracker/internal/api/middleware"
	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/account"
)

// AccountHandler is a thin REST adapter over the account application service.
// Get/GetBalance remain repo-backed (read-only extras outside the gRPC surface).
type AccountHandler struct {
	svc         *account.Service
	accountRepo *repository.AccountRepository
}

func NewAccountHandler(accountRepo *repository.AccountRepository) *AccountHandler {
	return &AccountHandler{
		svc:         account.New(accountRepo),
		accountRepo: accountRepo,
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

	accounts, err := h.svc.List(r.Context(), familyID, includeInactive)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := make([]dto.AccountResponse, len(accounts))
	for i, a := range accounts {
		resp[i] = accountToDTO(a)
	}
	writeSuccess(w, http.StatusOK, dto.AccountListResponse{Accounts: resp})
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
// @Failure      409 {object} dto.ErrorResponse "Account with this name already exists"
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

	in := account.CreateAccountInput{
		Name:           req.Name,
		Type:           req.Type,
		Currency:       req.Currency,
		InitialBalance: req.InitialBalance,
		Description:    derefString(req.Description),
	}

	acc, err := h.svc.Create(r.Context(), familyID, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, accountToDTO(acc))
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

	acc, err := h.accountRepo.GetByID(r.Context(), accountID)
	if err != nil || acc.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}

	writeSuccess(w, http.StatusOK, mapAccount(acc))
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
// @Failure      404 {object} dto.ErrorResponse "Account not found"
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

	var req dto.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	acc, err := h.svc.Update(r.Context(), familyID, accountID, account.UpdateAccountInput{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, accountToDTO(acc))
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
// @Failure      404 {object} dto.ErrorResponse "Account not found"
// @Failure      409 {object} dto.ErrorResponse "Account inactive or has transactions"
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

	if err := h.svc.Delete(r.Context(), familyID, accountID); err != nil {
		writeDomainError(w, err)
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
	if err != nil || account.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Account not found")
		return
	}

	writeSuccess(w, http.StatusOK, dto.AccountBalanceResponse{
		AccountID:      account.ID,
		AccountName:    account.Name,
		Currency:       account.Currency,
		CurrentBalance: account.CurrentBalance,
		BalanceDate:    account.UpdatedAt,
	})
}

// Helper functions

func writeMessage(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(dto.NewMessageResponse(message)); err != nil {
		_ = err
	}
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// accountToDTO maps a domain.Account (service result) to the REST DTO.
func accountToDTO(a domain.Account) dto.AccountResponse {
	var description *string
	if a.Description != "" {
		d := a.Description
		description = &d
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

// mapAccount maps a sqlc.Account to the REST DTO (used by repo-backed Get/GetBalance).
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
