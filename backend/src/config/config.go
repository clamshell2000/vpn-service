package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	Server     ServerConfig     `json:"server"`
	Database   DatabaseConfig   `json:"database"`
	JWT        JWTConfig        `json:"jwt"`
	WireGuard  WireGuardConfig  `json:"wireguard"`
	Monitoring MonitoringConfig `json:"monitoring"`
	APIAddr    string           `json:"apiAddr"`
}

// ServerConfig holds the server configuration
type ServerConfig struct {
	Port int    `json:"port"`
	Host string `json:"host"`
}

// DatabaseConfig holds the database configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// JWTConfig holds the JWT configuration
type JWTConfig struct {
	Secret     string `json:"secret"`
	Expiration int    `json:"expiration"` // in hours
}

// WireGuardConfig holds the WireGuard configuration
type WireGuardConfig struct {
	ConfigDir      string `json:"configDir"`
	DynamicPeerDir string `json:"dynamicPeerDir"`
	Interface      string `json:"interface"`
	ListenPort     int    `json:"listenPort"`
	PrivateKey     string `json:"privateKey"`
	PublicKey      string `json:"publicKey"`
	Address        string `json:"address"`
	DNS            string `json:"dns"`
	ServerIP       string `json:"serverIp"`
	ServerEndpoint string `json:"serverEndpoint"`
	AllowedIPs     string `json:"allowedIps"`
	MTU            int    `json:"mtu"`
	PreUp          string `json:"preUp"`
	PostUp         string `json:"postUp"`
	PreDown        string `json:"preDown"`
	PostDown       string `json:"postDown"`
}

// MonitoringConfig holds the monitoring configuration
type MonitoringConfig struct {
	LogDir           string `json:"logDir"`
	EnableAnalytics  bool   `json:"enableAnalytics"`
	AnalyticsLogFile string `json:"analyticsLogFile"`
	MetricsPort      int    `json:"metricsPort"`
	EnablePrometheus bool   `json:"enablePrometheus"`
}

// Load loads the configuration from the config file
func Load() (*Config, error) {
	// Default configuration
	config := &Config{
		APIAddr: "0.0.0.0:8080",
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Host: "localhost",
			Port: 5432,
			User: "postgres",
			Name: "vpn_service",
		},
		JWT: JWTConfig{
			Secret:     "change-me-in-production",
			Expiration: 24,
		},
		WireGuard: WireGuardConfig{
			ConfigDir:      "/etc/wireguard",
			DynamicPeerDir: "/etc/wireguard/dynamic-peers",
			Interface:      "wg0",
			ListenPort:     51820,
			Address:        "10.0.0.1/24",
			DNS:            "1.1.1.1,8.8.8.8",
			ServerIP:       "10.0.0.1",
			ServerEndpoint: "vpn.example.com",
			AllowedIPs:     "0.0.0.0/0, ::/0",
			MTU:            1420,
			PreUp:          "",
			PostUp:         "iptables -A FORWARD -i %i -j ACCEPT; iptables -t nat -A POSTROUTING -o eth0 -j MASQUERADE",
			PreDown:        "",
			PostDown:       "iptables -D FORWARD -i %i -j ACCEPT; iptables -t nat -D POSTROUTING -o eth0 -j MASQUERADE",
		},
		Monitoring: MonitoringConfig{
			LogDir:           "logs",
			EnableAnalytics:  true,
			AnalyticsLogFile: "logs/usage_analytics.log",
			MetricsPort:      9090,
			EnablePrometheus: true,
		},
	}

	// Check if config file exists
	configPath := getConfigPath()
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Create default config file
		return createDefaultConfig(configPath, config)
	}

	// Read config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

// getConfigPath returns the path to the config file
func getConfigPath() string {
	// Check if config path is set in environment variable
	configPath := os.Getenv("VPN_CONFIG_PATH")
	if configPath != "" {
		return configPath
	}

	// Default config path
	return filepath.Join("config", "config.json")
}

// createDefaultConfig creates a default config file
func createDefaultConfig(path string, config *Config) (*Config, error) {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}

	// Create config file
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Write config to file
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return nil, err
	}

	return config, nil
}
