@echo off
REM Local build script for MCP Proto Server (Windows)

echo Building MCP Proto Server...
echo.

REM Check if Python is available
where python >nul 2>nul
if %errorlevel% neq 0 (
    echo Error: Python is not installed
    exit /b 1
)

echo Using Python: 
python --version

REM Install dependencies
echo.
echo Installing dependencies...
python -m pip install -q --upgrade pip
python -m pip install -q -r requirements.txt

REM Check if PyInstaller is installed
python -c "import PyInstaller" >nul 2>nul
if %errorlevel% neq 0 (
    echo Installing PyInstaller...
    python -m pip install -q pyinstaller
)

REM Clean previous builds
echo.
echo Cleaning previous builds...
if exist build rmdir /s /q build
if exist dist rmdir /s /q dist

REM Build with PyInstaller
echo.
echo Building executable...
python -m PyInstaller mcp_proto_server.spec

REM Test the build
echo.
echo Build complete!
echo.

if exist dist\mcp-proto-server.exe (
    echo Testing executable...
    echo.
    dist\mcp-proto-server.exe --help
    echo.
    echo Executable works!
    echo.
    
    REM Create distributable archive
    echo Creating distributable archive...
    cd dist
    copy ..\install-windows.bat install-windows.bat >nul
    powershell -Command "Compress-Archive -Path mcp-proto-server.exe,install-windows.bat -DestinationPath mcp-proto-server-windows-amd64.zip -Force"
    cd ..
    
    echo.
    echo Distributable archive created!
    echo.
    echo Binary location: %cd%\dist\mcp-proto-server.exe
    echo Archive location: %cd%\dist\mcp-proto-server-windows-amd64.zip
    echo.
    echo To install using the script:
    echo   cd dist ^&^& install-windows.bat
    echo.
    echo To test with examples:
    echo   dist\mcp-proto-server.exe --root examples\
) else (
    echo Build failed - executable not found
    exit /b 1
)

