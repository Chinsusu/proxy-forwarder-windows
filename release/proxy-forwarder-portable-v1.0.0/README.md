# Proxy Forward (Windows 10, Local Only)

A small Go app that **turns public proxies into local ones** on `127.0.0.1:<port>`.
Each local proxy **forwards via its assigned upstream** (`ip:port:user:pass` or `ip:port`).
When upstream fails repeatedly, the local port is stopped (clients will fail fast, no leak).

## Features

- Local-only HTTP proxy listeners: `127.0.0.1:10001`, `10002`, ...
- Web UI on `http://127.0.0.1:17890` (never binds to public).
- Add/Delete/Start/Stop proxies; *Sync from API* (line-delimited or JSON array).
- Health check every 10s; if 3 consecutive fails → stop listener.
- State persists to `proxies.yaml`.
- Optional `ADMIN_TOKEN` to protect UI/API.
- Simple **firewall kill-switch** scripts included.

> Protocol: **Local listener is an HTTP proxy**. Upstream is expected to be an **HTTP proxy** with optional Basic Auth.
> If you need SOCKS5, ask and we’ll add a `mode: socks5` variant.

## Build (Windows 10)

Install Go >= 1.22, then:

```powershell
cd cmd\proxy-fwd
go build -trimpath -ldflags "-s -w" -o proxy-fwd.exe
```

Run:

```powershell
set ADMIN_TOKEN=changeme
set UI_ADDR=127.0.0.1:17890
set INITIAL_API=http://127.0.0.1:8080/proxies.txt   # optional
# or: set INITIAL_PROXIES=1.2.3.4:8080:user:pass,2.3.4.5:3128
.\proxy-fwd.exe
```

Open `http://127.0.0.1:17890` in your browser.

## API

- `GET /api/list`
- `POST /api/add` body: `ip:port:user:pass` (or `ip:port`)
- `POST /api/remove?id=<id>`
- `POST /api/start?id=<id>`
- `POST /api/stop?id=<id>`
- `GET /api/sync?url=<API>` → accepts **lines** or **JSON array**
- `GET /api/export-local` → lines of `127.0.0.1:port`

If `ADMIN_TOKEN` is set, include `X-Admin-Token: <token>` header.

## Firewall kill-switch (optional, recommended)

**Goal:** ensure apps cannot connect to the internet directly; they must use the local proxies (127.0.0.1:10001-20000).
We provide **app-scoped rules** for popular browsers plus an **allow** rule for `proxy-fwd.exe`.

> Adjust paths as needed in `scripts\firewall_rules.ps1` before running.

```powershell
# Run as Administrator
powershell -ExecutionPolicy Bypass -File .\scripts\firewall_rules.ps1
```

To rollback:

```powershell
Get-NetFirewallRule -DisplayName "ProxyFwd *" | Remove-NetFirewallRule
```

## Run as a Windows Service

Option A (built-in):

```powershell
$EXE="C:\ProxyFwd\proxy-fwd.exe"
New-Item -Force -ItemType Directory C:\ProxyFwd | Out-Null
Copy-Item .\proxy-fwd.exe $EXE
sc.exe create ProxyFwd binPath= ""$EXE"" start= auto DisplayName= "Proxy Forward (Local)"
sc.exe start ProxyFwd
```

Option B: Use NSSM (if you prefer). Set env vars and working directory in NSSM UI.

## Notes

- UI is **forced** to bind only on `127.0.0.1`. If you try to change, app will refuse to start.
- When upstream becomes unhealthy (3x fails), local port is stopped. Clients will error instead of leaking.
- Ports begin at **10001** and increment. They are reserved per upstream; when removed, port number is not recycled in this simple version.
- State file: `proxies.yaml` in the working directory.

---

Made for your Windows 10 workflow. If you want SOCKS5 mode, multiple UI themes, or hot port reuse, we can extend it.
