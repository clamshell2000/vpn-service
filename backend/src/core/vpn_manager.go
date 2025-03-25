package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
	"github.com/vpn-service/backend/vpn/wireguard"
)

// VPNManager manages VPN connections
type VPNManager struct {
	config        *config.Config
	serverManager *ServerManager
	peerManager   *wireguard.PeerManager
	mutex         sync.RWMutex
}

// NewVPNManager creates a new VPN manager
func NewVPNManager(cfg *config.Config, serverManager *ServerManager) *VPNManager {
	return &VPNManager{
		config:        cfg,
		serverManager: serverManager,
		peerManager:   wireguard.NewPeerManager(cfg),
		mutex:         sync.RWMutex{},
	}
}

// Connect connects a user to a VPN server
func (vm *VPNManager) Connect(userID, serverID, deviceType, deviceName string) (*wireguard.PeerConfig, string, error) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	// Get server
	server, err := vm.serverManager.GetServer(serverID)
	if err != nil {
		return nil, "", fmt.Errorf("server not found: %s", serverID)
	}

	// Check if server is online
	if server.Status != "online" {
		return nil, "", fmt.Errorf("server is not online: %s", serverID)
	}

	// Create peer
	peer, err := vm.peerManager.CreatePeer(userID, serverID, deviceType, deviceName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create peer: %v", err)
	}

	// Generate configuration
	config, err := vm.peerManager.GenerateConfig(peer)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate configuration: %v", err)
	}

	// Update server load
	vm.serverManager.UpdateServerLoad(serverID, server.Load+1)

	// Log analytics
	utils.LogAnalytics(userID, "vpn_connect", fmt.Sprintf("server=%s device=%s", serverID, deviceType))

	return peer, config, nil
}

// Disconnect disconnects a user from a VPN server
func (vm *VPNManager) Disconnect(userID, peerID string) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	// Get peer
	peer, err := vm.peerManager.GetPeer(userID, peerID)
	if err != nil {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	// Remove peer
	if err := vm.peerManager.RemovePeer(userID, peerID); err != nil {
		return fmt.Errorf("failed to remove peer: %v", err)
	}

	// Update server load
	vm.serverManager.UpdateServerLoad(peer.ServerID, 0)

	// Log analytics
	utils.LogAnalytics(userID, "vpn_disconnect", fmt.Sprintf("peer=%s", peerID))

	return nil
}

// GetStatus gets the status of a user's VPN connections
func (vm *VPNManager) GetStatus(userID string) ([]*wireguard.PeerInfo, error) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()

	// Get peers
	peers, err := vm.peerManager.GetPeers(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get peers: %v", err)
	}

	// Get peer info
	peerInfo := make([]*wireguard.PeerInfo, len(peers))
	for i, peer := range peers {
		// Get server
		server, err := vm.serverManager.GetServer(peer.ServerID)
		if err != nil {
			return nil, fmt.Errorf("server not found: %s", peer.ServerID)
		}

		// Create peer info
		peerInfo[i] = &wireguard.PeerInfo{
			ID:         peer.ID,
			ServerID:   peer.ServerID,
			ServerName: server.Name,
			DeviceType: peer.DeviceType,
			DeviceName: peer.DeviceName,
			IP:         peer.IP,
			CreatedAt:  peer.CreatedAt.Format(time.RFC3339),
			LastSeen:   time.Now().Format(time.RFC3339), // Mock for now
			BytesRx:    1024 * 1024 * 10,                // Mock for now
			BytesTx:    1024 * 1024 * 5,                 // Mock for now
		}
	}

	return peerInfo, nil
}

// GetConfig gets the configuration for a peer
func (vm *VPNManager) GetConfig(userID, peerID string) (string, error) {
	vm.mutex.RLock()
	defer vm.mutex.RUnlock()

	// Get peer
	peer, err := vm.peerManager.GetPeer(userID, peerID)
	if err != nil {
		return "", fmt.Errorf("peer not found: %s", peerID)
	}

	// Generate configuration
	config, err := vm.peerManager.GenerateConfig(peer)
	if err != nil {
		return "", fmt.Errorf("failed to generate configuration: %v", err)
	}

	return config, nil
}

// GetServers gets all VPN servers
func (vm *VPNManager) GetServers() []*Server {
	return vm.serverManager.GetServers()
}

// DynamicConnect connects a user to a VPN server with a dynamic IP
func (vm *VPNManager) DynamicConnect(userID, serverID, deviceType, deviceName string) (*wireguard.PeerConfig, string, error) {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	// Get server
	server, err := vm.serverManager.GetServer(serverID)
	if err != nil {
		return nil, "", fmt.Errorf("server not found: %s", serverID)
	}

	// Check if server is online
	if server.Status != "online" {
		return nil, "", fmt.Errorf("server is not online: %s", serverID)
	}

	// Create dynamic peer
	peer, err := vm.peerManager.CreateDynamicPeer(userID, serverID, deviceType, deviceName)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create dynamic peer: %v", err)
	}

	// Generate configuration
	config, err := vm.peerManager.GenerateConfig(peer)
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate configuration: %v", err)
	}

	// Update server load
	vm.serverManager.UpdateServerLoad(serverID, server.Load+1)

	// Log analytics
	utils.LogAnalytics(userID, "vpn_dynamic_connect", fmt.Sprintf("server=%s device=%s", serverID, deviceType))

	return peer, config, nil
}

// DynamicDisconnect disconnects a user from a VPN server with a dynamic IP
func (vm *VPNManager) DynamicDisconnect(userID, peerID string) error {
	vm.mutex.Lock()
	defer vm.mutex.Unlock()

	// Get peer
	peer, err := vm.peerManager.GetPeer(userID, peerID)
	if err != nil {
		return fmt.Errorf("peer not found: %s", peerID)
	}

	// Remove peer
	if err := vm.peerManager.RemoveDynamicPeer(userID, peerID); err != nil {
		return fmt.Errorf("failed to remove dynamic peer: %v", err)
	}

	// Update server load
	vm.serverManager.UpdateServerLoad(peer.ServerID, 0)

	// Log analytics
	utils.LogAnalytics(userID, "vpn_dynamic_disconnect", fmt.Sprintf("peer=%s", peerID))

	return nil
}
