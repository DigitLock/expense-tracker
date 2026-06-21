// Package account holds the application-layer service for the Account domain.
// All Account business rules live here; REST and gRPC are thin adapters over it.
package account

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
	"github.com/DigitLock/expense-tracker/internal/domain"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

// Repo is the narrow set of repository methods the service depends on.
// *repository.AccountRepository satisfies it.
type Repo interface {
	ListByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Account, error)
	ListAllByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.Account, error)
	GetByIDIncludingInactive(ctx context.Context, id uuid.UUID) (sqlc.Account, error)
	Create(ctx context.Context, input repository.CreateAccountInput) (sqlc.Account, error)
	Update(ctx context.Context, input repository.UpdateAccountInput) (sqlc.Account, error)
	Delete(ctx context.Context, id uuid.UUID) error
	HasTransactions(ctx context.Context, id uuid.UUID) (bool, error)
}

// Service implements Account use cases.
type Service struct {
	repo Repo
}

// New creates an account Service.
func New(repo Repo) *Service {
	return &Service{repo: repo}
}

// Input DTOs (transport-agnostic).

type CreateAccountInput struct {
	Name           string
	Type           string
	Currency       string
	InitialBalance decimal.Decimal
	Description    string
}

type UpdateAccountInput struct {
	Name        *string
	Description *string
	IsActive    *bool
}

var (
	validTypes      = map[string]bool{"cash": true, "checking": true, "savings": true}
	validCurrencies = map[string]bool{"RSD": true, "EUR": true}
)

// List returns the family's accounts (optionally including inactive).
func (s *Service) List(ctx context.Context, familyID uuid.UUID, includeInactive bool) ([]domain.Account, error) {
	var rows []sqlc.Account
	var err error
	if includeInactive {
		rows, err = s.repo.ListAllByFamily(ctx, familyID)
	} else {
		rows, err = s.repo.ListByFamily(ctx, familyID)
	}
	if err != nil {
		return nil, err
	}

	accounts := make([]domain.Account, len(rows))
	for i, r := range rows {
		accounts[i] = toDomain(r)
	}
	return accounts, nil
}

// Create validates and creates an account for the family.
func (s *Service) Create(ctx context.Context, familyID uuid.UUID, in CreateAccountInput) (domain.Account, error) {
	if in.Name == "" {
		return domain.Account{}, domain.Errorf(domain.ErrInvalidArgument, "name is required")
	}
	if in.Type == "" || !validTypes[in.Type] {
		return domain.Account{}, domain.Errorf(domain.ErrInvalidArgument, "invalid account type")
	}
	if in.Currency == "" || !validCurrencies[in.Currency] {
		return domain.Account{}, domain.Errorf(domain.ErrInvalidArgument, "invalid currency")
	}
	if in.InitialBalance.IsNegative() {
		return domain.Account{}, domain.Errorf(domain.ErrInvalidArgument, "initial balance cannot be negative")
	}

	row, err := s.repo.Create(ctx, repository.CreateAccountInput{
		FamilyID:       familyID,
		Name:           in.Name,
		Type:           in.Type,
		Currency:       in.Currency,
		InitialBalance: in.InitialBalance,
		Description:    textValue(in.Description),
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Account{}, domain.Errorf(domain.ErrAlreadyExists, "account with this name already exists")
		}
		return domain.Account{}, err
	}
	return toDomain(row), nil
}

// Update applies a partial update after verifying ownership.
func (s *Service) Update(ctx context.Context, familyID, id uuid.UUID, in UpdateAccountInput) (domain.Account, error) {
	existing, err := s.repo.GetByIDIncludingInactive(ctx, id)
	if err != nil {
		return domain.Account{}, domain.Errorf(domain.ErrNotFound, "account not found")
	}
	if existing.FamilyID != familyID {
		return domain.Account{}, domain.Errorf(domain.ErrPermissionDenied, "access denied")
	}
	if in.Name != nil && *in.Name == "" {
		return domain.Account{}, domain.Errorf(domain.ErrInvalidArgument, "name cannot be empty")
	}

	row, err := s.repo.Update(ctx, repository.UpdateAccountInput{
		ID:          id,
		FamilyID:    familyID,
		Name:        in.Name,
		Description: textPtr(in.Description),
		IsActive:    in.IsActive,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return domain.Account{}, domain.Errorf(domain.ErrAlreadyExists, "account with this name already exists")
		}
		return domain.Account{}, err
	}
	return toDomain(row), nil
}

// Delete soft-deletes an account after ownership and precondition checks.
func (s *Service) Delete(ctx context.Context, familyID, id uuid.UUID) error {
	existing, err := s.repo.GetByIDIncludingInactive(ctx, id)
	if err != nil {
		return domain.Errorf(domain.ErrNotFound, "account not found")
	}
	if existing.FamilyID != familyID {
		return domain.Errorf(domain.ErrPermissionDenied, "access denied")
	}
	if !existing.IsActive {
		return domain.Errorf(domain.ErrFailedPrecondition, "account is already inactive")
	}

	hasTx, err := s.repo.HasTransactions(ctx, id)
	if err != nil {
		return err
	}
	if hasTx {
		return domain.Errorf(domain.ErrFailedPrecondition, "cannot delete account with existing transactions")
	}

	return s.repo.Delete(ctx, id)
}

// --- helpers ---

func toDomain(a sqlc.Account) domain.Account {
	desc := ""
	if a.Description.Valid {
		desc = a.Description.String
	}
	return domain.Account{
		ID:             a.ID,
		Name:           a.Name,
		Type:           a.Type,
		Currency:       a.Currency,
		InitialBalance: a.InitialBalance,
		CurrentBalance: a.CurrentBalance,
		Description:    desc,
		IsActive:       a.IsActive,
		CreatedAt:      a.CreatedAt,
		UpdatedAt:      a.UpdatedAt,
	}
}

// textValue maps a create-time description: empty string => SQL NULL.
func textValue(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: s, Valid: true}
}

// textPtr maps an update-time *description: nil => leave unchanged (NULL marker),
// non-nil => set to the given value (including empty string).
func textPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
