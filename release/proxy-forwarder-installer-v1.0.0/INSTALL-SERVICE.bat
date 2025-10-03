@echo off
echo Proxy Forwarder Service Installer
echo.
pause
powershell -ExecutionPolicy Bypass -File install.ps1 -Service
pause
