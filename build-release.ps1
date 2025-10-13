# Release Build Script for Proxy Forward
# Builds versioned binary with metadata injection

param(
    [Parameter(Mandatory=$false)]
    [string]$Version = "1.2.0",
    
    [Parameter(Mandatory=$false)]
    [switch]$CreatePackage = $false,
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipTests = $false
)

$ErrorActionPreference = "Stop"

Write-Host "=== Proxy Forward Release Builder ===" -ForegroundColor Cyan
Write-Host ""

# Get build timestamp
$BuildTime = Get-Date -Format "yyyy-MM-dd_HH:mm:ss"
$BuildDate = Get-Date -Format "yyyy-MM-dd"

# Validate version format
if ($Version -notmatch '^\d+\.\d+\.\d+$') {
    Write-Host "[!] Invalid version format. Use MAJOR.MINOR.PATCH (e.g., 1.2.0)" -ForegroundColor Red
    exit 1
}

Write-Host "Version:    $Version" -ForegroundColor Green
Write-Host "Build Time: $BuildTime" -ForegroundColor Green
Write-Host ""

# Check Git
$gitExists = Get-Command git -ErrorAction SilentlyContinue
if (-not $gitExists) {
    Write-Host "[!] Git not found. Please install Git first." -ForegroundColor Yellow
    exit 1
}

# Check Go
$goExists = Get-Command go -ErrorAction SilentlyContinue
if (-not $goExists) {
    Write-Host "[!] Go not found. Please install Go first." -ForegroundColor Red
    exit 1
}

Write-Host "[OK] Git: $(git --version)" -ForegroundColor Green
Write-Host "[OK] Go:  $(go version)" -ForegroundColor Green
Write-Host ""

# Get project root
$ProjectRoot = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ProjectRoot

# Verify go.mod exists
if (-not (Test-Path "go.mod")) {
    Write-Host "[!] go.mod not found in current directory!" -ForegroundColor Red
    exit 1
}

# Run tests (unless skipped)
if (-not $SkipTests) {
    Write-Host "[*] Running tests..." -ForegroundColor Cyan
    go test ./... -v
    if ($LASTEXITCODE -ne 0) {
        Write-Host "[!] Tests failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "[OK] All tests passed!" -ForegroundColor Green
    Write-Host ""
}

# Download dependencies
Write-Host "[*] Downloading dependencies..." -ForegroundColor Cyan
go mod download
if ($LASTEXITCODE -ne 0) {
    Write-Host "[!] Failed to download dependencies!" -ForegroundColor Red
    exit 1
}
Write-Host "[OK] Dependencies downloaded!" -ForegroundColor Green
Write-Host ""

# Clean and create release directory
$ReleaseDir = ".\release\bin"
Write-Host "[*] Preparing release directory: $ReleaseDir" -ForegroundColor Cyan
Remove-Item -Recurse -Force $ReleaseDir -ErrorAction SilentlyContinue
New-Item -ItemType Directory -Force $ReleaseDir | Out-Null

# Build binary with version injection
$OutputFile = "$ReleaseDir\proxy-fwd-v$Version.exe"
$LdFlags = "-s -w -X main.version=$Version -X main.buildTime=$BuildTime"

Write-Host "[*] Building release binary..." -ForegroundColor Cyan
Write-Host "    Output: $OutputFile" -ForegroundColor White
Write-Host "    Flags:  $LdFlags" -ForegroundColor White
Write-Host ""

go build -trimpath -ldflags $LdFlags -o $OutputFile .\cmd\proxy-fwd

if ($LASTEXITCODE -ne 0) {
    Write-Host "[!] Build failed!" -ForegroundColor Red
    exit 1
}

# Check if file was created
if (-not (Test-Path $OutputFile)) {
    Write-Host "[!] Binary not found after build!" -ForegroundColor Red
    exit 1
}

$FileSize = (Get-Item $OutputFile).Length / 1MB
Write-Host "[OK] Build successful!" -ForegroundColor Green
Write-Host ""
Write-Host "Binary:     $OutputFile" -ForegroundColor Cyan
Write-Host "Size:       $($FileSize.ToString('F2')) MB" -ForegroundColor White
Write-Host "Version:    $Version" -ForegroundColor White
Write-Host "Build Time: $BuildTime" -ForegroundColor White
Write-Host ""

# Copy to release root for convenience
$ReleaseRootFile = ".\release\proxy-fwd.exe"
Copy-Item $OutputFile $ReleaseRootFile -Force
Write-Host "[*] Copied to: $ReleaseRootFile" -ForegroundColor Cyan
Write-Host ""

# Create portable package if requested
if ($CreatePackage) {
    Write-Host "[*] Creating portable package..." -ForegroundColor Cyan
    
    $PackageName = "proxy-forwarder-portable-v$Version.zip"
    $PackagePath = ".\release\$PackageName"
    
    # Remove old package if exists
    Remove-Item $PackagePath -ErrorAction SilentlyContinue
    
    # Create temporary staging directory
    $StagingDir = ".\release\staging"
    Remove-Item -Recurse -Force $StagingDir -ErrorAction SilentlyContinue
    New-Item -ItemType Directory -Force $StagingDir | Out-Null
    
    # Copy files to staging
    Copy-Item $OutputFile "$StagingDir\proxy-fwd.exe" -Force
    Copy-Item ".\README.md" "$StagingDir\" -Force -ErrorAction SilentlyContinue
    Copy-Item ".\QUICKSTART.md" "$StagingDir\" -Force -ErrorAction SilentlyContinue
    Copy-Item ".\BUILD.md" "$StagingDir\" -Force -ErrorAction SilentlyContinue
    
    # Copy scripts directory
    if (Test-Path ".\scripts") {
        Copy-Item ".\scripts" "$StagingDir\scripts" -Recurse -Force
    }
    
    # Create version info file
    $VersionInfo = @"
Proxy Forward v$Version
Build Date: $BuildDate
Build Time: $BuildTime

Files:
- proxy-fwd.exe       : Main executable
- README.md           : User documentation
- QUICKSTART.md       : Quick start guide
- BUILD.md            : Build instructions
- scripts/            : Helper scripts

Installation:
1. Extract all files to desired location
2. Run proxy-fwd.exe
3. Open http://127.0.0.1:17890 in browser

For more info, see README.md
"@
    
    $VersionInfo | Out-File "$StagingDir\VERSION.txt" -Encoding UTF8
    
    # Create package
    Compress-Archive -Path "$StagingDir\*" -DestinationPath $PackagePath -Force
    
    # Clean up staging
    Remove-Item -Recurse -Force $StagingDir
    
    $PackageSize = (Get-Item $PackagePath).Length / 1MB
    Write-Host "[OK] Package created!" -ForegroundColor Green
    Write-Host "    File: $PackagePath" -ForegroundColor Cyan
    Write-Host "    Size: $($PackageSize.ToString('F2')) MB" -ForegroundColor White
    Write-Host ""
}

# Summary
Write-Host "=== Build Complete ===" -ForegroundColor Green
Write-Host ""
Write-Host "Release artifacts:" -ForegroundColor Cyan
Write-Host "  - Binary:  $OutputFile" -ForegroundColor White
Write-Host "  - Copy:    $ReleaseRootFile" -ForegroundColor White
if ($CreatePackage) {
    Write-Host "  - Package: .\release\proxy-forwarder-portable-v$Version.zip" -ForegroundColor White
}
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Test the binary:" -ForegroundColor White
Write-Host "     cd release\bin" -ForegroundColor Gray
Write-Host "     .\proxy-fwd-v$Version.exe" -ForegroundColor Gray
Write-Host ""
Write-Host "  2. Create git tag:" -ForegroundColor White
Write-Host "     git tag -a v$Version -m `"Release v$Version`"" -ForegroundColor Gray
Write-Host "     git push origin v$Version" -ForegroundColor Gray
Write-Host ""
Write-Host "  3. Create GitHub release and upload:" -ForegroundColor White
Write-Host "     - proxy-fwd-v$Version.exe" -ForegroundColor Gray
if ($CreatePackage) {
    Write-Host "     - proxy-forwarder-portable-v$Version.zip" -ForegroundColor Gray
}
Write-Host ""
