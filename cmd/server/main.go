package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DigitLock/expense-tracker/internal/api"
	"github.com/DigitLock/expense-tracker/internal/config"
	"github.com/DigitLock/expense-tracker/internal/database"
	"github.com/DigitLock/expense-tracker/internal/repository"
)

func main() {
	ctx := context.Background()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// Validate required config
	if cfg.JWT.Secret == "" {
		log.Println("⚠️  WARNING: JWT_SECRET not set, using insecure default for development")
		cfg.JWT.Secret = "dev-secret-change-in-production"
	}

	// Connect to database (используем существующий database.New)
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("✅ Connected to database")

	// Initialize repositories
	repos := repository.New(db.Pool)
	log.Println("✅ Repositories initialized")

	// Setup router
	router := api.NewRouter(cfg, db.Pool, repos)

	// Create and start server
	server := api.NewServer(&cfg.Server, router)

	// Graceful shutdown
	go func() {
		if err := server.Start(); err != nil && err.Error() != "http: Server closed" {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("❌ Server forced to shutdown: %v", err)
	}

	log.Println("✅ Server stopped gracefully")
}
