# Build script cho proxy-fwd (Windows 10)
# Tự động kiểm tra Git và build binary

Write-Host "=== Proxy Forward Builder ===" -ForegroundColor Cyan

# Kiểm tra Git
$gitExists = Get-Command git -ErrorAction SilentlyContinue
if (-not $gitExists) {
    Write-Host ""
    Write-Host "[!] Git chua duoc cai dat. Go modules can Git de tai dependencies." -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Cach cai Git nhanh:" -ForegroundColor Green
    Write-Host "  1. Dung winget:   winget install --id Git.Git -e --source winget" -ForegroundColor White
    Write-Host "  2. Dung choco:    choco install git -y" -ForegroundColor White
    Write-Host "  3. Download tu:   https://git-scm.com/download/win" -ForegroundColor White
    Write-Host ""
    Write-Host "Sau khi cai xong, khoi dong lai PowerShell va chay lai script nay." -ForegroundColor Cyan
    Write-Host ""
    
    # Thử cài qua winget nếu có
    $wingetExists = Get-Command winget -ErrorAction SilentlyContinue
    if ($wingetExists) {
        $answer = Read-Host "Ban co muon tu dong cai Git qua winget khong? (y/n)"
        if ($answer -eq "y" -or $answer -eq "Y") {
            Write-Host "Dang cai dat Git..." -ForegroundColor Yellow
            winget install --id Git.Git -e --source winget --silent
            Write-Host ""
            Write-Host "[OK] Git da duoc cai dat. Vui long KHOI DONG LAI PowerShell va chay lai script nay." -ForegroundColor Green
            exit 0
        }
    }
    
    exit 1
}

Write-Host "[OK] Git da san sang: $(git --version)" -ForegroundColor Green

# Đảm bảo đang ở thư mục gốc
$rootDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $rootDir

# Kiểm tra go.mod
if (-not (Test-Path "go.mod")) {
    Write-Host "[!] Khong tim thay go.mod trong thu muc hien tai!" -ForegroundColor Red
    exit 1
}

Write-Host "[*] Dang tai dependencies..." -ForegroundColor Cyan

# Thiết lập GOPROXY
$env:GOPROXY = "https://proxy.golang.org,direct"

# Tải dependencies
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "[!] Tai dependencies that bai. Thu mirror khac..." -ForegroundColor Yellow
    $env:GOPROXY = "https://goproxy.cn,direct"
    go mod download
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[!] Van khong tai duoc. Kiem tra ket noi mang." -ForegroundColor Red
        exit 1
    }
}

Write-Host "[OK] Dependencies da duoc tai!" -ForegroundColor Green

# Build binary
Write-Host "[*] Dang build proxy-fwd.exe..." -ForegroundColor Cyan

$outputPath = ".\cmd\proxy-fwd\proxy-fwd.exe"
go build -trimpath -ldflags "-s -w" -o $outputPath .\cmd\proxy-fwd

if ($LASTEXITCODE -ne 0) {
    Write-Host "[!] Build that bai!" -ForegroundColor Red
    exit 1
}

Write-Host "[OK] Build thanh cong!" -ForegroundColor Green
Write-Host ""
Write-Host "Binary: $outputPath" -ForegroundColor Cyan
Write-Host "Kich thuoc: $((Get-Item $outputPath).Length / 1MB) MB" -ForegroundColor White
Write-Host ""
Write-Host "=== Cach chay ===" -ForegroundColor Cyan
Write-Host "  cd .\cmd\proxy-fwd" -ForegroundColor White
Write-Host "  `$env:ADMIN_TOKEN='changeme'" -ForegroundColor White
Write-Host "  `$env:UI_ADDR='127.0.0.1:17890'" -ForegroundColor White
Write-Host "  .\proxy-fwd.exe" -ForegroundColor White
Write-Host ""
Write-Host "Mo trinh duyet: http://127.0.0.1:17890" -ForegroundColor Green
