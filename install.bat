@echo off
setlocal EnableExtensions
title Say Less Installer

echo ================================
echo  Say Less Windows Installer
echo ================================
echo.

set "SOURCE=%~dp0sale.exe"
set "INSTALL_DIR=%LOCALAPPDATA%\Programs\SayLess"
set "BIN_DIR=%INSTALL_DIR%\bin"
set "TARGET=%BIN_DIR%\sale.exe"

if not exist "%SOURCE%" (
    echo sale.exe was not found beside install.bat.
    echo Extract the complete Windows release zip before running this script.
    echo.
    pause
    exit /b 1
)

echo Installing for the current user to:
echo   %BIN_DIR%
echo.

if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"
if errorlevel 1 (
    echo Failed to create the installation folder.
    pause
    exit /b 1
)

copy /Y "%SOURCE%" "%TARGET%" >nul
if errorlevel 1 (
    echo Failed to copy sale.exe.
    pause
    exit /b 1
)

echo Adding Say Less to your user PATH...
powershell -NoProfile -ExecutionPolicy Bypass -Command "$bin = Join-Path $env:LOCALAPPDATA 'Programs\SayLess\bin'; $path = [Environment]::GetEnvironmentVariable('Path', 'User'); $entries = @($path -split ';' ^| Where-Object { $_ }); if ($entries -notcontains $bin) { [Environment]::SetEnvironmentVariable('Path', (($entries + $bin) -join ';'), 'User') }" >nul
if errorlevel 1 (
    echo.
    echo sale.exe was installed, but PATH could not be updated automatically.
    echo Add this folder to your user PATH manually:
    echo   %BIN_DIR%
    echo.
    pause
    exit /b 1
)

echo.
echo ================================
echo  Installation Complete!
echo ================================
echo.
echo Open a new terminal, then run:
echo   sale --version
echo.
pause
