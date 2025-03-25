package monitoring

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// ServerMetrics represents metrics for a server
type ServerMetrics struct {
	ServerID      string    `json:"server_id"`
	CPU           float64   `json:"cpu"`
	Memory        float64   `json:"memory"`
	Bandwidth     float64   `json:"bandwidth"`
	Connections   int       `json:"connections"`
	LastUpdated   time.Time `json:"last_updated"`
	UptimeSeconds int64     `json:"uptime_seconds"`
}

// MetricsManager manages server metrics
type MetricsManager struct {
	config     *config.Config
	metrics    map[string]*ServerMetrics
	mutex      sync.RWMutex
	logFile    *os.File
	isEnabled  bool
	ticker     *time.Ticker
	done       chan bool
}

// NewMetricsManager creates a new metrics manager
func NewMetricsManager(cfg *config.Config) (*MetricsManager, error) {
	// Create metrics manager
	mm := &MetricsManager{
		config:    cfg,
		metrics:   make(map[string]*ServerMetrics),
		mutex:     sync.RWMutex{},
		isEnabled: cfg.Monitoring.EnableMetrics,
		done:      make(chan bool),
	}

	// If metrics is disabled, return early
	if !mm.isEnabled {
		utils.LogInfo("Metrics monitoring is disabled")
		return mm, nil
	}

	// Create log directory if it doesn't exist
	logDir := cfg.Monitoring.LogDir
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file
	logFilePath := filepath.Join(logDir, cfg.Monitoring.MetricsLogFile)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	mm.logFile = logFile
	utils.LogInfo("Metrics monitoring initialized, logging to %s", logFilePath)

	// Start metrics collection
	mm.startCollection()

	return mm, nil
}

// UpdateServerMetrics updates metrics for a server
func (mm *MetricsManager) UpdateServerMetrics(serverID string, cpu, memory, bandwidth float64, connections int) {
	// If metrics is disabled, return early
	if !mm.isEnabled {
		return
	}

	// Update metrics
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	metrics, ok := mm.metrics[serverID]
	if !ok {
		// Create new metrics
		metrics = &ServerMetrics{
			ServerID:      serverID,
			UptimeSeconds: 0,
		}
		mm.metrics[serverID] = metrics
	}

	// Update metrics
	metrics.CPU = cpu
	metrics.Memory = memory
	metrics.Bandwidth = bandwidth
	metrics.Connections = connections
	metrics.LastUpdated = time.Now()

	// Log metrics
	mm.logMetrics(metrics)
}

// GetServerMetrics gets metrics for a server
func (mm *MetricsManager) GetServerMetrics(serverID string) (*ServerMetrics, error) {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	metrics, ok := mm.metrics[serverID]
	if !ok {
		return nil, fmt.Errorf("no metrics for server: %s", serverID)
	}

	return metrics, nil
}

// GetAllServerMetrics gets metrics for all servers
func (mm *MetricsManager) GetAllServerMetrics() map[string]*ServerMetrics {
	mm.mutex.RLock()
	defer mm.mutex.RUnlock()

	// Copy metrics
	metrics := make(map[string]*ServerMetrics)
	for id, m := range mm.metrics {
		metrics[id] = m
	}

	return metrics
}

// Close closes the metrics manager
func (mm *MetricsManager) Close() error {
	// If metrics is disabled, return early
	if !mm.isEnabled {
		return nil
	}

	// Stop collection
	mm.stopCollection()

	// Close log file
	if mm.logFile != nil {
		return mm.logFile.Close()
	}

	return nil
}

// startCollection starts metrics collection
func (mm *MetricsManager) startCollection() {
	// Start ticker
	mm.ticker = time.NewTicker(60 * time.Second)

	// Start collection goroutine
	go func() {
		for {
			select {
			case <-mm.ticker.C:
				mm.collectMetrics()
			case <-mm.done:
				return
			}
		}
	}()
}

// stopCollection stops metrics collection
func (mm *MetricsManager) stopCollection() {
	// Stop ticker
	if mm.ticker != nil {
		mm.ticker.Stop()
	}

	// Signal done
	mm.done <- true
}

// collectMetrics collects metrics for all servers
func (mm *MetricsManager) collectMetrics() {
	// In a real implementation, this would collect metrics from the servers
	// For now, we'll just update the uptime
	mm.mutex.Lock()
	defer mm.mutex.Unlock()

	for _, metrics := range mm.metrics {
		metrics.UptimeSeconds += 60
	}
}

// logMetrics logs metrics to the log file
func (mm *MetricsManager) logMetrics(metrics *ServerMetrics) {
	// Marshal metrics to JSON
	data, err := json.Marshal(metrics)
	if err != nil {
		utils.LogError("Failed to marshal metrics: %v", err)
		return
	}

	// Write to log file
	if _, err := mm.logFile.Write(append(data, '\n')); err != nil {
		utils.LogError("Failed to write metrics to log file: %v", err)
	}
}
