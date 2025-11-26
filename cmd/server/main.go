package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/DigitLock/expense-tracker/internal/config"
	"github.com/DigitLock/expense-tracker/internal/database"
	"github.com/DigitLock/expense-tracker/internal/repository"
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

	log.Println("Connected to database successfully!")

	// Initialize repositories
	repos := repository.New(db.Pool)
	log.Println("Repositories initialized!")

	// === Test Repository Layer ===
	log.Println("\n📊 Testing Repository layer with seed data...")

	// Test: List all families
	families, err := repos.Families.List(ctx)
	if err != nil {
		log.Printf("Families.List error: %v", err)
	} else {
		log.Printf("Found %d families", len(families))
		for _, f := range families {
			log.Printf("   - %s (base currency: %s)", f.Name, f.BaseCurrency)
		}
	}

	// Test with first family
	if len(families) > 0 {
		familyID := families[0].ID

		// Test: List users
		users, err := repos.Users.ListByFamily(ctx, familyID)
		if err != nil {
			log.Printf("Users.ListByFamily error: %v", err)
		} else {
			log.Printf("Found %d users in family '%s'", len(users), families[0].Name)
			for _, u := range users {
				log.Printf("   - %s (%s)", u.Name, u.Email)
			}
		}

		// Test: Authenticate user (with demo credentials)
		user, err := repos.Users.Authenticate(ctx, "demo@example.com", "Demo123!")
		if err != nil {
			log.Printf("Users.Authenticate error: %v", err)
		} else {
			log.Printf("Authentication successful for: %s", user.Name)
		}

		// Test: List accounts
		accounts, err := repos.Accounts.ListByFamily(ctx, familyID)
		if err != nil {
			log.Printf("Accounts.ListByFamily error: %v", err)
		} else {
			log.Printf("Found %d accounts", len(accounts))
			for _, a := range accounts {
				log.Printf("   - %s: %s %s (type: %s)", a.Name, a.CurrentBalance.StringFixed(2), a.Currency, a.Type)
			}
		}

		// Test: Get total balance
		totalBalance, accountCount, err := repos.Accounts.GetTotalBalanceByFamily(ctx, familyID)
		if err != nil {
			log.Printf("Accounts.GetTotalBalanceByFamily error: %v", err)
		} else {
			log.Printf("Total balance across %d accounts: %s", accountCount, totalBalance.StringFixed(2))
		}

		// Test: List categories
		categories, err := repos.Categories.ListByFamily(ctx, familyID)
		if err != nil {
			log.Printf("Categories.ListByFamily error: %v", err)
		} else {
			log.Printf("Found %d categories", len(categories))
		}

		// Test: List transactions with pagination
		transactions, err := repos.Transactions.ListPaginated(ctx, familyID, 5, 0)
		if err != nil {
			log.Printf("Transactions.ListPaginated error: %v", err)
		} else {
			log.Printf("Found %d recent transactions (showing max 5)", len(transactions))
			for _, t := range transactions {
				log.Printf("   - %s: %s %s (%s)",
					t.TransactionDate.Time.Format("2006-01-02"),
					t.Amount.StringFixed(2),
					t.Currency,
					t.Type)
			}
		}

		// Test: Get exchange rate
		rate, err := repos.ExchangeRates.GetLatestRate(ctx, "EUR", "RSD", families[0].CreatedAt)
		if err != nil {
			log.Printf("ExchangeRates.GetLatestRate: %v (may not have rates in seed data)", err)
		} else {
			log.Printf("EUR/RSD rate: %s", rate.Rate.StringFixed(4))
		}
	}

	log.Println("\nRepository layer working!")
	log.Printf("Expense Tracker server ready (will listen on port %d)", cfg.Server.Port)

	// Wait for shutdown
	<-ctx.Done()
	log.Println("Server stopped gracefully")
}
