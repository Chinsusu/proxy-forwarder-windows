# 📦 Hướng Dẫn Cài Đặt Proxy Forwarder

## 🎯 3 Cách Sử Dụng

### 1️⃣ **Portable (Khuyến nghị cho người dùng thường)**
✅ Không cần cài đặt, chạy ngay  
✅ Dễ dàng mang đi  
✅ Không cần quyền Administrator

**Các bước:**
1. Tải file `proxy-forwarder-portable-v1.0.0.zip`
2. Giải nén ra thư mục bất kỳ
3. Double-click `run.bat` hoặc `proxy-fwd.exe`
4. Mở trình duyệt: http://127.0.0.1:17890

**Files trong package:**
- `proxy-fwd.exe` - Chương trình chính
- `run.bat` - Script khởi chạy nhanh
- `proxies.yaml` - File cấu hình (tự tạo sau lần đầu chạy)

---

### 2️⃣ **Installer (Cài đặt vào hệ thống)**
✅ Tự động khởi động cùng Windows  
✅ Hoặc cài như Windows Service  
✅ Quản lý chuyên nghiệp

**Các bước:**
1. Tải file `proxy-forwarder-installer-v1.0.0.zip`
2. Giải nén
3. **Chọn 1 trong 2 cách:**

#### Cách A: Auto-start khi login (Đơn giản)
- Click phải `INSTALL.bat` → **Run as Administrator**
- Chương trình sẽ tự chạy mỗi khi bạn login Windows

#### Cách B: Cài như Windows Service (Chuyên nghiệp)
- Click phải `INSTALL-SERVICE.bat` → **Run as Administrator**
- Service sẽ tự chạy ngay cả khi chưa login

**Quản lý sau khi cài:**
```powershell
# Nếu cài như service:
Start-Service ProxyForwarder      # Start
Stop-Service ProxyForwarder       # Stop
Get-Service ProxyForwarder        # Check status
Restart-Service ProxyForwarder    # Restart
```

**Gỡ cài đặt:**
- Click phải `UNINSTALL.bat` → **Run as Administrator**

---

### 3️⃣ **Standalone EXE (Siêu đơn giản)**
✅ Chỉ 1 file duy nhất  
✅ Click là chạy  
✅ Nhỏ gọn (~7MB)

**Các bước:**
1. Tải file `proxy-forwarder-v1.0.0.exe`
2. Double-click để chạy
3. Mở trình duyệt: http://127.0.0.1:17890

---

## 🚀 Sau Khi Cài Đặt

### Bước 1: Mở Web UI
Truy cập: **http://127.0.0.1:17890**

### Bước 2: Thêm Proxy
Nhập proxy theo format:
- Có auth: `ip:port:user:pass`
- Không auth: `ip:port`

**Ví dụ:**
```
27.79.52.120:11314:user123:pass456
103.161.178.72:8080
```

### Bước 3: Sử Dụng
Proxy local sẽ available tại:
- `127.0.0.1:10001` → Proxy đầu tiên
- `127.0.0.1:10002` → Proxy thứ hai
- `127.0.0.1:10003` → Proxy thứ ba
- ...

### Bước 4: Config Ứng Dụng
Trong app/browser, set proxy thành:
```
HTTP Proxy: 127.0.0.1
Port: 10001 (hoặc 10002, 10003...)
```

---

## 🎨 Các Tính Năng Trong UI

✅ **Add Single Proxy** - Thêm từng proxy  
✅ **Bulk Add** - Thêm nhiều proxy cùng lúc (mỗi dòng 1 proxy)  
✅ **Sync from API** - Sync từ API endpoint  
✅ **Export Local Ports** - Xuất danh sách local ports  
✅ **Search** - Tìm kiếm proxy  
✅ **Auto Refresh** - Tự động refresh mỗi 5s  
✅ **Start/Stop/Delete** - Quản lý từng proxy  

---

## ⚙️ Cấu Hình Nâng Cao (Optional)

### Đổi Port của UI
Mặc định UI chạy port 17890. Để đổi:
```powershell
# Windows
$env:UI_ADDR = "127.0.0.1:18000"
.\proxy-fwd.exe
```

### Bật Authentication
Để bảo vệ UI với token:
```powershell
$env:ADMIN_TOKEN = "your-secret-token-here"
.\proxy-fwd.exe
```

Khi bật, cần nhập token vào ô "Admin Token" trong UI.

### Sync Tự Động Từ API
```powershell
$env:INITIAL_API = "https://api.example.com/proxies"
.\proxy-fwd.exe
```

---

## 🔥 Firewall

Installer tự động thêm rule cho port 17890.

Nếu cần thêm thủ công:
```powershell
New-NetFirewallRule -DisplayName "Proxy Forwarder UI" `
    -Direction Inbound `
    -Action Allow `
    -Protocol TCP `
    -LocalPort 17890
```

---

## 🐛 Troubleshooting

### UI không mở được?
1. Kiểm tra port 17890 có bị chiếm không:
   ```powershell
   netstat -ano | findstr :17890
   ```
2. Thử đổi port khác (dùng biến `UI_ADDR`)

### Proxy không kết nối?
1. Kiểm tra upstream proxy có hoạt động không
2. Xem logs trong terminal/command prompt
3. Kiểm tra format proxy: `ip:port:user:pass`

### Service không start?
1. Kiểm tra logs:
   ```powershell
   Get-EventLog -LogName Application -Source ProxyForwarder -Newest 10
   ```
2. Thử run manual để xem lỗi:
   ```powershell
   cd "C:\Program Files\ProxyForwarder"
   .\proxy-fwd.exe
   ```

---

## 📂 Vị Trí Files

### Portable:
- Mọi thứ trong folder bạn giải nén
- `proxies.yaml` cùng folder với exe

### Installed:
- Program: `C:\Program Files\ProxyForwarder\`
- Data: `C:\Program Files\ProxyForwarder\data\`
- Config: `C:\Program Files\ProxyForwarder\data\proxies.yaml`

---

## 🔗 Links

- **GitHub**: https://github.com/Chinsusu/proxy-forwarder-windows
- **Issues**: https://github.com/Chinsusu/proxy-forwarder-windows/issues
- **Releases**: https://github.com/Chinsusu/proxy-forwarder-windows/releases

---

## 💡 Tips

1. **Nhiều proxy**: Bulk add nhanh hơn là add từng cái
2. **API sync**: Nếu có nhiều máy, sync từ 1 API trung tâm
3. **Health check**: App tự động tắt proxy khi upstream die (3 fails)
4. **Copy local port**: Click nút "Copy" bên cạnh mỗi local port
5. **Auto refresh**: Bật để real-time monitor status

---

Made with ❤️ in Vietnam 🇻🇳
