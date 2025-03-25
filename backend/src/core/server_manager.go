package core

import (
	"fmt"
	"sync"
	"time"

	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// Server represents a VPN server
type Server struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Country     string    `json:"country"`
	City        string    `json:"city"`
	IP          string    `json:"ip"`
	Load        int       `json:"load"`
	Capacity    int       `json:"capacity"`
	Status      string    `json:"status"`
	LastUpdated time.Time `json:"lastUpdated"`
}

// ServerManager manages VPN servers
type ServerManager struct {
	config  *config.Config
	servers map[string]*Server
	mutex   sync.RWMutex
}

// NewServerManager creates a new server manager
func NewServerManager(cfg *config.Config) *ServerManager {
	sm := &ServerManager{
		config:  cfg,
		servers: make(map[string]*Server),
		mutex:   sync.RWMutex{},
	}

	// Initialize with default servers
	sm.initializeServers()

	return sm
}

// initializeServers initializes the server list
func (sm *ServerManager) initializeServers() {
	// In a real implementation, this would load servers from a database
	// For now, we'll just add some mock servers
	servers := []*Server{
		{
			ID:          "us-east-1",
			Name:        "US East (N. Virginia)",
			Country:     "United States",
			City:        "Virginia",
			IP:          "192.168.1.1",
			Load:        0,
			Capacity:    100,
			Status:      "online",
			LastUpdated: time.Now(),
		},
		{
			ID:          "us-west-1",
			Name:        "US West (N. California)",
			Country:     "United States",
			City:        "California",
			IP:          "192.168.1.2",
			Load:        0,
			Capacity:    100,
			Status:      "online",
			LastUpdated: time.Now(),
		},
		{
			ID:          "eu-west-1",
			Name:        "EU (Ireland)",
			Country:     "Ireland",
			City:        "Dublin",
			IP:          "192.168.1.3",
			Load:        0,
			Capacity:    100,
			Status:      "online",
			LastUpdated: time.Now(),
		},
		{
			ID:          "ap-northeast-1",
			Name:        "Asia Pacific (Tokyo)",
			Country:     "Japan",
			City:        "Tokyo",
			IP:          "192.168.1.4",
			Load:        0,
			Capacity:    100,
			Status:      "maintenance",
			LastUpdated: time.Now(),
		},
	}

	// Add servers to map
	for _, server := range servers {
		sm.servers[server.ID] = server
	}
}

// GetServer gets a server by ID
func (sm *ServerManager) GetServer(id string) (*Server, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	server, ok := sm.servers[id]
	if !ok {
		return nil, fmt.Errorf("server not found: %s", id)
	}

	return server, nil
}

// GetServers gets all servers
func (sm *ServerManager) GetServers() []*Server {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	servers := make([]*Server, 0, len(sm.servers))
	for _, server := range sm.servers {
		servers = append(servers, server)
	}

	return servers
}

// GetServersByCountry gets servers by country
func (sm *ServerManager) GetServersByCountry(country string) []*Server {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	servers := make([]*Server, 0)
	for _, server := range sm.servers {
		if server.Country == country {
			servers = append(servers, server)
		}
	}

	return servers
}

// UpdateServerStatus updates a server's status
func (sm *ServerManager) UpdateServerStatus(id, status string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	server, ok := sm.servers[id]
	if !ok {
		return fmt.Errorf("server not found: %s", id)
	}

	server.Status = status
	server.LastUpdated = time.Now()

	// Log analytics
	utils.LogAnalytics("system", "server_status_update", fmt.Sprintf("server=%s status=%s", id, status))

	return nil
}

// UpdateServerLoad updates a server's load
func (sm *ServerManager) UpdateServerLoad(id string, load int) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	server, ok := sm.servers[id]
	if !ok {
		return fmt.Errorf("server not found: %s", id)
	}

	server.Load = load
	server.LastUpdated = time.Now()

	return nil
}

// GetOptimalServer gets the optimal server for a user
func (sm *ServerManager) GetOptimalServer(country string) (*Server, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var candidates []*Server

	// If country is specified, filter by country
	if country != "" {
		candidates = sm.GetServersByCountry(country)
		if len(candidates) == 0 {
			// If no servers in the requested country, fall back to all servers
			utils.LogWarning("No servers found in country %s, falling back to all servers", country)
			for _, server := range sm.servers {
				if server.Status == "online" {
					candidates = append(candidates, server)
				}
			}
		}
	} else {
		// Otherwise, consider all online servers
		for _, server := range sm.servers {
			if server.Status == "online" {
				candidates = append(candidates, server)
			}
		}
	}

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no available servers")
	}

	// Find the server with the lowest load
	var optimalServer *Server
	lowestLoad := -1

	for _, server := range candidates {
		// Skip servers at capacity
		if server.Load >= server.Capacity {
			continue
		}

		// Initialize or update if we find a server with lower load
		if lowestLoad == -1 || server.Load < lowestLoad {
			optimalServer = server
			lowestLoad = server.Load
		}
	}

	if optimalServer == nil {
		return nil, fmt.Errorf("all servers are at capacity")
	}

	return optimalServer, nil
}

// AddServer adds a new server
func (sm *ServerManager) AddServer(server *Server) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if server already exists
	if _, ok := sm.servers[server.ID]; ok {
		return fmt.Errorf("server already exists: %s", server.ID)
	}

	// Set last updated time
	server.LastUpdated = time.Now()

	// Add server
	sm.servers[server.ID] = server

	// Log analytics
	utils.LogAnalytics("system", "server_added", fmt.Sprintf("server=%s", server.ID))

	return nil
}

// RemoveServer removes a server
func (sm *ServerManager) RemoveServer(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Check if server exists
	if _, ok := sm.servers[id]; !ok {
		return fmt.Errorf("server not found: %s", id)
	}

	// Remove server
	delete(sm.servers, id)

	// Log analytics
	utils.LogAnalytics("system", "server_removed", fmt.Sprintf("server=%s", id))

	return nil
}

// MonitorServers periodically checks server status
func (sm *ServerManager) MonitorServers() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		sm.checkServerStatus()
	}
}

// checkServerStatus checks the status of all servers
func (sm *ServerManager) checkServerStatus() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	for id, server := range sm.servers {
		// In a real implementation, this would ping the server or check its health endpoint
		// For now, we'll just simulate a check
		if utils.RandomBool(0.95) { // 95% chance of being online
			if server.Status != "online" {
				server.Status = "online"
				server.LastUpdated = time.Now()
				utils.LogInfo("Server %s is now online", id)
			}
		} else {
			if server.Status != "offline" {
				server.Status = "offline"
				server.LastUpdated = time.Now()
				utils.LogWarning("Server %s is now offline", id)
			}
		}
	}
}
