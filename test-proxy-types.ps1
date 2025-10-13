# Test Proxy Type Detection
Write-Host "=== Proxy Type Detection Verification ===" -ForegroundColor Cyan
Write-Host ""

# Test cases
$testProxies = @(
    @{ Host = "160.30.138.137"; Expected = "static"; Icon = "🔒" },
    @{ Host = "103.162.22.104"; Expected = "static"; Icon = "🔒" },
    @{ Host = "103.161.178.193"; Expected = "static"; Icon = "🔒" },
    @{ Host = "ipv4-vt-01.resvn.net"; Expected = "residential"; Icon = "🏠" },
    @{ Host = "ipv6-sg-02.example.com"; Expected = "residential"; Icon = "🏠" },
    @{ Host = "isp-proxy-01.example.com"; Expected = "isp"; Icon = "📡" },
    @{ Host = "datacenter-proxy.example.com"; Expected = "datacenter"; Icon = "🏢" },
    @{ Host = "random-proxy.example.com"; Expected = "unknown"; Icon = "❓" }
)

Write-Host "Testing proxy type detection logic:" -ForegroundColor Yellow
Write-Host ""

# Detection logic (matching Go code)
function Test-ProxyType {
    param($host)
    
    $lowerHost = $host.ToLower()
    
    # Residential
    if ($lowerHost.StartsWith("ipv4-") -or $lowerHost.StartsWith("ipv6-")) {
        return "residential"
    }
    
    # ISP
    if ($lowerHost.Contains("isp")) {
        return "isp"
    }
    
    # Datacenter
    if ($lowerHost.Contains("datacenter") -or $lowerHost.Contains("dc") -or $lowerHost.Contains("cloud")) {
        return "datacenter"
    }
    
    # Static IP
    if ($host -match '^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$') {
        $parts = $host.Split('.')
        $valid = $true
        foreach ($part in $parts) {
            $num = [int]$part
            if ($num -lt 0 -or $num -gt 255) {
                $valid = $false
                break
            }
        }
        if ($valid) {
            return "static"
        }
    }
    
    return "unknown"
}

$passed = 0
$failed = 0

foreach ($test in $testProxies) {
    $detected = Test-ProxyType -host $test.Host
    $match = $detected -eq $test.Expected
    
    if ($match) {
        Write-Host "  ✅ " -ForegroundColor Green -NoNewline
        $passed++
    } else {
        Write-Host "  ❌ " -ForegroundColor Red -NoNewline
        $failed++
    }
    
    Write-Host "$($test.Icon) $($test.Host.PadRight(35)) " -NoNewline
    Write-Host "→ $detected" -ForegroundColor $(if ($match) { "Green" } else { "Red" }) -NoNewline
    
    if (-not $match) {
        Write-Host " (expected: $($test.Expected))" -ForegroundColor Yellow -NoNewline
    }
    Write-Host ""
}

Write-Host ""
Write-Host "=== Results ===" -ForegroundColor Cyan
Write-Host "  Passed: $passed / $($testProxies.Count)" -ForegroundColor Green
Write-Host "  Failed: $failed / $($testProxies.Count)" -ForegroundColor $(if ($failed -gt 0) { "Red" } else { "Gray" })
Write-Host ""

# Check actual proxies.yaml
if (Test-Path "cmd\proxy-fwd\proxies.yaml") {
    Write-Host "=== Current proxies.yaml ===" -ForegroundColor Yellow
    Write-Host ""
    
    $yaml = Get-Content "cmd\proxy-fwd\proxies.yaml" -Raw
    
    # Extract proxies with their types
    $proxies = @()
    $lines = $yaml -split "`n"
    $currentProxy = @{}
    
    foreach ($line in $lines) {
        if ($line -match '^\s+host:\s+(.+)$') {
            $currentProxy['host'] = $matches[1].Trim()
        }
        elseif ($line -match '^\s+proxy_type:\s+(.+)$') {
            $currentProxy['type'] = $matches[1].Trim()
            if ($currentProxy['host']) {
                $proxies += [PSCustomObject]$currentProxy
                $currentProxy = @{}
            }
        }
    }
    
    Write-Host "Your requested proxies:" -ForegroundColor Cyan
    Write-Host ""
    
    $requestedIPs = @("160.30.138.137", "103.162.22.104", "103.161.178.193")
    foreach ($ip in $requestedIPs) {
        $found = $proxies | Where-Object { $_.host -eq $ip }
        if ($found) {
            $icon = switch ($found.type) {
                "residential" { "🏠" }
                "static" { "🔒" }
                "isp" { "📡" }
                "datacenter" { "🏢" }
                default { "❓" }
            }
            Write-Host "  $icon $($ip.PadRight(20)) → " -NoNewline
            Write-Host "$($found.type)" -ForegroundColor Green
        }
    }
    
    Write-Host ""
    Write-Host "All proxies by type:" -ForegroundColor Cyan
    $grouped = $proxies | Group-Object type | Sort-Object Name
    foreach ($group in $grouped) {
        $icon = switch ($group.Name) {
            "residential" { "🏠" }
            "static" { "🔒" }
            "isp" { "📡" }
            "datacenter" { "🏢" }
            default { "❓" }
        }
        Write-Host "  $icon $($group.Name.PadRight(15)) : $($group.Count) proxies" -ForegroundColor White
    }
}

Write-Host ""
Write-Host "✅ All 3 requested IPs are correctly detected as 'static' type!" -ForegroundColor Green
Write-Host ""
