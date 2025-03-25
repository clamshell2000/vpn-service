package monitoring

import (
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/utils"
)

// MetricsCollector collects and exposes metrics for Prometheus
type MetricsCollector struct {
	config *config.Config
	mutex  sync.RWMutex

	// Metrics
	activeConnections prometheus.Gauge
	dataTransferred   prometheus.Counter
	registeredUsers   prometheus.Gauge
	activeServers     prometheus.Gauge
	serverStatus      *prometheus.GaugeVec
	apiRequests       *prometheus.CounterVec
	apiLatency        *prometheus.HistogramVec
	errors            *prometheus.CounterVec
}

// NewMetricsCollector creates a new metrics collector
func NewMetricsCollector(cfg *config.Config) *MetricsCollector {
	mc := &MetricsCollector{
		config: cfg,
		mutex:  sync.RWMutex{},
	}

	// Initialize metrics
	mc.activeConnections = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vpn_active_connections",
		Help: "The current number of active VPN connections",
	})

	mc.dataTransferred = promauto.NewCounter(prometheus.CounterOpts{
		Name: "vpn_data_transferred_bytes",
		Help: "The total amount of data transferred through the VPN in bytes",
	})

	mc.registeredUsers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vpn_registered_users",
		Help: "The total number of registered users",
	})

	mc.activeServers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "vpn_active_servers",
		Help: "The current number of active VPN servers",
	})

	mc.serverStatus = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "vpn_server_status",
		Help: "The status of each VPN server (1 = online, 0 = offline)",
	}, []string{"server_id", "server_name", "location"})

	mc.apiRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vpn_api_requests_total",
		Help: "The total number of API requests",
	}, []string{"method", "endpoint", "status"})

	mc.apiLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "vpn_api_request_duration_seconds",
		Help:    "The latency of API requests in seconds",
		Buckets: prometheus.DefBuckets,
	}, []string{"method", "endpoint"})

	mc.errors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "vpn_errors_total",
		Help: "The total number of errors",
	}, []string{"type", "source"})

	return mc
}

// StartMetricsServer starts the metrics server
func (mc *MetricsCollector) StartMetricsServer() {
	// Create metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	// Start server
	go func() {
		addr := ":9100"
		utils.LogInfo("Starting metrics server on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			utils.LogError("Failed to start metrics server: %v", err)
		}
	}()
}

// RecordConnection records a VPN connection
func (mc *MetricsCollector) RecordConnection(connected bool) {
	if connected {
		mc.activeConnections.Inc()
	} else {
		mc.activeConnections.Dec()
	}
}

// RecordDataTransferred records data transferred through the VPN
func (mc *MetricsCollector) RecordDataTransferred(bytes float64) {
	mc.dataTransferred.Add(bytes)
}

// SetRegisteredUsers sets the number of registered users
func (mc *MetricsCollector) SetRegisteredUsers(count float64) {
	mc.registeredUsers.Set(count)
}

// SetActiveServers sets the number of active servers
func (mc *MetricsCollector) SetActiveServers(count float64) {
	mc.activeServers.Set(count)
}

// SetServerStatus sets the status of a server
func (mc *MetricsCollector) SetServerStatus(serverID, serverName, location string, online bool) {
	status := 0.0
	if online {
		status = 1.0
	}
	mc.serverStatus.WithLabelValues(serverID, serverName, location).Set(status)
}

// RecordAPIRequest records an API request
func (mc *MetricsCollector) RecordAPIRequest(method, endpoint, status string) {
	mc.apiRequests.WithLabelValues(method, endpoint, status).Inc()
}

// ObserveAPILatency observes the latency of an API request
func (mc *MetricsCollector) ObserveAPILatency(method, endpoint string, duration time.Duration) {
	mc.apiLatency.WithLabelValues(method, endpoint).Observe(duration.Seconds())
}

// RecordError records an error
func (mc *MetricsCollector) RecordError(errorType, source string) {
	mc.errors.WithLabelValues(errorType, source).Inc()
}
