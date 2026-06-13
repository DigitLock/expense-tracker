# Load environment variables from .env (in repo root)
include .env
export

# Database connection string assembled from .env variables
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

MIGRATIONS_DIR := database/migrations

.PHONY: run build migrate-up migrate-down migrate-version sqlc-generate test

# Run the server from the repo root
run:
	go run ./cmd/server/

# Build the server binary
build:
	go build -o bin/server ./cmd/server/

# Apply all pending migrations
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" up

# Roll back the most recent migration
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" down 1

# Show current migration version and dirty state
migrate-version:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_URL)" version

# Regenerate sqlc code
sqlc-generate:
	sqlc generate

# Run the Go test suite
test:
	go test ./...
