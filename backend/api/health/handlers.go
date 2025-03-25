package health

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/vpn-service/backend/db"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Timestamp string            `json:"timestamp"`
	Version   string            `json:"version"`
	Services  map[string]string `json:"services"`
}

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	// Create response
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   config.Version,
		Services:  make(map[string]string),
	}

	// Check database
	if err := checkDatabase(); err != nil {
		response.Status = "degraded"
		response.Services["database"] = "unhealthy: " + err.Error()
	} else {
		response.Services["database"] = "healthy"
	}

	// Check WireGuard
	if err := checkWireGuard(); err != nil {
		response.Status = "degraded"
		response.Services["wireguard"] = "unhealthy: " + err.Error()
	} else {
		response.Services["wireguard"] = "healthy"
	}

	// Set content type
	w.Header().Set("Content-Type", "application/json")

	// Set status code
	if response.Status == "ok" {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	// Write response
	if err := json.NewEncoder(w).Encode(response); err != nil {
		utils.LogError("Failed to encode health response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

// ReadinessHandler handles readiness check requests
func ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check if service is ready
	if !isReady() {
		http.Error(w, "Service is not ready", http.StatusServiceUnavailable)
		return
	}

	// Service is ready
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is ready"))
}

// LivenessHandler handles liveness check requests
func LivenessHandler(w http.ResponseWriter, r *http.Request) {
	// Check if service is alive
	if !isAlive() {
		http.Error(w, "Service is not alive", http.StatusServiceUnavailable)
		return
	}

	// Service is alive
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Service is alive"))
}

// checkDatabase checks if the database is healthy
func checkDatabase() error {
	// Ping database
	if db.DB == nil {
		return utils.NewError("database connection not initialized")
	}

	return db.DB.Ping()
}

// checkWireGuard checks if WireGuard is healthy
func checkWireGuard() error {
	// In a real implementation, this would check if WireGuard is running
	// For now, we'll just return nil
	return nil
}

// isReady checks if the service is ready to accept requests
func isReady() bool {
	// Check database
	if err := checkDatabase(); err != nil {
		return false
	}

	// Check WireGuard
	if err := checkWireGuard(); err != nil {
		return false
	}

	return true
}

// isAlive checks if the service is alive
func isAlive() bool {
	// In a real implementation, this would check if the service is alive
	// For now, we'll just return true
	return true
}
