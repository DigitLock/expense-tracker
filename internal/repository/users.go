package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

// Common errors
var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailExists        = errors.New("email already exists")
)

// UserRepository handles user data operations
type UserRepository struct {
	queries *sqlc.Queries
	pool    *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(queries *sqlc.Queries, pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{queries: queries, pool: pool}
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return r.queries.GetUser(ctx, id)
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

// ListByFamily retrieves all users in a family
func (r *UserRepository) ListByFamily(ctx context.Context, familyID uuid.UUID) ([]sqlc.User, error) {
	return r.queries.ListUsersByFamily(ctx, familyID)
}

// CreateUserInput contains data for creating a new user
type CreateUserInput struct {
	FamilyID uuid.UUID
	Email    string
	Name     string
	Password string // Plain text - will be hashed
}

// Create creates a new user with hashed password
func (r *UserRepository) Create(ctx context.Context, input CreateUserInput) (sqlc.User, error) {
	// Hash password with bcrypt (cost 12)
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return sqlc.User{}, err
	}

	return r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           uuid.New(),
		FamilyID:     input.FamilyID,
		Email:        input.Email,
		Name:         input.Name,
		PasswordHash: string(hashedPassword),
		Role:         "member",
	})
}

// RegisterInput contains data for self-registration (new family + owner user)
type RegisterInput struct {
	Email      string
	Password   string // plain text - hashed before insert
	Name       string
	FamilyName string
}

// Register creates a new family, its owner user, and seeds default categories
// atomically in a single transaction. On a duplicate email (unique violation)
// the whole transaction is rolled back and ErrEmailExists is returned.
func (r *UserRepository) Register(ctx context.Context, input RegisterInput) (sqlc.User, error) {
	// Hash the password before opening the transaction (bcrypt is CPU-bound).
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 12)
	if err != nil {
		return sqlc.User{}, err
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.User{}, err
	}
	defer tx.Rollback(ctx) // no-op after a successful commit

	q := r.queries.WithTx(tx)

	family, err := q.CreateFamily(ctx, sqlc.CreateFamilyParams{
		ID:           uuid.New(),
		Name:         input.FamilyName,
		BaseCurrency: "RSD",
	})
	if err != nil {
		return sqlc.User{}, err
	}

	user, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           uuid.New(),
		FamilyID:     family.ID,
		Email:        input.Email,
		Name:         input.Name,
		PasswordHash: string(hashedPassword),
		Role:         "owner",
	})
	if err != nil {
		// Duplicate email -> unique constraint violation (SQLSTATE 23505).
		// Detected here rather than via a pre-check so soft-deleted duplicates
		// (which GetUserByEmail filters out) are still caught.
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return sqlc.User{}, ErrEmailExists
		}
		return sqlc.User{}, err
	}

	if err := q.SeedDefaultCategories(ctx, family.ID); err != nil {
		return sqlc.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.User{}, err
	}

	return user, nil
}

// Update updates user profile (not password)
func (r *UserRepository) Update(ctx context.Context, id uuid.UUID, name, email string) (sqlc.User, error) {
	return r.queries.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:    id,
		Name:  name,
		Email: email,
	})
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		return err
	}

	return r.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: string(hashedPassword),
	})
}

// Delete soft-deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.queries.DeleteUser(ctx, id)
}

// Authenticate verifies email and password, returns user if valid
func (r *UserRepository) Authenticate(ctx context.Context, email, password string) (sqlc.User, error) {
	user, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return sqlc.User{}, ErrInvalidCredentials
	}

	// Compare password with hash
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return sqlc.User{}, ErrInvalidCredentials
	}

	return user, nil
}
