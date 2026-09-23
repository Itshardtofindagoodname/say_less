@echo off
echo Building Say Less for all platforms...
echo.

if not exist dist mkdir dist
if errorlevel 1 (
    echo Failed to create dist directory
    exit /b 1
)

set GOOS=windows
set GOARCH=amd64
echo Building Windows amd64...
go build -o dist/sale-windows-amd64.exe ./cmd/sale
if errorlevel 1 (
    echo Failed to build Windows amd64
    exit /b 1
)

echo Building Windows installer (say_less.exe)...
copy /Y dist\sale-windows-amd64.exe cmd\installer\sale.exe >nul
if errorlevel 1 goto :windows_fail
copy /Y assets\say_less_logo.png cmd\installer\say_less_logo.png >nul
if errorlevel 1 goto :windows_fail
copy /Y assets\say_less_portrait_cover.png cmd\installer\say_less_portrait_cover.png >nul
if errorlevel 1 goto :windows_fail
copy /Y assets\font\GoogleSansCodeNerdFontPropo-Regular.ttf cmd\installer\GoogleSansCodeNerdFontPropo-Regular.ttf >nul
if errorlevel 1 goto :windows_fail
copy /Y assets\font\GoogleSansCodeNerdFontPropo-SemiBold.ttf cmd\installer\GoogleSansCodeNerdFontPropo-SemiBold.ttf >nul
if errorlevel 1 goto :windows_fail

go build -tags saylesspayload -ldflags "-H=windowsgui -s -w" -o dist/say_less.exe ./cmd/installer
if errorlevel 1 goto :windows_fail

del cmd\installer\sale.exe >nul 2>nul
del cmd\installer\say_less_logo.png >nul 2>nul
del cmd\installer\say_less_portrait_cover.png >nul 2>nul
del cmd\installer\GoogleSansCodeNerdFontPropo-Regular.ttf >nul 2>nul
del cmd\installer\GoogleSansCodeNerdFontPropo-SemiBold.ttf >nul 2>nul
echo say_less.exe built successfully

echo Syncing installer into the repository root...
copy /Y dist\say_less.exe say_less.exe >nul
if errorlevel 1 (
    echo Failed to sync say_less.exe
    exit /b 1
)

set GOOS=darwin
set GOARCH=amd64
echo Building macOS amd64...
go build -o dist/sale-macos-amd64 ./cmd/sale
if errorlevel 1 (
    echo Failed to build macOS amd64
    exit /b 1
)

set GOOS=darwin
set GOARCH=arm64
echo Building macOS arm64...
go build -o dist/sale-macos-arm64 ./cmd/sale
if errorlevel 1 (
    echo Failed to build macOS arm64
    exit /b 1
)

set GOOS=linux
set GOARCH=amd64
echo Building Linux x64...
go build -o dist/sale-linux-amd64 ./cmd/sale
if errorlevel 1 (
    echo Failed to build Linux amd64
    exit /b 1
)

set GOOS=linux
set GOARCH=arm64
echo Building Linux arm64...
go build -o dist/sale-linux-arm64 ./cmd/sale
if errorlevel 1 (
    echo Failed to build Linux arm64
    exit /b 1
)

echo.
echo All builds complete! Binaries are in dist/
echo.
echo Files:
dir /b dist\
exit /b 0

:windows_fail
echo Failed to build Windows installer
del cmd\installer\sale.exe >nul 2>nul
del cmd\installer\say_less_logo.png >nul 2>nul
del cmd\installer\say_less_portrait_cover.png >nul 2>nul
del cmd\installer\GoogleSansCodeNerdFontPropo-Regular.ttf >nul 2>nul
del cmd\installer\GoogleSansCodeNerdFontPropo-SemiBold.ttf >nul 2>nul
exit /b 1
