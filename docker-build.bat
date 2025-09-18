@echo off
setlocal
set VERSION=v1.0.0
docker build -t github.com/hellobchain/gateway-server:%VERSION% -f .\docker\Dockerfile .