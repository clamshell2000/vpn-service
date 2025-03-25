#!/bin/bash

# This script sets up the necessary iptables rules for the VPN service
# It ensures that traffic from VPN clients can access the internet

# Get the network interface
INTERFACE=$(ip route | grep default | awk '{print $5}')
echo "Using network interface: $INTERFACE"

# Enable IP forwarding
echo "Enabling IP forwarding..."
echo 1 > /proc/sys/net/ipv4/ip_forward

# Set up iptables MASQUERADE rule
echo "Setting up iptables MASQUERADE rule..."
sudo iptables -t nat -A POSTROUTING -s 10.0.0.0/24 -o $INTERFACE -j MASQUERADE

# Make IP forwarding persistent
echo "Making IP forwarding persistent..."
echo "net.ipv4.ip_forward = 1" | sudo tee -a /etc/sysctl.conf
sudo sysctl -p

# Save iptables rules
echo "Saving iptables rules..."
if command -v iptables-save > /dev/null; then
    sudo iptables-save | sudo tee /etc/iptables/rules.v4
else
    echo "iptables-save not found. Installing iptables-persistent..."
    sudo apt-get update
    sudo apt-get install -y iptables-persistent
    sudo iptables-save | sudo tee /etc/iptables/rules.v4
fi

echo "Setup complete!"
echo "VPN clients should now be able to access the internet through the VPN."
