package core

import (
	"fmt"
	"sync"

	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// Server represents a VPN server
type Server struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Location string `json:"location"`
	IP       string `json:"ip"`
	Status   string `json:"status"`
	Load     int    `json:"load"`
}

// ServerManager manages VPN servers
type ServerManager struct {
	config  *config.Config
	servers map[string]*Server
	mutex   sync.RWMutex
}

// NewServerManager creates a new server manager
func NewServerManager(cfg *config.Config) *ServerManager {
	return &ServerManager{
		config:  cfg,
		servers: make(map[string]*Server),
		mutex:   sync.RWMutex{},
	}
}

// LoadServers loads servers from configuration
func (sm *ServerManager) LoadServers() error {
	// In a real implementation, this would load servers from a database
	// For now, we'll create some mock servers
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Clear existing servers
	sm.servers = make(map[string]*Server)

	// Add mock servers
	sm.servers["server-1"] = &Server{
		ID:       "server-1",
		Name:     "US East",
		Location: "New York",
		IP:       "us-east.vpn-service.com",
		Status:   "online",
		Load:     0,
	}

	sm.servers["server-2"] = &Server{
		ID:       "server-2",
		Name:     "US West",
		Location: "San Francisco",
		IP:       "us-west.vpn-service.com",
		Status:   "online",
		Load:     0,
	}

	sm.servers["server-3"] = &Server{
		ID:       "server-3",
		Name:     "Europe",
		Location: "Amsterdam",
		IP:       "eu.vpn-service.com",
		Status:   "online",
		Load:     0,
	}

	utils.LogInfo("Loaded %d servers", len(sm.servers))
	return nil
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

// UpdateServerStatus updates a server's status
func (sm *ServerManager) UpdateServerStatus(id, status string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	server, ok := sm.servers[id]
	if !ok {
		return fmt.Errorf("server not found: %s", id)
	}

	server.Status = status
	utils.LogInfo("Updated server status: %s -> %s", id, status)
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
	utils.LogInfo("Updated server load: %s -> %d", id, load)
	return nil
}

// GetBestServer gets the best server based on load
func (sm *ServerManager) GetBestServer() (*Server, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	if len(sm.servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}

	var bestServer *Server
	bestLoad := -1

	for _, server := range sm.servers {
		if server.Status != "online" {
			continue
		}

		if bestLoad == -1 || server.Load < bestLoad {
			bestServer = server
			bestLoad = server.Load
		}
	}

	if bestServer == nil {
		return nil, fmt.Errorf("no online servers available")
	}

	return bestServer, nil
}

// AddServer adds a new server
func (sm *ServerManager) AddServer(server *Server) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	sm.servers[server.ID] = server
	utils.LogInfo("Added server: %s (%s)", server.ID, server.Name)
}

// UpdateServer updates an existing server
func (sm *ServerManager) UpdateServer(server *Server) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, ok := sm.servers[server.ID]; !ok {
		return fmt.Errorf("server not found: %s", server.ID)
	}

	sm.servers[server.ID] = server
	utils.LogInfo("Updated server: %s (%s)", server.ID, server.Name)
	return nil
}

// DeleteServer deletes a server
func (sm *ServerManager) DeleteServer(id string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if _, ok := sm.servers[id]; !ok {
		return fmt.Errorf("server not found: %s", id)
	}

	delete(sm.servers, id)
	utils.LogInfo("Deleted server: %s", id)
	return nil
}
