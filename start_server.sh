#!/bin/bash

# MCP Proto Server - Startup Script

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "  MCP Proto Server - Starting"
echo "=========================================="
echo ""

# Check Python version
echo -n "Checking Python version... "
PYTHON_VERSION=$(python3 --version 2>&1 | awk '{print $2}')
REQUIRED_VERSION="3.11"

if python3 -c "import sys; exit(0 if sys.version_info >= (3, 11) else 1)"; then
    echo -e "${GREEN}✓${NC} Python $PYTHON_VERSION"
else
    echo -e "${RED}✗${NC} Python $PYTHON_VERSION (requires >= 3.11)"
    exit 1
fi

# Check if dependencies are installed
echo -n "Checking dependencies... "
if python3 -c "import mcp, rapidfuzz, protobuf" 2>/dev/null; then
    echo -e "${GREEN}✓${NC} All dependencies installed"
else
    echo -e "${YELLOW}!${NC} Installing dependencies..."
    pip install -q -r requirements.txt
    echo -e "${GREEN}✓${NC} Dependencies installed"
fi

# Determine proto root
if [ -n "$1" ]; then
    PROTO_ROOT="$1"
elif [ -n "$PROTO_ROOT" ]; then
    PROTO_ROOT="$PROTO_ROOT"
else
    PROTO_ROOT="examples/"
fi

# Check if directory exists
if [ ! -d "$PROTO_ROOT" ]; then
    echo -e "${RED}✗${NC} Directory not found: $PROTO_ROOT"
    echo ""
    echo "Usage:"
    echo "  ./start_server.sh [proto_root_dir]"
    echo ""
    echo "Example:"
    echo "  ./start_server.sh examples/"
    echo "  ./start_server.sh ~/my-project/protos"
    exit 1
fi

# Count proto files
PROTO_COUNT=$(find "$PROTO_ROOT" -name "*.proto" 2>/dev/null | wc -l | tr -d ' ')
echo -e "Proto root: ${GREEN}$PROTO_ROOT${NC}"
echo -e "Proto files found: ${GREEN}$PROTO_COUNT${NC}"
echo ""

if [ "$PROTO_COUNT" -eq 0 ]; then
    echo -e "${YELLOW}Warning:${NC} No .proto files found in $PROTO_ROOT"
    echo ""
fi

# Start server
echo "Starting MCP Proto Server..."
echo "Press Ctrl+C to stop"
echo "=========================================="
echo ""

exec python3 mcp_proto_server.py --root "$PROTO_ROOT"

