#!/bin/bash

# This script is used to set up the environment for the application.

# Create /var/run/checkmk directory
sudo mkdir -p /var/run/checkmk

# Create /var/www/remotely directory
sudo mkdir -p /var/www/remotely

# Update the apt package list
sudo apt update && sudo apt upgrade -y

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

# Generate a random secret for Checkmk automation
export AUTOMATION_SECRET=$(openssl rand -base64 32)

# Write API info and secret to the .env file
echo "API_USER=cmkadmin" > .env
echo "AUTOMATION_SECRET=$AUTOMATION_SECRET" >> .env
echo "SITE_NAME=main" >> .env
echo "API_URL=http://localhost:5000/main/check_mk/api/1.0" >> .env

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

# Build Docker image
docker build -t slate-checkmk ./server

# Run Docker Services
docker-compose up -d

# Start the server
sudo systemctl start slatermm

# Enable the server to start on boot
sudo systemctl enable slatermm

# Get the IP address of the server
HOST_IP=$(hostname -I | awk '{print $1}')

# Inform the user of Remotely url
echo "Remotely is now available at http://$HOST_IP:4000"

# Prompt for Remotely API token
echo "Enter API token secret created in Remotely: "
read REMOTELY_API_TOKEN
echo "Enter API token ID: "
read REMOTELY_API_ID

# Write Remotely API URL and token to the .env file
echo "REMOTELY_API_URL=http://localhost:4000/api" >> .env
echo "REMOTELY_API_TOKEN=$REMOTELY_API_TOKEN" >> .env
echo "REMOTELY_API_ID=$REMOTELY_API_ID" >> .env

# Make a GET request to the Remotely API to get the Remotely Windows agent
curl -H "X-Api-Key: $REMOTELY_API_ID:$REMOTELY_API_TOKEN" -OJ -o ../agent/ "$REMOTE_API_URL/ClientDownloads/WindowsInstaller"