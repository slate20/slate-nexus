# Run this script with administrator privileges
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator"))  
{  
    Write-Host "Running without admin privileges, elevating to admin"
    \$arguments = "& '" + \$myinvocation.mycommand.definition + "'"
    Start-Process powershell -Verb runAs -ArgumentList \$arguments
    Break
}

# Change to the directory where the script is located
Write-Host "Changing directory to $PSScriptRoot"
Set-Location \$PSScriptRoot

# Prompt for server IP and append "https://"
Write-Host "Enter the server FQDN: "
\$serverUrl = Read-Host
\$remotelyUrl = "https://remotely." + \$serverUrl
\$serverUrl = "https://api." + \$serverUrl

# Create the directory structure
\$installPath = "C:\Program Files\SlateNexus"
\$remotelyPath = "\$installPath\Remotely"

Write-Host "Creating directory structure at $installPath and $remotelyPath"
New-Item -ItemType Directory -Path \$installPath
New-Item -ItemType Directory -Path \$remotelyPath

# Copy the agent executable to the installation directory
Write-Host "Copying agent executable to $installPath"
Copy-Item ".\slate-nexus-agent.exe" -Destination "\$installPath\slate-nexus-agent.exe"

# move to the new directory
Write-Host "Changing directory to $installPath"
Set-Location \$installPath

# Create the config.json file
\$config = @{
    "server_url" = "\$serverUrl"
    "host_id" = 0
    "api_key" = "$API_KEY"
}

# Convert to JSON and save to file using UTF-8
Write-Host "Saving config to config.json"
\$config | ConvertTo-Json | Out-File "config.json" -Encoding utf8

# Function to install the Slate Nexus agent
function Install-SlateNexusAgent {
    \$agentPath = "C:\Program Files\SlateNexus\slate-nexus-agent.exe"

    Write-Host "Installing Slate Nexus agent..."

    # Create the Windows service
    Write-Host "Creating Windows service SlateNexusAgent"
    New-Service -Name "SlateNexusAgent" -BinaryPathName \$agentPath -DisplayName "Slate Nexus Agent" -StartupType Automatic

    Write-Host "Slate Nexus agent installed successfully"

    # Start the service
    Write-Host "Starting Windows service SlateNexusAgent"
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
Write-Host "Installing Slate Nexus agent"
Install-SlateNexusAgent

# Wait for 10 seconds
Write-Host "Waiting 10 seconds..."
Start-Sleep -Seconds 10

Write-Host "Installing Remotely agent"
Install-Remotely

if (\$LASTEXITCODE -eq 1) {
    Write-Host "An error occurred during installation"
    Read-Host -Prompt "Press Enter to exit"
    exit
}

Write-Host "Installation completed successfully"

# pause
Read-Host -Prompt "Press Enter to exit"

