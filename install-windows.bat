@echo off
REM MCP Proto Server - Windows Installation Script

echo.
echo MCP Proto Server - Windows Installer
echo.

REM Check if the executable exists
if not exist "mcp-proto-server.exe" (
    echo Error: mcp-proto-server.exe not found in current directory
    echo Please run this script from the directory containing the executable
    pause
    exit /b 1
)

REM Set target directory
set "TARGET_DIR=%ProgramFiles%\mcp-proto"
set "TARGET_PATH=%TARGET_DIR%\mcp-proto-server.exe"

echo Installing to %TARGET_PATH%...
echo.
echo This requires administrator privileges.
echo Please approve the UAC prompt if it appears.
echo.

REM Create directory (will prompt for admin if needed)
if not exist "%TARGET_DIR%" (
    mkdir "%TARGET_DIR%" 2>nul
    if errorlevel 1 (
        echo Error: Failed to create %TARGET_DIR%
        echo Please run this script as Administrator
        pause
        exit /b 1
    )
)

REM Copy executable
copy /Y "mcp-proto-server.exe" "%TARGET_PATH%" >nul
if errorlevel 1 (
    echo Error: Failed to copy executable
    echo Please run this script as Administrator
    pause
    exit /b 1
)

echo.
echo Installation complete!
echo.
echo The executable has been installed to:
echo   %TARGET_PATH%
echo.
echo To use from command line, add to PATH:
echo   1. Open System Properties ^> Environment Variables
echo   2. Add "%TARGET_DIR%" to PATH
echo.
echo Next: Configure Cursor by adding to %%APPDATA%%\Cursor\mcp.json
echo.
pause

