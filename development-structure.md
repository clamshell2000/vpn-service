# VPN Service Development Structure

## Overview

This document provides a comprehensive overview of the VPN service architecture, components, and development structure. It serves as an onboarding guide for new developers joining the project.

**Last Updated:** 2025-03-25

## System Architecture

The VPN service follows a modular, microservices-based architecture with the following key components:

```
┌─────────────────────────────────────────────────────────────────┐
│                      Client Devices                             │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Internet                                 │
└───────────────────────────────┬─────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                     AWS EC2 Instance                            │
│                                                                 │
│  ┌─────────────────┐      ┌──────────────┐     ┌────────────┐   │
│  │   Nginx Proxy   │──────►  API Service  │────►  Database  │   │
│  └────────┬────────┘      └──────┬───────┘     └────────────┘   │
│           │                      │                              │
│           │                      │                              │
│  ┌────────▼────────┐      ┌──────▼───────┐     ┌────────────┐   │
│  │  WireGuard VPN  │◄─────►  VPN Manager │────►    Redis    │   │
│  └─────────────────┘      └──────────────┘     └────────────┘   │
│                                                                 │
│  ┌─────────────────┐      ┌──────────────┐     ┌────────────┐   │
│  │   Prometheus    │◄─────►    Grafana   │────►Node Exporter│   │
│  └─────────────────┘      └──────────────┘     └────────────┘   │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

### Key Components

1. **WireGuard VPN Server**: Core VPN technology providing secure, encrypted connections
2. **API Service**: RESTful API for managing VPN connections, users, and peers
3. **Database Layer**: PostgreSQL for persistent storage of user and configuration data
4. **Redis Cache**: For session management and high-performance data caching
5. **Nginx Reverse Proxy**: Handles SSL termination and routes traffic to appropriate services
6. **Monitoring Stack**: Prometheus and Grafana for metrics collection and visualization

## Directory Structure

The project follows a well-organized directory structure:

```
/home/ubuntu/vpn-service/
├── backend/                  # Backend services and API
│   ├── api/                  # API handlers and routes
│   │   ├── auth/             # Authentication handlers
│   │   ├── middleware/       # API middleware (JWT, logging)
│   │   └── vpn/              # VPN connection handlers
│   ├── config/               # Configuration settings
│   ├── core/                 # Core business logic
│   ├── data/                 # Data storage
│   │   └── wg_configs/       # WireGuard configurations
│   │       ├── peer1/        # Individual peer configurations
│   │       ├── templates/    # Configuration templates
│   │       └── wg_confs/     # Server configurations
│   ├── db/                   # Database models and migrations
│   │   ├── migrations/       # Database migration files
│   │   └── models/           # Database models
│   ├── logs/                 # Application logs
│   ├── monitoring/           # Monitoring configurations
│   ├── tests/                # Test suite
│   │   ├── e2e/              # End-to-end tests
│   │   ├── mocks/            # Mock data for testing
│   │   └── unit/             # Unit tests
│   ├── utils/                # Utility functions
│   └── vpn/                  # VPN management logic
│       └── wireguard/        # WireGuard-specific code
│           ├── config_templates/ # WireGuard config templates
│           └── utils/        # WireGuard utilities
├── infrastructure/           # Infrastructure configuration
│   ├── docker/               # Docker configuration
│   │   └── docker-compose.yml # Main Docker Compose file
│   ├── monitoring/           # Monitoring configuration
│   │   ├── grafana/          # Grafana dashboards and config
│   │   │   ├── dashboards/   # Dashboard definitions
│   │   │   └── provisioning/ # Provisioning configuration
│   │   └── prometheus/       # Prometheus configuration
│   └── nginx/                # Nginx configuration
│       ├── conf.d/           # Nginx site configurations
│       └── ssl/              # SSL certificates
└── scripts/                  # Utility scripts
    ├── check-status.sh       # Service status checker
    ├── create-peer.sh        # WireGuard peer creation
    ├── maintenance.sh        # Maintenance tasks
    └── troubleshoot.sh       # Troubleshooting utilities
```

## Core Backend Components

### VPN Manager (`backend/core/vpn_manager.go`)

The VPN Manager is responsible for:
- Managing VPN connections and disconnections
- Tracking active connections
- Handling connection statistics
- Managing connection pools

### Server Manager (`backend/core/server_manager.go`)

The Server Manager handles:
- Load balancing between multiple VPN servers (if configured)
- Server health monitoring
- Server provisioning and deprovisioning
- Routing clients to appropriate servers

### WireGuard Peer Manager (`backend/vpn/wireguard/peer_manager.go`)

The Peer Manager is responsible for:
- Creating new WireGuard peers
- Managing peer configurations
- Handling peer lifecycle (creation, update, deletion)
- Generating configuration files and QR codes for client setup

## API Endpoints

The API follows RESTful principles and is organized into logical groups:

### Authentication Endpoints

```
POST /api/auth/login         # Authenticate user and get JWT token
POST /api/auth/register      # Register new user
POST /api/auth/refresh       # Refresh JWT token
GET  /api/auth/profile       # Get user profile
PUT  /api/auth/profile       # Update user profile
DELETE /api/auth/logout      # Logout and invalidate token
```

### VPN Management Endpoints

```
GET  /api/vpn/status         # Get VPN connection status
POST /api/vpn/connect        # Connect to VPN
POST /api/vpn/disconnect     # Disconnect from VPN
GET  /api/vpn/peers          # List all peers
POST /api/vpn/peers          # Create new peer
GET  /api/vpn/peers/{id}     # Get peer details
DELETE /api/vpn/peers/{id}   # Delete peer
GET  /api/vpn/peers/{id}/config   # Get peer configuration
GET  /api/vpn/peers/{id}/qrcode   # Get peer QR code
```

### Server Management Endpoints

```
GET  /api/servers            # List all servers
GET  /api/servers/status     # Get server status
POST /api/servers            # Add new server
PUT  /api/servers/{id}       # Update server
DELETE /api/servers/{id}     # Remove server
```

### Health & Monitoring Endpoints

```
GET  /api/health             # Basic health check
GET  /api/health/detailed    # Detailed health information
```

## Docker Infrastructure

The entire application is containerized using Docker, with services defined in `infrastructure/docker/docker-compose.yml`:

### Services

1. **API Service (`vpn-api`)**
   - Image: `nginx:alpine`
   - Ports: 8080:80
   - Serves the API endpoints

2. **WireGuard VPN (`vpn-wireguard`)**
   - Image: `linuxserver/wireguard:latest`
   - Ports: 51820:51820/udp
   - Provides VPN connectivity
   - Environment variables:
     - SERVERURL=54.254.241.55
     - SERVERPORT=51820
     - PEERS=1
     - PEERDNS=1.1.1.1,8.8.8.8

3. **Database (`vpn-db`)**
   - Image: `postgres:14`
   - Stores user and configuration data
   - Environment variables:
     - POSTGRES_USER=postgres
     - POSTGRES_PASSWORD=postgres
     - POSTGRES_DB=vpn_service

4. **Redis (`vpn-redis`)**
   - Image: `redis:alpine`
   - Provides caching and session management

5. **Nginx Proxy (`vpn-nginx`)**
   - Image: `nginx:latest`
   - Ports: 80:80, 443:443
   - Handles SSL termination and request routing

6. **Prometheus (`vpn-prometheus`)**
   - Image: `prom/prometheus:latest`
   - Ports: 9090:9090
   - Collects metrics from services

7. **Grafana (`vpn-grafana`)**
   - Image: `grafana/grafana:latest`
   - Ports: 3000:3000
   - Provides visualization of metrics
   - Credentials: admin/admin

8. **Node Exporter (`vpn-node-exporter`)**
   - Image: `prom/node-exporter:latest`
   - Collects host-level metrics

9. **Redis Exporter (`vpn-redis-exporter`)**
   - Image: `oliver006/redis_exporter:latest`
   - Collects Redis metrics

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

## Monitoring & Logging

### Prometheus Configuration

Prometheus is configured to scrape metrics from various services:
- API service
- WireGuard
- Nginx
- Redis
- Node Exporter

The configuration is defined in `infrastructure/monitoring/prometheus/prometheus.yml`.

### Grafana Dashboards

Grafana is provisioned with pre-configured dashboards:
- VPN Service Dashboard: Shows VPN connection statistics
- Node Exporter Dashboard: Shows host metrics
- Redis Dashboard: Shows Redis performance metrics

Dashboard definitions are stored in `infrastructure/monitoring/grafana/dashboards/`.

### Logging

Logs are collected in the following locations:
- API logs: `/backend/logs/api.log`
- Access logs: `/backend/logs/access.log`
- Error logs: `/backend/logs/error.log`
- Nginx logs: `/backend/logs/nginx/`

## Management Scripts

### Check Status Script (`scripts/check-status.sh`)

This script checks the status of all services and provides a comprehensive health report:
- Container status
- Service accessibility
- WireGuard interface configuration
- Peer connections

### Create Peer Script (`scripts/create-peer.sh`)

This script simplifies the creation of new WireGuard peers:
- Creates a new peer with the specified name
- Generates configuration files
- Creates QR code for mobile setup
- Updates WireGuard configuration

### Maintenance Script (`scripts/maintenance.sh`)

This script performs routine maintenance tasks:
- Checks for and applies updates
- Rotates logs
- Performs database maintenance
- Checks disk space

### Troubleshooting Script (`scripts/troubleshoot.sh`)

This script helps diagnose and fix common issues:
- Checks for common configuration problems
- Verifies network connectivity
- Tests database connections
- Validates WireGuard configuration

## Security Considerations

### Network Security

- WireGuard provides secure, encrypted VPN connections
- Nginx handles SSL termination with modern cipher suites
- Firewall rules restrict access to sensitive ports
- IP-based restrictions for administrative interfaces

### Authentication & Authorization

- JWT-based authentication for API access
- Role-based access control for administrative functions
- Secure password storage with bcrypt hashing
- Token expiration and refresh mechanisms

### Data Protection

- Database encryption for sensitive data
- Secure storage of WireGuard keys
- Regular backups of critical data
- Proper handling of user data in compliance with privacy regulations

## Development Workflow

### Setting Up Development Environment

1. Clone the repository
2. Install Docker and Docker Compose
3. Run `docker-compose up -d` from the `infrastructure/docker` directory
4. Access the API at http://localhost:8080
5. Access Grafana at http://localhost:3000 (admin/admin)

### Making Changes

1. Follow the modular architecture when adding new features
2. Add appropriate tests for new functionality
3. Update documentation to reflect changes
4. Submit changes for code review

### Testing

1. Run unit tests with `go test ./...`
2. Run integration tests with `go test -tags=integration ./...`
3. Manually test changes using the provided scripts

## Deployment

### Production Deployment

1. Update the `SERVERURL` environment variable to the production server's public IP
2. Ensure all security groups and firewall rules are properly configured
3. Deploy using `docker-compose up -d`
4. Verify deployment with the check-status script

### Scaling Considerations

- The architecture supports horizontal scaling by adding more VPN servers
- Load balancing is handled by the Server Manager
- Database can be scaled separately if needed
- Redis can be configured for high availability

## Troubleshooting

### Common Issues

1. **VPN Connection Issues**
   - Check WireGuard configuration
   - Verify firewall rules
   - Ensure the server URL is correct

2. **API Access Issues**
   - Check Nginx configuration
   - Verify JWT authentication
   - Check database connectivity

3. **Monitoring Issues**
   - Verify Prometheus configuration
   - Check Grafana datasource configuration
   - Ensure metrics exporters are running

## Contact Information

For additional support or questions, please contact:
- VPN System Administrator: admin@vpnservice.com
- Backend Development Lead: dev@vpnservice.com
