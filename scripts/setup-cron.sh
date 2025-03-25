#!/bin/bash

# Setup cron job for VPN service maintenance
# This script creates a cron job to run the maintenance script daily

# Set up color codes
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}[ERROR]${NC} Please run as root"
    exit 1
fi

echo -e "${GREEN}[INFO]${NC} Setting up cron job for VPN service maintenance..."

# Create cron job entry
CRON_ENTRY="0 2 * * * /home/ubuntu/vpn-service/scripts/maintenance.sh >> /home/ubuntu/vpn-service/logs/maintenance.log 2>&1"

# Check if cron job already exists
if crontab -l | grep -q "maintenance.sh"; then
    echo -e "${GREEN}[INFO]${NC} Cron job already exists"
else
    # Add cron job
    (crontab -l 2>/dev/null; echo "$CRON_ENTRY") | crontab -
    echo -e "${GREEN}[INFO]${NC} Cron job added successfully"
fi

# Create log directory if it doesn't exist
mkdir -p /home/ubuntu/vpn-service/logs

echo -e "${GREEN}[INFO]${NC} Cron job setup complete"
echo -e "${GREEN}[INFO]${NC} The maintenance script will run daily at 2:00 AM"

exit 0
