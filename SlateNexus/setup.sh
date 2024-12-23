#!/bin/bash

# This script is used to set up the environment for the application.

# Function to generate a random string
generate_random_string() {
    openssl rand -base64 32
}

# Create needed directories if they don't exist
sudo mkdir -p /var/www/remotely
sudo mkdir -p /etc/ssl/Nexus

# Update the apt package list
sudo apt update && sudo apt upgrade -y

# Install zip
sudo apt install -y zip

# Prompt for server FQDN
read -p "Enter your FQDN for this server (e.g., nexus.example.com): " NEXUS_FQDN

# Get the server IP address
NEXUS_IP=$(hostname -I | awk '{print $1}')

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
chmod +x slatenexus

# Generate a random secrets and set other variables
REMOTELY_FQDN="remotely.$NEXUS_FQDN"
REMOTELY_API_URL="https://$REMOTELY_FQDN/api"
PG_DB=NEXUS_db
PG_USER=Nexus
PG_PASS=$(generate_random_string)
NEXUS_API_KEY=$(generate_random_string)
AUTHENTIK_SECRET=$(generate_random_string)
AK_BT_PASS=$(generate_random_string)
AK_BT_TOKEN=$(generate_random_string)
AK_BT_EMAIL=admin@example.com

# Create the agent installer scripts
./create_installers.sh "$NEXUS_API_KEY"

# Write all variables to the .env file
cat << EOF > .env
NEXUS_FQDN=$NEXUS_FQDN
NEXUS_IP=$NEXUS_IP
REMOTELY_FQDN=$REMOTELY_FQDN
REMOTELY_API_URL=$REMOTELY_API_URL
PG_DB=$PG_DB
PG_USER=$PG_USER
PG_PASS=$PG_PASS
NEXUS_API_KEY=$NEXUS_API_KEY
AUTHENTIK_SECRET_KEY=$AUTHENTIK_SECRET
AK_BT_PASS=$AK_BT_PASS
AK_BT_TOKEN=$AK_BT_TOKEN
AK_BT_EMAIL=$AK_BT_EMAIL
EOF

echo "Environment variables set in .env file"

# Create a systemd service file for the server
echo "[Unit]
Description=Slate Nexus Server

[Service]
ExecStart=$(pwd)/slatenexus
Restart=always
User=$(whoami)
Group=$(whoami)
Environment=PATH=$/usr/bin:/usr/local/bin
Environment=GO_ENV=production
WorkingDirectory=$(pwd)

[Install]
WantedBy=multi-user.target" | sudo tee /etc/systemd/system/slatenexus.service

# Reload the systemd daemon
sudo systemctl daemon-reload

# Run Docker Services
docker-compose up -d

# Start the server
sudo systemctl start slatenexus

# Enable the server to start on boot
sudo systemctl enable slatenexus

# Inform the user of Remotely url
echo "Remotely is now available at https://$REMOTELY_FQDN"

# Prompt for Remotely API token
echo "Enter API token secret created in Remotely: "
read REMOTELY_API_TOKEN
echo "Enter API token ID: "
read REMOTELY_API_ID

# Update the .env file with the Remotely API token and ID
echo "REMOTELY_API_TOKEN=$REMOTELY_API_TOKEN" >> .env
echo "REMOTELY_API_ID=$REMOTELY_API_ID" >> .env

# Make a GET request to the Remotely API to get the Remotely Windows agent
curl -H "X-Api-Key: $REMOTELY_API_ID:$REMOTELY_API_TOKEN" "$REMOTELY_API_URL/ClientDownloads/WindowsInstaller" -OJ
mv Install-Remotely.ps1 ../agent/

# Wait 60 seconds for authentik to finish setting up
echo "Waiting for other services to finish setting up..."
for i in {1..60}; do
    echo -n "."
    sleep 1
done

# Copy logo.png from dashboard/assets to media dir
cp ../dashboard/assets/logo.png ./media

# Run the authentik_config script
source .env
echo "Configuring Authentik..."
sudo ./authentik_config.sh