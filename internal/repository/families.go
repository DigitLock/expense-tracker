package repository

import (
	"context"

	"github.com/google/uuid"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

// FamilyRepository handles family data operations
type FamilyRepository struct {
	queries *sqlc.Queries
}

// NewFamilyRepository creates a new FamilyRepository
func NewFamilyRepository(queries *sqlc.Queries) *FamilyRepository {
	return &FamilyRepository{queries: queries}
}

// GetByID retrieves a family by ID
func (r *FamilyRepository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.Family, error) {
	return r.queries.GetFamily(ctx, id)
}

// GetByName retrieves a family by name
func (r *FamilyRepository) GetByName(ctx context.Context, name string) (sqlc.Family, error) {
	return r.queries.GetFamilyByName(ctx, name)
}

// List retrieves all active families
func (r *FamilyRepository) List(ctx context.Context) ([]sqlc.Family, error) {
	return r.queries.ListFamilies(ctx)
}

// Create creates a new family
func (r *FamilyRepository) Create(ctx context.Context, name, baseCurrency string) (sqlc.Family, error) {
	return r.queries.CreateFamily(ctx, sqlc.CreateFamilyParams{
		ID:           uuid.New(),
		Name:         name,
		BaseCurrency: baseCurrency,
	})
}

// Update updates a family
func (r *FamilyRepository) Update(ctx context.Context, id uuid.UUID, name, baseCurrency string) (sqlc.Family, error) {
	return r.queries.UpdateFamily(ctx, sqlc.UpdateFamilyParams{
		ID:           id,
		Name:         name,
		BaseCurrency: baseCurrency,
	})
}

// Delete soft-deletes a family
func (r *FamilyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteFamily(ctx, id)
}
