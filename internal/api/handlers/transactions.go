package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/DigitLock/expense-tracker/internal/api/middleware"
	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	transactionRepo *repository.TransactionRepository
	accountRepo     *repository.AccountRepository
	categoryRepo    *repository.CategoryRepository
	userRepo        *repository.UserRepository
	validate        *validator.Validate
}

func NewTransactionHandler(
	transactionRepo *repository.TransactionRepository,
	accountRepo *repository.AccountRepository,
	categoryRepo *repository.CategoryRepository,
	userRepo *repository.UserRepository,
) *TransactionHandler {
	return &TransactionHandler{
		transactionRepo: transactionRepo,
		accountRepo:     accountRepo,
		categoryRepo:    categoryRepo,
		userRepo:        userRepo,
		validate:        newValidator(),
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
// @Failure      400 {object} dto.ErrorResponse "Invalid query parameters"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /transactions [get]
func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	// Parse query params
	typeFilter := r.URL.Query().Get("type")
	accountIDStr := r.URL.Query().Get("account_id")
	month := r.URL.Query().Get("month")

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 100 {
		perPage = 50
	}

	filter := repository.TransactionFilter{
		FamilyID: familyID,
		Limit:    int32(perPage),
		Offset:   int32((page - 1) * perPage),
	}

	if typeFilter == "income" || typeFilter == "expense" {
		filter.Type = &typeFilter
	}

	if accountIDStr != "" {
		if accountID, err := uuid.Parse(accountIDStr); err == nil {
			filter.AccountID = &accountID
		}
	}

	// Parse month filter (YYYY-MM)
	if month != "" {
		if startDate, err := time.Parse("2006-01", month); err == nil {
			endDate := startDate.AddDate(0, 1, -1) // Last day of month
			filter.StartDate = &startDate
			filter.EndDate = &endDate
		}
	}

	transactions, total, err := h.transactionRepo.ListFiltered(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch transactions")
		return
	}

	response := dto.TransactionListResponse{
		Transactions: make([]dto.TransactionResponse, len(transactions)),
		Pagination: dto.PaginationMeta{
			Page:       page,
			PerPage:    perPage,
			Total:      int(total),
			TotalPages: (int(total) + perPage - 1) / perPage,
		},
	}

	for i, t := range transactions {
		response.Transactions[i] = h.mapTransaction(r.Context(), t)
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

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, formatValidationErrors(err))
		return
	}

	// Business validation
	if errors := req.ValidateBusiness(); len(errors) > 0 {
		writeValidationError(w, errors)
		return
	}

	// Validate account belongs to family
	account, err := h.accountRepo.GetByID(r.Context(), req.AccountID)
	if err != nil || account.FamilyID != familyID {
		writeValidationError(w, []dto.ValidationError{
			{Field: "account_id", Message: "Account not found"},
		})
		return
	}

	// Transaction currency must match account currency
	if req.Currency != account.Currency {
		writeValidationError(w, []dto.ValidationError{
			{Field: "currency", Message: "Transaction currency must match account currency (" + account.Currency + ")"},
		})
		return
	}

	// Validate category belongs to family and matches type
	category, err := h.categoryRepo.GetByID(r.Context(), req.CategoryID)
	if err != nil || category.FamilyID != familyID {
		writeValidationError(w, []dto.ValidationError{
			{Field: "category_id", Message: "Category not found"},
		})
		return
	}
	if category.Type != req.Type {
		writeValidationError(w, []dto.ValidationError{
			{Field: "category_id", Message: "Category type must match transaction type"},
		})
		return
	}

	date, _ := time.Parse("2006-01-02", req.Date)

	transaction, err := h.transactionRepo.Create(r.Context(), repository.CreateTransactionInput{
		FamilyID:        familyID,
		AccountID:       req.AccountID,
		CategoryID:      req.CategoryID,
		Type:            req.Type,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Description:     req.Description,
		TransactionDate: date,
		CreatedBy:       userID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create transaction: "+err.Error())
		return
	}

	writeSuccess(w, http.StatusCreated, h.mapTransaction(r.Context(), transaction))
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

	transaction, err := h.transactionRepo.GetByID(r.Context(), transactionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Transaction not found")
		return
	}

	if transaction.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Transaction not found")
		return
	}

	writeSuccess(w, http.StatusOK, h.mapTransaction(r.Context(), transaction))
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
// @Failure      404 {object} dto.ErrorResponse "Transaction not found or does not belong to user's family"
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

	existing, err := h.transactionRepo.GetByID(r.Context(), transactionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Transaction not found")
		return
	}
	if existing.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Transaction not found")
		return
	}

	var req dto.UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, formatValidationErrors(err))
		return
	}

	if errors := req.ValidateBusiness(); len(errors) > 0 {
		writeValidationError(w, errors)
		return
	}

	// Build update input with existing values as defaults
	categoryID := existing.CategoryID
	if req.CategoryID != nil {
		// Validate new category
		category, err := h.categoryRepo.GetByID(r.Context(), *req.CategoryID)
		if err != nil || category.FamilyID != familyID {
			writeValidationError(w, []dto.ValidationError{
				{Field: "category_id", Message: "Category not found"},
			})
			return
		}
		categoryID = *req.CategoryID
	}

	amount := existing.Amount
	if req.Amount != nil {
		amount = *req.Amount
	}

	currency := existing.Currency
	if req.Currency != nil {
		// Validate currency matches the transaction's account
		account, err := h.accountRepo.GetByID(r.Context(), existing.AccountID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to validate account currency")
			return
		}
		if *req.Currency != account.Currency {
			writeValidationError(w, []dto.ValidationError{
				{Field: "currency", Message: "Transaction currency must match account currency (" + account.Currency + ")"},
			})
			return
		}
		currency = *req.Currency
	}

	description := ""
	if existing.Description.Valid {
		description = existing.Description.String
	}
	if req.Description != nil {
		description = *req.Description
	}

	transactionDate := existing.TransactionDate.Time
	if req.Date != nil {
		transactionDate, _ = time.Parse("2006-01-02", *req.Date)
	}

	transaction, err := h.transactionRepo.Update(r.Context(), repository.UpdateTransactionInput{
		ID:              transactionID,
		CategoryID:      categoryID,
		Amount:          amount,
		Currency:        currency,
		Description:     description,
		TransactionDate: transactionDate,
		UpdatedBy:       userID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update transaction")
		return
	}

	writeSuccess(w, http.StatusOK, h.mapTransaction(r.Context(), transaction))
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
// @Failure      404 {object} dto.ErrorResponse "Transaction not found or does not belong to user's family"
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

	existing, err := h.transactionRepo.GetByID(r.Context(), transactionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Transaction not found")
		return
	}
	if existing.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Transaction not found")
		return
	}

	if err := h.transactionRepo.Delete(r.Context(), transactionID, userID); err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete transaction")
		return
	}

	writeMessage(w, http.StatusOK, "Transaction deleted successfully")
}

// Helper functions

func (h *TransactionHandler) mapTransaction(ctx context.Context, t sqlc.Transaction) dto.TransactionResponse {
	response := dto.TransactionResponse{
		ID:           t.ID,
		Type:         t.Type,
		Amount:       t.Amount,
		Currency:     t.Currency,
		AmountBase:   t.AmountBase,
		BaseCurrency: "RSD", // MVP: always RSD
		Date:         t.TransactionDate.Time.Format("2006-01-02"),
		CreatedAt:    t.CreatedAt,
	}

	// Get description
	if t.Description.Valid {
		response.Description = &t.Description.String
	}

	// Get category info
	if category, err := h.categoryRepo.GetByID(ctx, t.CategoryID); err == nil {
		response.Category = dto.TransactionCategoryInfo{
			ID:   category.ID,
			Name: category.Name,
			Type: category.Type,
		}
	}

	// Get account info
	if account, err := h.accountRepo.GetByID(ctx, t.AccountID); err == nil {
		response.Account = dto.TransactionAccountInfo{
			ID:   account.ID,
			Name: account.Name,
			Type: account.Type,
		}
	}

	// Get creator name
	if user, err := h.userRepo.GetByID(ctx, t.CreatedBy); err == nil {
		response.CreatedBy = user.Name
	}

	return response
}
