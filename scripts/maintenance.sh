#!/bin/bash

# VPN Service Maintenance Script
# This script performs maintenance tasks for the VPN service
# - Checks for updates
# - Restarts services if needed
# - Performs health checks
# - Rotates logs

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

# Set working directory
cd /home/ubuntu/vpn-service

# Perform health checks
perform_health_checks() {
    print_message "Performing health checks..."
    
    # Check if API service is running
    if curl -s http://localhost:8080/api/health > /dev/null; then
        print_message "API service is running"
    else
        print_warning "API service is not responding"
        restart_service "api"
    fi
    
    # Check if WireGuard is running
    if systemctl is-active --quiet wg-quick@wg0; then
        print_message "WireGuard service is running"
    else
        print_warning "WireGuard service is not running"
        restart_service "wireguard"
    fi
    
    # Check if Prometheus is running
    if curl -s http://localhost:9090/-/healthy > /dev/null; then
        print_message "Prometheus service is running"
    else
        print_warning "Prometheus service is not responding"
        restart_service "prometheus"
    fi
    
    # Check if Grafana is running
    if curl -s http://localhost:3000/api/health > /dev/null; then
        print_message "Grafana service is running"
    else
        print_warning "Grafana service is not responding"
        restart_service "grafana"
    fi
    
    # Check if Redis is running
    if redis-cli ping > /dev/null 2>&1; then
        print_message "Redis service is running"
    else
        print_warning "Redis service is not responding"
        restart_service "redis"
    fi
}

# Restart a service
restart_service() {
    local service=$1
    print_message "Restarting $service service..."
    docker-compose -f infrastructure/docker/docker-compose.yml restart $service
    sleep 5
    
    # Verify service is now running
    case $service in
        api)
            if curl -s http://localhost:8080/api/health > /dev/null; then
                print_message "$service service restarted successfully"
            else
                print_error "Failed to restart $service service"
            fi
            ;;
        wireguard)
            if docker ps | grep -q "vpn-wireguard"; then
                print_message "$service service restarted successfully"
            else
                print_error "Failed to restart $service service"
            fi
            ;;
        prometheus)
            if curl -s http://localhost:9090/-/healthy > /dev/null; then
                print_message "$service service restarted successfully"
            else
                print_error "Failed to restart $service service"
            fi
            ;;
        grafana)
            if curl -s http://localhost:3000/api/health > /dev/null; then
                print_message "$service service restarted successfully"
            else
                print_error "Failed to restart $service service"
            fi
            ;;
        redis)
            if redis-cli ping > /dev/null 2>&1; then
                print_message "$service service restarted successfully"
            else
                print_error "Failed to restart $service service"
            fi
            ;;
        *)
            docker-compose -f infrastructure/docker/docker-compose.yml restart $service
            print_message "$service service restart attempted"
            ;;
    esac
}

# Rotate logs
rotate_logs() {
    print_message "Rotating logs..."
    
    # Check if logrotate is installed
    if ! command -v logrotate &> /dev/null; then
        print_warning "logrotate not found. Installing..."
        apt-get update
        apt-get install -y logrotate
    fi
    
    # Create logrotate configuration if it doesn't exist
    if [ ! -f /etc/logrotate.d/vpn-service ]; then
        print_message "Creating logrotate configuration..."
        cat > /etc/logrotate.d/vpn-service << EOF
/home/ubuntu/vpn-service/backend/logs/*.log {
    daily
    missingok
    rotate 14
    compress
    delaycompress
    notifempty
    create 0640 root root
    sharedscripts
    postrotate
        docker-compose -f /home/ubuntu/vpn-service/infrastructure/docker/docker-compose.yml restart api
    endscript
}
EOF
    fi
    
    # Run logrotate
    logrotate -f /etc/logrotate.d/vpn-service
    print_message "Log rotation complete"
}

# Check for updates
check_for_updates() {
    print_message "Checking for updates..."
    
    # Pull latest changes from git
    if git pull | grep -q "Already up to date"; then
        print_message "No updates available"
    else
        print_message "Updates found, rebuilding services..."
        docker-compose -f infrastructure/docker/docker-compose.yml build
        docker-compose -f infrastructure/docker/docker-compose.yml up -d
        print_message "Services updated and restarted"
    fi
}

# Check disk space
check_disk_space() {
    print_message "Checking disk space..."
    
    # Get disk usage
    DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')
    
    if [ "$DISK_USAGE" -gt 90 ]; then
        print_error "Disk usage is critical: ${DISK_USAGE}%"
        
        # Clean up docker
        print_message "Cleaning up Docker resources..."
        docker system prune -af --volumes
        
        # Clean up logs
        print_message "Cleaning up old logs..."
        find /home/ubuntu/vpn-service/backend/logs -name "*.log.*" -type f -mtime +30 -delete
    elif [ "$DISK_USAGE" -gt 80 ]; then
        print_warning "Disk usage is high: ${DISK_USAGE}%"
        
        # Clean up logs
        print_message "Cleaning up old logs..."
        find /home/ubuntu/vpn-service/backend/logs -name "*.log.*" -type f -mtime +60 -delete
    else
        print_message "Disk usage is normal: ${DISK_USAGE}%"
    fi
}

# Backup configuration
backup_config() {
    print_message "Backing up configuration..."
    
    # Create backup directory
    BACKUP_DIR="/home/ubuntu/vpn-service/backups/$(date +%Y%m%d)"
    mkdir -p "$BACKUP_DIR"
    
    # Backup configuration files
    cp -r /home/ubuntu/vpn-service/backend/config "$BACKUP_DIR/"
    cp -r /home/ubuntu/vpn-service/backend/data/wg_configs "$BACKUP_DIR/"
    
    # Backup database
    docker exec vpn-db pg_dump -U postgres vpn_service > "$BACKUP_DIR/database.sql"
    
    # Compress backup
    tar -czf "$BACKUP_DIR.tar.gz" "$BACKUP_DIR"
    rm -rf "$BACKUP_DIR"
    
    # Keep only the last 7 backups
    ls -t /home/ubuntu/vpn-service/backups/*.tar.gz | tail -n +8 | xargs -r rm
    
    print_message "Backup completed: $BACKUP_DIR.tar.gz"
}

# Main function
main() {
    print_message "Starting VPN service maintenance at $(date)"
    
    # Create necessary directories
    mkdir -p /home/ubuntu/vpn-service/backups
    
    # Run maintenance tasks
    check_disk_space
    perform_health_checks
    rotate_logs
    check_for_updates
    backup_config
    
    print_message "Maintenance completed at $(date)"
}

# Run main function
main

exit 0
