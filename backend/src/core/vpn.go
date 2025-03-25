package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
	"github.com/vpn-service/backend/vpn/wireguard"
)

// VPNManager manages VPN operations
type VPNManager struct {
	config     *config.Config
	peerManager *wireguard.PeerManager
}

// NewVPNManager creates a new VPN manager
func NewVPNManager(cfg *config.Config) (*VPNManager, error) {
	// Create peer manager
	peerManager, err := wireguard.NewPeerManager(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create peer manager: %v", err)
	}

	return &VPNManager{
		config:     cfg,
		peerManager: peerManager,
	}, nil
}

// ConnectUser connects a user to the VPN
func (vm *VPNManager) ConnectUser(userID, serverID, deviceType string) (string, string, error) {
	// Create peer
	peer, err := vm.peerManager.CreatePeer(userID, serverID, deviceType)
	if err != nil {
		return "", "", fmt.Errorf("failed to create peer: %v", err)
	}

	// Generate configuration
	config, err := vm.peerManager.GenerateConfig(peer)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate configuration: %v", err)
	}

	// Apply configuration
	if err := vm.applyConfiguration(); err != nil {
		return "", "", fmt.Errorf("failed to apply configuration: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(userID, "vpn_connect", fmt.Sprintf("server=%s device=%s", serverID, deviceType))

	return config, peer.ID, nil
}

// DisconnectUser disconnects a user from the VPN
func (vm *VPNManager) DisconnectUser(userID, peerID string) error {
	// Delete peer
	if err := vm.peerManager.DeletePeer(userID, peerID); err != nil {
		return fmt.Errorf("failed to delete peer: %v", err)
	}

	// Apply configuration
	if err := vm.applyConfiguration(); err != nil {
		return fmt.Errorf("failed to apply configuration: %v", err)
	}

	// Log analytics
	utils.LogAnalytics(userID, "vpn_disconnect", fmt.Sprintf("peer=%s", peerID))

	return nil
}

// GetUserStatus gets the VPN status for a user
func (vm *VPNManager) GetUserStatus(userID string) (map[string]interface{}, error) {
	// List peers
	peers, err := vm.peerManager.ListPeers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list peers: %v", err)
	}

	// Check if user has any peers
	if len(peers) == 0 {
		return map[string]interface{}{
			"connected": false,
		}, nil
	}

	// Get connection status
	status, err := vm.getConnectionStatus(peers[0])
	if err != nil {
		return nil, fmt.Errorf("failed to get connection status: %v", err)
	}

	return status, nil
}

// applyConfiguration applies the WireGuard configuration
func (vm *VPNManager) applyConfiguration() error {
	// Check if WireGuard is running in a container
	inContainer := os.Getenv("CONTAINER") == "1"

	// Command to reload WireGuard
	var cmd *exec.Cmd
	if inContainer {
		// In container, use wg-quick
		cmd = exec.Command("wg-quick", "down", vm.config.WireGuard.Interface)
		cmd.Run() // Ignore errors, interface might not be up
		cmd = exec.Command("wg-quick", "up", vm.config.WireGuard.Interface)
	} else {
		// On host, use wg syncconf
		configPath := filepath.Join(vm.config.WireGuard.ConfigDir, vm.config.WireGuard.Interface+".conf")
		cmd = exec.Command("wg", "syncconf", vm.config.WireGuard.Interface, configPath)
	}

	// Run command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to apply configuration: %v, output: %s", err, output)
	}

	return nil
}

// getConnectionStatus gets the connection status for a peer
func (vm *VPNManager) getConnectionStatus(peer *wireguard.PeerConfig) (map[string]interface{}, error) {
	// Get WireGuard status
	cmd := exec.Command("wg", "show", vm.config.WireGuard.Interface)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get WireGuard status: %v", err)
	}

	// Parse output
	lines := strings.Split(string(output), "\n")
	connected := false
	bytesReceived := int64(0)
	bytesSent := int64(0)

	for i, line := range lines {
		if strings.Contains(line, peer.PublicKey) {
			connected = true
			// Try to get transfer stats from the next line
			if i+1 < len(lines) {
				transferLine := lines[i+1]
				if strings.Contains(transferLine, "transfer:") {
					parts := strings.Split(transferLine, " ")
					if len(parts) >= 3 {
						// Parse received bytes
						receivedStr := strings.TrimSuffix(parts[1], "B")
						bytesReceived = parseBytes(receivedStr)

						// Parse sent bytes
						sentStr := strings.TrimSuffix(parts[2], "B")
						bytesSent = parseBytes(sentStr)
					}
				}
			}
			break
		}
	}

	// Build status
	status := map[string]interface{}{
		"connected":     connected,
		"peerId":        peer.ID,
		"serverId":      peer.ServerID,
		"ip":            peer.IP,
		"bytesReceived": bytesReceived,
		"bytesSent":     bytesSent,
	}

	return status, nil
}

// parseBytes parses a byte string (e.g., "1.5KiB") to bytes
func parseBytes(s string) int64 {
	// This is a simple implementation, a real one would handle units properly
	return 0
}
