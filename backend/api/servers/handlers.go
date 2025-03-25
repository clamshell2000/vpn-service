package servers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/vpn-service/backend/src/core"
	"github.com/vpn-service/backend/src/utils"
)

// ServerManager is the server manager instance
var ServerManager *core.ServerManager

// ServerRequest represents a server creation/update request
type ServerRequest struct {
	Name     string `json:"name"`
	Location string `json:"location"`
	IP       string `json:"ip"`
}

// ListServersHandler handles server listing requests
func ListServersHandler(w http.ResponseWriter, r *http.Request) {
	// Get servers
	servers := ServerManager.GetServers()

	// Return servers
	utils.WriteJSONResponse(w, http.StatusOK, servers)
}

// GetServerHandler handles server retrieval requests
func GetServerHandler(w http.ResponseWriter, r *http.Request) {
	// Get server ID from URL
	vars := mux.Vars(r)
	serverID := vars["id"]

	// Get server
	server, err := ServerManager.GetServer(serverID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "Server not found")
		return
	}

	// Return server
	utils.WriteJSONResponse(w, http.StatusOK, server)
}

// CreateServerHandler handles server creation requests
func CreateServerHandler(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req ServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validateServerRequest(req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Create server
	server := &core.Server{
		ID:       utils.GenerateUUID(),
		Name:     req.Name,
		Location: req.Location,
		IP:       req.IP,
		Status:   "offline",
		Load:     0,
	}

	// Add server
	ServerManager.AddServer(server)

	// Return server
	utils.WriteJSONResponse(w, http.StatusCreated, server)
}

// UpdateServerHandler handles server update requests
func UpdateServerHandler(w http.ResponseWriter, r *http.Request) {
	// Get server ID from URL
	vars := mux.Vars(r)
	serverID := vars["id"]

	// Get server
	server, err := ServerManager.GetServer(serverID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "Server not found")
		return
	}

	// Parse request
	var req ServerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate request
	if err := validateServerRequest(req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update server
	server.Name = req.Name
	server.Location = req.Location
	server.IP = req.IP

	// Save server
	if err := ServerManager.UpdateServer(server); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to update server")
		return
	}

	// Return server
	utils.WriteJSONResponse(w, http.StatusOK, server)
}

// DeleteServerHandler handles server deletion requests
func DeleteServerHandler(w http.ResponseWriter, r *http.Request) {
	// Get server ID from URL
	vars := mux.Vars(r)
	serverID := vars["id"]

	// Delete server
	if err := ServerManager.DeleteServer(serverID); err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "Server not found")
		return
	}

	// Return success
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// UpdateServerStatusHandler handles server status update requests
func UpdateServerStatusHandler(w http.ResponseWriter, r *http.Request) {
	// Get server ID from URL
	vars := mux.Vars(r)
	serverID := vars["id"]

	// Get status from URL
	status := vars["status"]
	if status != "online" && status != "offline" && status != "maintenance" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid status")
		return
	}

	// Update server status
	if err := ServerManager.UpdateServerStatus(serverID, status); err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "Server not found")
		return
	}

	// Return success
	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "success"})
}

// validateServerRequest validates a server request
func validateServerRequest(req ServerRequest) error {
	// Validate name
	if strings.TrimSpace(req.Name) == "" {
		return utils.NewError("name is required")
	}

	// Validate location
	if strings.TrimSpace(req.Location) == "" {
		return utils.NewError("location is required")
	}

	// Validate IP
	if strings.TrimSpace(req.IP) == "" {
		return utils.NewError("IP is required")
	}

	return nil
}
