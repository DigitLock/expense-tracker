package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/DigitLock/expense-tracker/internal/api/middleware"
	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/DigitLock/expense-tracker/internal/service/transaction"
)

// TransactionHandler is a thin REST adapter over the transaction service.
// The account/category/user repos are retained only to enrich the REST
// response (nested account/category objects and creator name) — the response
// shape is unchanged. Get remains repo-backed (read-only extra).
type TransactionHandler struct {
	svc             *transaction.Service
	transactionRepo *repository.TransactionRepository
	accountRepo     *repository.AccountRepository
	categoryRepo    *repository.CategoryRepository
	userRepo        *repository.UserRepository
}

func NewTransactionHandler(
	transactionRepo *repository.TransactionRepository,
	accountRepo *repository.AccountRepository,
	categoryRepo *repository.CategoryRepository,
	userRepo *repository.UserRepository,
) *TransactionHandler {
	return &TransactionHandler{
		svc:             transaction.New(transactionRepo, accountRepo, categoryRepo),
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		userRepo:        userRepo,
	}
}

// List godoc
// @Summary      List transactions
// @Description  Get paginated list of transactions for the authenticated user's family with optional filters
// @Tags         Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        type query string false "Filter by type: income or expense" Enums(income, expense)
// @Param        account_id query string false "Filter by account ID (UUID)" format(uuid)
// @Param        month query string false "Filter by month in YYYY-MM format" example(2025-12)
// @Param        page query int false "Page number (starts from 1)" default(1) minimum(1)
// @Param        per_page query int false "Items per page" default(50) minimum(1) maximum(100)
// @Success      200 {object} dto.SuccessResponse{data=dto.TransactionListResponse} "Paginated list of transactions"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /transactions [get]
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))

	filter := transaction.ListFilter{
		Month:   r.URL.Query().Get("month"),
		Page:    page,
		PerPage: perPage,
	}
	if typeFilter := r.URL.Query().Get("type"); typeFilter == "income" || typeFilter == "expense" {
		filter.Type = &typeFilter
	}
	if accountIDStr := r.URL.Query().Get("account_id"); accountIDStr != "" {
		if accountID, err := uuid.Parse(accountIDStr); err == nil {
			filter.AccountID = &accountID
		}
	}

	result, err := h.svc.List(r.Context(), familyID, filter)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	response := dto.TransactionListResponse{
		Transactions: make([]dto.TransactionResponse, len(result.Transactions)),
		Pagination: dto.PaginationMeta{
			Page:       result.Page,
			PerPage:    result.PerPage,
			Total:      result.Total,
			TotalPages: result.TotalPages,
		},
	}
	for i, t := range result.Transactions {
		response.Transactions[i] = h.mapDomainTransaction(r.Context(), t)
	}

	writeSuccess(w, http.StatusOK, response)
}

// Create godoc
// @Summary      Create transaction
// @Description  Create a new income or expense transaction with automatic currency conversion to base currency (RSD)
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateTransactionRequest true "Transaction creation data"
// @Success      201 {object} dto.SuccessResponse{data=dto.TransactionResponse} "Transaction created successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid request body, validation error, or business rule violation"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /transactions [post]
func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User context not found")
		return
	}

	var req dto.CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	tx, err := h.svc.Create(r.Context(), familyID, userID, transaction.CreateTransactionInput{
		Type:        req.Type,
		Amount:      req.Amount,
		Currency:    req.Currency,
		CategoryID:  req.CategoryID.String(),
		AccountID:   req.AccountID.String(),
		Description: req.Description,
		Date:        req.Date,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, h.mapDomainTransaction(r.Context(), tx))
}

// Get godoc
// @Summary      Get transaction by ID
// @Description  Retrieve detailed information about a specific transaction
// @Tags         Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Transaction ID (UUID)" format(uuid)
// @Success      200 {object} dto.SuccessResponse{data=dto.TransactionResponse} "Transaction details"
// @Failure      400 {object} dto.ErrorResponse "Invalid transaction ID format"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Transaction not found or does not belong to user's family"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /transactions/{id} [get]
func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	transactionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid transaction ID format")
		return
	}

	t, err := h.transactionRepo.GetByID(r.Context(), transactionID)
	if err != nil || t.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Transaction not found")
		return
	}

	writeSuccess(w, http.StatusOK, h.mapSqlcTransaction(r.Context(), t))
}

// Update godoc
// @Summary      Update transaction
// @Description  Update transaction information (partial update - only provided fields will be updated). Account balance is automatically recalculated.
// @Tags         Transactions
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Transaction ID (UUID)" format(uuid)
// @Param        request body dto.UpdateTransactionRequest true "Transaction update data (partial)"
// @Success      200 {object} dto.SuccessResponse{data=dto.TransactionResponse} "Transaction updated successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid request body, validation error, or business rule violation"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Transaction not found"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /transactions/{id} [patch]
func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User context not found")
		return
	}

	transactionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid transaction ID format")
		return
	}

	var req dto.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	// REST supports updating category/amount/currency/description/date (not
	// account_id or type — those are only exposed over gRPC).
	in := transaction.UpdateTransactionInput{
		Amount:      req.Amount,
		Currency:    req.Currency,
		Description: req.Description,
		Date:        req.Date,
	}
	if req.CategoryID != nil {
		s := req.CategoryID.String()
		in.CategoryID = &s
	}

	tx, err := h.svc.Update(r.Context(), familyID, userID, transactionID, in)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, h.mapDomainTransaction(r.Context(), tx))
}

// Delete godoc
// @Summary      Delete transaction
// @Description  Soft delete a transaction (marks as inactive, account balance is automatically recalculated)
// @Tags         Transactions
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Transaction ID (UUID)" format(uuid)
// @Success      200 {object} dto.MessageResponse "Transaction deleted successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid transaction ID format"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Transaction not found"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /transactions/{id} [delete]
func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "User context not found")
		return
	}

	transactionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid transaction ID format")
		return
	}

	if err := h.svc.Delete(r.Context(), familyID, userID, transactionID); err != nil {
		writeDomainError(w, err)
		return
	}

	writeMessage(w, http.StatusOK, "Transaction deleted successfully")
}

// Helper functions

// mapDomainTransaction maps a service-layer domain.Transaction to the REST DTO,
// enriching with nested category/account objects and creator name.
func (h *TransactionHandler) mapDomainTransaction(ctx context.Context, t domain.Transaction) dto.TransactionResponse {
	resp := dto.TransactionResponse{
		ID:           t.ID,
		Type:         t.Type,
		Amount:       t.Amount,
		Currency:     t.Currency,
		AmountBase:   t.AmountBase,
		BaseCurrency: t.BaseCurrency,
		Date:         t.Date.Format("2006-01-02"),
		CreatedAt:    t.CreatedAt,
	}
	if t.Description != "" {
		d := t.Description
		resp.Description = &d
	}
	h.enrich(ctx, &resp, t.CategoryID, t.AccountID, t.CreatedBy)
	return resp
}

// mapSqlcTransaction maps a repo sqlc.Transaction to the REST DTO (used by Get).
func (h *TransactionHandler) mapSqlcTransaction(ctx context.Context, t sqlc.Transaction) dto.TransactionResponse {
	resp := dto.TransactionResponse{
		ID:           t.ID,
		Type:         t.Type,
		Amount:       t.Amount,
		Currency:     t.Currency,
		AmountBase:   t.AmountBase,
		BaseCurrency: "RSD",
		Date:         t.TransactionDate.Time.Format("2006-01-02"),
		CreatedAt:    t.CreatedAt,
	}
	if t.Description.Valid {
		resp.Description = &t.Description.String
	}
	h.enrich(ctx, &resp, t.CategoryID, t.AccountID, t.CreatedBy)
	return resp
}

// enrich fills the nested category/account objects and creator name.
func (h *TransactionHandler) enrich(ctx context.Context, resp *dto.TransactionResponse, categoryID, accountID, createdBy uuid.UUID) {
	if category, err := h.categoryRepo.GetByID(ctx, categoryID); err == nil {
		resp.Category = dto.TransactionCategoryInfo{ID: category.ID, Name: category.Name, Type: category.Type}
	}
	if account, err := h.accountRepo.GetByID(ctx, accountID); err == nil {
		resp.Account = dto.TransactionAccountInfo{ID: account.ID, Name: account.Name, Type: account.Type}
	}
	if user, err := h.userRepo.GetByID(ctx, createdBy); err == nil {
		resp.CreatedBy = user.Name
	}
}
