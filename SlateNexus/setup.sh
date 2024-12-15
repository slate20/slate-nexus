#!/bin/bash

# This script is used to set up the environment for the application.

# Create needed directories if they don't exist
sudo mkdir -p /var/www/remotely
sudo mkdir -p /etc/ssl/Nexus

# Update the apt package list
sudo apt update && sudo apt upgrade -y

# Prompt for server FQDN
read -p "Enter your FQDN for this server (e.g., nexus.example.com): " fqdn

# Get the server IP address
host_ip=$(hostname -I | awk '{print $1}')

# Write domain info to the .env file
echo "NEXUS_IP=$host_ip" >> .env
echo "NEXUS_FQDN=$fqdn" >> .env
echo "REMOTELY_FQDN=remotely.$fqdn" >> .env

# Check if cert.pem and key.pem are already present
if [ ! -f /etc/ssl/Nexus/cert.pem ] || [ ! -f /etc/ssl/Nexus/key.pem ]; then
    echo "Please run the generate_ssl.sh script or add your existing certificate and key to /etc/ssl/Nexus/. Re run this script once complete."
    exit 1
fi

# Install Docker
sudo apt install -y docker.io

# Install Docker Compose
sudo apt install -y docker-compose

# Restart Docker service
sudo service docker restart

# Navigate to the server directory
cd server

# Set permissions for the server binary
chmod +x slatermm

# Generate a random secret for Nexus API, Postgres, and Authentik
export PG_PASS=$(openssl rand -base64 32)
export API_KEY=$(openssl rand -base64 32)
export AUTHENTIK_SECRET=$(openssl rand -base64 32)

# Write API info and secret to the .env file
echo "PG_PASS=$PG_PASS" >> .env
echo "NEXUS_API_KEY=$API_KEY" >> .env
echo "AUTHENTIK_SECRET_KEY=$AUTHENTIK_SECRET" >> .env

# Create the agent installer scripts
./create_installers.sh "$API_KEY"

# Create a systemd service file for the server
echo "[Unit]
Description=SlaterMM Server

[Service]
ExecStart=$(pwd)/slatermm
Restart=always
User=$(whoami)
Group=$(whoami)
Environment=PATH=$/usr/bin:/usr/local/bin
Environment=GO_ENV=production
WorkingDirectory=$(pwd)

[Install]
WantedBy=multi-user.target" | sudo tee /etc/systemd/system/slatermm.service

# Reload the systemd daemon
sudo systemctl daemon-reload

# Run Docker Services
docker-compose up -d

# Start the server
sudo systemctl start slatermm

# Enable the server to start on boot
sudo systemctl enable slatermm

# Inform the user of Remotely url
echo "Remotely is now available at ${REMOTELY_FQDN}"

# Prompt for Remotely API token
echo "Enter API token secret created in Remotely: "
read REMOTELY_API_TOKEN
echo "Enter API token ID: "
read REMOTELY_API_ID

# Write Remotely API URL and token to the .env file
echo "REMOTELY_API_URL= https://$REMOTELY_FQDN/api" >> .env
echo "REMOTELY_API_TOKEN=$REMOTELY_API_TOKEN" >> .env
echo "REMOTELY_API_ID=$REMOTELY_API_ID" >> .env

# Make a GET request to the Remotely API to get the Remotely Windows agent
source .env
curl -H "X-Api-Key: $REMOTELY_API_ID:$REMOTELY_API_TOKEN" "$REMOTELY_API_URL/ClientDownloads/WindowsInstaller" -OJ
mv Install-Remotely.ps1 ../agent/