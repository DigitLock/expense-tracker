package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

// Repositories contains all repository instances
type Repositories struct {
	Families      *FamilyRepository
	Users         *UserRepository
	Accounts      *AccountRepository
	Categories    *CategoryRepository
	Transactions  *TransactionRepository
	ExchangeRates *ExchangeRateRepository

	// Keep reference to pool for transactions
	pool *pgxpool.Pool
}

// New creates all repositories with shared database connection
func New(pool *pgxpool.Pool) *Repositories {
	queries := sqlc.New(pool)

	return &Repositories{
		Families:      NewFamilyRepository(queries),
		Users:         NewUserRepository(queries),
		Accounts:      NewAccountRepository(queries),
		Categories:    NewCategoryRepository(queries),
		Transactions:  NewTransactionRepository(queries, pool),
		ExchangeRates: NewExchangeRateRepository(queries),
		pool:          pool,
	}
}
