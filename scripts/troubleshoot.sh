#!/bin/bash

# VPN Service Troubleshooting Script
# This script helps diagnose common issues with the VPN service

# Set up color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored messages
print_header() {
    echo -e "\n${BLUE}=== $1 ===${NC}\n"
}

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

# Set working directory
cd /home/ubuntu/vpn-service

# Check Docker services
check_docker_services() {
    print_header "Checking Docker Services"
    
    # Get list of services
    SERVICES=("vpn-api" "vpn-wireguard" "vpn-db" "vpn-nginx" "vpn-prometheus" "vpn-grafana" "vpn-redis" "vpn-node-exporter" "vpn-redis-exporter")
    
    for service in "${SERVICES[@]}"; do
        if docker ps | grep -q "$service"; then
            print_message "$service is running"
        else
            print_error "$service is not running"
            if docker ps -a | grep -q "$service"; then
                print_warning "Container exists but is not running. Checking logs..."
                docker logs --tail 20 "$service"
            else
                print_error "Container does not exist"
            fi
        fi
    done
}

# Check network configuration
check_network_config() {
    print_header "Checking Network Configuration"
    
    # Check IP forwarding
    if [ "$(cat /proc/sys/net/ipv4/ip_forward)" -eq 1 ]; then
        print_message "IP forwarding is enabled"
    else
        print_error "IP forwarding is disabled"
        print_message "Run the following command to enable IP forwarding:"
        echo "echo 1 > /proc/sys/net/ipv4/ip_forward"
        echo "echo \"net.ipv4.ip_forward = 1\" >> /etc/sysctl.conf"
        echo "sysctl -p"
    fi
    
    # Check MASQUERADE rule
    if iptables -t nat -L POSTROUTING | grep -q MASQUERADE; then
        print_message "MASQUERADE rule is configured"
    else
        print_error "MASQUERADE rule is not configured"
        print_message "Run the following command to configure MASQUERADE:"
        echo "iptables -t nat -A POSTROUTING -s 10.0.0.0/24 -o \$(ip route | grep default | awk '{print \$5}') -j MASQUERADE"
    fi
    
    # Check WireGuard interface
    if ip a | grep -q wg0; then
        print_message "WireGuard interface (wg0) is up"
        ip -brief a show wg0
    else
        print_error "WireGuard interface (wg0) is not up"
    fi
}

# Check WireGuard configuration
check_wireguard_config() {
    print_header "Checking WireGuard Configuration"
    
    # Check if WireGuard config directory exists
    if [ -d "/home/ubuntu/vpn-service/backend/data/wg_configs" ]; then
        print_message "WireGuard config directory exists"
        
        # Check if wg0.conf exists
        if [ -f "/home/ubuntu/vpn-service/backend/data/wg_configs/wg0.conf" ]; then
            print_message "wg0.conf exists"
            
            # Check if private key is set
            if grep -q "PrivateKey" "/home/ubuntu/vpn-service/backend/data/wg_configs/wg0.conf"; then
                print_message "PrivateKey is configured"
            else
                print_error "PrivateKey is not configured in wg0.conf"
            fi
            
            # Check if Address is set
            if grep -q "Address" "/home/ubuntu/vpn-service/backend/data/wg_configs/wg0.conf"; then
                print_message "Address is configured"
            else
                print_error "Address is not configured in wg0.conf"
            fi
            
            # Check if ListenPort is set
            if grep -q "ListenPort" "/home/ubuntu/vpn-service/backend/data/wg_configs/wg0.conf"; then
                print_message "ListenPort is configured"
            else
                print_error "ListenPort is not configured in wg0.conf"
            fi
        else
            print_error "wg0.conf does not exist"
        fi
    else
        print_error "WireGuard config directory does not exist"
    fi
    
    # Check WireGuard peer configurations
    if [ -d "/home/ubuntu/vpn-service/backend/data/wg_configs/peers" ]; then
        PEER_COUNT=$(ls -1 /home/ubuntu/vpn-service/backend/data/wg_configs/peers | wc -l)
        print_message "Found $PEER_COUNT peer configuration(s)"
    else
        print_warning "No peer configurations found"
    fi
}

# Check API service
check_api_service() {
    print_header "Checking API Service"
    
    # Check if API is responding
    if curl -s http://localhost:8080/api/health > /dev/null; then
        print_message "API service is responding"
        API_RESPONSE=$(curl -s http://localhost:8080/api/health)
        echo "Response: $API_RESPONSE"
    else
        print_error "API service is not responding"
        
        # Check API logs
        print_warning "Checking API logs..."
        docker logs --tail 20 vpn-api
    fi
    
    # Check API configuration
    if [ -f "/home/ubuntu/vpn-service/backend/config/config.json" ]; then
        print_message "API configuration file exists"
    else
        print_error "API configuration file does not exist"
    fi
}

# Check database
check_database() {
    print_header "Checking Database"
    
    # Check if database is running
    if docker exec vpn-db pg_isready -U postgres > /dev/null 2>&1; then
        print_message "Database is running"
        
        # Check if vpn_service database exists
        if docker exec vpn-db psql -U postgres -lqt | cut -d \| -f 1 | grep -qw vpn_service; then
            print_message "vpn_service database exists"
            
            # Check tables
            TABLES=$(docker exec vpn-db psql -U postgres -d vpn_service -c "\\dt" | grep -c "public")
            print_message "Found $TABLES table(s) in the database"
        else
            print_error "vpn_service database does not exist"
        fi
    else
        print_error "Database is not running"
    fi
}

# Check monitoring
check_monitoring() {
    print_header "Checking Monitoring"
    
    # Check if Prometheus is running
    if curl -s http://localhost:9090/-/healthy > /dev/null; then
        print_message "Prometheus is running"
        
        # Check targets
        TARGET_STATUS=$(curl -s http://localhost:9090/api/v1/targets | grep -o '"health":"up"' | wc -l)
        print_message "$TARGET_STATUS target(s) are up"
    else
        print_error "Prometheus is not running"
    fi
    
    # Check if Grafana is running
    if curl -s http://localhost:3000/api/health > /dev/null; then
        print_message "Grafana is running"
    else
        print_error "Grafana is not running"
    fi
}

# Check logs
check_logs() {
    print_header "Checking Logs"
    
    # Check API logs
    if [ -f "/home/ubuntu/vpn-service/backend/logs/api.log" ]; then
        print_message "API log file exists"
        print_message "Recent errors in API logs:"
        grep -i "error" /home/ubuntu/vpn-service/backend/logs/api.log | tail -5
    else
        print_warning "API log file does not exist"
    fi
    
    # Check analytics logs
    if [ -f "/home/ubuntu/vpn-service/backend/logs/usage_analytics.log" ]; then
        print_message "Analytics log file exists"
        print_message "Recent analytics entries:"
        tail -5 /home/ubuntu/vpn-service/backend/logs/usage_analytics.log
    else
        print_warning "Analytics log file does not exist"
    fi
    
    # Check Nginx logs
    if [ -f "/home/ubuntu/vpn-service/backend/logs/nginx/error.log" ]; then
        print_message "Nginx error log file exists"
        print_message "Recent errors in Nginx logs:"
        grep -i "error" /home/ubuntu/vpn-service/backend/logs/nginx/error.log | tail -5
    else
        print_warning "Nginx error log file does not exist"
    fi
}

# Fix common issues
fix_common_issues() {
    print_header "Fixing Common Issues"
    
    # Enable IP forwarding
    if [ "$(cat /proc/sys/net/ipv4/ip_forward)" -ne 1 ]; then
        print_message "Enabling IP forwarding..."
        echo 1 > /proc/sys/net/ipv4/ip_forward
        echo "net.ipv4.ip_forward = 1" >> /etc/sysctl.conf
        sysctl -p
    fi
    
    # Add MASQUERADE rule if missing
    if ! iptables -t nat -L POSTROUTING | grep -q MASQUERADE; then
        print_message "Adding MASQUERADE rule..."
        INTERFACE=$(ip route | grep default | awk '{print $5}')
        iptables -t nat -A POSTROUTING -s 10.0.0.0/24 -o $INTERFACE -j MASQUERADE
        
        # Save iptables rules
        if command -v iptables-save > /dev/null; then
            mkdir -p /etc/iptables
            iptables-save > /etc/iptables/rules.v4
        fi
    fi
    
    # Restart services if needed
    if ! docker ps | grep -q "vpn-api"; then
        print_message "Restarting API service..."
        docker-compose -f infrastructure/docker/docker-compose.yml restart api
    fi
    
    if ! docker ps | grep -q "vpn-wireguard"; then
        print_message "Restarting WireGuard service..."
        docker-compose -f infrastructure/docker/docker-compose.yml restart wireguard
    fi
    
    # Create log directories if missing
    mkdir -p /home/ubuntu/vpn-service/backend/logs
    mkdir -p /home/ubuntu/vpn-service/backend/logs/nginx
    
    print_message "Common issues fixed"
}

# Display system information
show_system_info() {
    print_header "System Information"
    
    # OS information
    print_message "Operating System:"
    cat /etc/os-release | grep "PRETTY_NAME" | cut -d= -f2 | tr -d '"'
    
    # Kernel information
    print_message "Kernel Version:"
    uname -r
    
    # CPU information
    print_message "CPU Information:"
    lscpu | grep "Model name" | cut -d: -f2 | sed 's/^[ \t]*//'
    
    # Memory information
    print_message "Memory Information:"
    free -h | grep "Mem" | awk '{print "Total: " $2 ", Used: " $3 ", Free: " $4}'
    
    # Disk information
    print_message "Disk Information:"
    df -h / | awk 'NR==2 {print "Total: " $2 ", Used: " $3 ", Free: " $4 ", Usage: " $5}'
    
    # Network information
    print_message "Network Information:"
    ip -brief a | grep -v "lo"
    
    # Docker information
    print_message "Docker Information:"
    docker version --format '{{.Server.Version}}'
    
    # WireGuard information
    print_message "WireGuard Information:"
    if command -v wg > /dev/null; then
        wg --version
    else
        echo "WireGuard tools not installed"
    fi
}

# Main function
main() {
    print_header "VPN Service Troubleshooting"
    print_message "Starting troubleshooting at $(date)"
    
    # Run checks
    check_docker_services
    check_network_config
    check_wireguard_config
    check_api_service
    check_database
    check_monitoring
    check_logs
    
    # Ask if user wants to fix common issues
    echo -e "\n${YELLOW}Do you want to attempt to fix common issues? (y/n)${NC}"
    read -r answer
    if [[ "$answer" =~ ^[Yy]$ ]]; then
        fix_common_issues
    fi
    
    # Show system information
    show_system_info
    
    print_header "Troubleshooting Complete"
    print_message "Completed at $(date)"
    print_message "If issues persist, please check the documentation or contact support."
}

# Run main function
main

exit 0
