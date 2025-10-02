# Hướng dẫn Build (Windows 10)

## Yêu cầu

1. **Go >= 1.22** ([tải tại đây](https://go.dev/dl/))
2. **Git for Windows** ([tải tại đây](https://git-scm.com/download/win))

> ⚠️ **Quan trọng**: Go modules cần Git để tải dependencies từ GitHub. Không thể build nếu thiếu Git.

---

## Cách 1: Sử dụng build script (Khuyến nghị)

Script `build.ps1` sẽ tự động kiểm tra Git, tải dependencies và build binary:

```powershell
# Chạy từ thư mục gốc project
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

Script sẽ:
- ✅ Kiểm tra Git đã cài chưa
- ✅ Đề nghị cài Git tự động qua winget (nếu có)
- ✅ Tải dependencies qua GOPROXY
- ✅ Build binary vào `cmd\proxy-fwd\proxy-fwd.exe`

---

## Cách 2: Build thủ công

### Bước 1: Cài đặt Git

**Option A - Winget** (Windows 10 2004+):
```powershell
winget install --id Git.Git -e --source winget
```

**Option B - Chocolatey**:
```powershell
choco install git -y
```

**Option C - Manual**: Tải từ https://git-scm.com/download/win

⚠️ **Sau khi cài Git, phải khởi động lại PowerShell!**

### Bước 2: Verify Git

```powershell
git --version
# Phải hiển thị: git version 2.x.x
```

### Bước 3: Tải dependencies

Đảm bảo đang ở **thư mục gốc** (nơi có `go.mod`):

```powershell
cd C:\Users\Administrator\Documents\proxy-fwd-windows

# Thiết lập Go proxy
go env -w GOPROXY=https://proxy.golang.org,direct

# Tải dependencies
go mod download

# (Tuỳ chọn) Nếu mạng chặn, thử mirror Trung Quốc:
# go env -w GOPROXY=https://goproxy.cn,direct
# go mod download
```

✅ Sau lệnh này, file `go.sum` sẽ được tự động tạo.

### Bước 4: Build binary

```powershell
# Build từ thư mục gốc (khuyến nghị)
go build -trimpath -ldflags "-s -w" -o .\cmd\proxy-fwd\proxy-fwd.exe .\cmd\proxy-fwd

# HOẶC build trực tiếp trong cmd\proxy-fwd (sau khi đã có go.sum)
cd cmd\proxy-fwd
go build -trimpath -ldflags "-s -w" -o proxy-fwd.exe
```

✅ Binary sẽ xuất hiện tại: `cmd\proxy-fwd\proxy-fwd.exe`

---

## Troubleshooting

### ❌ `missing go.sum entry for module`

**Nguyên nhân**: Chưa chạy `go mod download` từ thư mục gốc (nơi có `go.mod`).

**Giải pháp**:
```powershell
# Đi tới thư mục gốc
cd C:\Users\Administrator\Documents\proxy-fwd-windows

# Tải dependencies
go mod download

# Sau đó build
go build -trimpath -ldflags "-s -w" -o .\cmd\proxy-fwd\proxy-fwd.exe .\cmd\proxy-fwd
```

### ❌ `git: executable file not found in %PATH%`

**Nguyên nhân**: Git chưa được cài đặt hoặc chưa có trong PATH.

**Giải pháp**:
1. Cài Git (xem hướng dẫn ở trên)
2. Khởi động lại PowerShell
3. Verify: `git --version`

### ❌ `GOPROXY` conflict warning

**Nguyên nhân**: Biến môi trường hệ thống đang override `go env`.

**Giải pháp**:
```powershell
# Set trực tiếp trong session hiện tại
$env:GOPROXY = "https://proxy.golang.org,direct"
go mod download
```

### ❌ Network timeout khi tải dependencies

**Giải pháp**: Thử mirror khác:

```powershell
# Mirror Trung Quốc (nhanh hơn ở châu Á)
go env -w GOPROXY=https://goproxy.cn,direct
go mod download

# Mirror khác
# go env -w GOPROXY=https://goproxy.io,direct
```

### ❌ Antivirus/Defender chặn build

**Giải pháp**:
1. Tạm thời tắt Real-time Protection
2. Hoặc thêm exception cho `go.exe` và thư mục project

---

## Kiểm tra build thành công

```powershell
cd cmd\proxy-fwd

# Kiểm tra file tồn tại
Get-Item .\proxy-fwd.exe

# Kiểm tra version (sẽ hiển thị usage nếu chạy không có args)
.\proxy-fwd.exe
```

---

## Chạy application

```powershell
cd cmd\proxy-fwd

# Thiết lập biến môi trường (tuỳ chọn)
$env:ADMIN_TOKEN = "changeme"
$env:UI_ADDR = "127.0.0.1:17890"
# $env:INITIAL_API = "http://127.0.0.1:8080/proxies.txt"
# $env:INITIAL_PROXIES = "1.2.3.4:8080:user:pass,2.3.4.5:3128"

# Chạy
.\proxy-fwd.exe
```

Mở trình duyệt: **http://127.0.0.1:17890**

---

## Build optimization

Để binary nhỏ hơn:

```powershell
# Tắt debug info + strip symbols
go build -trimpath -ldflags "-s -w" -o proxy-fwd.exe

# (Tuỳ chọn) Dùng UPX để compress thêm
# upx --best --lzma proxy-fwd.exe
```

---

## Cấu trúc thư mục

```
proxy-fwd-windows/
├── go.mod              ← Module definition
├── go.sum              ← Dependencies checksums (auto-generated)
├── README.md           ← User guide
├── BUILD.md            ← Build instructions (file này)
├── build.ps1           ← Automated build script
├── cmd/
│   └── proxy-fwd/
│       ├── main.go     ← Application code
│       └── proxy-fwd.exe  ← Built binary (after build)
└── scripts/
    └── firewall_rules.ps1  ← Firewall kill-switch
```

⚠️ **Lưu ý**: Luôn build từ thư mục **gốc** (nơi có `go.mod`), không build từ `cmd\proxy-fwd` trước khi có `go.sum`.

---

## Chạy như Windows Service

Sau khi build thành công:

```powershell
# Copy binary tới thư mục cố định
$targetDir = "C:\ProxyFwd"
New-Item -Force -ItemType Directory $targetDir | Out-Null
Copy-Item .\cmd\proxy-fwd\proxy-fwd.exe "$targetDir\proxy-fwd.exe"

# Tạo service
sc.exe create ProxyFwd binPath= "`"$targetDir\proxy-fwd.exe`"" start= auto DisplayName= "Proxy Forward (Local)"

# Start service
sc.exe start ProxyFwd

# Kiểm tra status
sc.exe query ProxyFwd
```

Gỡ service:
```powershell
sc.exe stop ProxyFwd
sc.exe delete ProxyFwd
```

---

## Next steps

- 📖 Xem `README.md` để biết cách sử dụng
- 🔥 Chạy `scripts\firewall_rules.ps1` để thiết lập kill-switch (cần Admin)
- 🌐 Mở UI tại http://127.0.0.1:17890 để quản lý proxies
