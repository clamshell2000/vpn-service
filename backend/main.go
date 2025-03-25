package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/vpn-service/backend/api/auth"
	"github.com/vpn-service/backend/api/middleware"
	"github.com/vpn-service/backend/api/vpn"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/core"
	"github.com/vpn-service/backend/src/db"
	"github.com/vpn-service/backend/src/monitoring"
	"github.com/vpn-service/backend/src/utils"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	if err := utils.InitLogger(cfg.Monitoring.LogDir); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer utils.CloseLogger()

	// Initialize database
	if err := db.Initialize(cfg.Database); err != nil {
		utils.LogFatal("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.RunMigrations(); err != nil {
		utils.LogFatal("Failed to run migrations: %v", err)
	}

	// Initialize metrics collector
	metricsCollector := monitoring.NewCollector(cfg)
	monitoring.MetricsCollector = metricsCollector
	metricsCollector.StartMetricsServer()

	// Initialize managers
	serverManager := core.NewServerManager(cfg)
	vpnManager := core.NewVPNManager(cfg, serverManager)

	// Set VPN manager for API handlers
	vpn.VPNManager = vpnManager

	// Start server monitoring in background
	go serverManager.MonitorServers()

	// Initialize router
	router := mux.NewRouter()

	// Set up middleware
	router.Use(middleware.LoggingMiddleware)
	router.Use(middleware.MetricsMiddleware)

	// Public routes
	router.HandleFunc("/api/health", healthCheckHandler).Methods("GET")
	
	// Auth routes
	authRouter := router.PathPrefix("/api/auth").Subrouter()
	auth.RegisterRoutes(authRouter)

	// VPN routes (protected)
	vpnRouter := router.PathPrefix("/api/vpn").Subrouter()
	vpnRouter.Use(middleware.JWTAuthMiddleware)
	vpn.RegisterRoutes(vpnRouter)

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
	utils.LogInfo("Starting API server on %s", cfg.APIAddr)
	srv := &http.Server{
		Addr:         cfg.APIAddr,
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.LogError("Failed to start server: %v", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Shutdown server
	utils.LogInfo("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		utils.LogError("Server shutdown failed: %v", err)
	}

	utils.LogInfo("Server shutdown complete")
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","version":"1.0.0"}`))
}
