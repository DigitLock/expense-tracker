package api

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DigitLock/expense-tracker/internal/api/handlers"
	"github.com/DigitLock/expense-tracker/internal/api/middleware"
	"github.com/DigitLock/expense-tracker/internal/auth"
	"github.com/DigitLock/expense-tracker/internal/config"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

func NewRouter(cfg *config.Config, db *pgxpool.Pool, repos *repository.Repositories) *chi.Mux {
	r := chi.NewRouter()

	// --- JWT Service ---
	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpirationHours)

	// --- Global Middleware ---
	r.Use(middleware.Recovery)
	r.Use(middleware.Logging)
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   cfg.Server.AllowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// --- Handlers ---
	healthHandler := handlers.NewHealthHandler(db)
	authHandler := handlers.NewAuthHandler(repos.Users, jwtService)
	accountHandler := handlers.NewAccountHandler(repos.Accounts)
	categoryHandler := handlers.NewCategoryHandler(repos.Categories)
	transactionHandler := handlers.NewTransactionHandler(
		repos.Transactions,
		repos.Accounts,
		repos.Categories,
		repos.Users,
	)
	reportHandler := handlers.NewReportHandler(
		repos.Transactions,
		repos.Accounts,
		repos.Categories,
	)
	currencyHandler := handlers.NewCurrencyHandler(repos.ExchangeRates)

	// --- Public Routes (no auth required) ---
	r.Get("/health", healthHandler.Health)
	r.Get("/ready", healthHandler.Ready)

	// --- API v1 Routes ---
	r.Route("/api/v1", func(r chi.Router) {
		// Public routes (no auth)
		r.Group(func(r chi.Router) {
			r.Post("/auth/login", authHandler.Login)
		})

		// Protected routes (require JWT)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(jwtService))

			// Accounts
			r.Route("/accounts", func(r chi.Router) {
				r.Get("/", accountHandler.List)
				r.Post("/", accountHandler.Create)
				r.Get("/{id}", accountHandler.Get)
				r.Patch("/{id}", accountHandler.Update)
				r.Delete("/{id}", accountHandler.Delete)
				r.Get("/{id}/balance", accountHandler.GetBalance)
			})

			// Categories
			r.Route("/categories", func(r chi.Router) {
				r.Get("/", categoryHandler.List)
				r.Post("/", categoryHandler.Create)
				r.Get("/{id}", categoryHandler.Get)
				r.Patch("/{id}", categoryHandler.Update)
				r.Delete("/{id}", categoryHandler.Delete)
			})

			// Transactions
			r.Route("/transactions", func(r chi.Router) {
				r.Get("/", transactionHandler.List)
				r.Post("/", transactionHandler.Create)
				r.Get("/{id}", transactionHandler.Get)
				r.Patch("/{id}", transactionHandler.Update)
				r.Delete("/{id}", transactionHandler.Delete)
			})

			// Reports
			r.Route("/reports", func(r chi.Router) {
				r.Get("/spending-by-category", reportHandler.SpendingByCategory)
				r.Get("/monthly-summary", reportHandler.MonthlySummary)
			})

			// Currencies
			r.Route("/currencies", func(r chi.Router) {
				r.Get("/rates", currencyHandler.GetRates)
				r.Get("/convert", currencyHandler.Convert)
			})
		})
	})

	return r
}
