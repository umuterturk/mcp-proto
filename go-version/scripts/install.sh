#!/bin/bash
# Installation script for MCP Proto Server

set -e

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}=== MCP Proto Server Installation ===${NC}\n"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64)
        ARCH="amd64"
        ;;
    aarch64|arm64)
        ARCH="arm64"
        ;;
    *)
        echo -e "${RED}Unsupported architecture: $ARCH${NC}"
        exit 1
        ;;
esac

BINARY_NAME="mcp-proto-server-${OS}-${ARCH}"
if [ "$OS" = "windows" ]; then
    BINARY_NAME="${BINARY_NAME}.exe"
fi

INSTALL_DIR="${HOME}/.local/bin"
BINARY_PATH="${INSTALL_DIR}/mcp-proto-server"

echo -e "${BLUE}Detected:${NC} $OS/$ARCH"
echo -e "${BLUE}Looking for:${NC} dist/$BINARY_NAME\n"

# Check if dist binary exists
if [ ! -f "dist/$BINARY_NAME" ]; then
    echo -e "${YELLOW}Binary not found in dist/. Building...${NC}\n"
    make build-all
fi

# Create install directory
mkdir -p "$INSTALL_DIR"

# Copy binary
echo -e "${BLUE}Installing to:${NC} $BINARY_PATH"
cp "dist/$BINARY_NAME" "$BINARY_PATH"
chmod +x "$BINARY_PATH"

# Verify installation
if [ -f "$BINARY_PATH" ]; then
    echo -e "${GREEN}✓ Installation successful!${NC}\n"
    
    # Check if in PATH
    if echo "$PATH" | grep -q "$INSTALL_DIR"; then
        echo -e "${GREEN}✓ $INSTALL_DIR is in your PATH${NC}"
    else
        echo -e "${YELLOW}⚠ $INSTALL_DIR is not in your PATH${NC}"
        echo -e "${YELLOW}  Add this to your ~/.bashrc or ~/.zshrc:${NC}"
        echo -e "    export PATH=\"\$HOME/.local/bin:\$PATH\"\n"
    fi
    
    # Show version
    echo -e "${BLUE}Installed version:${NC}"
    "$BINARY_PATH" -version
    echo ""
    
    # Show next steps
    echo -e "${BLUE}=== Next Steps ===${NC}\n"
    echo "1. Test the installation:"
    echo "   mcp-proto-server -version"
    echo ""
    echo "2. Run with your proto files:"
    echo "   mcp-proto-server -root /path/to/protos"
    echo ""
    echo "3. Configure in Cursor/Claude Desktop:"
    echo "   See examples in mcp_config_examples/"
    echo ""
else
    echo -e "${RED}✗ Installation failed${NC}"
    exit 1
fi

















