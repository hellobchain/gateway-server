@echo off
setlocal
cd /d "%~dp0cmd"
go mod tidy
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -trimpath -ldflags="-s -w" -o ..\bin\gateway-server.bin