package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/DigitLock/expense-tracker/internal/config"
)

type Server struct {
	httpServer *http.Server
	config     *config.ServerConfig
}

func NewServer(cfg *config.ServerConfig, router *chi.Mux) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Port),
			Handler:      router,
			ReadTimeout:  time.Duration(cfg.ReadTimeout) * time.Second,
			WriteTimeout: time.Duration(cfg.WriteTimeout) * time.Second,
		},
		config: cfg,
	}
}

func (s *Server) Start() error {
	log.Printf("Server starting on port %d", s.config.Port)
	log.Printf("Health check: http://localhost:%d/health", s.config.Port)
	log.Printf("API base URL: http://localhost:%d/api/v1", s.config.Port)

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Server shutting down...")
	return s.httpServer.Shutdown(ctx)
}
