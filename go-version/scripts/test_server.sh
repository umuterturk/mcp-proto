#!/bin/bash
# Test script for MCP Proto Server
# This script sends JSON-RPC requests to test the server

set -e

BINARY="../mcp-proto-server"
PROTO_ROOT="../examples"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== MCP Proto Server Test ===${NC}\n"

# Check if binary exists
if [ ! -f "$BINARY" ]; then
    echo -e "${RED}Error: Binary not found at $BINARY${NC}"
    echo "Run 'make build' first"
    exit 1
fi

# Start server in background
echo -e "${BLUE}Starting server...${NC}"
$BINARY -root $PROTO_ROOT 2>/dev/null &
SERVER_PID=$!

# Give server time to start
sleep 1

# Function to send JSON-RPC request
send_request() {
    local request="$1"
    local description="$2"
    
    echo -e "\n${BLUE}Test: $description${NC}"
    echo "$request" | timeout 2s $BINARY -root $PROTO_ROOT 2>/dev/null || true
}

# Function to cleanup
cleanup() {
    if [ ! -z "$SERVER_PID" ]; then
        kill $SERVER_PID 2>/dev/null || true
    fi
}

trap cleanup EXIT

echo -e "\n${BLUE}Running tests...${NC}\n"

# Test 1: Initialize
echo -e "${BLUE}1. Testing initialize...${NC}"
REQUEST='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test-client","version":"1.0"}}}'
echo "$REQUEST" | $BINARY -root $PROTO_ROOT 2>/dev/null | head -1 | jq '.' 2>/dev/null && echo -e "${GREEN}✓ Initialize OK${NC}" || echo -e "${RED}✗ Initialize failed${NC}"

# Test 2: List tools
echo -e "\n${BLUE}2. Testing tools/list...${NC}"
REQUEST='{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}'
echo "$REQUEST" | $BINARY -root $PROTO_ROOT 2>/dev/null | head -1 | jq '.result.tools | length' 2>/dev/null && echo -e "${GREEN}✓ Tools list OK (should show 3 tools)${NC}" || echo -e "${RED}✗ Tools list failed${NC}"

# Test 3: Search proto
echo -e "\n${BLUE}3. Testing search_proto...${NC}"
REQUEST='{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"search_proto","arguments":{"query":"User","limit":5,"min_score":60}}}'
echo "$REQUEST" | $BINARY -root $PROTO_ROOT 2>/dev/null | head -1 | jq -r '.result.content[0].text' 2>/dev/null | head -5 && echo -e "${GREEN}✓ Search OK${NC}" || echo -e "${RED}✗ Search failed${NC}"

# Test 4: Get service
echo -e "\n${BLUE}4. Testing get_service_definition...${NC}"
REQUEST='{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_service_definition","arguments":{"name":"UserService","resolve_types":true,"max_depth":5}}}'
echo "$REQUEST" | $BINARY -root $PROTO_ROOT 2>/dev/null | head -1 | jq -r '.result.content[0].text' 2>/dev/null | head -10 && echo -e "${GREEN}✓ Get service OK${NC}" || echo -e "${RED}✗ Get service failed${NC}"

echo -e "\n${GREEN}=== Tests Complete ===${NC}\n"

# Show binary info
echo -e "${BLUE}Binary info:${NC}"
$BINARY -version
echo ""
ls -lh $BINARY

cleanup

















