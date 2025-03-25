# VPN Service - Important Information for Backend Developers

## System Overview
This document provides comprehensive information for backend developers to interact with the VPN service API and related components.

**Last Updated:** 2025-03-25

## Quick Reference

### Core Service URLs
- **VPN Server:** 54.254.241.55:51820/udp
- **API Base URL:** http://54.254.241.55:8080
- **Admin Panel:** https://54.254.241.55/api/admin
- **Health Check:** http://54.254.241.55:8080/api/health

### Monitoring URLs
**Note: These services may require SSH tunneling or VPN connection to access if they're not publicly exposed:**
- **Grafana (local access):** http://localhost:3000 (credentials: admin/admin)
- **Prometheus (local access):** http://localhost:9090
- **Node Exporter (local access):** http://localhost:9100
- **Redis Exporter (local access):** http://localhost:9121

### Remote Access to Monitoring Services
To access monitoring services from outside the server, use SSH tunneling:

```bash
# For Prometheus
ssh -L 9090:localhost:9090 ubuntu@54.254.241.55

# For Grafana
ssh -L 3000:localhost:3000 ubuntu@54.254.241.55

# Then access in your browser:
# http://localhost:9090 (Prometheus)
# http://localhost:3000 (Grafana)
```

Alternatively, connect to the VPN first, then access these services.

### Database Information
- **PostgreSQL (internal):** localhost:5432
  - Database: vpn_service
  - Username: postgres
  - Password: postgres
- **Redis (internal):** localhost:6379

## API Endpoints Reference

### Authentication
```
POST /api/auth/login
POST /api/auth/register
POST /api/auth/refresh
GET /api/auth/profile
PUT /api/auth/profile
DELETE /api/auth/logout
```

### VPN Management
```
GET /api/vpn/status
POST /api/vpn/connect
POST /api/vpn/disconnect
GET /api/vpn/peers
POST /api/vpn/peers
GET /api/vpn/peers/{id}
DELETE /api/vpn/peers/{id}
GET /api/vpn/peers/{id}/config
GET /api/vpn/peers/{id}/qrcode
```

### Server Management
```
GET /api/servers
GET /api/servers/status
POST /api/servers
PUT /api/servers/{id}
DELETE /api/servers/{id}
```

### Health & Monitoring
```
GET /api/health
GET /api/health/detailed
```

## Authentication

The API uses JWT for authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <token>
```

Authentication flow:
1. Call `/api/auth/login` with credentials to get a JWT token
2. Use the token in subsequent requests
3. Refresh the token with `/api/auth/refresh` when needed

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_login TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE
);
```

### VPN Peers Table
```sql
CREATE TABLE vpn_peers (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id),
    peer_name VARCHAR(50) NOT NULL,
    public_key VARCHAR(255) UNIQUE NOT NULL,
    allowed_ips VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_handshake TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE,
    data_received BIGINT DEFAULT 0,
    data_sent BIGINT DEFAULT 0
);
```

### Servers Table
```sql
CREATE TABLE servers (
    id SERIAL PRIMARY KEY,
    server_name VARCHAR(50) NOT NULL,
    server_url VARCHAR(255) NOT NULL,
    server_port INTEGER NOT NULL,
    location VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    current_load FLOAT DEFAULT 0.0,
    max_peers INTEGER DEFAULT 100,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## WireGuard Configuration

### Server Configuration
The WireGuard server is configured with the following parameters:
- Server URL: 54.254.241.55
- Port: 51820/udp
- Internal Subnet: 10.13.13.0/24
- Server IP: 10.13.13.1
- DNS: 1.1.1.1, 8.8.8.8

### Peer Configuration
Each peer is assigned an IP from the 10.13.13.0/24 subnet, starting from 10.13.13.2.

Sample peer configuration:
```
[Interface]
Address = 10.13.13.2
PrivateKey = <private_key>
ListenPort = 51820
DNS = 1.1.1.1,8.8.8.8

[Peer]
PublicKey = <server_public_key>
PresharedKey = <preshared_key>
Endpoint = 54.254.241.55:51820
AllowedIPs = 0.0.0.0/0, ::/0
```

## Docker Infrastructure

### Container Structure
- **vpn-api**: Nginx container serving the API
- **vpn-wireguard**: WireGuard VPN server
- **vpn-db**: PostgreSQL database
- **vpn-nginx**: Nginx reverse proxy
- **vpn-prometheus**: Prometheus metrics collection
- **vpn-grafana**: Grafana visualization
- **vpn-redis**: Redis cache
- **vpn-node-exporter**: Node metrics exporter
- **vpn-redis-exporter**: Redis metrics exporter

### Docker Compose Configuration
The complete Docker Compose configuration is available at `/home/ubuntu/vpn-service/infrastructure/docker/docker-compose.yml`. Here's a sample of the configuration:

```yml
version: '3.8'

services:
  # API service
  api:
    image: nginx:alpine
    container_name: vpn-api
    restart: unless-stopped
    ports:
      - "8080:80"
    volumes:
      - ../../backend/logs:/var/log/nginx
    networks:
      - vpn-network

  # WireGuard VPN service
  wireguard:
    image: linuxserver/wireguard:latest
    container_name: vpn-wireguard
    cap_add:
      - NET_ADMIN
      - SYS_MODULE
    environment:
      - PUID=1000
      - PGID=1000
      - TZ=UTC
      - SERVERURL=54.254.241.55
      - SERVERPORT=51820
      - PEERS=1
      - PEERDNS=1.1.1.1,8.8.8.8
      - PEER1_NAME=test-client
    volumes:
      - ../../backend/data/wg_configs:/config
    ports:
      - "51820:51820/udp"
    sysctls:
      - net.ipv4.conf.all.src_valid_mark=1
    restart: unless-stopped
    networks:
      - vpn-network

  # Other services omitted for brevity...

volumes:
  postgres_data:
  prometheus_data:
  grafana_data:
  redis_data:

networks:
  vpn-network:
    driver: bridge
```

To deploy the entire stack:
```bash
cd /home/ubuntu/vpn-service/infrastructure/docker
docker-compose up -d
```

### Network Configuration
All containers are connected to the `vpn-network` bridge network.

## Monitoring

### Grafana
- **URL (local access)**: http://localhost:3000
- **Username**: admin
- **Password**: admin
- **Dashboard**: VPN Service Dashboard

### Prometheus
- **URL (local access)**: http://localhost:9090
- **Metrics Path**: /metrics
- **Scrape Interval**: 15s

### Exposing Monitoring Services Publicly (Optional)
If you need to expose monitoring services publicly, update the Nginx configuration to add reverse proxy settings:

```nginx
# Add to /home/ubuntu/vpn-service/infrastructure/nginx/conf.d/default.conf

# Prometheus
server {
    listen 80;
    server_name prometheus.yourdomain.com;
    
    location / {
        proxy_pass http://vpn-prometheus:9090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

# Grafana
server {
    listen 80;
    server_name grafana.yourdomain.com;
    
    location / {
        proxy_pass http://vpn-grafana:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

Then update your DNS settings and restart Nginx:
```bash
docker restart vpn-nginx
```

### IP-Restricted Access to Monitoring Services
For security reasons, you can restrict access to monitoring services to specific IP addresses. This has been configured for Prometheus:

```nginx
# Prometheus access configuration - restricted to specific IP
server {
    listen 9090;
    server_name localhost;

    # Allow access only from the specified IP
    allow 116.106.201.170;
    deny all;

    location / {
        proxy_pass http://vpn-prometheus:9090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

With this configuration:
- Prometheus is accessible at http://54.254.241.55:9090 only from IP 116.106.201.170
- All other IPs will receive a 403 Forbidden response

To add or change allowed IPs, edit the Nginx configuration and restart the container:
```bash
docker restart vpn-nginx
```

### AWS Security Group Configuration for Prometheus Access

To allow external access to Prometheus (port 9090), you need to add a rule to your AWS security group:

1. Log in to the AWS Management Console
2. Navigate to EC2 > Security Groups
3. Select the security group associated with your VPN server instance
4. Click "Edit inbound rules"
5. Add a new rule with the following settings:
   - Type: Custom TCP
   - Protocol: TCP
   - Port range: 9090
   - Source: Custom IP (116.106.201.170/32)
   - Description: Prometheus access for specific IP
6. Click "Save rules"

This will allow traffic from your IP address (116.106.201.170) to reach port 9090 on the server.

### Prometheus Direct Access Configuration

If you're still having issues accessing Prometheus, you may need to modify the Prometheus configuration to bind to all network interfaces instead of just localhost:

1. Edit the Prometheus configuration:
```bash
nano /home/ubuntu/vpn-service/infrastructure/monitoring/prometheus/prometheus.yml
```

2. Add or modify the web configuration section:
```yaml
# Add this to the global section
web:
  listen_address: 0.0.0.0:9090
```

3. Restart the Prometheus container:
```bash
docker restart vpn-prometheus
```

## Management Scripts

### Create Peer
To create a new WireGuard peer:
```bash
./scripts/create-peer.sh <peer_name>
```

This script:
1. Checks if the WireGuard container is running
2. Gets the current number of peers
3. Creates a new peer with the specified name
4. Generates QR code for mobile device setup
5. Copies configuration to an accessible location

### Check Status
To check the status of all services:
```bash
./scripts/check-status.sh
```

This script verifies:
1. Docker service status
2. Container status for all services
3. Service accessibility on expected ports
4. WireGuard interface configuration
5. Peer connections

## Development Guidelines

### API Response Format
All API endpoints return responses in the following format:
```json
{
  "success": true|false,
  "data": { ... },
  "error": null|"error message",
  "timestamp": "2025-03-25T08:47:20Z"
}
```

### Error Handling
HTTP status codes:
- 200: Success
- 400: Bad request
- 401: Unauthorized
- 403: Forbidden
- 404: Not found
- 500: Internal server error

### Logging
Logs are stored in:
- API logs: /backend/logs/api.log
- Access logs: /backend/logs/access.log
- Error logs: /backend/logs/error.log
- Nginx logs: /backend/logs/nginx/

### Testing
- Unit tests: /backend/tests/unit/
- E2E tests: /backend/tests/e2e/
- Mock responses: /backend/tests/mocks/

## Security Considerations

### API Security
- All endpoints except /api/auth/login and /api/auth/register require authentication
- HTTPS is enforced for all API requests
- Rate limiting is applied to prevent brute force attacks
- Input validation is performed on all requests

### VPN Security
- WireGuard uses modern cryptography (Curve25519, ChaCha20, Poly1305)
- Each peer has unique keys
- Preshared keys add an additional layer of security
- Firewall rules restrict access to the VPN server

## Troubleshooting

### Common Issues
1. **API Connection Refused**
   - Check if the API container is running
   - Verify network configuration

2. **VPN Connection Timeout**
   - Ensure WireGuard port (51820/udp) is open
   - Check server endpoint configuration

3. **Database Connection Issues**
   - Verify PostgreSQL container is running
   - Check database credentials

4. **JWT Authentication Failures**
   - Token may be expired (use refresh endpoint)
   - Check if user is active in the database

### Debugging Tools
- Docker logs: `docker logs <container_name>`
- WireGuard status: `docker exec vpn-wireguard wg show`
- Database query: `docker exec vpn-db psql -U postgres -d vpn_service -c "SELECT * FROM users;"`

## Backup and Recovery

### Database Backup
```bash
docker exec vpn-db pg_dump -U postgres vpn_service > backup.sql
```

### WireGuard Configuration Backup
```bash
tar -czvf wg_backup.tar.gz /home/ubuntu/vpn-service/backend/data/wg_configs
```

## Contact Information

For additional support or questions, please contact:
- VPN System Administrator: admin@vpnservice.com
- Backend Development Lead: dev@vpnservice.com
