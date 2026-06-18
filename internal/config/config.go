package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	Server   ServerConfig
	JWT      JWTConfig
	Currency CurrencyConfig
}

type DatabaseConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Port            int
	GRPCPort        int
	AllowedOrigins  []string
	ReadTimeout     int
	WriteTimeout    int
	ShutdownTimeout int
}

type JWTConfig struct {
	Secret          string
	ExpirationHours int
}

// CurrencyConfig holds settings for the Currency Rate Service gRPC client.
type CurrencyConfig struct {
	ServiceAddr  string        // host:port of the currency-rate-service gRPC endpoint
	SyncInterval time.Duration // how often rates are synced (used by the st4 scheduler)
	SyncTimeout  time.Duration // per-call timeout for a rate fetch
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		d.User, d.Password, d.Host, d.Port, d.DBName, d.SSLMode,
	)
}

func Load() (*Config, error) {
	// Load .env file (ignore error if not exists - for production)
	_ = godotenv.Load()

	dbPort, err := strconv.Atoi(getEnv("DB_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("invalid DB_PORT: %w", err)
	}

	serverPort, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		return nil, fmt.Errorf("invalid SERVER_PORT: %w", err)
	}

	grpcPort, err := strconv.Atoi(getEnv("GRPC_PORT", "50051"))
	if err != nil {
		return nil, fmt.Errorf("invalid GRPC_PORT: %w", err)
	}

	jwtExpiration, err := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_EXPIRATION_HOURS: %w", err)
	}

	currencySyncInterval, err := getEnvDuration("CURRENCY_SYNC_INTERVAL", "6h")
	if err != nil {
		return nil, fmt.Errorf("invalid CURRENCY_SYNC_INTERVAL: %w", err)
	}

	currencySyncTimeout, err := getEnvDuration("CURRENCY_SYNC_TIMEOUT", "10s")
	if err != nil {
		return nil, fmt.Errorf("invalid CURRENCY_SYNC_TIMEOUT: %w", err)
	}

	// Parse CORS origins
	originsStr := getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000,http://localhost:5173")
	origins := strings.Split(originsStr, ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return &Config{
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     dbPort,
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", "expense_tracker"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		Server: ServerConfig{
			Port:            serverPort,
			GRPCPort:        grpcPort,
			AllowedOrigins:  origins,
			ReadTimeout:     15,
			WriteTimeout:    15,
			ShutdownTimeout: 30,
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", ""),
			ExpirationHours: jwtExpiration,
		},
		Currency: CurrencyConfig{
			ServiceAddr:  getEnv("CURRENCY_SERVICE_ADDR", "localhost:50052"),
			SyncInterval: currencySyncInterval,
			SyncTimeout:  currencySyncTimeout,
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvDuration reads a time.Duration env var (e.g. "6h", "10s"), falling back
// to defaultValue when unset.
func getEnvDuration(key, defaultValue string) (time.Duration, error) {
	return time.ParseDuration(getEnv(key, defaultValue))
}
