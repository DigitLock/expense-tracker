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

// GetByID retrieves an active category by ID
func (r *CategoryRepository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Category, error) {
	return r.queries.GetCategory(ctx, id)
}

// GetByIDIncludingInactive retrieves a category by ID (even if inactive)
func (r *CategoryRepository) GetByIDIncludingInactive(ctx context.Context, id uuid.UUID) (sqlc.Category, error) {
	return r.queries.GetCategoryIncludingInactive(ctx, id)
}

// ListByFamily retrieves all active categories in a family
func (r *CategoryRepository) ListByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Category, error) {
	return r.queries.ListCategoriesByFamily(ctx, familyID)
}

// ListAllByFamily retrieves all categories in a family (including inactive)
func (r *CategoryRepository) ListAllByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Category, error) {
	return r.queries.ListAllCategoriesByFamily(ctx, familyID)
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

// UpdateCategoryInput contains data for updating a category (partial update)
type UpdateCategoryInput struct {
	ID       uuid.UUID
	Name     *string
	ParentID *uuid.UUID // use special value to clear parent
	IsActive *bool
	// Special flag to indicate we want to clear parent_id (set to NULL)
	ClearParent bool
}

// Update updates category details (partial update)
func (r *CategoryRepository) Update(ctx context.Context, input UpdateCategoryInput) (sqlc.Category, error) {
	// First get current category (including inactive to allow reactivation)
	current, err := r.queries.GetCategoryIncludingInactive(ctx, input.ID)
	if err != nil {
		return sqlc.Category{}, err
	}

	// Apply updates (keep current values if not provided)
	name := current.Name
	if input.Name != nil {
		name = *input.Name
	}

	parentID := current.ParentID
	if input.ClearParent {
		parentID = pgtype.UUID{Valid: false}
	} else if input.ParentID != nil {
		parentID = pgtype.UUID{Bytes: *input.ParentID, Valid: true}
	}

	isActive := current.IsActive
	if input.IsActive != nil {
		isActive = *input.IsActive
	}

	return r.queries.UpdateCategory(ctx, sqlc.UpdateCategoryParams{
		ID:       input.ID,
		Name:     name,
		ParentID: parentID,
		IsActive: isActive,
	})
}

// Delete soft-deletes a category
func (r *CategoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteCategory(ctx, id)
}
