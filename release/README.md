# Proxy Forward v1.0.0

HTTP/HTTPS proxy forwarder with CloudMini integration and web UI.

## Features

- ✅ Forward remote proxies to local ports (127.0.0.1:10001+)
- ✅ Web-based UI for proxy management
- ✅ CloudMini API integration (order & sync residential proxies)
- ✅ Proxy pool management (stop/start proxies)
- ✅ Port reuse optimization
- ✅ Config persistence (proxies.yaml)
- ✅ Health monitoring with auto-restart
- ✅ Exit IP checking

## Installation

1. Download `proxy-fwd-v1.0.0-windows-amd64.exe`
2. Place it in a folder (e.g., `C:\ProxyForward\`)
3. Run the executable

## Usage

### Start the server

```cmd
proxy-fwd-v1.0.0-windows-amd64.exe
```

The server will:
- Listen on `http://127.0.0.1:17890` (web UI)
- Load config from `proxies.yaml` (same directory as exe)
- Auto-create config file on first run

### Access Web UI

Open browser: `http://127.0.0.1:17890`

### Tabs

1. **Proxies**: Active proxies with assigned local ports
2. **Order**: Order new CloudMini proxies
3. **Pool**: Stopped/unassigned proxies

### Add Proxies

**Method 1: Manual Add**
- Format: `ip:port:user:pass` or `ip:port`
- Click "Add Proxy"

**Method 2: Bulk Add**
- Click "Bulk Add" button
- Paste multiple proxies (one per line)

**Method 3: CloudMini Sync**
- Go to **Pool tab**
- Enter CloudMini API token
- Click "Sync All Proxy-Res"
- All residential proxies will be added to pool

**Method 4: CloudMini Order**
- Go to **Order tab**
- Enter CloudMini API token
- Click "Load Regions"
- Select region and quantity
- Click "Order Now"

### Proxy Operations

- **Start**: Start proxy and assign local port
- **Stop**: Stop proxy and release port (move to pool)
- **Check IP**: Verify exit IP through proxy
- **Copy**: Copy local proxy address to clipboard

### Environment Variables

```cmd
# Optional: Set admin token for API access
set ADMIN_TOKEN=your-secret-token

# Optional: Custom UI port
set UI_ADDR=127.0.0.1:8080
```

## Configuration

The `proxies.yaml` file is automatically created in the same directory as the executable.

Example:
```yaml
items:
  - id: example-proxy-1
    host: proxy.example.com
    port: 8080
    user: username
    pass: password
    local_port: 10001
    status: live
next: 10002
```

## Port Range

- UI: `127.0.0.1:17890` (default)
- Proxies: `127.0.0.1:10001` - `127.0.0.1:xxxxx`

## Troubleshooting

**UI not accessible:**
- Check if port 17890 is available
- Try: `netstat -ano | findstr 17890`

**Proxy won't start:**
- Check upstream proxy credentials
- Verify remote proxy is online
- Check firewall settings

**Config not loading:**
- Ensure `proxies.yaml` is in same folder as exe
- Check file permissions
- Review console logs for errors

## Building from Source

```bash
git clone https://github.com/Chinsusu/proxy-forwarder-windows.git
cd proxy-forwarder-windows
go build -ldflags "-s -w" -o proxy-fwd.exe ./cmd/proxy-fwd
```

## License

MIT License

## Support

GitHub: https://github.com/Chinsusu/proxy-forwarder-windows
