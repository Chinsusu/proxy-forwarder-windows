# WARP.md

This file provides guidance to WARP (warp.dev) when working with code in this repository.

## Project Overview

This is a Windows-focused Go application called **Proxy Forward** that converts public HTTP proxies into local ones. The application creates HTTP proxy listeners on `127.0.0.1:10001+` that forward traffic through upstream proxies with automatic health checking and failover.

## Development Commands

### Build
- **Development build**: `powershell -ExecutionPolicy Bypass -File .\build.ps1`
- **Release build**: `powershell -ExecutionPolicy Bypass -File .\build-release.ps1 -Version "1.2.0"`
- **Release with package**: `powershell -ExecutionPolicy Bypass -File .\build-release.ps1 -Version "1.2.0" -CreatePackage`
- **Quick manual build**: `go build -o .\cmd\proxy-fwd\proxy-fwd.exe .\cmd\proxy-fwd`

### Run
```powershell
cd cmd\proxy-fwd
$env:ADMIN_TOKEN = "changeme"
$env:UI_ADDR = "127.0.0.1:17890"
.\proxy-fwd.exe
```

### Environment Variables
- `UI_ADDR`: Web UI address (default: "127.0.0.1:17890", must be localhost)
- `ADMIN_TOKEN`: Optional token to protect UI/API access
- `INITIAL_API`: URL to sync proxies from on startup
- `INITIAL_PROXIES`: Comma-separated list of proxies to add on startup

### Dependency Management
- **Prerequisites**: Go >= 1.22.3 and Git (required for Go modules)
- **Download dependencies**: `go mod download` (from project root)
- **Set GOPROXY if needed**: `go env -w GOPROXY=https://proxy.golang.org,direct`

## Architecture

### Core Components

**Manager (`manager.go`)**
- Central orchestrator that manages all proxy items
- Handles state persistence to `proxies.yaml`
- Manages port allocation (starting from 10001)
- Thread-safe with RWMutex protection

**ProxyItem (`types.go`)**
- Runtime wrapper around upstream proxy configuration
- Contains HTTP server, listener, health monitoring goroutine
- Tracks running state and provides graceful shutdown

**Proxy Engine (`proxy.go`)**
- Uses `github.com/elazarl/goproxy` for HTTP proxy functionality
- Custom `ConnectDial` function for HTTPS tunneling through upstream
- Health monitoring with automatic failover (3 consecutive failures)

**Web UI (`ui.go`)**
- Single-file embedded HTML/CSS/JavaScript interface
- Bootstrap-style responsive design with Tailwind CSS
- Three main tabs: Proxies (active), Pool (stopped), Order (CloudMini integration)

**HTTP Handlers (`handlers.go`)**
- REST API endpoints for proxy management
- Admin token authentication if configured
- JSON responses for programmatic access

### Key Design Patterns

**State Management**
- All proxy configurations persist to YAML file
- Port allocation with gap-filling (reuses released ports)
- Status tracking: "creating" → "live" → "dead"/"stopped"

**Health Monitoring**
- Each proxy runs independent health check goroutine
- Tests connectivity via `http://www.gstatic.com/generate_204`
- Automatic shutdown on repeated failures to prevent IP leaks

**Security Model**
- Forces localhost-only binding (127.0.0.1) for UI
- Optional admin token protection
- Firewall integration scripts for additional isolation

## File Structure

```
cmd/proxy-fwd/           # Main application directory
├── main.go             # Application entry point, signal handling
├── types.go            # Core data structures and constants
├── manager.go          # Proxy lifecycle management
├── proxy.go            # HTTP proxy implementation with health checks
├── handlers.go         # REST API endpoints
├── ui.go              # Embedded web interface
├── cloudmini.go        # CloudMini API integration
└── cloudmini_handlers.go # CloudMini API handlers
build.ps1              # Development build script
build-release.ps1      # Release build script with versioning
go.mod                 # Go module definition
CHANGELOG.md          # Version history
.project-rules.md     # Project build and versioning rules
scripts/firewall_rules.ps1  # Firewall kill-switch configuration
```

### State File
- **Location**: `<executable_directory>/proxies.yaml`
- **Content**: Proxy configurations and next port counter
- **Backup**: Atomic writes via temporary file

## Important Implementation Details

### Port Management
- Local ports start at 10001 and increment
- Gap-filling algorithm reuses released ports before incrementing
- Ports are released when proxies stop (move to pool)

### Health Checking
- Interval: 10 seconds
- Failure threshold: 3 consecutive failures
- Test URL: `http://www.gstatic.com/generate_204` (lightweight 204 response)
- Auto-shutdown prevents IP leakage on upstream failure

### Proxy Forwarding
- HTTP requests: Standard upstream proxy via Transport.Proxy
- HTTPS tunneling: Custom ConnectDial with manual CONNECT handling
- Basic authentication: Automatic header injection for upstream auth

### CloudMini Integration
- Order API for purchasing proxies
- Region-based proxy selection
- Automatic conversion to internal format
- Pool vs immediate start options

## Development Guidelines

### Code Organization
- Keep UI embedded in single file for deployment simplicity
- Maintain thread safety with proper mutex usage
- Use atomic file operations for state persistence
- Prefer graceful shutdown patterns with context cancellation

### Testing Proxy Functionality
- Use `curl -x 127.0.0.1:10001 http://httpbin.org/ip` to test proxy
- Check health status via web UI or `/api/list` endpoint
- Monitor logs for health check failures and automatic shutdowns

### Windows Service Deployment
```powershell
$targetDir = "C:\ProxyFwd"
New-Item -Force -ItemType Directory $targetDir | Out-Null
Copy-Item .\proxy-fwd.exe "$targetDir\proxy-fwd.exe"
sc.exe create ProxyFwd binPath= "`"$targetDir\proxy-fwd.exe`"" start= auto DisplayName= "Proxy Forward"
```

### Firewall Kill-Switch
- Script location: `scripts/firewall_rules.ps1`
- Must run as Administrator
- Blocks direct internet access for browsers, forces proxy usage
- Removal: `Get-NetFirewallRule -DisplayName "ProxyFwd *" | Remove-NetFirewallRule`

## API Reference

### REST Endpoints
All endpoints use `/api/` prefix on UI address (default `127.0.0.1:17890`).

- `GET /api/list` - List all proxy configurations
- `POST /api/add` - Add proxy (body: `ip:port:user:pass` or `ip:port`)
- `POST /api/add-pool` - Add to pool without starting
- `POST /api/remove?id=<id>` - Remove proxy
- `POST /api/start?id=<id>` - Start stopped proxy
- `POST /api/stop?id=<id>` - Stop running proxy
- `GET /api/sync?url=<url>` - Sync from external API
- `GET /api/export-local` - Export local endpoints
- `GET /api/check-ip?id=<id>` - Check exit IP of proxy

### Authentication
If `ADMIN_TOKEN` is set, include header: `X-Admin-Token: <token>`

## Versioning

### Current Version
- **Version**: 1.2.0
- **Release Date**: 2025-10-13
- **Changes**: Removed residential-only filter, sync all proxy types

### Version Injection
Release builds inject version info:
```powershell
# Build with version
.\build-release.ps1 -Version "1.2.0"

# Build with package
.\build-release.ps1 -Version "1.2.0" -CreatePackage
```

Output:
- `release\bin\proxy-fwd-v1.2.0.exe` - Versioned binary
- `release\proxy-forwarder-portable-v1.2.0.zip` - Portable package

### Release Process
1. Update `CHANGELOG.md` with changes
2. Run release build: `.\build-release.ps1 -Version "X.Y.Z" -CreatePackage`
3. Test binary: `cd release\bin && .\proxy-fwd-vX.Y.Z.exe`
4. Create git tag: `git tag -a vX.Y.Z -m "Release vX.Y.Z"`
5. Push tag: `git push origin vX.Y.Z`
6. Create GitHub release with artifacts

See `.project-rules.md` for complete release guidelines.

## Common Issues

### Build Problems
- **Git required**: Go modules need Git to fetch dependencies from GitHub
- **Missing go.sum**: Run `go mod download` from project root before building
- **Network issues**: Try different GOPROXY mirrors (goproxy.cn, goproxy.io)

### Runtime Issues
- **Port conflicts**: Change `UI_ADDR` environment variable
- **Upstream failures**: Check proxy credentials and connectivity
- **Health check failures**: Monitor logs, proxies auto-shutdown after 3 failures
- **Permission issues**: Firewall scripts require Administrator privileges