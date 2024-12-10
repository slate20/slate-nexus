#!/bin/bash

# This script is used to set up the environment for the application.

# Create needed directories if they don't exist
sudo mkdir -p /var/www/remotely
sudo mkdir -p /etc/ssl/Nexus

# Update the apt package list
sudo apt update && sudo apt upgrade -y

# Get all host IP addresses
IP_ADDRESSES=$(hostname -I | xargs -n1 | sed 's/^/IP:/g' | paste -sd,)

# Check if cert.pem and key.pem are already present
if [ ! -f /etc/ssl/Nexus/cert.pem ] || [ ! -f /etc/ssl/Nexus/key.pem ]; then
    # Prompt user to choose between using Let's Encrypt (certbot), self-signed certificate, or no certificate
    echo "Which method would you like to use for SSL certificates?"
    echo "1. Let's Encrypt (must have a valid FQDN)"
    echo "2. Self-signed certificate"
    echo "3. I've placed my own certificate in /etc/ssl/Nexus"
    echo "4. No certificate"
    read -p "Enter your choice (1/2/3): " choice

    if [ "$choice" = "1" ]; then
        # Install certbot
        sudo apt install -y certbot python3-certbot-nginx

        # Prompt user for FQDN
        read -p "Enter your FQDN: " fqdn

        # Generate Let's Encrypt certificate
        sudo certbot --nginx -d $fqdn

        # Copy the certificate and key to the appropriate location
        sudo cp /etc/letsencrypt/live/$fqdn/fullchain.pem /etc/ssl/Nexus/cert.pem
        sudo cp /etc/letsencrypt/live/$fqdn/privkey.pem /etc/ssl/Nexus/key.pem

        echo "Certificate setup complete."

    elif [ "$choice" = "2" ]; then
        # Generate self-signed certificate
        openssl req -x509 -newkey rsa:4096 -days 365 -nodes -out /etc/ssl/Nexus/cert.pem -keyout /etc/ssl/Nexus/key.pem -subj "/C=US/ST=Texas/L=Dallas/O=SlateNexus" -addext "subjectAltName=IP:${HOST_IP}"

        echo "Certificate setup complete."

    elif [ "$choice" = "3" ]; then
        echo "Certificates must be named cert.pem and key.pem. Please rename your certificates and then re-run this setup."
        exit 1

    elif [ "$choice" = "4" ]; then
        echo "No certificate will be used."
    else
        # If invalid choice, re-prompt user
        echo "Invalid choice. Please enter 1, 2, or 3."
        exit 1
    fi
# If .pem files exist but are named incorrectly
else
    echo "Certificates already exist. Skipping certificate setup."
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

# Generate a random secret for automation and for postgres
export AUTOMATION_SECRET=$(openssl rand -base64 32)
export POSTGRES_PASSWORD=$(openssl rand -base64 32)

# Write API info and secret to the .env file
echo "AUTOMATION_SECRET=$AUTOMATION_SECRET" >> .env
echo "POSTGRES_PASSWORD=$POSTGRES_PASSWORD" >> .env

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
echo "REMOTELY_API_URL=http://$HOST_IP:4000/api" >> .env
echo "REMOTELY_API_TOKEN=$REMOTELY_API_TOKEN" >> .env
echo "REMOTELY_API_ID=$REMOTELY_API_ID" >> .env

# Make a GET request to the Remotely API to get the Remotely Windows agent
source .env
curl -H "X-Api-Key: $REMOTELY_API_ID:$REMOTELY_API_TOKEN" "$REMOTELY_API_URL/ClientDownloads/WindowsInstaller" -OJ
mv Install-Remotely.ps1 ../agent/