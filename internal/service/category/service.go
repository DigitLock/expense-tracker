// Package category holds the application-layer service for the Category domain.
// All category business rules (hierarchy depth, parent-type match, cycle
// prevention, delete guards) live here; REST and gRPC are thin adapters.
package category

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

// Repo is the narrow set of category repository methods the service needs.
type Repo interface {
	GetByID(ctx context.Context, id uuid.UUID) (sqlc.Category, error)
	GetByIDIncludingInactive(ctx context.Context, id uuid.UUID) (sqlc.Category, error)
	ListByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Category, error)
	ListAllByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Category, error)
	ListByType(ctx context.Context, familyID uuid.UUID, categoryType string) ([]sqlc.Category, error)
	Create(ctx context.Context, input repository.CreateCategoryInput) (sqlc.Category, error)
	Update(ctx context.Context, input repository.UpdateCategoryInput) (sqlc.Category, error)
	Delete(ctx context.Context, id uuid.UUID) error
	HasChildren(ctx context.Context, parentID uuid.UUID) (bool, error)
	HasTransactions(ctx context.Context, id uuid.UUID) (bool, error)
}

type Service struct {
	repo Repo
}

func New(repo Repo) *Service {
	return &Service{repo: repo}
}

// Input DTOs (transport-agnostic). ParentID is a *string so "absent" (nil) is
// distinct from any value.
type CreateCategoryInput struct {
	Name     string
	Type     string
	ParentID *string
}

type UpdateCategoryInput struct {
	Name     *string
	ParentID *string
	IsActive *bool
}

func validType(t string) bool { return t == "income" || t == "expense" }

// List returns the family's categories, optionally filtered by type or
// including inactive ones (type filter takes precedence, matching prior REST).
func (s *Service) List(ctx context.Context, familyID uuid.UUID, typeFilter string, includeInactive bool) ([]domain.Category, error) {
	var rows []sqlc.Category
	var err error
	switch {
	case typeFilter == "income" || typeFilter == "expense":
		rows, err = s.repo.ListByType(ctx, familyID, typeFilter)
	case includeInactive:
		rows, err = s.repo.ListAllByFamily(ctx, familyID)
	default:
		rows, err = s.repo.ListByFamily(ctx, familyID)
	}
	if err != nil {
		return nil, err
	}
	cats := make([]domain.Category, len(rows))
	for i, r := range rows {
		cats[i] = toDomain(r)
	}
	return cats, nil
}

// Create validates and creates a category (optionally under a parent).
func (s *Service) Create(ctx context.Context, familyID uuid.UUID, in CreateCategoryInput) (domain.Category, error) {
	if in.Name == "" {
		return domain.Category{}, domain.Errorf(domain.ErrInvalidArgument, "name is required")
	}
	if !validType(in.Type) {
		return domain.Category{}, domain.Errorf(domain.ErrInvalidArgument, "invalid category type")
	}

	var parentID *uuid.UUID
	if in.ParentID != nil {
		pid, err := uuid.Parse(*in.ParentID)
		if err != nil {
			return domain.Category{}, domain.Errorf(domain.ErrInvalidArgument, "invalid parent_id")
		}
		if _, err := s.requireParent(ctx, familyID, pid, in.Type); err != nil {
			return domain.Category{}, err
		}
		parentID = &pid
	}

	row, err := s.repo.Create(ctx, repository.CreateCategoryInput{
		FamilyID: familyID,
		Name:     in.Name,
		Type:     in.Type,
		ParentID: parentID,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Category{}, domain.Errorf(domain.ErrAlreadyExists, "category with this name already exists")
		}
		return domain.Category{}, err
	}
	return toDomain(row), nil
}

// Update applies a partial update, enforcing hierarchy/cycle rules when the
// parent changes.
func (s *Service) Update(ctx context.Context, familyID, id uuid.UUID, in UpdateCategoryInput) (domain.Category, error) {
	existing, err := s.repo.GetByIDIncludingInactive(ctx, id)
	if err != nil {
		return domain.Category{}, domain.Errorf(domain.ErrNotFound, "category not found")
	}
	if existing.FamilyID != familyID {
		return domain.Category{}, domain.Errorf(domain.ErrPermissionDenied, "access denied")
	}
	if in.Name != nil && *in.Name == "" {
		return domain.Category{}, domain.Errorf(domain.ErrInvalidArgument, "name cannot be empty")
	}

	input := repository.UpdateCategoryInput{
		ID:       id,
		FamilyID: familyID,
		Name:     in.Name,
		IsActive: in.IsActive,
	}

	if in.ParentID != nil {
		pid, err := uuid.Parse(*in.ParentID)
		if err != nil {
			return domain.Category{}, domain.Errorf(domain.ErrInvalidArgument, "invalid parent_id")
		}
		// Cycle prevention: cannot be its own parent.
		if pid == id {
			return domain.Category{}, domain.Errorf(domain.ErrInvalidArgument, "category cannot be its own parent")
		}
		if _, err := s.requireParent(ctx, familyID, pid, existing.Type); err != nil {
			return domain.Category{}, err
		}
		// Moving a category that itself has children under a parent would create a
		// 3rd level (and covers making a subcategory the parent → cycle).
		hasChildren, err := s.repo.HasChildren(ctx, id)
		if err != nil {
			return domain.Category{}, err
		}
		if hasChildren {
			return domain.Category{}, domain.Errorf(domain.ErrInvalidArgument, "cannot move a parent category under another category (max 2 levels)")
		}
		input.ParentID = &pid
	}

	row, err := s.repo.Update(ctx, input)
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Category{}, domain.Errorf(domain.ErrAlreadyExists, "category with this name already exists")
		}
		return domain.Category{}, err
	}
	return toDomain(row), nil
}

// Delete soft-deletes a category after ownership and dependency checks.
func (s *Service) Delete(ctx context.Context, familyID, id uuid.UUID) error {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return domain.Errorf(domain.ErrNotFound, "category not found")
	}
	if existing.FamilyID != familyID {
		return domain.Errorf(domain.ErrPermissionDenied, "access denied")
	}

	hasChildren, err := s.repo.HasChildren(ctx, id)
	if err != nil {
		return err
	}
	if hasChildren {
		return domain.Errorf(domain.ErrFailedPrecondition, "category has active subcategories")
	}

	hasTx, err := s.repo.HasTransactions(ctx, id)
	if err != nil {
		return err
	}
	if hasTx {
		return domain.Errorf(domain.ErrFailedPrecondition, "cannot delete category with existing transactions")
	}

	return s.repo.Delete(ctx, id)
}

// --- helpers ---

// requireParent validates that the parent exists, belongs to the family, has a
// matching type, and is itself a root category (enforcing the 2-level limit).
func (s *Service) requireParent(ctx context.Context, familyID, parentID uuid.UUID, childType string) (sqlc.Category, error) {
	parent, err := s.repo.GetByID(ctx, parentID)
	if err != nil {
		return sqlc.Category{}, domain.Errorf(domain.ErrNotFound, "parent category not found")
	}
	if parent.FamilyID != familyID {
		return sqlc.Category{}, domain.Errorf(domain.ErrPermissionDenied, "access denied")
	}
	if parent.Type != childType {
		return sqlc.Category{}, domain.Errorf(domain.ErrInvalidArgument, "parent category type does not match")
	}
	if parent.ParentID.Valid {
		return sqlc.Category{}, domain.Errorf(domain.ErrInvalidArgument, "maximum 2 levels of hierarchy allowed (parent must be top-level)")
	}
	return parent, nil
}

func toDomain(c sqlc.Category) domain.Category {
	var parentID *uuid.UUID
	if c.ParentID.Valid {
		id := uuid.UUID(c.ParentID.Bytes)
		parentID = &id
	}
	return domain.Category{
		ID:        c.ID,
		Name:      c.Name,
		Type:      c.Type,
		ParentID:  parentID,
		IsActive:  c.IsActive,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
