@echo off
echo Creating release packages...
echo.

REM Build all platforms first
call build.bat
if %errorlevel% neq 0 (
    echo Build failed!
    exit /b 1
)

echo.
echo Creating release packages...
echo.

REM Windows release
echo Packaging Windows release...
mkdir release\sale-windows-amd64 2>nul
copy dist\sale-windows-amd64.exe release\sale-windows-amd64\sale.exe >nul
copy dist\say_less.exe release\sale-windows-amd64\say_less.exe >nul
copy install.bat release\sale-windows-amd64\install.bat >nul
copy README.md release\sale-windows-amd64\README.md >nul
cd release
tar -a -cf sale-windows-amd64.zip sale-windows-amd64
cd ..
rmdir /s /q release\sale-windows-amd64

REM macOS release (Intel)
echo Packaging macOS Intel release...
mkdir release\sale-macos-amd64 2>nul
copy dist\sale-macos-amd64 release\sale-macos-amd64\sale >nul
copy install.sh release\sale-macos-amd64\install.sh >nul
copy README.md release\sale-macos-amd64\README.md >nul
cd release
tar -cf sale-macos-amd64.tar sale-macos-amd64
gzip sale-macos-amd64.tar
cd ..
rmdir /s /q release\sale-macos-amd64

REM macOS release (Apple Silicon)
echo Packaging macOS Apple Silicon release...
mkdir release\sale-macos-arm64 2>nul
copy dist\sale-macos-arm64 release\sale-macos-arm64\sale >nul
copy install.sh release\sale-macos-arm64\install.sh >nul
copy README.md release\sale-macos-arm64\README.md >nul
cd release
tar -cf sale-macos-arm64.tar sale-macos-arm64
gzip sale-macos-arm64.tar
cd ..
rmdir /s /q release\sale-macos-arm64

REM Linux release
echo Packaging Linux release...
mkdir release\sale-linux-amd64 2>nul
copy dist\sale-linux-amd64 release\sale-linux-amd64\sale >nul
copy install.sh release\sale-linux-amd64\install.sh >nul
copy README.md release\sale-linux-amd64\README.md >nul
cd release
tar -cf sale-linux-amd64.tar sale-linux-amd64
gzip sale-linux-amd64.tar
cd ..
rmdir /s /q release\sale-linux-amd64

echo.
echo ================================
echo  Release packages created!
echo ================================
echo.
echo Files in release/:
dir /b release\
echo.
