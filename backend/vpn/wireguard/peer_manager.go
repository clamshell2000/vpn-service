package wireguard

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

var (
	// peerMutex ensures thread-safe peer operations
	peerMutex sync.Mutex
)

// PeerManager handles WireGuard peer operations
type PeerManager struct {
	config *config.Config
}

// PeerConfig represents a WireGuard peer configuration
type PeerConfig struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId"`
	ServerID   string    `json:"serverId"`
	DeviceType string    `json:"deviceType"`
	DeviceName string    `json:"deviceName"`
	PublicKey  string    `json:"publicKey"`
	PrivateKey string    `json:"privateKey"`
	IP         string    `json:"ip"`
	ServerIP   string    `json:"serverIp"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	Dynamic    bool      `json:"dynamic"`
}

// PeerInfo represents information about a WireGuard peer
type PeerInfo struct {
	ID         string `json:"id"`
	ServerID   string `json:"serverId"`
	ServerName string `json:"serverName"`
	DeviceType string `json:"deviceType"`
	DeviceName string `json:"deviceName"`
	IP         string `json:"ip"`
	CreatedAt  string `json:"createdAt"`
	LastSeen   string `json:"lastSeen"`
	BytesRx    int64  `json:"bytesRx"`
	BytesTx    int64  `json:"bytesTx"`
}

// NewPeerManager creates a new peer manager
func NewPeerManager(cfg *config.Config) *PeerManager {
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(cfg.WireGuard.ConfigDir, 0755); err != nil {
		utils.LogError("Failed to create config directory: %v", err)
	}

	// Create dynamic peer directory if it doesn't exist
	if err := os.MkdirAll(cfg.WireGuard.DynamicPeerDir, 0755); err != nil {
		utils.LogError("Failed to create dynamic peer directory: %v", err)
	}

	return &PeerManager{
		config: cfg,
	}
}

// CreatePeer creates a new WireGuard peer
func (pm *PeerManager) CreatePeer(userID, serverID, deviceType, deviceName string) (*PeerConfig, error) {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	// Generate peer ID
	peerID := utils.GenerateUUID()

	// Generate key pair
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %v", err)
	}

	// Allocate IP address
	ip, err := pm.allocateIP()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP address: %v", err)
	}

	// Create peer config
	peer := &PeerConfig{
		ID:         peerID,
		UserID:     userID,
		ServerID:   serverID,
		DeviceType: deviceType,
		DeviceName: deviceName,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		IP:         ip,
		ServerIP:   pm.config.WireGuard.ServerIP,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Dynamic:    false,
	}

	// Save peer config
	if err := pm.savePeerConfig(peer); err != nil {
		return nil, fmt.Errorf("failed to save peer config: %v", err)
	}

	// Apply configuration
	if err := pm.applyConfiguration(); err != nil {
		return nil, fmt.Errorf("failed to apply configuration: %v", err)
	}

	return peer, nil
}

// CreateDynamicPeer creates a new dynamic WireGuard peer
func (pm *PeerManager) CreateDynamicPeer(userID, serverID, deviceType, deviceName string) (*PeerConfig, error) {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	// Generate peer ID
	peerID := utils.GenerateUUID()

	// Generate key pair
	privateKey, publicKey, err := generateKeyPair()
	if err != nil {
		return nil, fmt.Errorf("failed to generate key pair: %v", err)
	}

	// Allocate IP address
	ip, err := pm.allocateIP()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP address: %v", err)
	}

	// Create peer config
	peer := &PeerConfig{
		ID:         peerID,
		UserID:     userID,
		ServerID:   serverID,
		DeviceType: deviceType,
		DeviceName: deviceName,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		IP:         ip,
		ServerIP:   pm.config.WireGuard.ServerIP,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
		Dynamic:    true,
	}

	// Save peer config
	if err := pm.saveDynamicPeerConfig(peer); err != nil {
		return nil, fmt.Errorf("failed to save dynamic peer config: %v", err)
	}

	// Apply configuration
	if err := pm.applyConfiguration(); err != nil {
		return nil, fmt.Errorf("failed to apply configuration: %v", err)
	}

	return peer, nil
}

// RemovePeer removes a WireGuard peer
func (pm *PeerManager) RemovePeer(userID, peerID string) error {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	// Get peer config
	peer, err := pm.getPeerConfig(userID, peerID)
	if err != nil {
		return fmt.Errorf("failed to get peer config: %v", err)
	}

	// Delete peer config
	if err := pm.deletePeerConfig(peer); err != nil {
		return fmt.Errorf("failed to delete peer config: %v", err)
	}

	// Apply configuration
	if err := pm.applyConfiguration(); err != nil {
		return fmt.Errorf("failed to apply configuration: %v", err)
	}

	return nil
}

// RemoveDynamicPeer removes a dynamic WireGuard peer
func (pm *PeerManager) RemoveDynamicPeer(userID, peerID string) error {
	peerMutex.Lock()
	defer peerMutex.Unlock()

	// Get peer config
	peer, err := pm.getDynamicPeerConfig(userID, peerID)
	if err != nil {
		return fmt.Errorf("failed to get dynamic peer config: %v", err)
	}

	// Delete peer config
	if err := pm.deleteDynamicPeerConfig(peer); err != nil {
		return fmt.Errorf("failed to delete dynamic peer config: %v", err)
	}

	// Apply configuration
	if err := pm.applyConfiguration(); err != nil {
		return fmt.Errorf("failed to apply configuration: %v", err)
	}

	return nil
}

// GetPeer gets a WireGuard peer
func (pm *PeerManager) GetPeer(userID, peerID string) (*PeerConfig, error) {
	// Try to get static peer first
	peer, err := pm.getPeerConfig(userID, peerID)
	if err == nil {
		return peer, nil
	}

	// If not found, try to get dynamic peer
	return pm.getDynamicPeerConfig(userID, peerID)
}

// GetPeers gets all WireGuard peers for a user
func (pm *PeerManager) GetPeers(userID string) ([]*PeerConfig, error) {
	// Get static peers
	staticPeers, err := pm.getStaticPeers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get static peers: %v", err)
	}

	// Get dynamic peers
	dynamicPeers, err := pm.getDynamicPeers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get dynamic peers: %v", err)
	}

	// Combine peers
	peers := append(staticPeers, dynamicPeers...)
	return peers, nil
}

// getStaticPeers gets all static WireGuard peers for a user
func (pm *PeerManager) getStaticPeers(userID string) ([]*PeerConfig, error) {
	// Get user directory
	userDir := filepath.Join(pm.config.WireGuard.ConfigDir, userID)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return []*PeerConfig{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read user directory: %v", err)
	}

	// Get peer configs
	peers := []*PeerConfig{}
	for _, entry := range entries {
		if entry.IsDir() {
			peerID := entry.Name()
			peer, err := pm.getPeerConfig(userID, peerID)
			if err != nil {
				utils.LogError("Failed to get peer config: %v", err)
				continue
			}
			peers = append(peers, peer)
		}
	}

	return peers, nil
}

// getDynamicPeers gets all dynamic WireGuard peers for a user
func (pm *PeerManager) getDynamicPeers(userID string) ([]*PeerConfig, error) {
	// Get user directory
	userDir := filepath.Join(pm.config.WireGuard.DynamicPeerDir, userID)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return []*PeerConfig{}, nil
	}

	// Read directory
	entries, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read user directory: %v", err)
	}

	// Get peer configs
	peers := []*PeerConfig{}
	for _, entry := range entries {
		if entry.IsDir() {
			peerID := entry.Name()
			peer, err := pm.getDynamicPeerConfig(userID, peerID)
			if err != nil {
				utils.LogError("Failed to get dynamic peer config: %v", err)
				continue
			}
			peers = append(peers, peer)
		}
	}

	return peers, nil
}

// GenerateConfig generates a WireGuard configuration for a peer
func (pm *PeerManager) GenerateConfig(peer *PeerConfig) (string, error) {
	// Get template based on device type
	template, err := getConfigTemplate(peer.DeviceType)
	if err != nil {
		return "", fmt.Errorf("failed to get config template: %v", err)
	}

	// Replace placeholders
	config := template
	config = replaceConfigPlaceholders(config, map[string]string{
		"PRIVATE_KEY":        peer.PrivateKey,
		"CLIENT_IP":          peer.IP,
		"SERVER_PUBLIC_KEY":  pm.config.WireGuard.PublicKey,
		"SERVER_ENDPOINT":    fmt.Sprintf("%s:%d", pm.config.WireGuard.ServerEndpoint, pm.config.WireGuard.ListenPort),
		"DNS":                pm.config.WireGuard.DNS,
		"ALLOWED_IPS":        pm.config.WireGuard.AllowedIPs,
		"PERSISTENT_KEEPALIVE": "25",
	})

	return config, nil
}

// savePeerConfig saves a peer configuration
func (pm *PeerManager) savePeerConfig(peer *PeerConfig) error {
	// Create user directory if it doesn't exist
	userDir := filepath.Join(pm.config.WireGuard.ConfigDir, peer.UserID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %v", err)
	}

	// Create peer directory if it doesn't exist
	peerDir := filepath.Join(userDir, peer.ID)
	if err := os.MkdirAll(peerDir, 0755); err != nil {
		return fmt.Errorf("failed to create peer directory: %v", err)
	}

	// Save peer metadata
	metadataPath := filepath.Join(peerDir, "metadata.json")
	if err := utils.WriteJSONToFile(metadataPath, peer); err != nil {
		return fmt.Errorf("failed to save peer metadata: %v", err)
	}

	return nil
}

// saveDynamicPeerConfig saves a dynamic peer configuration
func (pm *PeerManager) saveDynamicPeerConfig(peer *PeerConfig) error {
	// Create user directory if it doesn't exist
	userDir := filepath.Join(pm.config.WireGuard.DynamicPeerDir, peer.UserID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %v", err)
	}

	// Create peer directory if it doesn't exist
	peerDir := filepath.Join(userDir, peer.ID)
	if err := os.MkdirAll(peerDir, 0755); err != nil {
		return fmt.Errorf("failed to create peer directory: %v", err)
	}

	// Save peer metadata
	metadataPath := filepath.Join(peerDir, "metadata.json")
	if err := utils.WriteJSONToFile(metadataPath, peer); err != nil {
		return fmt.Errorf("failed to save peer metadata: %v", err)
	}

	return nil
}

// getPeerConfig gets a peer configuration
func (pm *PeerManager) getPeerConfig(userID, peerID string) (*PeerConfig, error) {
	// Get peer metadata path
	metadataPath := filepath.Join(pm.config.WireGuard.ConfigDir, userID, peerID, "metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("peer not found: %s", peerID)
	}

	// Read peer metadata
	var peer PeerConfig
	if err := utils.ReadJSONFromFile(metadataPath, &peer); err != nil {
		return nil, fmt.Errorf("failed to read peer metadata: %v", err)
	}

	return &peer, nil
}

// getDynamicPeerConfig gets a dynamic peer configuration
func (pm *PeerManager) getDynamicPeerConfig(userID, peerID string) (*PeerConfig, error) {
	// Get peer metadata path
	metadataPath := filepath.Join(pm.config.WireGuard.DynamicPeerDir, userID, peerID, "metadata.json")
	if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("dynamic peer not found: %s", peerID)
	}

	// Read peer metadata
	var peer PeerConfig
	if err := utils.ReadJSONFromFile(metadataPath, &peer); err != nil {
		return nil, fmt.Errorf("failed to read dynamic peer metadata: %v", err)
	}

	return &peer, nil
}

// deletePeerConfig deletes a peer configuration
func (pm *PeerManager) deletePeerConfig(peer *PeerConfig) error {
	// Get peer directory
	peerDir := filepath.Join(pm.config.WireGuard.ConfigDir, peer.UserID, peer.ID)
	if _, err := os.Stat(peerDir); os.IsNotExist(err) {
		return fmt.Errorf("peer directory not found: %s", peerDir)
	}

	// Delete peer directory
	if err := os.RemoveAll(peerDir); err != nil {
		return fmt.Errorf("failed to delete peer directory: %v", err)
	}

	return nil
}

// deleteDynamicPeerConfig deletes a dynamic peer configuration
func (pm *PeerManager) deleteDynamicPeerConfig(peer *PeerConfig) error {
	// Get peer directory
	peerDir := filepath.Join(pm.config.WireGuard.DynamicPeerDir, peer.UserID, peer.ID)
	if _, err := os.Stat(peerDir); os.IsNotExist(err) {
		return fmt.Errorf("dynamic peer directory not found: %s", peerDir)
	}

	// Delete peer directory
	if err := os.RemoveAll(peerDir); err != nil {
		return fmt.Errorf("failed to delete dynamic peer directory: %v", err)
	}

	return nil
}

// allocateIP allocates an IP address for a peer
func (pm *PeerManager) allocateIP() (string, error) {
	// In a real implementation, this would allocate an IP from a pool
	// For now, we'll just return a mock IP
	return "10.0.0.2/32", nil
}

// applyConfiguration applies the WireGuard configuration
func (pm *PeerManager) applyConfiguration() error {
	// In a real implementation, this would apply the configuration to WireGuard
	// For now, we'll just log it
	utils.LogInfo("Applying WireGuard configuration...")
	return nil
}

// generateKeyPair generates a WireGuard key pair
func generateKeyPair() (string, string, error) {
	// In a real implementation, this would use wg-quick to generate keys
	// For now, we'll just return mock keys
	privateKey := "YAnV4SnPYEA+jS6nQtxF5lS3jj0gqXBVVeP9tz/bP2A="
	publicKey := "zzz3UBcqiV9RsYCzJWOU5VVVNk3VtQECQXXPnQiEfQQ="
	return privateKey, publicKey, nil
}

// getConfigTemplate gets a configuration template for a device type
func getConfigTemplate(deviceType string) (string, error) {
	// Map device type to template file
	templateFile := "generic.conf"
	switch strings.ToLower(deviceType) {
	case "android":
		templateFile = "android.conf"
	case "ios", "iphone", "ipad":
		templateFile = "ios.conf"
	case "windows":
		templateFile = "windows.conf"
	case "mac", "macos":
		templateFile = "mac.conf"
	}

	// Read template file
	templatePath := filepath.Join("vpn/wireguard/config_templates", templateFile)
	content, err := os.ReadFile(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %v", err)
	}

	return string(content), nil
}

// replaceConfigPlaceholders replaces placeholders in a configuration template
func replaceConfigPlaceholders(template string, replacements map[string]string) string {
	result := template
	for key, value := range replacements {
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), value)
	}
	return result
}

// GenerateQRCode generates a QR code for a WireGuard configuration
func GenerateQRCode(config string) (string, error) {
	// In a real implementation, this would generate a QR code
	// For now, we'll just return a mock QR code
	return "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==", nil
}
