#!/bin/bash

# Prompt user for FQDN
read -p "Enter your FQDN for this server (e.g., nexus.example.com): " fqdn

# Provide provider options
echo "Select your DNS provider for automated SSL generation:"
echo "1) Cloudflare"
echo "2) AWS Route 53"
echo "3) Google Domains"
echo "4) GoDaddy"
echo "5) Manual DNS challenge"
read -p "Enter the number corresponding to your provider: " provider_choice

# Set variables based on the choice
case $provider_choice in
    1)
        provider="Cloudflare"
        package="python3-certbot-dns-cloudflare"
        ;;
    2)
        provider="AWS Route 53"
        package="python3-certbot-dns-route53"
        ;;
    3)
        provider="Google Domains"
        package="python3-certbot-dns-google"
        ;;
    4)
        provider="GoDaddy"
        package="python3-venv" # Virtualenv is required for the GoDaddy plugin
        ;;
    5)
        provider="Manual DNS challenge"
        package=""
        ;;
    *)
        echo "Invalid option. Exiting."
        exit 1
        ;;
esac

# Confirm installation
echo "You selected $provider."
if [ -n "$package" ]; then
    echo "The following package(s) will be installed: $package"
fi
read -p "Do you want to proceed? (y/n): " confirm
if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
    echo "Installation canceled."
    exit 0
fi

# Install Certbot and the selected plugin
if [ "$provider_choice" -ne 5 ]; then
    sudo apt update
    sudo apt install -y certbot $package
fi

# Handle GoDaddy's virtual environment setup
if [ "$provider_choice" -eq 4 ]; then
    # Create a virtual environment for the GoDaddy plugin
    sudo mkdir -p /opt/certbot-godaddy
    sudo python3 -m venv /opt/certbot-godaddy
    sudo /opt/certbot-godaddy/bin/pip install certbot-dns-godaddy
fi

# Directory to store credentials securely
cred_dir="/etc/ssl/Nexus/credentials"
sudo mkdir -p "$cred_dir"
sudo chmod 700 "$cred_dir"

# Execute the appropriate Certbot command
case $provider_choice in
    1)  # Cloudflare
        echo "Please enter your Cloudflare API token."
        read -p "Enter your Cloudflare API token: " cloudflare_token
        cred_file="$cred_dir/cloudflare.ini"
        echo "dns_cloudflare_api_token = $cloudflare_token" | sudo tee "$cred_file" > /dev/null
        sudo chmod 600 "$cred_file"
        sudo certbot certonly --dns-cloudflare --dns-cloudflare-credentials "$cred_file" -d "*.$fqdn" -d "$fqdn" --agree-tos
        ;;
    2)  # AWS Route 53
        echo "Ensure your AWS credentials are configured on this system."
        sudo certbot certonly --dns-route53 -d "*.$fqdn" -d "$fqdn" --agree-tos
        ;;
    3)  # Google Domains
        echo "Please provide your Google JSON credentials file."
        read -p "Enter the path to your Google JSON credentials file: " google_json
        cred_file="$cred_dir/google.json"
        sudo cp "$google_json" "$cred_file"
        sudo chmod 600 "$cred_file"
        sudo certbot certonly --dns-google --dns-google-credentials "$cred_file" -d "*.$fqdn" -d "$fqdn" --agree-tos
        ;;
    4)  # GoDaddy
        echo "Please provide your GoDaddy API credentials."
        read -p "Enter your GoDaddy API key: " godaddy_key
        read -p "Enter your GoDaddy API secret: " godaddy_secret
        cred_file="$cred_dir/godaddy.ini"
        echo "dns_godaddy_key = $godaddy_key" | sudo tee "$cred_file" > /dev/null
        echo "dns_godaddy_secret = $godaddy_secret" | sudo tee -a "$cred_file" > /dev/null
        sudo chmod 600 "$cred_file"
        sudo /opt/certbot-godaddy/bin/certbot certonly --authenticator dns-godaddy --dns-godaddy-credentials "$cred_file" --dns-godaddy-propagation-seconds 900 -d "*.$fqdn" -d "$fqdn" --agree-tos
        deactivate
        ;;
    5)  # Manual DNS challenge
        echo "You will need to create a DNS TXT record. Certbot will provide the details during this process."
        sudo certbot certonly --manual --preferred-challenges=dns -d "*.$fqdn" -d "$fqdn" --agree-tos --manual-public-ip-logging-ok
        ;;
esac

# Check if the certificate was successfully generated
cert_path="/etc/letsencrypt/live/$fqdn"
if [ -d "$cert_path" ]; then
    echo "Certificate successfully generated."
    sudo mkdir -p /etc/ssl/Nexus
    sudo cp "$cert_path/fullchain.pem" /etc/ssl/Nexus/cert.pem
    sudo cp "$cert_path/privkey.pem" /etc/ssl/Nexus/key.pem
    echo "Certificate and key have been copied to /etc/ssl/Nexus."
    echo "Please restart your server or services to use the new certificate."
else
    echo "Certificate generation failed. Please check the certbot output for details."
fi
