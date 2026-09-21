@echo off
echo ================================
echo  Say Less Language Installer
echo ================================
echo.

REM Check if running as administrator
net session >nul 2>&1
if %errorlevel% neq 0 (
    echo This installer requires administrator privileges.
    echo Please right-click and select "Run as administrator".
    echo.
    pause
    exit /b 1
)

set INSTALL_DIR=%ProgramFiles%\SayLess
set BIN_DIR=%INSTALL_DIR%\bin

echo Installing Say Less to %INSTALL_DIR%...
echo.

REM Create installation directory
if not exist "%INSTALL_DIR%" mkdir "%INSTALL_DIR%"
if not exist "%BIN_DIR%" mkdir "%BIN_DIR%"

REM Copy the binary
echo Copying sale.exe...
copy /Y "%~dp0sale.exe" "%BIN_DIR%\sale.exe" >nul
if %errorlevel% neq 0 (
    echo Failed to copy sale.exe
    pause
    exit /b 1
)

REM Add to PATH
echo Adding to PATH...
setx PATH "%PATH%;%BIN_DIR%" /M >nul 2>&1
if %errorlevel% neq 0 (
    echo Warning: Could not add to system PATH automatically.
    echo Please add %BIN_DIR% to your PATH manually.
)

echo.
echo ================================
echo  Installation Complete!
echo ================================
echo.
echo You can now use 'sale' from any terminal.
echo.
echo Quick start:
echo   sale new my-project
echo   cd my-project
echo   sale run src/main.sl
echo.
echo For web projects:
echo   sale create --web my-web-app
echo   cd my-web-app
echo   sale dev
echo.
pause
