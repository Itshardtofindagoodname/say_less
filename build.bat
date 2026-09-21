@echo off
echo Building Say Less for all platforms...
echo.

set GOOS=windows
set GOARCH=amd64
echo Building Windows amd64...
go build -o dist/sale-windows-amd64.exe ./cmd/sale
if %errorlevel% neq 0 (
    echo Failed to build Windows amd64
    exit /b 1
)

set GOOS=darwin
set GOARCH=amd64
echo Building macOS amd64...
go build -o dist/sale-macos-amd64 ./cmd/sale
if %errorlevel% neq 0 (
    echo Failed to build macOS amd64
    exit /b 1
)

set GOOS=darwin
set GOARCH=arm64
echo Building macOS arm64...
go build -o dist/sale-macos-arm64 ./cmd/sale
if %errorlevel% neq 0 (
    echo Failed to build macOS arm64
    exit /b 1
)

set GOOS=linux
set GOARCH=amd64
echo Building Linux amd64...
go build -o dist/sale-linux-amd64 ./cmd/sale
if %errorlevel% neq 0 (
    echo Failed to build Linux amd64
    exit /b 1
)

set GOOS=linux
set GOARCH=arm64
echo Building Linux arm64...
go build -o dist/sale-linux-arm64 ./cmd/sale
if %errorlevel% neq 0 (
    echo Failed to build Linux arm64
    exit /b 1
)

echo.
echo All builds complete! Binaries are in dist/
echo.
echo Files:
dir /b dist\
