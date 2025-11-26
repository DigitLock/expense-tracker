package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

// CategoryRepository handles category data operations
type CategoryRepository struct {
	queries *sqlc.Queries
}

// NewCategoryRepository creates a new CategoryRepository
func NewCategoryRepository(queries *sqlc.Queries) *CategoryRepository {
	return &CategoryRepository{queries: queries}
}

// GetByID retrieves a category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Category, error) {
	return r.queries.GetCategory(ctx, id)
}

// ListByFamily retrieves all categories in a family
func (r *CategoryRepository) ListByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Category, error) {
	return r.queries.ListCategoriesByFamily(ctx, familyID)
}

// ListByType retrieves categories by type (income/expense)
func (r *CategoryRepository) ListByType(ctx context.Context, familyID uuid.UUID, categoryType string) ([]sqlc.Category, error) {
	return r.queries.ListCategoriesByType(ctx, sqlc.ListCategoriesByTypeParams{
		FamilyID: familyID,
		Type:     categoryType,
	})
}

// ListRootCategories retrieves top-level categories (no parent)
func (r *CategoryRepository) ListRootCategories(ctx context.Context, familyID uuid.UUID) ([]sqlc.Category, error) {
	return r.queries.ListRootCategories(ctx, familyID)
}

// ListChildCategories retrieves child categories of a parent
func (r *CategoryRepository) ListChildCategories(ctx context.Context, parentID uuid.UUID) ([]sqlc.Category, error) {
	return r.queries.ListChildCategories(ctx, pgtype.UUID{Bytes: parentID, Valid: true})
}

// CreateCategoryInput contains data for creating a new category
type CreateCategoryInput struct {
	FamilyID uuid.UUID
	Name     string
	Type     string     // income, expense
	ParentID *uuid.UUID // nil for root categories
}

// Create creates a new category
func (r *CategoryRepository) Create(ctx context.Context, input CreateCategoryInput) (sqlc.Category, error) {
	var parentID pgtype.UUID
	if input.ParentID != nil {
		parentID = pgtype.UUID{Bytes: *input.ParentID, Valid: true}
	}

	return r.queries.CreateCategory(ctx, sqlc.CreateCategoryParams{
		ID:       uuid.New(),
		FamilyID: input.FamilyID,
		Name:     input.Name,
		Type:     input.Type,
		ParentID: parentID,
	})
}

// Update updates a category
func (r *CategoryRepository) Update(ctx context.Context, id uuid.UUID, name string, parentID *uuid.UUID) (sqlc.Category, error) {
	var pgParentID pgtype.UUID
	if parentID != nil {
		pgParentID = pgtype.UUID{Bytes: *parentID, Valid: true}
	}

	return r.queries.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:       id,
		Name:     name,
		ParentID: pgParentID,
	})
}

// Delete soft-deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteCategory(ctx, id)
}
