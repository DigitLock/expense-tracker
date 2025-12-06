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
// @Summary List categories
// @Description Returns all categories for the authenticated user's family
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Param type query string false "Filter by type: income or expense"
// @Param include_inactive query bool false "Include inactive categories"
// @Success 200 {object} dto.SuccessResponse{data=dto.CategoryListResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/categories [get]
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
// @Summary Create category
// @Description Creates a new category for the authenticated user's family
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateCategoryRequest true "Category data"
// @Success 201 {object} dto.SuccessResponse{data=dto.CategoryResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Router /api/v1/categories [post]
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
// @Summary Get category
// @Description Returns a specific category by ID
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.CategoryResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/categories/{id} [get]
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

	// Check category belongs to user's family
	if category.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	writeSuccess(w, http.StatusOK, mapCategory(category))
}

// Update godoc
// @Summary Update category
// @Description Updates an existing category (partial update)
// @Tags categories
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param request body dto.UpdateCategoryRequest true "Category data"
// @Success 200 {object} dto.SuccessResponse{data=dto.CategoryResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/categories/{id} [patch]
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

	// Check category exists and belongs to family
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

	// Build update input
	input := repository.UpdateCategoryInput{ID: categoryID}
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
// @Summary Delete category
// @Description Soft deletes a category
// @Tags categories
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Success 200 {object} dto.MessageResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Router /api/v1/categories/{id} [delete]
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

	// Check category exists and belongs to family
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

// --- Helper functions ---

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
