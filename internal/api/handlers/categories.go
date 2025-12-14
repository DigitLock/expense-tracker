package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/DigitLock/expense-tracker/internal/api/middleware"
	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/dto"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

type CategoryHandler struct {
	categoryRepo *repository.CategoryRepository
	validate     *validator.Validate
}

func NewCategoryHandler(categoryRepo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		categoryRepo: categoryRepo,
		validate:     validator.New(),
	}
}

// List godoc
// @Summary      List categories
// @Description  Get all transaction categories for the authenticated user's family with optional filters for type and inactive categories. Supports hierarchical structure (parent-child).
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        type query string false "Filter by type: income or expense" Enums(income, expense)
// @Param        include_inactive query bool false "Include inactive (deleted) categories" default(false)
// @Success      200 {object} dto.SuccessResponse{data=dto.CategoryListResponse} "List of categories"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /categories [get]
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	includeInactive := r.URL.Query().Get("include_inactive") == "true"
	typeFilter := r.URL.Query().Get("type")

	var dbCategories []sqlc.Category
	var err error

	// Apply filters
	if typeFilter != "" && (typeFilter == "income" || typeFilter == "expense") {
		dbCategories, err = h.categoryRepo.ListByType(r.Context(), familyID, typeFilter)
	} else if includeInactive {
		dbCategories, err = h.categoryRepo.ListAllByFamily(r.Context(), familyID)
	} else {
		dbCategories, err = h.categoryRepo.ListByFamily(r.Context(), familyID)
	}

	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to fetch categories")
		return
	}

	categories := mapCategories(dbCategories)
	writeSuccess(w, http.StatusOK, dto.CategoryListResponse{Categories: categories})
}

// Create godoc
// @Summary      Create category
// @Description  Create a new transaction category (income or expense) with optional parent for hierarchical structure
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body dto.CreateCategoryRequest true "Category creation data"
// @Success      201 {object} dto.SuccessResponse{data=dto.CategoryResponse} "Category created successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid request body or validation error"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /categories [post]
func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	var req dto.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, formatValidationErrors(err))
		return
	}

	// Validate parent_id if provided
	if req.ParentID != nil {
		parent, err := h.categoryRepo.GetByID(r.Context(), *req.ParentID)
		if err != nil {
			writeValidationError(w, []dto.ValidationError{
				{Field: "parent_id", Message: "Parent category not found"},
			})
			return
		}
		// Parent must belong to same family
		if parent.FamilyID != familyID {
			writeValidationError(w, []dto.ValidationError{
				{Field: "parent_id", Message: "Parent category not found"},
			})
			return
		}
		// Parent must be same type
		if parent.Type != req.Type {
			writeValidationError(w, []dto.ValidationError{
				{Field: "parent_id", Message: "Parent category must be the same type"},
			})
			return
		}
	}

	// Checking for inactive duplicate BEFORE creating
	inactiveCategory, err := h.categoryRepo.FindInactiveByName(
		r.Context(),
		familyID,
		req.Name,
		req.Type,
		req.ParentID,
	)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to check existing categories")
		return
	}

	if inactiveCategory != nil {
		// Found inactive category - return conflict with restore option
		var parentID *uuid.UUID
		if inactiveCategory.ParentID.Valid {
			id := uuid.UUID(inactiveCategory.ParentID.Bytes)
			parentID = &id
		}

		writeErrorWithData(w, http.StatusConflict, "CATEGORY_INACTIVE_EXISTS",
			"A category with this name was previously deleted. Would you like to restore it?",
			map[string]interface{}{
				"inactive_category_id": inactiveCategory.ID,
				"name":                 inactiveCategory.Name,
				"type":                 inactiveCategory.Type,
				"parent_id":            parentID,
				"deleted_at":           inactiveCategory.UpdatedAt,
			})
		return
	}

	category, err := h.categoryRepo.Create(r.Context(), repository.CreateCategoryInput{
		FamilyID: familyID,
		Name:     req.Name,
		Type:     req.Type,
		ParentID: req.ParentID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to create category")
		return
	}

	writeSuccess(w, http.StatusCreated, mapCategory(category))
}

// Get godoc
// @Summary      Get category by ID
// @Description  Retrieve detailed information about a specific category
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Category ID (UUID)" format(uuid)
// @Success      200 {object} dto.SuccessResponse{data=dto.CategoryResponse} "Category details"
// @Failure      400 {object} dto.ErrorResponse "Invalid category ID format"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Category not found or does not belong to user's family"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /categories/{id} [get]
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	categoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID format")
		return
	}

	category, err := h.categoryRepo.GetByID(r.Context(), categoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	// Category belongs to user's family
	if category.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	writeSuccess(w, http.StatusOK, mapCategory(category))
}

// Update godoc
// @Summary      Update category
// @Description  Update category information including name, parent, or active status (partial update)
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Category ID (UUID)" format(uuid)
// @Param        request body dto.UpdateCategoryRequest true "Category update data (partial)"
// @Success      200 {object} dto.SuccessResponse{data=dto.CategoryResponse} "Category updated successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid request body or validation error (e.g., circular parent reference)"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Category not found or does not belong to user's family"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /categories/{id} [patch]
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	categoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID format")
		return
	}

	// Category exists and belongs to family
	existing, err := h.categoryRepo.GetByIDIncludingInactive(r.Context(), categoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}
	if existing.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	var req dto.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		writeValidationError(w, formatValidationErrors(err))
		return
	}

	// Validate parent_id if provided
	if req.ParentID != nil {
		// Can't be own parent
		if *req.ParentID == categoryID {
			writeValidationError(w, []dto.ValidationError{
				{Field: "parent_id", Message: "Category cannot be its own parent"},
			})
			return
		}

		parent, err := h.categoryRepo.GetByID(r.Context(), *req.ParentID)
		if err != nil {
			writeValidationError(w, []dto.ValidationError{
				{Field: "parent_id", Message: "Parent category not found"},
			})
			return
		}
		if parent.FamilyID != familyID {
			writeValidationError(w, []dto.ValidationError{
				{Field: "parent_id", Message: "Parent category not found"},
			})
			return
		}
		if parent.Type != existing.Type {
			writeValidationError(w, []dto.ValidationError{
				{Field: "parent_id", Message: "Parent category must be the same type"},
			})
			return
		}
	}

	// Updated input
	input := repository.UpdateCategoryInput{
		ID:       categoryID,
		FamilyID: familyID,
	}
	if req.Name != nil {
		input.Name = req.Name
	}
	if req.ParentID != nil {
		input.ParentID = req.ParentID
	}
	if req.IsActive != nil {
		input.IsActive = req.IsActive
	}

	category, err := h.categoryRepo.Update(r.Context(), input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to update category")
		return
	}

	writeSuccess(w, http.StatusOK, mapCategory(category))
}

// Delete godoc
// @Summary      Delete category
// @Description  Soft delete a category (marks as inactive, preserves transaction history)
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Category ID (UUID)" format(uuid)
// @Success      200 {object} dto.MessageResponse "Category deleted successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid category ID format"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized - invalid or missing JWT token"
// @Failure      404 {object} dto.ErrorResponse "Category not found or does not belong to user's family"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /categories/{id} [delete]
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	categoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID format")
		return
	}

	// Category exists and belongs to family
	existing, err := h.categoryRepo.GetByID(r.Context(), categoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}
	if existing.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	if err := h.categoryRepo.Delete(r.Context(), categoryID); err != nil {
		writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", "Failed to delete category")
		return
	}

	writeMessage(w, http.StatusOK, "Category deleted successfully")
}

// Restore godoc
// @Summary      Restore deleted category
// @Description  Reactivate a soft-deleted category
// @Tags         Categories
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "Category ID (UUID)" format(uuid)
// @Success      200 {object} dto.SuccessResponse{data=dto.CategoryResponse} "Category restored successfully"
// @Failure      400 {object} dto.ErrorResponse "Invalid category ID format"
// @Failure      401 {object} dto.ErrorResponse "Unauthorized"
// @Failure      404 {object} dto.ErrorResponse "Category not found or already active"
// @Failure      500 {object} dto.ErrorResponse "Internal server error"
// @Router       /categories/{id}/restore [post]
func (h *CategoryHandler) Restore(w http.ResponseWriter, r *http.Request) {
	familyID, ok := middleware.GetFamilyID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Family context not found")
		return
	}

	categoryID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "Invalid category ID format")
		return
	}

	// Verify category belongs to family before restoring
	existing, err := h.categoryRepo.GetByIDIncludingInactive(r.Context(), categoryID)
	if err != nil {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}
	if existing.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	// Restore the category
	category, err := h.categoryRepo.Restore(r.Context(), categoryID)
	if err != nil {
		if err.Error() == "category not found or already active" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found or already active")
		} else {
			writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
		}
		return
	}

	writeSuccess(w, http.StatusOK, mapCategory(*category))
}

// Helper functions

func mapCategory(c sqlc.Category) dto.CategoryResponse {
	var parentID *uuid.UUID
	if c.ParentID.Valid {
		// Convert [16]byte to uuid.UUID
		id := uuid.UUID(c.ParentID.Bytes)
		parentID = &id
	}

	return dto.CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		ParentID:  parentID,
		IsActive:  c.IsActive,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func mapCategories(categories []sqlc.Category) []dto.CategoryResponse {
	result := make([]dto.CategoryResponse, len(categories))
	for i, c := range categories {
		result[i] = mapCategory(c)
	}
	return result
}

func writeErrorWithData(w http.ResponseWriter, status int, code, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
			"data":    data,
		},
	}); err != nil {
	}
}
