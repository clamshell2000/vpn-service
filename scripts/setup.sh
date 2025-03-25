#!/bin/bash

# VPN Service Setup Script
# This script sets up the VPN service and configures the necessary components

# Set up color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Function to print colored messages
print_message() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    print_error "Please run as root"
    exit 1
fi

# Detect network interface
INTERFACE=$(ip route | grep default | awk '{print $5}')
if [ -z "$INTERFACE" ]; then
    print_error "Could not detect network interface"
    exit 1
fi

print_message "Detected network interface: $INTERFACE"

# Update wg0.conf with the correct interface
print_message "Updating WireGuard configuration with the correct interface..."
sed -i "s/enX0/$INTERFACE/g" /home/ubuntu/vpn-service/backend/data/wg_configs/wg0.conf

# Generate WireGuard keys
print_message "Generating WireGuard keys..."
if ! command -v wg &> /dev/null; then
    print_message "Installing WireGuard tools..."
    apt-get update
    apt-get install -y wireguard-tools
fi

# Generate server keys
PRIVATE_KEY=$(wg genkey)
PUBLIC_KEY=$(echo $PRIVATE_KEY | wg pubkey)

# Update wg0.conf with the server private key
print_message "Updating WireGuard configuration with server keys..."
sed -i "s/SERVER_PRIVATE_KEY_PLACEHOLDER/$PRIVATE_KEY/g" /home/ubuntu/vpn-service/backend/data/wg_configs/wg0.conf

# Create config.json with the server public key
print_message "Creating configuration file..."
mkdir -p /home/ubuntu/vpn-service/backend/config
cat > /home/ubuntu/vpn-service/backend/config/config.json << EOF
{
  "server": {
    "port": 8080,
    "host": "0.0.0.0"
  },
  "database": {
    "host": "db",
    "port": 5432,
    "user": "postgres",
    "password": "postgres",
    "name": "vpn_service"
  },
  "jwt": {
    "secret": "$(openssl rand -hex 32)",
    "expiration": 24
  },
  "wireguard": {
    "configDir": "/etc/wireguard",
    "dynamicPeerDir": "/etc/wireguard/dynamic-peers",
    "interface": "wg0",
    "listenPort": 51820,
    "privateKey": "$PRIVATE_KEY",
    "publicKey": "$PUBLIC_KEY",
    "address": "10.0.0.1/24",
    "dns": "1.1.1.1,8.8.8.8"
  },
  "monitoring": {
    "logDir": "logs",
    "enableAnalytics": true,
    "analyticsLogFile": "logs/usage_analytics.log"
  }
}
EOF

# Set up iptables rules
print_message "Setting up iptables rules..."
echo 1 > /proc/sys/net/ipv4/ip_forward
iptables -t nat -A POSTROUTING -s 10.0.0.0/24 -o $INTERFACE -j MASQUERADE

# Make IP forwarding persistent
print_message "Making IP forwarding persistent..."
echo "net.ipv4.ip_forward = 1" | tee -a /etc/sysctl.conf
sysctl -p

# Save iptables rules
print_message "Saving iptables rules..."
if command -v iptables-save > /dev/null; then
    mkdir -p /etc/iptables
    iptables-save > /etc/iptables/rules.v4
else
    print_message "Installing iptables-persistent..."
    apt-get update
    apt-get install -y iptables-persistent
    iptables-save > /etc/iptables/rules.v4
fi

# Create SSL certificates for Nginx
print_message "Creating SSL certificates for Nginx..."
mkdir -p /home/ubuntu/vpn-service/infrastructure/nginx/ssl
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
    -keyout /home/ubuntu/vpn-service/infrastructure/nginx/ssl/server.key \
    -out /home/ubuntu/vpn-service/infrastructure/nginx/ssl/server.crt \
    -subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"

# Make scripts executable
print_message "Making scripts executable..."
chmod +x /home/ubuntu/vpn-service/scripts/*.sh

# Start the services
print_message "Starting services..."
cd /home/ubuntu/vpn-service/infrastructure/docker
docker-compose up -d

print_message "Setup complete!"
print_message "VPN service is now running at http://localhost:8080"
print_message "WireGuard VPN is available at UDP port 51820"
print_message "Server public key: $PUBLIC_KEY"
