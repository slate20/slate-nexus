# move to SlateNexus directory
Set-Location "C:\Program Files\SlateNexus"

# get server_url from config.json file
$config = Get-Content "C:\Program Files\SlateNexus\config.json" | ConvertFrom-Json
$serverUrl = $config.server_url.Replace("http://", "https://")

# Get the certificate from the server
$response = [Net.HttpWebRequest]::Create($serverUrl)
$response.AllowAutoRedirect = $false
$response.Timeout = 5000
$response.GetResponse() | Out-Null

$certificate = $response.ServicePoint.Certificate

$certPath = "C:\Program Files\SlateNexus\cert.pem"

# Write the certificate to a file
$certBytes = $certificate.Export([Security.Cryptography.X509Certificates.X509ContentType]::Cert)
[IO.File]::WriteAllBytes($certPath, $certBytes)

# Import the certificate to the root store
Import-Certificate -FilePath $certPath -CertStoreLocation Cert:\LocalMachine\Root

# Stop the SlateNexusAgent service
Stop-Service "SlateNexusAgent"

# Update the config.json file
$config.server_url = $serverUrl
$config | ConvertTo-Json | Out-File "config.json" -Encoding utf8

# Start the SlateNexusAgent service
Start-Service "SlateNexusAgent"