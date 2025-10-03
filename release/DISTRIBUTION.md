# 🚀 Proxy Forwarder v1.0.0 - Distribution Package

## 📦 Gói Phân Phối

### 1. Portable Package (Khuyến nghị)
**File**: `proxy-forwarder-portable-v1.0.0.zip` (3.09 MB)

**Đối tượng**: Người dùng thường, không cần cài đặt

**Cách dùng**:
1. Giải nén file ZIP
2. Double-click `run.bat` hoặc `proxy-fwd.exe`
3. Mở http://127.0.0.1:17890
4. Thêm proxy và sử dụng!

**Ưu điểm**:
- ✅ Không cần cài đặt
- ✅ Portable, mang đi được
- ✅ Không cần quyền Administrator
- ✅ Dễ dàng xóa (chỉ cần xóa folder)

---

### 2. Installer Package (Chuyên nghiệp)
**File**: `proxy-forwarder-installer-v1.0.0.zip` (3.09 MB)

**Đối tượng**: Người dùng muốn cài vào hệ thống

**Cách dùng**:
1. Giải nén file ZIP
2. Click phải `INSTALL.bat` → Run as Administrator
   HOẶC
   Click phải `INSTALL-SERVICE.bat` → Run as Administrator
3. Chương trình tự động khởi động
4. Mở http://127.0.0.1:17890

**Ưu điểm**:
- ✅ Tự động start khi login Windows
- ✅ Hoặc cài như Windows Service (chạy background)
- ✅ Quản lý chuyên nghiệp
- ✅ Firewall rule tự động
- ✅ Gỡ cài đặt dễ dàng (`UNINSTALL.bat`)

**Vị trí cài đặt**: `C:\Program Files\ProxyForwarder\`

---

### 3. Standalone Executable (Siêu đơn giản)
**File**: `proxy-forwarder-v1.0.0.exe` (7.41 MB)

**Đối tượng**: Người dùng muốn chạy nhanh, 1 file duy nhất

**Cách dùng**:
1. Tải file EXE
2. Double-click để chạy
3. Mở http://127.0.0.1:17890

**Ưu điểm**:
- ✅ Chỉ 1 file
- ✅ Không cần giải nén
- ✅ Click là chạy
- ✅ Nhỏ gọn

---

## 🎯 So Sánh Các Gói

| Tiêu chí | Portable | Installer | Standalone |
|----------|----------|-----------|------------|
| Cài đặt | ❌ Không cần | ✅ Cần Admin | ❌ Không cần |
| Auto-start | ❌ | ✅ | ❌ |
| Service mode | ❌ | ✅ | ❌ |
| Size | 3.09 MB | 3.09 MB | 7.41 MB |
| Portable | ✅ | ❌ | ✅ |
| Khuyến nghị | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | ⭐⭐⭐ |

---

## 🚀 Quick Start (Dành cho tất cả gói)

### Bước 1: Chạy chương trình
Theo cách của mỗi gói (xem bên trên)

### Bước 2: Truy cập Web UI
Mở trình duyệt: **http://127.0.0.1:17890**

### Bước 3: Thêm Proxy
Format: `ip:port:user:pass` hoặc `ip:port`

Ví dụ:
```
27.79.52.120:11314:user123:pass456
103.161.178.72:8080
```

### Bước 4: Sử dụng Local Proxy
Proxy local được tạo tại:
- `127.0.0.1:10001` → Proxy #1
- `127.0.0.1:10002` → Proxy #2
- `127.0.0.1:10003` → Proxy #3

Config app/browser của bạn để dùng các proxy này.

---

## ✨ Tính Năng

### UI Features
- ✅ **Modern Web Interface** - Giao diện đẹp với Tailwind CSS
- ✅ **Vietnamese Support** - Font Inter hỗ trợ tiếng Việt hoàn hảo
- ✅ **Dark Purple Theme** - Gradient tím chuyên nghiệp
- ✅ **Responsive Design** - Tương thích mọi kích thước màn hình

### Proxy Management
- ✅ **Add Single** - Thêm từng proxy
- ✅ **Bulk Add** - Thêm nhiều proxy cùng lúc
- ✅ **API Sync** - Đồng bộ từ API endpoint
- ✅ **Export** - Xuất danh sách local ports
- ✅ **Search** - Tìm kiếm nhanh
- ✅ **Auto Refresh** - Tự động cập nhật mỗi 5s

### Advanced Features
- ✅ **HTTP/HTTPS Forwarding** - Hỗ trợ cả HTTP và HTTPS
- ✅ **Health Monitoring** - Kiểm tra sức khỏe proxy
- ✅ **Auto Shutdown** - Tự tắt proxy chết (3 fails)
- ✅ **State Persistence** - Lưu cấu hình vào `proxies.yaml`
- ✅ **Admin Token** - Bảo vệ UI với token (optional)

---

## 📖 Tài Liệu

### Trong Repository
- `README.md` - Tổng quan project
- `INSTALL-GUIDE.md` - Hướng dẫn cài đặt chi tiết
- `QUICKSTART.md` - Hướng dẫn nhanh
- `BUILD.md` - Hướng dẫn build từ source

### Online
- **GitHub**: https://github.com/Chinsusu/proxy-forwarder-windows
- **Issues**: Report bugs tại Issues tab
- **Wiki**: Tài liệu mở rộng (coming soon)

---

## 🔧 Cấu Hình Nâng Cao

### Environment Variables

```powershell
# Đổi port UI (mặc định 17890)
$env:UI_ADDR = "127.0.0.1:18000"

# Bật authentication
$env:ADMIN_TOKEN = "your-secret-token"

# Sync tự động từ API
$env:INITIAL_API = "https://api.example.com/proxies"

# Initial proxy list
$env:INITIAL_PROXIES = "ip:port:user:pass,ip2:port2"
```

---

## 🐛 Troubleshooting

### Port 17890 bị chiếm?
```powershell
netstat -ano | findstr :17890
# Đổi port bằng UI_ADDR
```

### Proxy không connect?
1. Check upstream proxy hoạt động không
2. Xem logs trong console
3. Check format: `ip:port:user:pass`

### Cần uninstall?
- Portable: Xóa folder
- Installer: Run `UNINSTALL.bat` as Admin
- Standalone: Xóa file EXE

---

## 📊 System Requirements

- **OS**: Windows 10/11 (64-bit)
- **RAM**: 50MB minimum
- **Disk**: 10MB minimum
- **Network**: Internet connection (để dùng upstream proxy)

---

## 🔒 Security

- ✅ UI chỉ listen trên `127.0.0.1` (local-only)
- ✅ Optional admin token authentication
- ✅ No telemetry, no tracking
- ✅ Open source, audit được

---

## 💡 Tips & Tricks

1. **Nhiều proxy?** → Dùng Bulk Add
2. **Nhiều máy?** → Dùng API Sync
3. **Production?** → Dùng Service mode với ADMIN_TOKEN
4. **Testing?** → Dùng Portable mode
5. **Share với team?** → Export local ports

---

## 🤝 Support

### Community
- GitHub Issues: Bug reports & feature requests
- Discussions: Q&A and community help

### Commercial Support
Contact: chinsusu@example.com

---

## 📜 License

MIT License - Free for personal and commercial use

---

## 🙏 Credits

**Built with**:
- [Go](https://golang.org/) - Programming language
- [goproxy](https://github.com/elazarl/goproxy) - HTTP proxy library
- [yaml.v3](https://github.com/go-yaml/yaml) - YAML parser
- [Tailwind CSS](https://tailwindcss.com/) - UI framework
- [Inter Font](https://fonts.google.com/specimen/Inter) - Typography

---

## 📈 Changelog

### v1.0.0 (2025-10-02)
- ✅ Initial release
- ✅ HTTP/HTTPS proxy forwarding
- ✅ Modern web UI
- ✅ Vietnamese language support
- ✅ Bulk operations
- ✅ Health monitoring
- ✅ Multiple distribution formats

---

Made with ❤️ in Vietnam 🇻🇳

**Share this project**: https://github.com/Chinsusu/proxy-forwarder-windows
