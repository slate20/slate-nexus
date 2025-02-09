#!/bin/bash

# Check for sudo
if [ "$EUID" -ne 0 ]; then
    echo "This script must be run with sudo."
    exit 1
fi

# Define Zammad services
ZAMMAD_SERVICES="zammad-backup zammad-elasticsearch zammad-init zammad-memcached zammad-nginx zammad-postgresql zammad-railsserver zammad-redis zammad-scheduler zammad-websocket"

# Stop and remove Zammad containers
echo "Stopping and removing Zammad containers..."
docker-compose stop $ZAMMAD_SERVICES
docker-compose rm -f $ZAMMAD_SERVICES

# Recreate and start Zammad containers
echo "Recreating and starting Zammad containers..."
docker-compose up -d $ZAMMAD_SERVICES

# Restart Nginx service
echo "Restarting Nginx service..."
docker-compose restart nginx

echo "Zammad services have been restarted."