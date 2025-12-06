package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DigitLock/expense-tracker/internal/dto"
)

const AppVersion = "1.0.0"

type HealthHandler struct {
	db *pgxpool.Pool
}

func NewHealthHandler(db *pgxpool.Pool) *HealthHandler {
	return &HealthHandler{db: db}
}

// Health godoc
// @Summary      Health check
// @Description  Returns the overall health status of the service including database connectivity
// @Tags         Health
// @Produce      json
// @Success      200 {object} dto.HealthResponse "Service is healthy"
// @Failure      503 {object} dto.HealthResponse "Service is degraded (database unhealthy)"
// @Router       /health [get]
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "connected"
	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "disconnected"
	}

	status := "healthy"
	if dbStatus == "disconnected" {
		status = "degraded"
	}

	response := dto.HealthResponse{
		Status:    status,
		Timestamp: time.Now().UTC(),
		Version:   AppVersion,
		Database:  dbStatus,
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(response)
}

// Ready godoc
// @Summary      Readiness check
// @Description  Kubernetes readiness probe endpoint. Returns 200 if service is ready to accept traffic, 503 otherwise.
// @Tags         Health
// @Produce      plain
// @Success      200 {string} string "ready"
// @Failure      503 {string} string "not ready"
// @Router       /ready [get]
func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("not ready"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ready"))
}
