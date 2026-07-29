# Create build directory if it does not exist
if (!(Test-Path -Path "build")) {
    Write-Host "Creating build directory..." -ForegroundColor Green
    New-Item -ItemType Directory -Path "build" | Out-Null
}

Write-Host "Building for Windows..." -ForegroundColor Cyan
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o build/jdterm.exe cmd/jdterm/main.go

Write-Host "Building for Ubuntu (Linux)..." -ForegroundColor Cyan
$env:GOOS="linux"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o build/jdterm_linux cmd/jdterm/main.go

Write-Host "Build complete! Executables are in the /build directory" -ForegroundColor Green


# -ldflags="-s -w": Loại bỏ thông tin debug để thu gọn dung lượng file tối đa.