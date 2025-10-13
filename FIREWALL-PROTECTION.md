# Firewall Protection - WebRTC Leak Prevention

## 🛡️ Overview

Proxy Forward v1.3.0+ includes **automatic firewall protection** to prevent WebRTC leaks and ensure all traffic goes through proxies.

### What it does:
1. ✅ **Allows proxy-fwd.exe** to access the internet
2. ✅ **Allows browsers** to access localhost proxy ports (127.0.0.1:10001-20000)
3. ❌ **Blocks browsers** from accessing the internet directly
4. 🛡️ **Prevents WebRTC leaks** by blocking UDP/STUN/TURN requests

---

## 🚀 Quick Start

### Automatic (Recommended)

Firewall protection is **enabled by default** when running as Administrator:

```powershell
# Run as Administrator (Right-click → Run as Administrator)
.\proxy-fwd-v1.3.0.exe
```

**Output:**
```
[Firewall] Setting up firewall rules...
[Firewall] ✅ Allow rule created for proxy-fwd.exe
[Firewall] ✅ Firewall rules created for 2 browser(s)
[Firewall] 🛡️  WebRTC leak protection active
```

### Manual Control

**Disable firewall protection:**
```powershell
$env:ENABLE_FIREWALL = "false"
.\proxy-fwd-v1.3.0.exe
```

**Check if running as Admin:**
```powershell
# This command checks admin status
net session
# If succeeds → You are Admin
# If fails → You need to run as Admin
```

---

## 📋 Supported Browsers

Firewall rules are automatically created for:
- ✅ Google Chrome
- ✅ Microsoft Edge
- ✅ Mozilla Firefox
- ✅ Brave Browser

Both 32-bit and 64-bit versions are detected.

---

## 🔍 Verify Protection

### Method 1: Check Firewall Rules

```powershell
# List all ProxyFwd firewall rules
Get-NetFirewallRule -Group "ProxyFwd Rules" | Select-Object DisplayName, Action

# Expected output:
# DisplayName                                  Action
# -----------                                  ------
# ProxyFwd Allow Out                           Allow
# ProxyFwd Allow Localhost for chrome.exe      Allow
# ProxyFwd Block Internet for chrome.exe       Block
# ProxyFwd Allow Localhost for msedge.exe      Allow
# ProxyFwd Block Internet for msedge.exe       Block
```

### Method 2: Test WebRTC Leaks

Visit these websites **with browser configured to use proxy**:
- https://browserleaks.com/webrtc
- https://ipleak.net/
- https://www.dnsleaktest.com/

✅ **Pass**: Should show proxy IP, not your real IP  
❌ **Fail**: Shows your real IP → Firewall not active or misconfigured

### Method 3: API Status Check

```powershell
# Check firewall status via API
curl http://127.0.0.1:17890/api/firewall/status

# Response:
# {
#   "enabled": true,
#   "is_admin": true,
#   "rule_count": 5,
#   "rules": [...]
# }
```

---

## 🔧 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `ENABLE_FIREWALL` | `true` | Enable/disable automatic firewall setup |
| `UI_ADDR` | `127.0.0.1:17890` | UI server address |
| `ADMIN_TOKEN` | *(empty)* | Optional admin token |

### Examples

**Disable firewall:**
```powershell
$env:ENABLE_FIREWALL = "false"
.\proxy-fwd.exe
```

**Enable firewall (default):**
```powershell
$env:ENABLE_FIREWALL = "true"
.\proxy-fwd.exe
```

---

## 🧹 Manual Cleanup

If proxy-fwd doesn't cleanup rules automatically:

```powershell
# Remove all ProxyFwd firewall rules
Get-NetFirewallRule -Group "ProxyFwd Rules" | Remove-NetFirewallRule

# Or use the script
.\scripts\cleanup_firewall.ps1
```

---

## ⚠️ Important Notes

### Administrator Privileges Required

Firewall rules **require Administrator privileges**:
- ✅ Run as Admin → Rules created automatically
- ❌ Run as User → Warning logged, no protection

**Warning message when not Admin:**
```
[Firewall] Warning: Not running as Administrator. Firewall rules cannot be created.
[Firewall] To enable firewall protection, run as Administrator or manually execute: scripts\firewall_rules.ps1
```

### Automatic Cleanup

Firewall rules are **automatically removed** when proxy-fwd exits gracefully:
- ✅ Ctrl+C → Rules cleaned up
- ✅ Normal exit → Rules cleaned up
- ❌ Process killed → Rules remain (manual cleanup needed)

### Port Range

Firewall allows browser access to:
- **Range**: 127.0.0.1:10001-20000
- **Protocol**: TCP only
- **Direction**: Outbound only

If you need different ports, edit `firewall.go`:
```go
const portRange = "10001-20000"  // Change this
```

---

## 🔬 Technical Details

### Firewall Rules Created

For each browser found:

1. **Allow Localhost Rule**
   ```
   Name: ProxyFwd Allow Localhost for <browser>
   Direction: Outbound
   Action: Allow
   Program: C:\...\browser.exe
   Remote Address: 127.0.0.1
   Remote Port: 10001-20000
   Protocol: TCP
   ```

2. **Block Internet Rule**
   ```
   Name: ProxyFwd Block Internet for <browser>
   Direction: Outbound
   Action: Block
   Program: C:\...\browser.exe
   Remote Address: Any
   Protocol: TCP
   ```

For proxy-fwd.exe:

3. **Allow Outbound Rule**
   ```
   Name: ProxyFwd Allow Out
   Direction: Outbound
   Action: Allow
   Program: C:\...\proxy-fwd.exe
   Remote Address: Any
   Protocol: TCP
   ```

All rules are grouped under: `"ProxyFwd Rules"`

### Rule Priority

Windows Firewall evaluates rules in this order:
1. **Block rules** (highest priority)
2. **Allow rules**
3. Default policy

Our setup:
- Allow localhost:10001-20000 (specific)
- Block everything else (broad)

Result: Browsers can **only** access localhost proxies.

---

## 🧪 Testing

### Test 1: Browser Can Access Proxy

```powershell
# Configure browser to use: 127.0.0.1:10001
# Visit any website → Should work
```

### Test 2: Browser Cannot Access Direct Internet

```powershell
# Close all proxies in proxy-fwd
# Try to visit any website → Should fail with "Unable to connect"
```

### Test 3: WebRTC Blocked

```powershell
# Configure browser proxy
# Visit: https://browserleaks.com/webrtc
# Result: Should show proxy IP only, no real IP leaked
```

---

## 🐛 Troubleshooting

### Issue: Rules not created

**Symptoms:**
```
[Firewall] Warning: Not running as Administrator
```

**Solution:**
1. Close proxy-fwd
2. Right-click `proxy-fwd.exe`
3. Select "Run as Administrator"

---

### Issue: Browser still leaks IP

**Possible causes:**
1. Browser not in supported list
2. Rules not applied yet
3. VPN/other firewall interfering

**Solution:**
```powershell
# Check which browsers are detected
Get-NetFirewallRule -Group "ProxyFwd Rules" | Select-Object DisplayName

# Manually add your browser
New-NetFirewallRule -DisplayName "ProxyFwd Block for MyBrowser" `
  -Direction Outbound `
  -Program "C:\Path\To\MyBrowser.exe" `
  -Action Block `
  -Group "ProxyFwd Rules"
```

---

### Issue: Rules remain after exit

**Symptoms:**
Firewall rules still present after closing proxy-fwd.

**Solution:**
```powershell
# Manual cleanup
Get-NetFirewallRule -Group "ProxyFwd Rules" | Remove-NetFirewallRule
```

---

### Issue: Cannot access internet

**Symptoms:**
Everything blocked, even with proxies running.

**Solution:**
```powershell
# Disable firewall temporarily
$env:ENABLE_FIREWALL = "false"
.\proxy-fwd.exe

# Or remove rules
Get-NetFirewallRule -Group "ProxyFwd Rules" | Remove-NetFirewallRule
```

---

## 📚 Additional Resources

- **Manual Script**: `scripts\firewall_rules.ps1`
- **Cleanup Script**: `scripts\cleanup_firewall.ps1`
- **Windows Firewall Docs**: https://docs.microsoft.com/en-us/windows/security/threat-protection/windows-firewall/

---

## 🔒 Security Best Practices

1. ✅ **Always run as Administrator** for firewall protection
2. ✅ **Test WebRTC leaks** after setup
3. ✅ **Use ADMIN_TOKEN** for UI access control
4. ✅ **Keep proxy-fwd updated** for latest security fixes
5. ✅ **Disable WebRTC in browser** as additional layer (see README.md)

---

**Last Updated**: 2025-10-13  
**Version**: 1.3.0+
