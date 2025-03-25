package api

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/cors"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// Server represents the API server
type Server struct {
	config *config.Config
	server *http.Server
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, router http.Handler) *Server {
	// Set up CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400,
	})
	handler := c.Handler(router)

	// Create server
	server := &http.Server{
		Addr:         cfg.APIAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		config: cfg,
		server: server,
	}
}

// Start starts the API server
func (s *Server) Start() error {
	utils.LogInfo("API server listening on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the API server
func (s *Server) Shutdown(ctx context.Context) error {
	utils.LogInfo("Shutting down API server...")
	return s.server.Shutdown(ctx)
}
