#!/bin/bash
# Local build script for MCP Proto Server

set -e

echo "🔨 Building MCP Proto Server..."
echo ""

# Check if Python is available
if ! command -v python &> /dev/null && ! command -v python3 &> /dev/null; then
    echo "❌ Error: Python is not installed"
    exit 1
fi

PYTHON_CMD=$(command -v python3 2>/dev/null || command -v python)
echo "✓ Using Python: $PYTHON_CMD"

# Check Python version
PYTHON_VERSION=$($PYTHON_CMD --version 2>&1 | awk '{print $2}')
echo "✓ Python version: $PYTHON_VERSION"

# Install dependencies
echo ""
echo "📦 Installing dependencies..."
$PYTHON_CMD -m pip install -q --upgrade pip
$PYTHON_CMD -m pip install -q -r requirements.txt

# Check if PyInstaller is installed
if ! $PYTHON_CMD -c "import PyInstaller" &> /dev/null; then
    echo "📦 Installing PyInstaller..."
    $PYTHON_CMD -m pip install -q pyinstaller
fi

# Clean previous builds
echo ""
echo "🧹 Cleaning previous builds..."
rm -rf build/ dist/

# Build with PyInstaller
echo ""
echo "🔨 Building executable..."
$PYTHON_CMD -m PyInstaller mcp_proto_server.spec

# Test the build
echo ""
echo "✅ Build complete!"
echo ""
echo "Testing executable..."

if [ -f "dist/mcp-proto-server" ]; then
    chmod +x dist/mcp-proto-server
    echo ""
    ./dist/mcp-proto-server --help
    echo ""
    echo "✓ Executable works!"
    echo ""
    
    # Create distributable archive
    echo "📦 Creating distributable archive..."
    cd dist
    
    # Detect platform
    if [[ "$OSTYPE" == "darwin"* ]]; then
        ARCH=$(uname -m)
        if [[ "$ARCH" == "x86_64" ]]; then
            PLATFORM="macos-amd64"
        else
            PLATFORM="macos-arm64"
        fi
        cp ../install-macos.sh install.sh
    else
        PLATFORM="linux-amd64"
        cp ../install-linux.sh install.sh
    fi
    
    chmod +x install.sh
    tar -czf "mcp-proto-server-${PLATFORM}.tar.gz" mcp-proto-server install.sh
    cd ..
    
    echo ""
    echo "✓ Distributable archive created!"
    echo ""
    echo "📍 Binary location: $(pwd)/dist/mcp-proto-server"
    echo "📦 Archive location: $(pwd)/dist/mcp-proto-server-${PLATFORM}.tar.gz"
    echo ""
    echo "To install using the script:"
    echo "  cd dist && ./install.sh"
    echo ""
    echo "To test with examples:"
    echo "  ./dist/mcp-proto-server --root examples/"
else
    echo "❌ Build failed - executable not found"
    exit 1
fi

