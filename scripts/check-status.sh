#!/bin/bash

# Script to check the status of all VPN service components
# Author: Cascade
# Date: 2025-03-25

# Color definitions
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

echo "==========================================="
echo "        VPN Service Status Check           "
echo "==========================================="

# Function to check if a container is running
check_container() {
    local container_name=$1
    local status=$(docker inspect --format='{{.State.Status}}' $container_name 2>/dev/null)
    
    if [ "$status" == "running" ]; then
        echo -e "${GREEN}✓${NC} $container_name is ${GREEN}running${NC}"
        return 0
    elif [ -n "$status" ]; then
        echo -e "${RED}✗${NC} $container_name is ${RED}$status${NC}"
        return 1
    else
        echo -e "${RED}✗${NC} $container_name ${RED}does not exist${NC}"
        return 2
    fi
}

# Function to check if a TCP port is open
check_port() {
    local port=$1
    local service=$2
    local protocol=${3:-tcp}
    
    if [ "$protocol" == "tcp" ]; then
        if nc -z localhost $port >/dev/null 2>&1; then
            echo -e "${GREEN}✓${NC} $service is ${GREEN}accessible${NC} on port $port/$protocol"
            return 0
        else
            echo -e "${RED}✗${NC} $service is ${RED}not accessible${NC} on port $port/$protocol"
            return 1
        fi
    fi
}

# Check Docker service
if systemctl is-active --quiet docker; then
    echo -e "${GREEN}✓${NC} Docker service is ${GREEN}running${NC}"
else
    echo -e "${RED}✗${NC} Docker service is ${RED}not running${NC}"
    echo "Please start Docker service with: sudo systemctl start docker"
    exit 1
fi

echo -e "\n----- Container Status -----"
# Check all containers
containers=("vpn-api" "vpn-wireguard" "vpn-db" "vpn-nginx" "vpn-prometheus" "vpn-grafana" "vpn-redis" "vpn-node-exporter" "vpn-redis-exporter")
failed_containers=()

for container in "${containers[@]}"; do
    if ! check_container "$container"; then
        failed_containers+=("$container")
    fi
done

echo -e "\n----- Service Accessibility -----"
# Check service ports
check_port 8080 "API Service"
check_port 80 "Nginx HTTP"
check_port 443 "Nginx HTTPS"
check_port 9090 "Prometheus"
check_port 3000 "Grafana"
check_port 6379 "Redis"
check_port 9100 "Node Exporter"
check_port 9121 "Redis Exporter"

# Initialize failed_services array
failed_services=()

echo -e "\n----- WireGuard Status -----"
# Check WireGuard interfaces
if docker exec vpn-wireguard wg show >/dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} WireGuard interface is ${GREEN}configured${NC}"
    
    # Check if WireGuard is listening on port 51820
    LISTENING_PORT=$(docker exec vpn-wireguard wg show | grep "listening port" | awk '{print $3}')
    if [ "$LISTENING_PORT" == "51820" ]; then
        echo -e "${GREEN}✓${NC} WireGuard is ${GREEN}listening${NC} on port 51820/udp"
    else
        echo -e "${RED}✗${NC} WireGuard is ${RED}not listening${NC} on expected port 51820/udp (actual: $LISTENING_PORT)"
        failed_services+=("WireGuard VPN (port 51820)")
    fi
    
    echo -e "\nWireGuard Interface Details:"
    docker exec vpn-wireguard wg show | grep -v private
    
    # Check if there are any peers configured
    PEER_COUNT=$(docker exec vpn-wireguard wg show | grep -c "peer:")
    if [ $PEER_COUNT -gt 0 ]; then
        echo -e "\n${GREEN}✓${NC} WireGuard has ${GREEN}$PEER_COUNT${NC} peer(s) configured"
    else
        echo -e "\n${YELLOW}!${NC} WireGuard has ${YELLOW}no peers${NC} configured yet"
    fi
else
    echo -e "${RED}✗${NC} WireGuard interface is ${RED}not configured${NC}"
    failed_services+=("WireGuard Interface")
fi

echo -e "\n----- Summary -----"
if [ ${#failed_containers[@]} -eq 0 ] && [ ${#failed_services[@]} -eq 0 ]; then
    echo -e "${GREEN}All services are running correctly!${NC}"
else
    echo -e "${YELLOW}Some services have issues:${NC}"
    
    if [ ${#failed_containers[@]} -gt 0 ]; then
        echo -e "\nFailed containers:"
        for container in "${failed_containers[@]}"; do
            echo -e "  - ${RED}$container${NC}"
        done
    fi
    
    if [ ${#failed_services[@]} -gt 0 ]; then
        echo -e "\nFailed services:"
        for service in "${failed_services[@]}"; do
            echo -e "  - ${RED}$service${NC}"
        done
    fi
    
    echo -e "\nTroubleshooting tips:"
    echo "1. Check container logs with: docker logs <container-name>"
    echo "2. Restart a specific container: docker restart <container-name>"
    echo "3. Restart all services: cd infrastructure/docker && docker-compose down && docker-compose up -d"
    echo "4. Check network connectivity: docker network inspect docker_vpn-network"
fi

echo -e "\n----- Web Interfaces -----"
echo "Prometheus: http://localhost:9090"
echo "Grafana: http://localhost:3000 (admin/admin)"
echo "API: http://localhost:8080"
echo "VPN Admin: https://localhost/api/admin"

echo -e "\n==========================================="
