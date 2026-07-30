# Create build directory if it does not exist
if (!(Test-Path -Path "build")) {
    Write-Host "Creating build directory..." -ForegroundColor Green
    New-Item -ItemType Directory -Path "build" | Out-Null
}

# 1. Windows 64-bit
Write-Host "Building for Windows..." -ForegroundColor Cyan
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o build/jdterm.exe cmd/jdterm/main.go

# 2. Ubuntu / Linux PC 64-bit
Write-Host "Building for Ubuntu (Linux)..." -ForegroundColor Cyan
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o build/jdterm_linux cmd/jdterm/main.go

# 3. Raspberry Pi OS (64-bit: Pi 3, Pi 4, Pi 5)
Write-Host "Building for Raspberry Pi OS (ARM64)..." -ForegroundColor Cyan
$env:GOOS="linux"
$env:GOARCH="arm64"
go build -ldflags="-s -w" -o build/jdterm_rpi_arm64 cmd/jdterm/main.go

# 4. Raspberry Pi OS (32-bit: Các dòng Pi cũ hoặc PiOS 32-bit)
Write-Host "Building for Raspberry Pi OS (ARMv7 32-bit)..." -ForegroundColor Cyan
$env:GOOS="linux"
$env:GOARCH="arm"
$env:GOARM="7"
go build -ldflags="-s -w" -o build/jdterm_rpi_arm32 cmd/jdterm/main.go

Write-Host "Build complete! Executables are in the /build directory" -ForegroundColor Green

# Ghi chú:
# -ldflags="-s -w": Loại bỏ thông tin debug để thu gọn dung lượng file tối đa.