# jdterm

JD Serial Terminal WebApp

## Project Feature

## Run

```powershell
go run ./cmd/jdterm/main.go
```

## Build

### Windows PowerShell

```powershell

if (!(Test-Path -Path "build")) { New-Item -ItemType Directory -Path "build" | Out-Null }; $env:GOOS="windows"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o build/jdterm.exe cmd/jdterm/main.go; $env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o build/jdterm_linux cmd/jdterm/main.go

```

Hoặc thực thi lệnh

```powershell
./build.ps1
```
