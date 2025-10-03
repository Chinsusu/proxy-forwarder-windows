@echo off
echo Uninstalling...
pause
powershell -ExecutionPolicy Bypass -File install.ps1 -Uninstall
pause
