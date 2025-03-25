package wgutils

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/skip2/go-qrcode"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
	"github.com/vpn-service/backend/vpn/wireguard"
)

// ConnectionStatus represents the status of a VPN connection
type ConnectionStatus struct {
	Connected    bool      `json:"connected"`
	PeerID       string    `json:"peerId,omitempty"`
	ServerID     string    `json:"serverId,omitempty"`
	IP           string    `json:"ip,omitempty"`
	ConnectedAt  time.Time `json:"connectedAt,omitempty"`
	BytesReceived int64    `json:"bytesReceived"`
	BytesSent    int64     `json:"bytesSent"`
}

// GeneratePeerConfig generates a peer configuration
func GeneratePeerConfig(userID, serverID, deviceType string) (string, string, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return "", "", fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create peer manager
	pm, err := wireguard.NewPeerManager(cfg)
	if err != nil {
		return "", "", fmt.Errorf("failed to create peer manager: %v", err)
	}

	// Create peer
	peer, err := pm.CreatePeer(userID, serverID, deviceType)
	if err != nil {
		return "", "", fmt.Errorf("failed to create peer: %v", err)
	}

	// Generate configuration
	config, err := pm.GenerateConfig(peer)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate configuration: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(userID, "vpn_connect", fmt.Sprintf("server=%s device=%s", serverID, deviceType))

	return config, peer.ID, nil
}

// RemovePeerConfig removes a peer configuration
func RemovePeerConfig(userID, peerID string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create peer manager
	pm, err := wireguard.NewPeerManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to create peer manager: %v", err)
	}

	// Delete peer
	if err := pm.DeletePeer(userID, peerID); err != nil {
		return fmt.Errorf("failed to delete peer: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(userID, "vpn_disconnect", fmt.Sprintf("peer=%s", peerID))

	return nil
}

// GetConnectionStatus gets the connection status for a user
func GetConnectionStatus(userID string) (*ConnectionStatus, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create peer manager
	pm, err := wireguard.NewPeerManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer manager: %v", err)
	}

	// List peers
	peers, err := pm.ListPeers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list peers: %v", err)
	}

	// Check if user has any peers
	if len(peers) == 0 {
		return &ConnectionStatus{
			Connected: false,
		}, nil
	}

	// TODO: Check if peer is actually connected
	// For now, just assume the first peer is connected
	peer := peers[0]
	return &ConnectionStatus{
		Connected:    true,
		PeerID:       peer.ID,
		ServerID:     peer.ServerID,
		IP:           peer.IP,
		ConnectedAt:  peer.CreatedAt,
		BytesReceived: 0, // TODO: Get actual bytes received
		BytesSent:    0, // TODO: Get actual bytes sent
	}, nil
}

// GenerateDynamicPeerConfig generates a dynamic peer configuration
func GenerateDynamicPeerConfig(userID, serverID, deviceType string) (string, string, error) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return "", "", fmt.Errorf("failed to load configuration: %v", err)
	}

	// Generate peer ID
	peerID := utils.GenerateUUID()

	// Create dynamic peer directory
	peerDir := filepath.Join(cfg.WireGuard.DynamicPeerDir, userID, peerID)
	if err := os.MkdirAll(peerDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create dynamic peer directory: %v", err)
	}

	// Generate key pair
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return "", "", fmt.Errorf("failed to generate key pair: %v", err)
	}

	// Allocate IP address
	ip := allocateIP()

	// Create metadata
	metadata := map[string]interface{}{
		"id":         peerID,
		"userId":     userID,
		"serverId":   serverID,
		"deviceType": deviceType,
		"publicKey":  publicKey,
		"ip":         ip,
		"createdAt":  time.Now(),
	}

	// Save metadata
	metadataPath := filepath.Join(peerDir, "metadata.json")
	if err := utils.WriteJSONToFile(metadataPath, metadata); err != nil {
		return "", "", fmt.Errorf("failed to save metadata: %v", err)
	}

	// Generate configuration
	config := generateConfig(privateKey, ip, cfg, deviceType)

	// Save configuration
	configPath := filepath.Join(peerDir, "wg.conf")
	if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
		return "", "", fmt.Errorf("failed to save configuration: %v", err)
	}

	// Apply configuration
	if err := applyConfiguration(publicKey, ip); err != nil {
		return "", "", fmt.Errorf("failed to apply configuration: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(userID, "vpn_dynamic_connect", fmt.Sprintf("server=%s device=%s", serverID, deviceType))

	return config, peerID, nil
}

// RemoveDynamicPeerConfig removes a dynamic peer configuration
func RemoveDynamicPeerConfig(userID, peerID string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Get peer directory
	peerDir := filepath.Join(cfg.WireGuard.DynamicPeerDir, userID, peerID)
	if _, err := os.Stat(peerDir); os.IsNotExist(err) {
		return fmt.Errorf("peer not found")
	}

	// Read metadata
	metadataPath := filepath.Join(peerDir, "metadata.json")
	var metadata map[string]interface{}
	if err := utils.ReadJSONFromFile(metadataPath, &metadata); err != nil {
		return fmt.Errorf("failed to read metadata: %v", err)
	}

	// Remove peer from WireGuard
	publicKey, ok := metadata["publicKey"].(string)
	if !ok {
		return fmt.Errorf("invalid metadata: missing publicKey")
	}

	if err := removePeer(publicKey); err != nil {
		return fmt.Errorf("failed to remove peer: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(userID, "vpn_dynamic_disconnect", fmt.Sprintf("peer=%s", peerID))

	return nil
}

// GenerateQRCode generates a QR code for a WireGuard configuration
func GenerateQRCode(config string) (string, error) {
	// Generate QR code
	qr, err := qrcode.Encode(config, qrcode.Medium, 256)
	if err != nil {
		return "", fmt.Errorf("failed to generate QR code: %v", err)
	}

	// Convert to base64
	return "data:image/png;base64," + utils.Base64Encode(qr), nil
}

// Helper functions

// generateKeyPair generates a WireGuard key pair
func generateKeyPair() (string, string, error) {
	// TODO: Implement actual key pair generation
	// For now, just return mock keys
	privateKey := "mock-private-key"
	publicKey := "mock-public-key"
	return privateKey, publicKey, nil
}

// allocateIP allocates an IP address
func allocateIP() string {
	// TODO: Implement actual IP allocation
	// For now, just return a mock IP
	return fmt.Sprintf("10.0.0.%d/24", 100+time.Now().UnixNano()%100)
}

// generateConfig generates a WireGuard configuration
func generateConfig(privateKey, ip string, cfg *config.Config, deviceType string) string {
	// Get template based on device type
	template := getConfigTemplate(deviceType)

	// Replace placeholders
	config := template
	config = strings.Replace(config, "PRIVATE_KEY", privateKey, -1)
	config = strings.Replace(config, "CLIENT_IP", ip, -1)
	config = strings.Replace(config, "SERVER_PUBLIC_KEY", cfg.WireGuard.PublicKey, -1)
	config = strings.Replace(config, "SERVER_ENDPOINT", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.WireGuard.ListenPort), -1)
	config = strings.Replace(config, "DNS", cfg.WireGuard.DNS, -1)
	config = strings.Replace(config, "ALLOWED_IPS", "0.0.0.0/0, ::/0", -1)
	config = strings.Replace(config, "PERSISTENT_KEEPALIVE", "25", -1)

	return config
}

// getConfigTemplate gets a configuration template for a device type
func getConfigTemplate(deviceType string) string {
	// TODO: Load templates from files
	// For now, just return a basic template
	return `[Interface]
PrivateKey = PRIVATE_KEY
Address = CLIENT_IP
DNS = DNS

[Peer]
PublicKey = SERVER_PUBLIC_KEY
Endpoint = SERVER_ENDPOINT
AllowedIPs = ALLOWED_IPS
PersistentKeepalive = PERSISTENT_KEEPALIVE
`
}

// applyConfiguration applies a peer configuration
func applyConfiguration(publicKey, ip string) error {
	// TODO: Implement actual configuration application
	// For now, just log that we would apply the configuration
	utils.LogInfo("Applying peer configuration: publicKey=%s ip=%s", publicKey, ip)
	return nil
}

// removePeer removes a peer
func removePeer(publicKey string) error {
	// TODO: Implement actual peer removal
	// For now, just log that we would remove the peer
	utils.LogInfo("Removing peer: publicKey=%s", publicKey)
	return nil
}
