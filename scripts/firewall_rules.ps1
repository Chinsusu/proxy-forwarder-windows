# ProxyFwd firewall rules (Windows 10)
# Run PowerShell as Administrator.
# Adjust program paths/ports as needed.

$ports = "10001-20000"
$loopback = "127.0.0.1"

# === Allow proxy-fwd.exe to reach the internet ===
$proxyExe = "C:\Users\Administrator\Documents\proxy-fwd-windows\cmd\proxy-fwd\proxy-fwd.exe"
if (-not (Test-Path $proxyExe)) {
  Write-Host "WARNING: proxy-fwd.exe not found at $proxyExe" -ForegroundColor Yellow
  Write-Host "Please update the path in this script if needed." -ForegroundColor Yellow
} else {
  New-NetFirewallRule -DisplayName "ProxyFwd Allow Out" -Direction Outbound -Program $proxyExe -Action Allow -Profile Any -Protocol TCP -RemoteAddress Any -EdgeTraversalPolicy Block -Group "ProxyFwd Rules" | Out-Null
  Write-Host "✅ Allow rule created for proxy-fwd.exe" -ForegroundColor Green
}

# === Example: restrict Chrome/Edge/Firefox to only localhost:ports, block all other outbound ===
$browsers = @(
  "C:\Program Files\Google\Chrome\Application\chrome.exe",
  "C:\Program Files (x86)\Microsoft\Edge\Application\msedge.exe",
  "C:\Program Files\Mozilla Firefox\firefox.exe"
)

foreach ($app in $browsers) {
  if (Test-Path $app) {
    # Specific allow to 127.0.0.1:10001-20000
    New-NetFirewallRule -DisplayName "ProxyFwd Allow Localhost for $app" -Direction Outbound -Program $app -Action Allow -Profile Any -Protocol TCP -RemoteAddress $loopback -RemotePort $ports -Group "ProxyFwd Rules" | Out-Null
    # Broad block for everything else
    New-NetFirewallRule -DisplayName "ProxyFwd Block Internet for $app" -Direction Outbound -Program $app -Action Block -Profile Any -Protocol TCP -RemoteAddress Any -Group "ProxyFwd Rules" | Out-Null
  }
}

Write-Host "Firewall rules installed. To remove:"
Write-Host '  Get-NetFirewallRule -DisplayName "ProxyFwd *" | Remove-NetFirewallRule'
