# Test Proxy Type Filter
# This script starts the app and provides instructions for testing

Write-Host "=== Proxy Type Filter Test ===" -ForegroundColor Cyan
Write-Host ""

# Check if binary exists
if (-not (Test-Path "cmd\proxy-fwd\proxy-fwd.exe")) {
    Write-Host "[!] Binary not found. Run build.ps1 first." -ForegroundColor Red
    exit 1
}

Write-Host "[*] Current proxy state:" -ForegroundColor Yellow
if (Test-Path "cmd\proxy-fwd\proxies.yaml") {
    $yaml = Get-Content "cmd\proxy-fwd\proxies.yaml" -Raw
    
    # Count proxy types
    $residential = ([regex]::Matches($yaml, "proxy_type: residential")).Count
    $static = ([regex]::Matches($yaml, "proxy_type: static")).Count
    $isp = ([regex]::Matches($yaml, "proxy_type: isp")).Count
    $datacenter = ([regex]::Matches($yaml, "proxy_type: datacenter")).Count
    $unknown = ([regex]::Matches($yaml, "proxy_type: unknown")).Count
    
    Write-Host "  🏠 Residential: $residential" -ForegroundColor Green
    Write-Host "  🔒 Static:      $static" -ForegroundColor Gray
    Write-Host "  📡 ISP:         $isp" -ForegroundColor Blue
    Write-Host "  🏢 Datacenter:  $datacenter" -ForegroundColor Magenta
    Write-Host "  ❓ Unknown:     $unknown" -ForegroundColor Yellow
    Write-Host ""
} else {
    Write-Host "  No proxies.yaml found - starting fresh" -ForegroundColor Gray
    Write-Host ""
}

Write-Host "=== Starting Application ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Opening: http://127.0.0.1:17890" -ForegroundColor Green
Write-Host ""

# Instructions
Write-Host "=== Test Instructions ===" -ForegroundColor Yellow
Write-Host ""
Write-Host "1. Go to Pool tab" -ForegroundColor White
Write-Host "   - You should see all proxies with type badges" -ForegroundColor Gray
Write-Host ""
Write-Host "2. Start some proxies" -ForegroundColor White
Write-Host "   - Click 'Start All' or start individual proxies" -ForegroundColor Gray
Write-Host ""
Write-Host "3. Go to Proxies tab" -ForegroundColor White
Write-Host "   - See the 'Type' column with colored badges" -ForegroundColor Gray
Write-Host "   - 🏠 Green = Residential" -ForegroundColor Gray
Write-Host "   - 🔒 Gray = Static IP" -ForegroundColor Gray
Write-Host "   - 📡 Blue = ISP" -ForegroundColor Gray
Write-Host "   - 🏢 Purple = Datacenter" -ForegroundColor Gray
Write-Host ""
Write-Host "4. Test the filter dropdown" -ForegroundColor White
Write-Host "   - Select 'Residential' - should show only residential proxies" -ForegroundColor Gray
Write-Host "   - Select 'Static' - should show only static IP proxies" -ForegroundColor Gray
Write-Host "   - Select 'All Types' - shows everything" -ForegroundColor Gray
Write-Host ""
Write-Host "5. Test search with type" -ForegroundColor White
Write-Host "   - Type 'residential' in search box" -ForegroundColor Gray
Write-Host "   - Type 'static' in search box" -ForegroundColor Gray
Write-Host ""
Write-Host "Press Ctrl+C to stop the application" -ForegroundColor Cyan
Write-Host ""

# Start application
cd cmd\proxy-fwd

# Open browser after 2 seconds
Start-Sleep -Seconds 2
Start-Process "http://127.0.0.1:17890"

# Run app
.\proxy-fwd.exe
