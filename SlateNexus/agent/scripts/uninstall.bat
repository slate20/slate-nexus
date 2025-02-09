net stop SlateNexusAgent
sc delete SlateNexusAgent
@powershell -NoProfile -ExecutionPolicy Bypass -File "C:\Program Files\SlateNexus\Remotely\Install-Remotely.ps1" -uninstall -quiet
rmdir "C:\Program Files\SlateNexus" /S /Q
echo "Uninstalled SlateNexusAgent successfully"
pause 