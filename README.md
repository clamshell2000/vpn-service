# VPN Service

A complete VPN solution with user management, dynamic peer configuration, secure connections, and comprehensive monitoring.

## Features

- User registration and authentication with JWT
- Dynamic WireGuard peer management
- Server load balancing and health monitoring
- Device-specific configurations (iOS, Android, Windows, Mac, Generic)
- QR code generation for mobile devices
- Secure API endpoints with CORS support
- Comprehensive monitoring and metrics collection
- Grafana dashboards for visualization

## Project Structure

### Backend Components
- Core VPN logic: `backend/src/core`
- Utilities and helpers: `backend/src/utils`
- Configuration settings: `backend/src/config`
- Monitoring: `backend/src/monitoring`

### API
- Authentication handlers: `backend/api/auth`
- VPN connection handlers: `backend/api/vpn`
- JWT middleware: `backend/api/middleware`
- Metrics middleware: `backend/api/middleware`

### WireGuard Management
- Peer management: `backend/vpn/wireguard/peer_manager.go`
- Configuration templates: `backend/vpn/wireguard/config_templates`
- Utilities: `backend/vpn/wireguard/utils`

### Database
- Models: `backend/db/models`
- Migrations: `backend/db/migrations`
- Configurations: `backend/data/wg_configs`

### Infrastructure
- Docker configurations: `infrastructure/docker`
- Nginx configurations: `infrastructure/nginx`
- Monitoring configurations: `infrastructure/monitoring`

## Setup Instructions

1. Clone this repository
   ```bash
   git clone https://github.com/yourusername/vpn-service.git
   cd vpn-service
   ```

2. Configure environment variables (optional)
   ```bash
   cp .env.example .env
   # Edit .env file with your settings
   ```

3. Run the setup script as root
   ```bash
   sudo ./scripts/setup.sh
   ```

4. Start the services
   ```bash
   cd infrastructure/docker
   docker-compose up -d
   ```

5. Access the services
   - API: `https://localhost/api`
   - Prometheus: `http://localhost:9090`
   - Grafana: `http://localhost:3000` (default credentials: admin/admin)

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login and get JWT token
- `POST /api/auth/refresh` - Refresh JWT token

### VPN Management
- `GET /api/vpn/servers` - Get list of available VPN servers
- `POST /api/vpn/connect` - Connect to VPN
- `POST /api/vpn/disconnect` - Disconnect from VPN
- `GET /api/vpn/status` - Get connection status
- `GET /api/vpn/config` - Get WireGuard configuration
- `GET /api/vpn/qr` - Get QR code for configuration

## Monitoring

The VPN service includes comprehensive monitoring with Prometheus and Grafana:

### Metrics Collected
- Active connections
- Connection durations
- Data transferred (rx/tx)
- Server load
- API request counts and latencies
- Authentication errors
- Connection errors

### Dashboards
- VPN Overview - General service health and metrics
- Connection Statistics - Detailed connection metrics
- Server Performance - Server load and health metrics
- API Performance - API request metrics and errors

## Troubleshooting

### VPN Connectivity Issues
- Ensure the MASQUERADE rule is properly set up for your network interface
  ```bash
  sudo ./scripts/setup-iptables.sh
  ```
- Check DNS configuration in the WireGuard settings
- Verify that IP forwarding is enabled on the host
  ```bash
  cat /proc/sys/net/ipv4/ip_forward
  # Should output 1
  ```

### API Access Issues
- Check CORS configuration if accessing from client applications
- Verify JWT authentication is properly configured
- Check logs for detailed error messages
  ```bash
  docker logs vpn-api
  ```

### Monitoring Issues
- Ensure Prometheus can reach all targets
  ```bash
  curl http://localhost:9090/api/v1/targets
  ```
- Check Grafana datasource configuration
- Verify metrics are being collected
  ```bash
  curl http://localhost:8080/metrics
  ```

## License

MIT
