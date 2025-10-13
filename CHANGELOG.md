# Changelog

All notable changes to Proxy Forward will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Unit tests for core components
- SOCKS5 proxy support
- Proxy rotation/load balancing
- Usage statistics and analytics
- WebSocket for real-time UI updates

## [1.4.0] - 2025-10-13

### Added
- **🔥 Firewall Protection**: Automatic Windows Firewall rules to prevent WebRTC leaks
- Auto-detect and protect Chrome, Edge, Firefox, Brave browsers
- Environment variable `ENABLE_FIREWALL` (default: true)
- API endpoint `/api/firewall/status` to check protection status
- Automatic cleanup of firewall rules on graceful exit
- New module: `firewall.go` with complete firewall management
- Documentation: `FIREWALL-PROTECTION.md` with full setup guide

### Changed
- Requires Administrator privileges for full firewall protection
- Firewall enabled by default (can be disabled with `ENABLE_FIREWALL=false`)
- Enhanced security with browser isolation from direct internet access

### Security
- **WebRTC Leak Prevention**: Blocks browsers from bypassing proxy
- **Kill-Switch**: Prevents direct internet access when proxies fail
- **Automatic Protection**: No manual firewall configuration needed

### Technical Details
- Added `setupFirewall()` - Creates firewall rules on startup
- Added `cleanupFirewall()` - Removes rules on shutdown
- Added `isAdmin()` - Checks Administrator privileges
- Port range protected: 127.0.0.1:10001-20000
- Rules grouped under "ProxyFwd Rules" for easy management

## [1.3.0] - 2025-10-13

### Added
- **Location column** in Pool tab to display proxy location
- Search filter now includes location field for better filtering

### Changed
- **Proxy type**: Replaced 'isp' with 'privatev4' (icon: 🔐 lock)
- Updated type icons and colors for better visual clarity
- Improved Pool tab layout with location information

### Removed
- **Sync from API** section removed from Pool tab
- `/api/sync` endpoint and `syncFromAPI()` function removed
- `INITIAL_API` environment variable support removed
- Cleaned up unused imports in manager.go

### Technical Details
- Added `Location` field to `Upstream` struct in types.go
- Updated Pool table to display 6 columns (#, Proxy Address, Type, Location, Status, Action)
- Modified filterPool() to search by location alongside type and address
- Removed syncFromAPI() method from Manager
- Updated UI colspan values to match new column count

## [1.2.0] - 2025-10-13

### Changed
- **CloudMini Sync**: Removed residential-only filter in `handleCloudMiniSync`
- Now syncs **all proxy types** from CloudMini API (residential, ISP, static, datacenter)
- Improved logging messages to reflect all-proxy sync behavior

### Fixed
- Better hostname extraction for non-residential proxies
- More detailed debug logging for first 5 proxies in sync

### Technical Details
- Modified `cmd/proxy-fwd/cloudmini_handlers.go` line 244-274
- Removed `isResidential` filter logic
- Changed log from "proxy-res residential only" to "no filtering"

## [1.1.0] - 2025-10-12

### Added
- CloudMini API integration for ordering proxies
- CloudMini sync functionality to import existing proxies
- Proxy pool management (stopped proxies without ports)
- Bulk add/remove operations in UI
- Exit IP checking functionality
- Auto-refresh option in UI (5-second interval)
- Region selection for CloudMini orders
- Auto-start option for ordered proxies

### Changed
- Port allocation now reuses released ports (gap-filling algorithm)
- Stopped proxies release their ports and move to pool
- Improved UI with three tabs: Proxies, Pool, Order

### Fixed
- Memory leak in health check goroutines
- Race condition in port allocation
- State persistence issues on rapid start/stop

## [1.0.0] - 2025-10-01

### Added
- Initial release of Proxy Forward
- HTTP proxy forwarding from public to local (127.0.0.1:10001+)
- Web UI for proxy management (http://127.0.0.1:17890)
- Health monitoring with automatic failover (3 consecutive failures)
- State persistence to `proxies.yaml`
- Admin token authentication support
- Graceful shutdown with context cancellation
- Windows service support
- Firewall kill-switch scripts

### Features
- Local-only HTTP proxy listeners on 127.0.0.1
- Upstream proxy support with Basic Auth
- Add/Remove/Start/Stop proxy operations
- API sync from external endpoints
- Export local proxy list
- Automatic port assignment starting from 10001

### Security
- Forces localhost-only binding (127.0.0.1)
- Optional admin token protection via `X-Admin-Token` header
- Firewall integration for preventing direct internet access

### Architecture
- Manager pattern for proxy lifecycle management
- Thread-safe operations with RWMutex
- Atomic state file operations
- Independent health check goroutines per proxy
- Custom HTTPS tunneling via `ConnectDial`

---

## Release Notes Format

### Types of Changes
- `Added` for new features
- `Changed` for changes in existing functionality
- `Deprecated` for soon-to-be removed features
- `Removed` for now removed features
- `Fixed` for any bug fixes
- `Security` for vulnerability fixes

### Version Links
- [Unreleased]: https://github.com/yourusername/proxy-fwd-windows/compare/v1.2.0...HEAD
- [1.2.0]: https://github.com/yourusername/proxy-fwd-windows/compare/v1.1.0...v1.2.0
- [1.1.0]: https://github.com/yourusername/proxy-fwd-windows/compare/v1.0.0...v1.1.0
- [1.0.0]: https://github.com/yourusername/proxy-fwd-windows/releases/tag/v1.0.0
