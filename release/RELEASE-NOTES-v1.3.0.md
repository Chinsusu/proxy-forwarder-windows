# Proxy Forward v1.3.0 Release Notes

**Release Date:** October 13, 2025  
**Build:** v1.3.0

---

## 🎉 What's New

### ✨ New Features

- **📍 Location Column**: Pool tab now displays proxy location information
  - Shows location between Type and Status columns
  - Helps identify proxy geographical origin
  - Searchable in filter box

- **🔐 PrivateV4 Proxy Type**: Improved proxy type naming
  - Replaced 'isp' with more accurate 'privatev4' designation
  - New lock icon (🔐) for better visual recognition
  - Maintains blue color scheme for consistency

### 🔧 Improvements

- **Better Search**: Filter now includes location field
  - Search by proxy address, type, AND location
  - More precise filtering options
  - Faster proxy discovery

- **Cleaner UI**: Removed unused sync functionality
  - "Sync from API" section removed from Pool tab
  - Simplified interface for better user experience
  - Focus on CloudMini integration

### 🗑️ Removed

- `/api/sync` endpoint (unused feature)
- `INITIAL_API` environment variable support
- Deprecated sync functionality

---

## 📦 Download

### Standalone Executable
- **File**: `proxy-fwd-v1.3.0.exe`
- **Size**: 7.52 MB
- **SHA256**: (calculate after download)

### Portable Package
- **File**: `proxy-forwarder-portable-v1.3.0.zip`
- **Size**: 3.14 MB
- **Includes**: Binary + README + Scripts

---

## 🚀 Quick Start

### For First-Time Users

1. **Download** `proxy-fwd-v1.3.0.exe`
2. **Run** the executable (double-click or via CMD)
3. **Open** browser to http://127.0.0.1:17890
4. **Add proxies** via UI or CloudMini integration

### For Existing Users

1. **Stop** current proxy-fwd instance
2. **Replace** old exe with `proxy-fwd-v1.3.0.exe`
3. **Start** new version
4. Your existing `proxies.yaml` will be preserved

---

## 🔄 Upgrade Path

**From v1.2.0 to v1.3.0:**
- ✅ Direct upgrade supported
- ✅ No database migration needed
- ✅ Existing proxies remain in pool
- ⚠️ INITIAL_API env var no longer supported (use CloudMini Sync instead)

**From v1.1.0 or earlier:**
- ✅ Direct upgrade supported
- ✅ All pool proxies preserved
- ✅ Port assignments maintained

---

## 📋 System Requirements

- **OS**: Windows 10/11 (64-bit)
- **RAM**: 50-100 MB
- **Disk**: 10 MB
- **Network**: Internet connection for upstream proxies

---

## 🛠️ Installation Methods

### Method 1: Standalone (Recommended)

```powershell
# Download and run directly
.\proxy-fwd-v1.3.0.exe

# Access UI
Start-Process http://127.0.0.1:17890
```

### Method 2: Windows Service

```powershell
# Copy to permanent location
$targetDir = "C:\ProxyFwd"
New-Item -Force -ItemType Directory $targetDir
Copy-Item .\proxy-fwd-v1.3.0.exe "$targetDir\proxy-fwd.exe"

# Create and start service
sc.exe create ProxyFwd binPath= "`"$targetDir\proxy-fwd.exe`"" start= auto DisplayName= "Proxy Forward"
sc.exe start ProxyFwd
```

### Method 3: Portable Package

```powershell
# Extract ZIP
Expand-Archive .\proxy-forwarder-portable-v1.3.0.zip -DestinationPath .\ProxyFwd

# Navigate and run
cd ProxyFwd
.\proxy-fwd-v1.3.0.exe
```

---

## 🎯 Key Features (All Versions)

- ✅ **Local HTTP Proxy**: Forward public proxies to 127.0.0.1:10001+
- ✅ **Web UI**: Manage proxies via browser
- ✅ **Proxy Pool**: Store stopped proxies without port assignment
- ✅ **CloudMini Integration**: Order and sync proxies from CloudMini
- ✅ **Health Monitoring**: Auto-shutdown on 3 consecutive failures
- ✅ **State Persistence**: Proxies saved to proxies.yaml
- ✅ **Admin Token**: Optional authentication protection

---

## 📖 Documentation

- **README**: [README.md](../README.md)
- **Quick Start**: [QUICKSTART.md](../QUICKSTART.md)
- **Build Guide**: [BUILD.md](../BUILD.md)
- **Full Changelog**: [CHANGELOG.md](../CHANGELOG.md)

---

## 🐛 Known Issues

- None reported in v1.3.0

---

## 🔐 Security Notes

- ⚠️ UI binds to 127.0.0.1 ONLY (localhost-only access)
- ⚠️ Use ADMIN_TOKEN environment variable for protection
- ⚠️ Never expose port 17890 to public network
- ✅ All proxies run on local loopback interface

---

## 💬 Support

- **GitHub Issues**: https://github.com/Chinsusu/proxy-forwarder-windows/issues
- **GitHub Discussions**: https://github.com/Chinsusu/proxy-forwarder-windows/discussions

---

## 📜 License

MIT License - See [LICENSE](../LICENSE) for details

---

## 🙏 Acknowledgments

Built with:
- Go 1.22+
- goproxy (elazarl/goproxy)
- YAML v3 (gopkg.in/yaml.v3)
- Tailwind CSS (via CDN)

---

**Enjoy Proxy Forward v1.3.0! 🚀**
