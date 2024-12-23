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

# Call the function to create the Windows installer script
create_windows_installer