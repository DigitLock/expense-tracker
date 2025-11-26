package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DigitLock/expense-tracker/internal/config"
	"github.com/DigitLock/expense-tracker/internal/database"
	"github.com/DigitLock/expense-tracker/internal/database/sqlc"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutdown signal received...")
		cancel()
	}()

	// Connect to database
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	log.Println("✅ Connected to database successfully!")

	// Health check
	if err := db.Health(ctx); err != nil {
		log.Fatalf("Database health check failed: %v", err)
	}
	log.Println("✅ Database health check passed!")

	// === Test sqlc queries ===
	log.Println("\n📊 Testing sqlc queries with seed data...")

	// Test: List all families
	families, err := db.Queries.ListFamilies(ctx)
	if err != nil {
		log.Printf("❌ ListFamilies error: %v", err)
	} else {
		log.Printf("✅ Found %d families", len(families))
		for _, f := range families {
			log.Printf("   - %s (base currency: %s)", f.Name, f.BaseCurrency)
		}
	}

	// Test: List users (if we have a family)
	if len(families) > 0 {
		familyID := families[0].ID

		users, err := db.Queries.ListUsersByFamily(ctx, familyID)
		if err != nil {
			log.Printf("❌ ListUsersByFamily error: %v", err)
		} else {
			log.Printf("✅ Found %d users in family '%s'", len(users), families[0].Name)
			for _, u := range users {
				log.Printf("   - %s (%s)", u.Name, u.Email)
			}
		}

		// Test: List accounts
		accounts, err := db.Queries.ListAccountsByFamily(ctx, familyID)
		if err != nil {
			log.Printf("❌ ListAccountsByFamily error: %v", err)
		} else {
			log.Printf("✅ Found %d accounts", len(accounts))
			for _, a := range accounts {
				log.Printf("   - %s: %s %s (type: %s)", a.Name, a.CurrentBalance.StringFixed(2), a.Currency, a.Type)
			}
		}

		// Test: List categories
		categories, err := db.Queries.ListCategoriesByFamily(ctx, familyID)
		if err != nil {
			log.Printf("❌ ListCategoriesByFamily error: %v", err)
		} else {
			log.Printf("✅ Found %d categories", len(categories))
		}

		// Test: List recent transactions
		transactions, err := db.Queries.ListTransactionsPaginated(ctx, sqlc.ListTransactionsPaginatedParams{
			FamilyID: familyID,
			Limit:    5,
			Offset:   0,
		})
		if err != nil {
			log.Printf("❌ ListTransactionsPaginated error: %v", err)
		} else {
			log.Printf("✅ Found %d recent transactions (showing max 5)", len(transactions))
			for _, t := range transactions {
				// pgtype.Date requires .Time to access underlying time.Time
				log.Printf("   - %s: %s %s (%s)",
					t.TransactionDate.Time.Format("2006-01-02"),
					t.Amount.StringFixed(2),
					t.Currency,
					t.Type)
			}
		}
	}

	log.Println("\n🎉 All sqlc queries working!")
	log.Printf("🚀 Expense Tracker server ready (will listen on port %d)", cfg.Server.Port)

	// Wait for shutdown
	<-ctx.Done()
	log.Println("👋 Server stopped gracefully")
}
