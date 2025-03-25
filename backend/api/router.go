package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/vpn-service/backend/api/admin"
	"github.com/vpn-service/backend/api/auth"
	"github.com/vpn-service/backend/api/health"
	"github.com/vpn-service/backend/api/middleware"
	"github.com/vpn-service/backend/api/servers"
	"github.com/vpn-service/backend/api/vpn"
	"github.com/vpn-service/backend/monitoring"
	"github.com/vpn-service/backend/src/config"
	"github.com/vpn-service/backend/src/core"
	"github.com/vpn-service/backend/src/utils"
)

// Router is the API router
type Router struct {
	config          *config.Config
	router          *mux.Router
	userManager     *core.UserManager
	serverManager   *core.ServerManager
	vpnManager      *core.VPNManager
	metricsCollector *monitoring.MetricsCollector
}

// NewRouter creates a new API router
func NewRouter(cfg *config.Config, userManager *core.UserManager, serverManager *core.ServerManager, vpnManager *core.VPNManager, metricsCollector *monitoring.MetricsCollector) *Router {
	return &Router{
		config:          cfg,
		router:          mux.NewRouter(),
		userManager:     userManager,
		serverManager:   serverManager,
		vpnManager:      vpnManager,
		metricsCollector: metricsCollector,
	}
}

// Setup sets up the API router
func (r *Router) Setup() {
	// Set up middleware
	authMiddleware := middleware.NewAuthMiddleware(r.config)
	metricsMiddleware := middleware.NewMetricsMiddleware(r.metricsCollector)

	// Set up global middleware
	r.router.Use(metricsMiddleware.Middleware)

	// Set up managers
	auth.UserManager = r.userManager
	servers.ServerManager = r.serverManager
	admin.UserManager = r.userManager
	vpn.VPNManager = r.vpnManager

	// Health routes
	r.router.HandleFunc("/health", health.HealthHandler).Methods(http.MethodGet)
	r.router.HandleFunc("/readiness", health.ReadinessHandler).Methods(http.MethodGet)
	r.router.HandleFunc("/liveness", health.LivenessHandler).Methods(http.MethodGet)

	// Auth routes
	r.router.HandleFunc("/api/auth/register", auth.RegisterHandler).Methods(http.MethodPost)
	r.router.HandleFunc("/api/auth/login", auth.LoginHandler).Methods(http.MethodPost)
	r.router.HandleFunc("/api/auth/refresh", auth.RefreshHandler).Methods(http.MethodPost)

	// User routes (authenticated)
	userRouter := r.router.PathPrefix("/api/user").Subrouter()
	userRouter.Use(authMiddleware.Middleware)
	userRouter.HandleFunc("", auth.GetUserHandler).Methods(http.MethodGet)
	userRouter.HandleFunc("/password", auth.ChangePasswordHandler).Methods(http.MethodPost)

	// VPN routes (authenticated)
	vpnRouter := r.router.PathPrefix("/api/vpn").Subrouter()
	vpnRouter.Use(authMiddleware.Middleware)
	vpnRouter.HandleFunc("/connect", vpn.ConnectHandler).Methods(http.MethodPost)
	vpnRouter.HandleFunc("/disconnect", vpn.DisconnectHandler).Methods(http.MethodPost)
	vpnRouter.HandleFunc("/status", vpn.StatusHandler).Methods(http.MethodGet)
	vpnRouter.HandleFunc("/config", vpn.GetConfigHandler).Methods(http.MethodGet)
	vpnRouter.HandleFunc("/config/qrcode", vpn.GetQRCodeHandler).Methods(http.MethodGet)
	vpnRouter.HandleFunc("/servers", vpn.GetServersHandler).Methods(http.MethodGet)

	// Admin routes (authenticated + admin)
	adminRouter := r.router.PathPrefix("/api/admin").Subrouter()
	adminRouter.Use(authMiddleware.AdminMiddleware)

	// Admin user routes
	adminRouter.HandleFunc("/users", admin.ListUsersHandler).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users/{id}", admin.GetUserHandler).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users/{id}", admin.UpdateUserHandler).Methods(http.MethodPut)
	adminRouter.HandleFunc("/users/{id}", admin.DeleteUserHandler).Methods(http.MethodDelete)
	adminRouter.HandleFunc("/users/{id}/peers", admin.GetUserPeersHandler).Methods(http.MethodGet)
	adminRouter.HandleFunc("/users/{id}/peers/{peerID}", admin.DeleteUserPeerHandler).Methods(http.MethodDelete)

	// Admin server routes
	adminRouter.HandleFunc("/servers", servers.ListServersHandler).Methods(http.MethodGet)
	adminRouter.HandleFunc("/servers/{id}", servers.GetServerHandler).Methods(http.MethodGet)
	adminRouter.HandleFunc("/servers", servers.CreateServerHandler).Methods(http.MethodPost)
	adminRouter.HandleFunc("/servers/{id}", servers.UpdateServerHandler).Methods(http.MethodPut)
	adminRouter.HandleFunc("/servers/{id}", servers.DeleteServerHandler).Methods(http.MethodDelete)
	adminRouter.HandleFunc("/servers/{id}/status/{status}", servers.UpdateServerStatusHandler).Methods(http.MethodPut)

	utils.LogInfo("API router setup complete")
}

// ServeHTTP implements the http.Handler interface
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.router.ServeHTTP(w, req)
}

// Start starts the API server
func (r *Router) Start() error {
	// Start server
	addr := r.config.APIAddr
	utils.LogInfo("Starting API server on %s", addr)
	return http.ListenAndServe(addr, r.router)
}
