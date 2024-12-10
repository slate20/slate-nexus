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
Write-Host "Enter the server IP or Hostname (if DNS is configured): "
$serverUrl = Read-Host
$serverUrl = "https://" + $serverUrl

# Create the directory structure
$installPath = "C:\Program Files\SlateNexus"
$remotelyPath = "$installPath\Remotely"

New-Item -ItemType Directory -Path $installPath
New-Item -ItemType Directory -Path $remotelyPath

# Copy the agent executable to the installation directory
Copy-Item ".\slate-rmm-agent.exe" -Destination "$installPath\slate-rmm-agent.exe"

# move to the new directory
Set-Location $installPath

# Check if server certificate is self-signed
$response = [Net.HttpWebRequest]::Create($serverUrl)
$response.AllowAutoRedirect = $false
$response.Timeout = 5000
$response.GetResponse() | Out-Null

$certificate = $response.ServicePoint.Certificate

if ($certificate) {
    if ($certificate.Subject -eq $certificate.Issuer) {
        $selfSigned = $true
    } else {
        $selfSigned = $false
    }
} else {
    Write-Output "Unable to retrieve server certificate."
}

# If it is a self-signed certificate, export it from the server and add it to the Trusted Root CAs
if ($selfSigned) {
    $certPath = "$installPath\cert.pem"
    Write-Output "Server certificate is self-signed, exporting and adding to Trusted Root CAs..."
    $certBytes = $certificate.Export([Security.Cryptography.X509Certificates.X509ContentType]::Cert)
    [IO.File]::WriteAllBytes($certPath, $certBytes)
    
    Import-Certificate -FilePath $certPath -CertStoreLocation Cert:\LocalMachine\Root
    Write-Output "Server certificate added to Trusted Root CAs successfully."
} else {
    Write-Output "Server certificate is not self-signed."
}

# Create the config.json file
$config = @{
    "server_url" = $serverUrl
    "host_id" = 0
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

Write-Host "Installation completed successfully"

# pause
Read-Host -Prompt "Press Enter to exit"