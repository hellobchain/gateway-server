@echo off
setlocal
set VERSION=v1.0.0
docker build -t dm/gateway-server:%VERSION% -f .\docker\Dockerfile .