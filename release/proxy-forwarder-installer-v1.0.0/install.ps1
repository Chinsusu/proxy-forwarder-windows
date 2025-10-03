# Proxy Forwarder Windows Installer
# Run as Administrator: powershell -ExecutionPolicy Bypass -File install.ps1

param(
    [string]$InstallPath = "$env:ProgramFiles\ProxyForwarder",
    [switch]$Uninstall,
    [switch]$Service
)

$ErrorActionPreference = "Stop"

# Check admin privileges
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "❌ This script requires Administrator privileges!" -ForegroundColor Red
    Write-Host "Please run as Administrator" -ForegroundColor Yellow
    exit 1
}

$ServiceName = "ProxyForwarder"
$ExeName = "proxy-fwd.exe"
$ExePath = Join-Path $InstallPath $ExeName

# Uninstall
if ($Uninstall) {
    Write-Host "🗑️  Uninstalling Proxy Forwarder..." -ForegroundColor Cyan
    
    # Stop and remove service if exists
    $service = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($service) {
        Write-Host "Stopping service..." -ForegroundColor Yellow
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
        
        Write-Host "Removing service..." -ForegroundColor Yellow
        sc.exe delete $ServiceName
    }
    
    # Remove installation directory
    if (Test-Path $InstallPath) {
        Write-Host "Removing installation directory..." -ForegroundColor Yellow
        Remove-Item -Path $InstallPath -Recurse -Force
    }
    
    # Remove firewall rule
    Write-Host "Removing firewall rule..." -ForegroundColor Yellow
    Remove-NetFirewallRule -DisplayName "Proxy Forwarder UI" -ErrorAction SilentlyContinue
    
    Write-Host "✅ Uninstallation completed!" -ForegroundColor Green
    exit 0
}

# Install
Write-Host "🚀 Installing Proxy Forwarder..." -ForegroundColor Cyan
Write-Host "Installation path: $InstallPath" -ForegroundColor Gray

# Create installation directory
if (-not (Test-Path $InstallPath)) {
    Write-Host "Creating installation directory..." -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
}

# Copy executable
$sourcePath = ".\bin\proxy-fwd.exe"
if (-not (Test-Path $sourcePath)) {
    Write-Host "❌ Binary not found: $sourcePath" -ForegroundColor Red
    Write-Host "Please run 'go build -o bin\proxy-fwd.exe .\cmd\proxy-fwd' first" -ForegroundColor Yellow
    exit 1
}

Write-Host "Copying executable..." -ForegroundColor Yellow
Copy-Item -Path $sourcePath -Destination $ExePath -Force

# Create data directory
$dataPath = Join-Path $InstallPath "data"
if (-not (Test-Path $dataPath)) {
    New-Item -ItemType Directory -Path $dataPath -Force | Out-Null
}

# Add firewall rule
Write-Host "Adding firewall rule..." -ForegroundColor Yellow
Remove-NetFirewallRule -DisplayName "Proxy Forwarder UI" -ErrorAction SilentlyContinue
New-NetFirewallRule -DisplayName "Proxy Forwarder UI" `
    -Direction Inbound `
    -Action Allow `
    -Protocol TCP `
    -LocalPort 17890 `
    -Program $ExePath `
    -ErrorAction SilentlyContinue | Out-Null

if ($Service) {
    # Install as Windows Service
    Write-Host "Installing as Windows Service..." -ForegroundColor Yellow
    
    # Stop existing service
    $existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($existingService) {
        Write-Host "Stopping existing service..." -ForegroundColor Yellow
        Stop-Service -Name $ServiceName -Force -ErrorAction SilentlyContinue
        Start-Sleep -Seconds 2
        sc.exe delete $ServiceName
        Start-Sleep -Seconds 2
    }
    
    # Create service using NSSM or sc.exe
    $nssmPath = Join-Path $InstallPath "nssm.exe"
    
    if (-not (Test-Path $nssmPath)) {
        Write-Host "⚠️  NSSM not found. Using basic service registration..." -ForegroundColor Yellow
        Write-Host "For better service management, download NSSM from: https://nssm.cc/download" -ForegroundColor Gray
        
        # Create service using sc.exe (basic, no auto-restart)
        sc.exe create $ServiceName binPath= $ExePath start= auto
        sc.exe description $ServiceName "Proxy Forwarder - Local HTTP Proxy Manager"
    } else {
        # Use NSSM for better service management
        & $nssmPath install $ServiceName $ExePath
        & $nssmPath set $ServiceName AppDirectory $dataPath
        & $nssmPath set $ServiceName DisplayName "Proxy Forwarder"
        & $nssmPath set $ServiceName Description "Local HTTP Proxy Manager with Web UI"
        & $nssmPath set $ServiceName Start SERVICE_AUTO_START
    }
    
    # Start service
    Write-Host "Starting service..." -ForegroundColor Yellow
    Start-Service -Name $ServiceName
    Start-Sleep -Seconds 2
    
    $serviceStatus = Get-Service -Name $ServiceName
    if ($serviceStatus.Status -eq "Running") {
        Write-Host "✅ Service started successfully!" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Service installation completed but not running" -ForegroundColor Yellow
        Write-Host "You can start it manually or check logs" -ForegroundColor Gray
    }
} else {
    # Create startup shortcut
    Write-Host "Creating startup shortcut..." -ForegroundColor Yellow
    $startupPath = [System.IO.Path]::Combine($env:APPDATA, "Microsoft\Windows\Start Menu\Programs\Startup")
    $shortcutPath = Join-Path $startupPath "Proxy Forwarder.lnk"
    
    $WScriptShell = New-Object -ComObject WScript.Shell
    $Shortcut = $WScriptShell.CreateShortcut($shortcutPath)
    $Shortcut.TargetPath = $ExePath
    $Shortcut.WorkingDirectory = $dataPath
    $Shortcut.WindowStyle = 7  # Minimized
    $Shortcut.Save()
    
    Write-Host "✅ Startup shortcut created!" -ForegroundColor Green
}

Write-Host ""
Write-Host "✅ Installation completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "📝 Next steps:" -ForegroundColor Cyan
Write-Host "  1. Access Web UI: http://127.0.0.1:17890" -ForegroundColor White
Write-Host "  2. Add your upstream proxies" -ForegroundColor White
Write-Host "  3. Use local proxies: 127.0.0.1:10001, 10002, ..." -ForegroundColor White
Write-Host ""

if ($Service) {
    Write-Host "🔧 Service Management:" -ForegroundColor Cyan
    Write-Host "  Start:   Start-Service $ServiceName" -ForegroundColor Gray
    Write-Host "  Stop:    Stop-Service $ServiceName" -ForegroundColor Gray
    Write-Host "  Status:  Get-Service $ServiceName" -ForegroundColor Gray
    Write-Host ""
}

Write-Host "📂 Installation path: $InstallPath" -ForegroundColor Gray
Write-Host "💾 Data path: $dataPath" -ForegroundColor Gray
Write-Host ""
Write-Host "To uninstall, run: powershell -ExecutionPolicy Bypass -File install.ps1 -Uninstall" -ForegroundColor Gray
