#!/bin/bash
# MCP Proto Server - macOS Installation Script

set -e

echo "🚀 MCP Proto Server - macOS Installer"
echo ""

# Find the binary in the current directory
BINARY="mcp-proto-server"

if [ ! -f "$BINARY" ]; then
    echo "❌ Error: $BINARY not found in current directory"
    echo "Please run this script from the extracted archive directory"
    exit 1
fi

# Remove macOS quarantine attribute
echo "🔓 Removing macOS quarantine attribute..."
xattr -cr "$BINARY"

# Make executable
chmod +x "$BINARY"

# Check if we can write to /usr/local/bin
TARGET_DIR="/usr/local/bin"
TARGET_PATH="$TARGET_DIR/$BINARY"

if [ -w "$TARGET_DIR" ]; then
    # Can write without sudo
    echo "📦 Installing to $TARGET_PATH..."
    cp "$BINARY" "$TARGET_PATH"
else
    # Need sudo
    echo "📦 Installing to $TARGET_PATH (requires sudo)..."
    sudo cp "$BINARY" "$TARGET_PATH"
fi

echo ""
echo "✅ Installation complete!"
echo ""
echo "Verify installation:"
echo "  mcp-proto-server --help"
echo ""
echo "Next: Configure Cursor by adding to ~/Library/Application Support/Cursor/mcp.json"
echo ""

