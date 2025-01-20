#!/bin/bash

API_KEY=$1

# Create scripts dir if it doesn't exist
mkdir -p ../agent/scripts

# Function to create Windows installer script
create_windows_installer() {
    cat << EOF > ../agent/scripts/NexusAgent_Installer.ps1
# Run this script with administrator privileges
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator"))  
{  
    \$arguments = "& '" + \$myinvocation.mycommand.definition + "'"
    Start-Process powershell -Verb runAs -ArgumentList \$arguments
    Break
}

# Change to the directory where the script is located
Set-Location \$PSScriptRoot

# Prompt for server IP and append "https://"
Write-Host "Enter the server FQDN: "
\$serverUrl = Read-Host
\$remotelyUrl = "https://remotely." + \$serverUrl
\$serverUrl = "https://api." + \$serverUrl

# Create the directory structure
\$installPath = "C:\Program Files\SlateNexus"
\$remotelyPath = "\$installPath\Remotely"

New-Item -ItemType Directory -Path \$installPath
New-Item -ItemType Directory -Path \$remotelyPath

# Copy the agent executable to the installation directory
Copy-Item ".\slate-nexus-agent.exe" -Destination "\$installPath\slate-nexus-agent.exe"

# move to the new directory
Set-Location \$installPath

# Create the config.json file
\$config = @{
    "server_url" = "\$serverUrl"
    "host_id" = 0
    "api_key" = "$API_KEY"
}

# Convert to JSON and save to file using UTF-8
\$config | ConvertTo-Json | Out-File "config.json" -Encoding utf8

# Function to install the Slate Nexus agent
function Install-SlateNexusAgent {
    \$agentPath = "C:\Program Files\SlateNexus\slate-nexus-agent.exe"

    Write-Host "Installing Slate Nexus agent..."

    # Create the Windows service
    New-Service -Name "SlateNexusAgent" -BinaryPathName \$agentPath -DisplayName "Slate Nexus Agent" -StartupType Automatic

    Write-Host "Slate Nexus agent installed successfully"

    # Start the service
    Start-Service "SlateNexusAgent"
}

# Function to install Remotely
function Install-Remotely {
    \$remotelySoftwarePath = "C:\Program Files\Remotely\Remotely_Agent.exe"
    \$remotelyInstallerPath = "C:\Program Files\SlateNexus\Remotely\Install-Remotely.ps1"

    if (Test-Path \$remotelySoftwarePath) {
        Write-Host "Remotely is already installed"
        return
    }



    # Download the Remotely installer
    Write-Host "Downloading Remotely installer"
    \$headers = @{
        "Authorization" = "Bearer $API_KEY"
    }
    Invoke-WebRequest -Uri "\$serverUrl/download/remotely-win" -OutFile \$remotelyInstallerPath -Headers \$headers

    Write-Host "Installing Remotely agent..."
    & \$remotelyInstallerPath -serverurl (\$remotelyUrl) -install

    if (Test-Path \$remotelySoftwarePath) {
        Write-Host "Remotely agent installation verified"
    } else {
        Write-Host "Remotely agent not found after installation"
        exit 1
    }
}

# Main installation process
Install-SlateNexusAgent

# Wait for 10 seconds
Start-Sleep -Seconds 10

Install-Remotely

Write-Host "Installation completed successfully"

# pause
Read-Host -Prompt "Press Enter to exit"
EOF

    echo "NexusAgent_Installer.ps1 created successfully."
}

# Function to create Linux installer script
create_linux_installer() {
    cat << EOF > ../agent/scripts/NexusAgent_Installer.sh
#!/bin/bash

# Exit immediately if a command exits with a non-zero status
set -e

# Check if the script is running as root
if [ "\$(id -u)" -ne 0 ]; then
    echo "This script must be run as root."
    exit 1
fi

# Prompt for server FQDN
read -p "Enter the server FQDN: " server_url
remotely_url="https://remotely.\$server_url"
server_url="https://api.\$server_url"

# Create the directory structure
install_path="/opt/SlateNexus"
remotely_path="\$install_path/Remotely"

mkdir -p \$install_path
mkdir -p \$remotely_path

# Copy the agent executable to the installation directory
cp ./slate-nexus-agent \$install_path/slate-nexus-agent

# Create the config.json file
cat > \$install_path/config.json << EOL
{
    "server_url": "\$server_url",
    "host_id": 0,
    "api_key": "$API_KEY"
}
EOL

# Function to install the Slate Nexus agent
install_slate_nexus_agent() {
    echo "Installing Slate Nexus agent..."

    # Copy the systemd service file
    cat > /etc/systemd/system/SlateNexusAgent.service << EOL
[Unit]
Description=Slate Nexus Agent
After=network.target

[Service]
ExecStart=\$install_path/slate-nexus-agent
Restart=always
User=root
Group=root
Environment=PATH=/usr/bin:/usr/local/bin
WorkingDirectory=\$install_path

[Install]
WantedBy=multi-user.target

EOL

    # Reload systemd to load the service
    systemctl daemon-reload

    # Enable and start the service
    systemctl enable SlateNexusAgent.service
    systemctl start SlateNexusAgent.service

    echo "Slate Nexus agent installed successfully"
}

# Main installation process
install_slate_nexus_agent

# Wait for 10 seconds
sleep 10

# Optionally add host record
read -p "Do you want to create a local host record for the server(if not handled by DNS)? (Y/N)" create_host_record

if [ "\$create_host_record" =~ ^[Yy]$ ]; then
    read -p "Enter the server IP: " server_ip
    echo "\$server_ip    \$server_url" >> /etc/hosts
    echo "Host record created successfully"
fi

echo "Installation completed successfully"
EOF

    chmod +x ../agent/scripts/NexusAgent_Installer.sh
}

# Call the function to create the Windows installer script
create_windows_installer

# Call the function to create the Linux installer script
create_linux_installer

# Zip the Install-Remotely.ps1 script with the slate-nexus-agent.exe
zip -j ../agent/NexusAgent_win.zip ../agent/scripts/NexusAgent_Installer.ps1 ../agent/slate-nexus-agent.exe

# tar the NexusAgent_Installer.sh script with the slate-nexus-agent executable
tar -czf ../agent/NexusAgent_lin.tar.gz -C ../agent/scripts NexusAgent_Installer.sh slate-nexus-agent