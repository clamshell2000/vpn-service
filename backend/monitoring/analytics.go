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

// AnalyticsEvent represents an analytics event
type AnalyticsEvent struct {
	ID        string                 `json:"id"`
	UserID    string                 `json:"user_id"`
	EventType string                 `json:"event_type"`
	Data      string                 `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// AnalyticsManager manages analytics
type AnalyticsManager struct {
	config    *config.Config
	events    []*AnalyticsEvent
	mutex     sync.RWMutex
	logFile   *os.File
	isEnabled bool
}

// NewAnalyticsManager creates a new analytics manager
func NewAnalyticsManager(cfg *config.Config) (*AnalyticsManager, error) {
	// Create analytics manager
	am := &AnalyticsManager{
		config:    cfg,
		events:    make([]*AnalyticsEvent, 0),
		mutex:     sync.RWMutex{},
		isEnabled: cfg.Monitoring.EnableAnalytics,
	}

	// If analytics is disabled, return early
	if !am.isEnabled {
		utils.LogInfo("Analytics is disabled")
		return am, nil
	}

	// Create log directory if it doesn't exist
	logDir := cfg.Monitoring.LogDir
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file
	logFilePath := filepath.Join(logDir, cfg.Monitoring.AnalyticsLogFile)
	logFile, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}

	am.logFile = logFile
	utils.LogInfo("Analytics initialized, logging to %s", logFilePath)

	return am, nil
}

// TrackEvent tracks an analytics event
func (am *AnalyticsManager) TrackEvent(userID, eventType, data string) {
	// If analytics is disabled, return early
	if !am.isEnabled {
		return
	}

	// Create event
	event := &AnalyticsEvent{
		ID:        utils.GenerateUUID(),
		UserID:    userID,
		EventType: eventType,
		Data:      data,
		Timestamp: time.Now(),
		Metadata:  make(map[string]interface{}),
	}

	// Add metadata
	event.Metadata["ip"] = "127.0.0.1" // In a real implementation, this would be the user's IP

	// Add event to list
	am.mutex.Lock()
	am.events = append(am.events, event)
	am.mutex.Unlock()

	// Log event
	am.logEvent(event)
}

// GetEvents gets all events
func (am *AnalyticsManager) GetEvents() []*AnalyticsEvent {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// Copy events
	events := make([]*AnalyticsEvent, len(am.events))
	copy(events, am.events)

	return events
}

// GetEventsByUser gets events for a user
func (am *AnalyticsManager) GetEventsByUser(userID string) []*AnalyticsEvent {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	// Filter events
	events := make([]*AnalyticsEvent, 0)
	for _, event := range am.events {
		if event.UserID == userID {
			events = append(events, event)
		}
	}

	return events
}

// Close closes the analytics manager
func (am *AnalyticsManager) Close() error {
	// If analytics is disabled, return early
	if !am.isEnabled {
		return nil
	}

	// Close log file
	if am.logFile != nil {
		return am.logFile.Close()
	}

	return nil
}

// logEvent logs an event to the log file
func (am *AnalyticsManager) logEvent(event *AnalyticsEvent) {
	// Marshal event to JSON
	data, err := json.Marshal(event)
	if err != nil {
		utils.LogError("Failed to marshal event: %v", err)
		return
	}

	// Write to log file
	if _, err := am.logFile.Write(append(data, '\n')); err != nil {
		utils.LogError("Failed to write event to log file: %v", err)
	}
}
