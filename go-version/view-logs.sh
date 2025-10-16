#!/bin/bash

# Find the most recent log file
LATEST_LOG=$(ls -t /tmp/mcp-proto-server-*.log 2>/dev/null | head -1)

if [ -z "$LATEST_LOG" ]; then
    echo "No MCP proto server log files found in /tmp/"
    echo "The server may not have been started yet."
    exit 1
fi

echo "=== MCP Proto Server Logs ==="
echo "Log file: $LATEST_LOG"
echo "==================================="
echo ""

# Check if we should tail (follow) or just show the file
if [ "$1" == "-f" ] || [ "$1" == "--follow" ]; then
    echo "Following log file (Ctrl+C to stop)..."
    echo ""
    tail -f "$LATEST_LOG"
else
    # Show last 50 lines by default
    LINES=${1:-50}
    echo "Showing last $LINES lines (use -f to follow, or specify number of lines):"
    echo ""
    tail -n "$LINES" "$LATEST_LOG"
    echo ""
    echo "---"
    echo "To follow logs in real-time: $0 -f"
    echo "To see more lines: $0 100"
fi

