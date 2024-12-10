@echo off
net stop SlateNexusAgent
del "C:\Program Files\SlateNexus\slate-rmm-agent.exe"
copy "C:\Slatecap IT\slate-rmm-agent.exe" "C:\Program Files\SlateNexus\slate-rmm-agent.exe"
net start SlateNexusAgent