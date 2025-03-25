package models

import (
	"time"
)

// VPNPeer represents a WireGuard VPN peer
type VPNPeer struct {
	ID         string    `json:"id" db:"id"`
	UserID     string    `json:"userId" db:"user_id"`
	ServerID   string    `json:"serverId" db:"server_id"`
	DeviceType string    `json:"deviceType" db:"device_type"`
	PublicKey  string    `json:"publicKey" db:"public_key"`
	PrivateKey string    `json:"-" db:"private_key"` // Private key is not included in JSON
	IP         string    `json:"ip" db:"ip"`
	Active     bool      `json:"active" db:"active"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
	LastSeen   time.Time `json:"lastSeen,omitempty" db:"last_seen"`
}

// NewVPNPeer creates a new VPN peer
func NewVPNPeer(userID, serverID, deviceType, publicKey, privateKey, ip string) *VPNPeer {
	now := time.Now()
	return &VPNPeer{
		ID:         generatePeerUUID(),
		UserID:     userID,
		ServerID:   serverID,
		DeviceType: deviceType,
		PublicKey:  publicKey,
		PrivateKey: privateKey,
		IP:         ip,
		Active:     true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// generatePeerUUID generates a UUID for a peer
func generatePeerUUID() string {
	// This is a placeholder. In a real implementation, use a proper UUID library
	return "peer-" + time.Now().Format("20060102150405")
}
