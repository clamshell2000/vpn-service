#!/bin/bash

# Script to create a new WireGuard peer
# Author: Cascade
# Date: 2025-03-25

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

# Default values
PEER_NAME=${1:-"client1"}

echo "==========================================="
echo "        WireGuard Peer Creation           "
echo "==========================================="

# Check if WireGuard container is running
if ! docker ps | grep -q vpn-wireguard; then
    echo -e "${RED}Error:${NC} WireGuard container is not running"
    echo "Please start the WireGuard container first"
    exit 1
fi

# Get current number of peers from the WireGuard container
CURRENT_PEERS=$(docker exec vpn-wireguard bash -c 'wg show | grep -c "peer:"' || echo "0")
NEW_PEERS=$((CURRENT_PEERS + 1))
echo -e "Current peers: ${YELLOW}$CURRENT_PEERS${NC}, creating new peer #${GREEN}$NEW_PEERS${NC} named '${GREEN}$PEER_NAME${NC}'"

# Create a new container with updated peer count
echo -e "Creating new WireGuard container with updated peer configuration..."
docker stop vpn-wireguard

# Run a new container with the updated peer count
docker run -d \
  --name=vpn-wireguard \
  --cap-add=NET_ADMIN \
  --cap-add=SYS_MODULE \
  -e PUID=1000 \
  -e PGID=1000 \
  -e TZ=UTC \
  -e SERVERURL=54.254.241.55 \
  -e SERVERPORT=51820 \
  -e PEERS=$NEW_PEERS \
  -e PEER${NEW_PEERS}_NAME=$PEER_NAME \
  -e PEERDNS=1.1.1.1,8.8.8.8 \
  -e INTERNAL_SUBNET=10.13.13.0 \
  -p 51820:51820/udp \
  -v /home/ubuntu/vpn-service/backend/data/wg_configs:/config \
  --sysctl="net.ipv4.conf.all.src_valid_mark=1" \
  --restart=unless-stopped \
  --network=docker_vpn-network \
  linuxserver/wireguard:latest

# Wait for the container to initialize
echo "Waiting for peer configuration to be generated..."
sleep 10

# Check if peer was created successfully
PEER_DIR="/home/ubuntu/vpn-service/backend/data/wg_configs/peer${NEW_PEERS}"
if [ -d "$PEER_DIR" ]; then
    echo -e "${GREEN}Peer created successfully!${NC}"
    
    # Show configuration file
    echo -e "\n${YELLOW}Configuration file:${NC}"
    cat "$PEER_DIR/peer${NEW_PEERS}.conf"
    
    # Create QR code for mobile devices
    echo -e "\n${YELLOW}QR code for mobile devices:${NC}"
    if command -v qrencode &> /dev/null; then
        qrencode -t ansiutf8 < "$PEER_DIR/peer${NEW_PEERS}.conf"
    else
        echo -e "${YELLOW}QR code generation skipped. Install qrencode package to enable this feature.${NC}"
    fi
    
    # Copy configuration to a more accessible location with a friendly name
    mkdir -p /home/ubuntu/vpn-service/backend/data/peers
    cp -r "$PEER_DIR" "/home/ubuntu/vpn-service/backend/data/peers/${PEER_NAME}"
    cp "$PEER_DIR/peer${NEW_PEERS}.conf" "/home/ubuntu/vpn-service/backend/data/peers/${PEER_NAME}/${PEER_NAME}.conf"
    
    echo -e "\nConfiguration copied to: ${GREEN}/home/ubuntu/vpn-service/backend/data/peers/${PEER_NAME}/${PEER_NAME}.conf${NC}"
    echo -e "You can use this configuration to connect to the VPN from any WireGuard client"
else
    echo -e "${RED}Failed to create peer${NC}"
    echo "Check the WireGuard container logs for more information"
fi

echo -e "\n==========================================="
