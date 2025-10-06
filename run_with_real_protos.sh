#!/bin/bash
# Quick script to run MCP Proto Server with your actual proto files

cd /Users/umut.erturk/mycode/mcp-proto
exec python mcp_proto_server.py --root /Users/umut.erturk/Code/services-protobuf-resources "$@"

