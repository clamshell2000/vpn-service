package vpn

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vpn-service/backend/src/core"
	"github.com/vpn-service/backend/src/utils"
	"github.com/vpn-service/backend/vpn/wireguard"
)

// VPNManager is the VPN manager instance
var VPNManager *core.VPNManager

// RegisterRoutes registers the VPN routes
func RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/servers", GetServersHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("/connect", ConnectHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/disconnect", DisconnectHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/status", StatusHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("/config", GetConfigHandler).Methods("GET", "OPTIONS")
	router.HandleFunc("/qr", GetQRCodeHandler).Methods("GET", "OPTIONS")
	
	// Dynamic peer management
	router.HandleFunc("/dynamic/connect", DynamicConnectHandler).Methods("POST", "OPTIONS")
	router.HandleFunc("/dynamic/disconnect", DynamicDisconnectHandler).Methods("POST", "OPTIONS")
}

// Server represents a VPN server
type Server struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	IP       string `json:"ip"`
	Status   string `json:"status"`
	Load     int    `json:"load"`
}

// ConnectRequest represents a VPN connection request
type ConnectRequest struct {
	ServerID   string `json:"serverId"`
	DeviceType string `json:"deviceType"`
	DeviceName string `json:"deviceName"`
}

// DisconnectRequest represents a VPN disconnection request
type DisconnectRequest struct {
	PeerID string `json:"peerId"`
}

// ConnectResponse represents a VPN connection response
type ConnectResponse struct {
	Config    string `json:"config"`
	QRCode    string `json:"qrCode,omitempty"`
	PeerID    string `json:"peerId"`
	ServerIP  string `json:"serverIp"`
}

// StatusResponse represents a VPN status response
type StatusResponse struct {
	Connected bool                  `json:"connected"`
	Peers     []*wireguard.PeerInfo `json:"peers"`
}

// GetServersHandler returns a list of available VPN servers
func GetServersHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(string)

	// Get servers from server manager
	coreServers := VPNManager.GetServers()
	
	// Convert to API response format
	servers := make([]Server, len(coreServers))
	for i, server := range coreServers {
		servers[i] = Server{
			ID:       server.ID,
			Name:     server.Name,
			Location: server.Location,
			IP:       server.IP,
			Status:   server.Status,
			Load:     server.Load,
		}
	}

	utils.WriteJSONResponse(w, http.StatusOK, servers)
}

// ConnectHandler handles VPN connection requests
func ConnectHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(string)

	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request
	if req.ServerID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	// Default to generic device type if not specified
	deviceType := req.DeviceType
	if deviceType == "" {
		deviceType = "generic"
	}

	// Default device name
	deviceName := req.DeviceName
	if deviceName == "" {
		deviceName = deviceType
	}

	// Connect to VPN
	peer, config, err := VPNManager.Connect(userID, req.ServerID, deviceType, deviceName)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to connect to VPN: "+err.Error())
		return
	}

	// Generate QR code for mobile devices
	var qrCode string
	if deviceType == "android" || deviceType == "ios" {
		qrCode, err = wireguard.GenerateQRCode(config)
		if err != nil {
			// Non-fatal error, continue without QR code
			utils.LogError("Failed to generate QR code: %v", err)
		}
	}

	// Respond with configuration
	utils.WriteJSONResponse(w, http.StatusOK, ConnectResponse{
		Config:   config,
		QRCode:   qrCode,
		PeerID:   peer.ID,
		ServerIP: peer.ServerIP,
	})
}

// DisconnectHandler handles VPN disconnection requests
func DisconnectHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(string)

	var req DisconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request
	if req.PeerID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Peer ID is required")
		return
	}

	// Disconnect from VPN
	if err := VPNManager.Disconnect(userID, req.PeerID); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to disconnect from VPN: "+err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "disconnected"})
}

// StatusHandler returns the current VPN connection status
func StatusHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(string)

	// Get connection status
	peers, err := VPNManager.GetStatus(userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get connection status: "+err.Error())
		return
	}

	// Create response
	response := StatusResponse{
		Connected: len(peers) > 0,
		Peers:     peers,
	}

	utils.WriteJSONResponse(w, http.StatusOK, response)
}

// GetConfigHandler returns the WireGuard configuration for a peer
func GetConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(string)

	// Get peer ID from query
	peerID := r.URL.Query().Get("peerId")
	if peerID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Peer ID is required")
		return
	}

	// Get configuration
	config, err := VPNManager.GetConfig(userID, peerID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get configuration: "+err.Error())
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Disposition", "attachment; filename=\"wg0.conf\"")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(config))
}

// GetQRCodeHandler returns a QR code for a WireGuard configuration
func GetQRCodeHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(string)

	// Get peer ID from query
	peerID := r.URL.Query().Get("peerId")
	if peerID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Peer ID is required")
		return
	}

	// Get configuration
	config, err := VPNManager.GetConfig(userID, peerID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to get configuration: "+err.Error())
		return
	}

	// Generate QR code
	qrCode, err := wireguard.GenerateQRCode(config)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to generate QR code: "+err.Error())
		return
	}

	// Set content type
	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(qrCode))
}

// DynamicConnectHandler handles dynamic VPN connection requests
func DynamicConnectHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(string)

	var req ConnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request
	if req.ServerID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Server ID is required")
		return
	}

	// Default to generic device type if not specified
	deviceType := req.DeviceType
	if deviceType == "" {
		deviceType = "generic"
	}

	// Default device name
	deviceName := req.DeviceName
	if deviceName == "" {
		deviceName = deviceType
	}

	// Connect to VPN
	peer, config, err := VPNManager.DynamicConnect(userID, req.ServerID, deviceType, deviceName)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to connect to VPN: "+err.Error())
		return
	}

	// Generate QR code for mobile devices
	var qrCode string
	if deviceType == "android" || deviceType == "ios" {
		qrCode, err = wireguard.GenerateQRCode(config)
		if err != nil {
			// Non-fatal error, continue without QR code
			utils.LogError("Failed to generate QR code: %v", err)
		}
	}

	// Respond with configuration
	utils.WriteJSONResponse(w, http.StatusOK, ConnectResponse{
		Config:   config,
		QRCode:   qrCode,
		PeerID:   peer.ID,
		ServerIP: peer.ServerIP,
	})
}

// DynamicDisconnectHandler handles dynamic VPN disconnection requests
func DynamicDisconnectHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := r.Context().Value("userID").(string)

	var req DisconnectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate request
	if req.PeerID == "" {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Peer ID is required")
		return
	}

	// Disconnect from VPN
	if err := VPNManager.DynamicDisconnect(userID, req.PeerID); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to disconnect from VPN: "+err.Error())
		return
	}

	utils.WriteJSONResponse(w, http.StatusOK, map[string]string{"status": "disconnected"})
}
