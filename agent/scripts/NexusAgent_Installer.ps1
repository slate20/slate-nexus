# install.ps1

# Run this script with administrator privileges
if (-NOT ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole] "Administrator"))  
{  
  $arguments = "& '" + $myinvocation.mycommand.definition + "'"
  Start-Process powershell -Verb runAs -ArgumentList $arguments
  Break
}

# Change to the directory where the script is located
Set-Location $PSScriptRoot

# Prompt for server IP and append "https://"
Write-Host "Enter the server FQDN: "
$serverUrl = Read-Host
$serverUrl = "https://api." + $serverUrl

# Create the directory structure
$installPath = "C:\Program Files\SlateNexus"
$remotelyPath = "$installPath\Remotely"

New-Item -ItemType Directory -Path $installPath
New-Item -ItemType Directory -Path $remotelyPath

# Copy the agent executable to the installation directory
Copy-Item ".\slate-rmm-agent.exe" -Destination "$installPath\slate-rmm-agent.exe"

# move to the new directory
Set-Location $installPath

# Create the config.json file
$config = @{
    "server_url" = $serverUrl
    "host_id" = 0
    "api_key" = "BaSELBzSui1brU2ALQ0WxD3G+qah9lAIoPDJ6yeY8TY="
}

# Convert to JSON and save to file using UTF-8
$config | ConvertTo-Json | Out-File "config.json" -Encoding utf8

# Function to install the Slate-RMM agent
function Install-SlateRMMAgent {
    $agentPath = "C:\Program Files\SlateNexus\slate-rmm-agent.exe"
    
    Write-Host "Installing Slate-RMM agent..."

    # Create the Windows service
    New-Service -Name "SlateNexusAgent" -BinaryPathName $agentPath -DisplayName "Slate Nexus Agent" -StartupType Automatic

    Write-Host "Slate-RMM agent installed successfully"

    # Start the service
    Start-Service "SlateNexusAgent"
}

# Function to install Remotely
function Install-Remotely {
    $remotelySoftwarePath = "C:\Program Files\Remotely\Remotely_Agent.exe"
    $remotelyInstallerPath = "C:\Program Files\SlateNexus\Remotely\Install-Remotely.ps1"

    if (Test-Path $remotelySoftwarePath) {
        Write-Host "Remotely is already installed"
        return
    }



    # Download the Remotely installer
    Write-Host "Downloading Remotely agent..."
    Invoke-WebRequest -Uri "$serverUrl/api/download/remotely-win" -OutFile $remotelyInstallerPath

    Write-Host "Installing Remotely agent..."
    & $remotelyInstallerPath -serverurl ($serverUrl + ":4000") -install

    if (Test-Path $remotelySoftwarePath) {
        Write-Host "Remotely agent installation verified"
    } else {
        Write-Host "Remotely agent not found after installation"
        exit 1
    }
}

# Main installation process
Install-SlateRMMAgent

# Wait for 10 seconds
Start-Sleep -Seconds 10

Install-Remotely

# Prompt the user to ask if they want to create a host record for the server
Write-Host "Do you want to create a local host record for the server(if not handled by DNS)? (Y/N)"
$createHostRecord = Read-Host

if ($createHostRecord -eq "Y" -or $createHostRecord -eq "y") {
    # Prompt for the server IP
    Write-Host "Enter the server IP: "
    $serverIP = Read-Host

    # Create the host record
    $hostRecord = "$serverIP`t$serverUrl"
    $hostsFile = "$env:windir\System32\drivers\etc\hosts"

    # Check if the host record already exists
    if (-not (Get-Content $hostsFile | Select-String -Pattern $hostRecord -Quiet)) {
        # Add the new host record
        Add-Content -Path $hostsFile -Value $hostRecord -Force
        Write-Host "Host record created successfully"
    } else {
        Write-Host "Host record already exists"
    }
}

Write-Host "Installation completed successfully"

# pause
Read-Host -Prompt "Press Enter to exit"