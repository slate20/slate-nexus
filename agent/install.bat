@echo off
cd /d %~dp0
set /p SERVER_URL=Enter the server IP or Hostname (if DNS is configured):
mkdir "C:\Program Files\SlateNexus"
echo %SERVER_URL% > "C:\Program Files\SlateNexus\server_url.tmp"
copy slate-rmm-agent.exe "C:\Program Files\SlateNexus\slate-rmm-agent.exe"
sc create SlateNexusAgent binpath= "C:\Program Files\SlateNexus\slate-rmm-agent.exe" start= auto error= ignore type= own
sc start SlateNexusAgent
echo "Agent installed successfully"
pause