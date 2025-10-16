#!/bin/bash
# Example script to run MCP Proto Server with your actual proto files
# Edit this script to point to your proto directory

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROTO_DIR="${PROTO_DIR:-/path/to/your/proto/files}"

echo "Starting MCP Proto Server..."
echo "Proto directory: $PROTO_DIR"
echo ""

cd "$SCRIPT_DIR"
exec python mcp_proto_server.py --root "$PROTO_DIR" "$@"

