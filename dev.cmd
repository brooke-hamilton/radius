@echo off
setlocal
set "RADIUS_DEV_GOOS="
set "RADIUS_DEV_GOARCH="
if defined GOOS set "RADIUS_DEV_GOOS=%GOOS%"
if defined GOARCH set "RADIUS_DEV_GOARCH=%GOARCH%"
set "GOOS="
set "GOARCH="
pushd "%~dp0"
go run ./cmd/dev %*
set "exit_code=%ERRORLEVEL%"
popd
exit /b %exit_code%
