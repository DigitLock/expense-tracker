package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/DigitLock/expense-tracker/Documentation/swagger" // Swagger docs
	"github.com/DigitLock/expense-tracker/internal/api"
	"github.com/DigitLock/expense-tracker/internal/auth"
	"github.com/DigitLock/expense-tracker/internal/config"
	"github.com/DigitLock/expense-tracker/internal/currency"
	"github.com/DigitLock/expense-tracker/internal/database"
	grpcserver "github.com/DigitLock/expense-tracker/internal/grpc"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

// @title           Expense Tracker API
// @version         1.0
// @description     Personal and family finance management system with multi-currency support and automatic balance calculation
// @termsOfService  http://swagger.io/terms/

// @contact.name   Igor Kudinov
// @contact.email  digitlock@proton.me

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// @tag.name Authentication
// @tag.description User authentication and authorization

// @tag.name Accounts
// @tag.description Financial accounts management

// @tag.name Categories
// @tag.description Transaction categories management

// @tag.name Transactions
// @tag.description Income and expense transactions

// @tag.name Reports
// @tag.description Financial reports and analytics

// @tag.name Currencies
// @tag.description Currency exchange rates and conversion

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate required config
	if cfg.JWT.Secret == "" {
		log.Println("WARNING: JWT_SECRET not set, using insecure default for development")
		cfg.JWT.Secret = "dev-secret-change-in-production"
	}

	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Connected to database")

	repos := repository.New(db.Pool)
	log.Println("Repositories initialized")

	// appCtx is cancelled on shutdown to stop background workers (currency sync).
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// Background currency rate sync. A failure to build the client (e.g. bad
	// address) must not crash the server — log and run without the scheduler.
	// The Syncer is shared with the API (forced-sync endpoint); it is nil when
	// the client could not be built.
	var currencyClient *currency.Client
	var currencySyncer *currency.Syncer
	var currencyScheduler *currency.Scheduler
	if currencyClient, err = currency.NewClient(cfg.Currency.ServiceAddr, cfg.Currency.SyncTimeout); err != nil {
		log.Printf("WARNING: currency client init failed, rate sync disabled: %v", err)
		currencyClient = nil
	} else {
		currencySyncer = currency.NewSyncer(currencyClient, repos.ExchangeRates)
		currencyScheduler = currency.NewScheduler(currencySyncer, cfg.Currency.SyncInterval)
		currencyScheduler.Start(appCtx)
		log.Printf("Currency rate sync scheduled every %s (service %s)", cfg.Currency.SyncInterval, cfg.Currency.ServiceAddr)
	}

	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpirationHours)

	grpcSrv := grpcserver.NewServer(repos, jwtService, cfg.Server.GRPCPort)

	log.Printf("gRPC server starting on port %d\n", cfg.Server.GRPCPort)
	go func() {
		if err := grpcSrv.Start(); err != nil {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()

	router := api.NewRouter(cfg, db.Pool, repos, currencySyncer)

	server := api.NewServer(&cfg.Server, router)

	// Log Swagger UI URL
	log.Printf("Swagger UI: http://localhost:%d/swagger/index.html\n", cfg.Server.Port)
	log.Printf("Server starting on port %d\n", cfg.Server.Port)

	go func() {
		if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Stop the currency scheduler goroutine before tearing down its dependencies,
	// and wait for it to exit so we don't close the client mid-sync.
	appCancel()
	if currencyScheduler != nil {
		currencyScheduler.Wait()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	grpcSrv.Stop()

	// Close the currency gRPC client after the scheduler has been signalled to stop.
	if currencyClient != nil {
		_ = currencyClient.Close()
	}

	log.Println("Server stopped gracefully")
}
