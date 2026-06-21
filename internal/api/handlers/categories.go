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
	"github.com/DigitLock/expense-tracker/internal/service/category"
)

// CategoryHandler is a thin REST adapter over the category service. The repo is
// retained for read-only extras (Get, Restore) and the inactive-duplicate
// restore-offer on Create (REST-only UX), which are outside the gRPC surface.
type CategoryHandler struct {
	svc          *category.Service
	categoryRepo *repository.CategoryRepository
}

func NewCategoryHandler(categoryRepo *repository.CategoryRepository) *CategoryHandler {
	return &CategoryHandler{
		svc:          category.New(categoryRepo),
		categoryRepo: categoryRepo,
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

	categories, err := h.svc.List(r.Context(), familyID, typeFilter, includeInactive)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	resp := make([]dto.CategoryResponse, len(categories))
	for i, c := range categories {
		resp[i] = categoryToDTO(c)
	}
	writeSuccess(w, http.StatusOK, dto.CategoryListResponse{Categories: resp})
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
// @Failure      409 {object} dto.ErrorResponse "Category already exists (active or restorable inactive)"
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

	// REST-only UX: if a matching inactive category exists, offer to restore it
	// instead of creating a duplicate.
	if inactive, err := h.categoryRepo.FindInactiveByName(r.Context(), familyID, req.Name, req.Type, req.ParentID); err == nil && inactive != nil {
		var parentID *uuid.UUID
		if inactive.ParentID.Valid {
			id := uuid.UUID(inactive.ParentID.Bytes)
			parentID = &id
		}
		writeErrorWithData(w, http.StatusConflict, "CATEGORY_INACTIVE_EXISTS",
			"A category with this name was previously deleted. Would you like to restore it?",
			map[string]interface{}{
				"inactive_category_id": inactive.ID,
				"name":                 inactive.Name,
				"type":                 inactive.Type,
				"parent_id":            parentID,
				"deleted_at":           inactive.UpdatedAt,
			})
		return
	}

	cat, err := h.svc.Create(r.Context(), familyID, category.CreateCategoryInput{
		Name:     req.Name,
		Type:     req.Type,
		ParentID: uuidPtrToStr(req.ParentID),
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusCreated, categoryToDTO(cat))
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

	cat, err := h.categoryRepo.GetByID(r.Context(), categoryID)
	if err != nil || cat.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	writeSuccess(w, http.StatusOK, mapCategory(cat))
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
// @Failure      404 {object} dto.ErrorResponse "Category not found"
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

	var req dto.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Invalid request body")
		return
	}

	cat, err := h.svc.Update(r.Context(), familyID, categoryID, category.UpdateCategoryInput{
		Name:     req.Name,
		ParentID: uuidPtrToStr(req.ParentID),
		IsActive: req.IsActive,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, categoryToDTO(cat))
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
// @Failure      404 {object} dto.ErrorResponse "Category not found"
// @Failure      409 {object} dto.ErrorResponse "Category has subcategories or transactions"
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

	if err := h.svc.Delete(r.Context(), familyID, categoryID); err != nil {
		writeDomainError(w, err)
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

	existing, err := h.categoryRepo.GetByIDIncludingInactive(r.Context(), categoryID)
	if err != nil || existing.FamilyID != familyID {
		writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found")
		return
	}

	cat, err := h.categoryRepo.Restore(r.Context(), categoryID)
	if err != nil {
		if err.Error() == "category not found or already active" {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Category not found or already active")
		} else {
			writeError(w, http.StatusInternalServerError, "DATABASE_ERROR", err.Error())
		}
		return
	}

	writeSuccess(w, http.StatusOK, mapCategory(*cat))
}

// Helper functions

func uuidPtrToStr(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// categoryToDTO maps a service-layer domain.Category to the REST DTO.
func categoryToDTO(c domain.Category) dto.CategoryResponse {
	return dto.CategoryResponse{
		ID:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		ParentID:  c.ParentID,
		IsActive:  c.IsActive,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// mapCategory maps a repo sqlc.Category to the REST DTO (used by Get/Restore).
func mapCategory(c sqlc.Category) dto.CategoryResponse {
	var parentID *uuid.UUID
	if c.ParentID.Valid {
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

func writeErrorWithData(w http.ResponseWriter, status int, code, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
			"data":    data,
		},
	})
}
