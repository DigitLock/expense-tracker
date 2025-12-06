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
		validate:        validator.New(),
	}
}

// List godoc
// @Summary List transactions
// @Description Returns transactions for the authenticated user's family with pagination
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by type: income or expense"
// @Param account_id query string false "Filter by account ID"
// @Param month query string false "Filter by month (YYYY-MM)"
// @Param page query int false "Page number (default: 1)"
// @Param per_page query int false "Items per page (default: 50, max: 100)"
// @Success 200 {object} dto.SuccessResponse{data=dto.TransactionListResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/transactions [get]
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

	// Build filter
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

	// Get transactions
	transactions, total, err := h.transactionRepo.ListFiltered(r.Context(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch transactions")
		return
	}

	// Map to response
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
// @Summary Create transaction
// @Description Creates a new transaction
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateTransactionRequest true "Transaction data"
// @Success 201 {object} dto.SuccessResponse{data=dto.TransactionResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/transactions [post]
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

	// Struct validation
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

	// Parse date
	date, _ := time.Parse("2006-01-02", req.Date)

	// Create transaction
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
// @Summary Get transaction
// @Description Returns a specific transaction by ID
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.TransactionResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/transactions/{id} [get]
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
// @Summary Update transaction
// @Description Updates an existing transaction (partial update)
// @Tags transactions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Param request body dto.UpdateTransactionRequest true "Transaction data"
// @Success 200 {object} dto.SuccessResponse{data=dto.TransactionResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/transactions/{id} [patch]
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

	// Get existing transaction
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

	// Update transaction
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
// @Summary Delete transaction
// @Description Soft deletes a transaction
// @Tags transactions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Transaction ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/transactions/{id} [delete]
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

	// Check exists and belongs to family
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

// --- Helper functions ---

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
