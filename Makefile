# Load environment variables from .env (in repo root)
include .env
export

# Database connection string assembled from .env variables
DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)

# Dedicated integration test database (same host/user, separate DB)
TEST_DB_URL := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/expense_tracker_test?sslmode=$(DB_SSLMODE)

MIGRATIONS_DIR := database/migrations

.PHONY: run build migrate-up migrate-down migrate-version sqlc-generate test test-setup

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

# Apply migrations to the test database (@ hides the connection string)
test-setup:
	@migrate -path $(MIGRATIONS_DIR) -database "$(TEST_DB_URL)" up

# Run the Go test suite against the test database.
# -p 1 serializes package execution: the integration tests share one test
# database and truncate it, so they must not run concurrently.
test: test-setup
	TEST_DATABASE_URL="$(TEST_DB_URL)" go test ./... -v -p 1
