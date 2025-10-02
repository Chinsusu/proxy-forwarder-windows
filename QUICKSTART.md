# Quick Start (5 phút)

## 🎯 Mục đích

Biến **public HTTP proxy** (`ip:port:user:pass`) thành **local proxy** trên `127.0.0.1:10001+` (không cần auth).
Khi upstream chết → local listener tự ngắt (không leak IP).

---

## ⚡ Build & Run nhanh

### 1️⃣ Cài Git (chỉ cần 1 lần)

```powershell
winget install --id Git.Git -e --source winget
```

Khởi động lại PowerShell sau khi cài xong!

### 2️⃣ Build binary

```powershell
cd C:\Users\Administrator\Documents\proxy-fwd-windows
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

✅ Binary xuất hiện tại: `cmd\proxy-fwd\proxy-fwd.exe`

### 3️⃣ Chạy

```powershell
cd cmd\proxy-fwd
.\proxy-fwd.exe
```

### 4️⃣ Mở UI

Trình duyệt → **http://127.0.0.1:17890**

---

## 🔧 Sử dụng cơ bản

### Thêm proxy qua UI

1. Vào http://127.0.0.1:17890
2. Nhập vào ô: `1.2.3.4:8080:username:password` (hoặc `1.2.3.4:8080` nếu không cần auth)
3. Click **Thêm**
4. Proxy sẽ xuất hiện và được map thành `127.0.0.1:10001`

### Dùng proxy local

Cấu hình trình duyệt/app để dùng: **127.0.0.1:10001** (HTTP proxy, no auth)

### Health check

- App tự check upstream mỗi 10s
- Nếu fail 3 lần liên tiếp → **tự ngắt listener local** (fail-fast, không leak IP)

---

## 🛡️ Bật Firewall Kill-Switch (Tuỳ chọn)

Chặn trình duyệt connect trực tiếp ra Internet, buộc phải đi qua local proxy:

```powershell
# Mở PowerShell AS ADMINISTRATOR
cd C:\Users\Administrator\Documents\proxy-fwd-windows
powershell -ExecutionPolicy Bypass -File .\scripts\firewall_rules.ps1
```

Rule sẽ:
- ✅ Cho phép `proxy-fwd.exe` ra Internet
- ✅ Cho phép Chrome/Edge/Firefox connect tới `127.0.0.1:10001-20000`
- ❌ Chặn tất cả outbound khác của trình duyệt

Gỡ rule:
```powershell
Get-NetFirewallRule -DisplayName "ProxyFwd *" | Remove-NetFirewallRule
```

---

## 🔄 Sync nhiều proxy cùng lúc

### Qua API

Tạo file text `proxies.txt`:
```
1.2.3.4:8080:user1:pass1
2.3.4.5:3128:user2:pass2
5.6.7.8:8888
```

Host file đó (ví dụ: `http://127.0.0.1:8000/proxies.txt`), rồi trong UI:
1. Nhập URL vào ô API
2. Click **Sync từ API**

### Qua env variable (khi start)

```powershell
$env:INITIAL_API = "http://127.0.0.1:8000/proxies.txt"
.\proxy-fwd.exe
```

Hoặc:

```powershell
$env:INITIAL_PROXIES = "1.2.3.4:8080:user:pass,5.6.7.8:3128"
.\proxy-fwd.exe
```

---

## 📋 API Endpoints

Tất cả endpoint ở `http://127.0.0.1:17890/api/`:

| Endpoint | Method | Mô tả |
|----------|--------|-------|
| `/api/list` | GET | Danh sách proxies |
| `/api/add` | POST | Thêm proxy (body: `ip:port:user:pass`) |
| `/api/remove?id=xxx` | POST | Xóa proxy |
| `/api/start?id=xxx` | POST | Start proxy đã stop |
| `/api/stop?id=xxx` | POST | Stop proxy |
| `/api/sync?url=xxx` | GET | Sync từ API/URL |
| `/api/export-local` | GET | Export list `127.0.0.1:port` |

Nếu đặt `ADMIN_TOKEN`, thêm header: `X-Admin-Token: <token>`

---

## 🔐 Bảo vệ UI

```powershell
$env:ADMIN_TOKEN = "my-secret-token"
.\proxy-fwd.exe
```

Khi gọi API, thêm header:
```bash
curl -H "X-Admin-Token: my-secret-token" http://127.0.0.1:17890/api/list
```

---

## 🏃 Chạy như Windows Service

```powershell
# Copy binary
$targetDir = "C:\ProxyFwd"
New-Item -Force -ItemType Directory $targetDir | Out-Null
Copy-Item .\proxy-fwd.exe "$targetDir\proxy-fwd.exe"

# Tạo service
sc.exe create ProxyFwd binPath= "`"$targetDir\proxy-fwd.exe`"" start= auto DisplayName= "Proxy Forward"
sc.exe start ProxyFwd
```

Stop & xóa:
```powershell
sc.exe stop ProxyFwd
sc.exe delete ProxyFwd
```

---

## ❓ Troubleshooting

### Build không được?
→ Xem `BUILD.md` để biết chi tiết troubleshooting

### Port đã bị dùng?
```powershell
$env:UI_ADDR = "127.0.0.1:18000"  # Đổi port khác
.\proxy-fwd.exe
```

### Upstream proxy fail?
- Kiểm tra credentials đúng chưa
- Kiểm tra upstream có hoạt động không: `curl -x ip:port:user:pass http://google.com`
- App sẽ tự stop listener sau 3 lần fail (xem status trong UI)

### Muốn restart proxy sau khi fix?
- Vào UI, click **Start** lại proxy đã stop

---

## 📚 Đọc thêm

- `README.md` - Hướng dẫn chi tiết features
- `BUILD.md` - Hướng dẫn build và troubleshooting đầy đủ
- `scripts/firewall_rules.ps1` - Script firewall kill-switch

---

## 🎉 Xong!

Bây giờ bạn có:
- ✅ Local proxy không cần auth trên `127.0.0.1:10001+`
- ✅ Auto health-check và fail-safe (không leak IP)
- ✅ Web UI quản lý dễ dàng
- ✅ Firewall kill-switch (tuỳ chọn)

**Happy proxying!** 🚀
