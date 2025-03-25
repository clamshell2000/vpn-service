package monitoring

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/core"
	"github.com/vpn-service/backend/src/utils"
	"github.com/vpn-service/backend/vpn/wireguard"
)

var (
	// MetricsCollector is the global metrics collector instance
	MetricsCollector *Collector
)

// Collector collects metrics for the VPN service
type Collector struct {
	config *config.Config
	mutex  sync.RWMutex

	// Prometheus metrics
	activeConnections      prometheus.Gauge
	totalConnections       prometheus.Counter
	connectionDurations    prometheus.Histogram
	dataTransferred        *prometheus.CounterVec
	connectionsPerServer   *prometheus.GaugeVec
	connectionsPerCountry  *prometheus.GaugeVec
	connectionsPerDevice   *prometheus.GaugeVec
	serverLoad             *prometheus.GaugeVec
	connectionErrors       prometheus.Counter
	authenticationErrors   prometheus.Counter
	configurationRequests  prometheus.Counter
	qrCodeRequests         prometheus.Counter
	apiRequestDuration     *prometheus.HistogramVec
	apiRequestCount        *prometheus.CounterVec
}

// NewCollector creates a new metrics collector
func NewCollector(cfg *config.Config) *Collector {
	collector := &Collector{
		config: cfg,
		mutex:  sync.RWMutex{},

		activeConnections: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "vpn_active_connections",
			Help: "Number of active VPN connections",
		}),

		totalConnections: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_total_connections",
			Help: "Total number of VPN connections",
		}),

		connectionDurations: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "vpn_connection_durations_seconds",
			Help:    "Histogram of VPN connection durations in seconds",
			Buckets: prometheus.ExponentialBuckets(60, 2, 10), // 1min to ~17hrs
		}),

		dataTransferred: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vpn_data_transferred_bytes",
				Help: "Amount of data transferred through the VPN in bytes",
			},
			[]string{"direction"}, // "rx" or "tx"
		),

		connectionsPerServer: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "vpn_connections_per_server",
				Help: "Number of active connections per server",
			},
			[]string{"server_id", "server_name"},
		),

		connectionsPerCountry: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "vpn_connections_per_country",
				Help: "Number of active connections per country",
			},
			[]string{"country"},
		),

		connectionsPerDevice: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "vpn_connections_per_device",
				Help: "Number of active connections per device type",
			},
			[]string{"device_type"},
		),

		serverLoad: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "vpn_server_load",
				Help: "Current load of VPN servers",
			},
			[]string{"server_id", "server_name"},
		),

		connectionErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_connection_errors_total",
			Help: "Total number of VPN connection errors",
		}),

		authenticationErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_authentication_errors_total",
			Help: "Total number of authentication errors",
		}),

		configurationRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_configuration_requests_total",
			Help: "Total number of configuration requests",
		}),

		qrCodeRequests: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "vpn_qr_code_requests_total",
			Help: "Total number of QR code requests",
		}),

		apiRequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "vpn_api_request_duration_seconds",
				Help:    "Histogram of API request durations in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status"},
		),

		apiRequestCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "vpn_api_requests_total",
				Help: "Total number of API requests",
			},
			[]string{"method", "endpoint", "status"},
		),
	}

	// Register metrics with Prometheus
	prometheus.MustRegister(
		collector.activeConnections,
		collector.totalConnections,
		collector.connectionDurations,
		collector.dataTransferred,
		collector.connectionsPerServer,
		collector.connectionsPerCountry,
		collector.connectionsPerDevice,
		collector.serverLoad,
		collector.connectionErrors,
		collector.authenticationErrors,
		collector.configurationRequests,
		collector.qrCodeRequests,
		collector.apiRequestDuration,
		collector.apiRequestCount,
	)

	return collector
}

// StartMetricsServer starts the metrics server
func (c *Collector) StartMetricsServer() {
	if !c.config.Monitoring.EnablePrometheus {
		utils.LogInfo("Prometheus metrics server disabled")
		return
	}

	// Start metrics server
	go func() {
		metricsAddr := fmt.Sprintf(":%d", c.config.Monitoring.MetricsPort)
		utils.LogInfo("Starting metrics server on %s", metricsAddr)
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(metricsAddr, nil); err != nil {
			utils.LogError("Failed to start metrics server: %v", err)
		}
	}()
}

// IncrementActiveConnections increments the active connections counter
func (c *Collector) IncrementActiveConnections() {
	c.activeConnections.Inc()
}

// DecrementActiveConnections decrements the active connections counter
func (c *Collector) DecrementActiveConnections() {
	c.activeConnections.Dec()
}

// IncrementTotalConnections increments the total connections counter
func (c *Collector) IncrementTotalConnections() {
	c.totalConnections.Inc()
}

// ObserveConnectionDuration observes a connection duration
func (c *Collector) ObserveConnectionDuration(duration time.Duration) {
	c.connectionDurations.Observe(duration.Seconds())
}

// AddDataTransferred adds data transferred
func (c *Collector) AddDataTransferred(direction string, bytes float64) {
	c.dataTransferred.WithLabelValues(direction).Add(bytes)
}

// SetConnectionsPerServer sets the number of connections for a server
func (c *Collector) SetConnectionsPerServer(serverID, serverName string, count float64) {
	c.connectionsPerServer.WithLabelValues(serverID, serverName).Set(count)
}

// SetConnectionsPerCountry sets the number of connections for a country
func (c *Collector) SetConnectionsPerCountry(country string, count float64) {
	c.connectionsPerCountry.WithLabelValues(country).Set(count)
}

// SetConnectionsPerDevice sets the number of connections for a device type
func (c *Collector) SetConnectionsPerDevice(deviceType string, count float64) {
	c.connectionsPerDevice.WithLabelValues(deviceType).Set(count)
}

// SetServerLoad sets the load for a server
func (c *Collector) SetServerLoad(serverID, serverName string, load float64) {
	c.serverLoad.WithLabelValues(serverID, serverName).Set(load)
}

// IncrementConnectionErrors increments the connection errors counter
func (c *Collector) IncrementConnectionErrors() {
	c.connectionErrors.Inc()
}

// IncrementAuthenticationErrors increments the authentication errors counter
func (c *Collector) IncrementAuthenticationErrors() {
	c.authenticationErrors.Inc()
}

// IncrementConfigurationRequests increments the configuration requests counter
func (c *Collector) IncrementConfigurationRequests() {
	c.configurationRequests.Inc()
}

// IncrementQRCodeRequests increments the QR code requests counter
func (c *Collector) IncrementQRCodeRequests() {
	c.qrCodeRequests.Inc()
}

// ObserveAPIRequestDuration observes an API request duration
func (c *Collector) ObserveAPIRequestDuration(method, endpoint, status string, duration time.Duration) {
	c.apiRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration.Seconds())
}

// IncrementAPIRequestCount increments the API request count
func (c *Collector) IncrementAPIRequestCount(method, endpoint, status string) {
	c.apiRequestCount.WithLabelValues(method, endpoint, status).Inc()
}

// UpdateMetrics updates all metrics
func (c *Collector) UpdateMetrics(servers []*core.Server, connections map[string][]*wireguard.PeerInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Count active connections
	totalActive := 0
	serverCounts := make(map[string]int)
	countryCounts := make(map[string]int)
	deviceCounts := make(map[string]int)

	for _, peers := range connections {
		totalActive += len(peers)

		for _, peer := range peers {
			// Increment server counts
			serverCounts[peer.ServerID]++

			// Find server to get country
			for _, server := range servers {
				if server.ID == peer.ServerID {
					countryCounts[server.Country]++
					break
				}
			}

			// Increment device counts
			deviceCounts[peer.DeviceType]++
		}
	}

	// Update active connections
	c.activeConnections.Set(float64(totalActive))

	// Update connections per server
	for _, server := range servers {
		count := float64(serverCounts[server.ID])
		c.SetConnectionsPerServer(server.ID, server.Name, count)
		c.SetServerLoad(server.ID, server.Name, float64(server.Load))
	}

	// Update connections per country
	for country, count := range countryCounts {
		c.SetConnectionsPerCountry(country, float64(count))
	}

	// Update connections per device
	for deviceType, count := range deviceCounts {
		c.SetConnectionsPerDevice(deviceType, float64(count))
	}
}
