#!/usr/bin/env python3
"""
MCP Proto Server - Model Context Protocol server for Protocol Buffer files.

Exposes proto file structure to AI agents through MCP tools.
"""

import argparse
import asyncio
import logging
import os
import sys
from pathlib import Path
from typing import Optional

from mcp.server import Server
from mcp.server.stdio import stdio_server
from mcp.types import Tool, TextContent

from proto_indexer import ProtoIndex

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s',
    handlers=[logging.StreamHandler(sys.stderr)]
)
logger = logging.getLogger(__name__)


class MCPProtoServer:
    """MCP server for proto file indexing and search."""
    
    def __init__(self, proto_root: str, watch: bool = False):
        self.proto_root = proto_root
        self.watch = watch
        self.index = ProtoIndex()
        self.server = Server("proto-server")
        
        # Register handlers
        self._register_handlers()
        
    def _register_handlers(self):
        """Register MCP tool handlers."""
        
        @self.server.list_tools()
        async def list_tools() -> list[Tool]:
            """List available tools."""
            return [
                Tool(
                    name="search_proto",
                    description=(
                        "Fuzzy search across all proto definitions (services, messages, enums). "
                        "Searches in names, fields, RPC methods, and comments. "
                        "Returns structured results with match scores."
                    ),
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "query": {
                                "type": "string",
                                "description": "Search query (name, field, or keyword)"
                            },
                            "limit": {
                                "type": "number",
                                "description": "Maximum number of results (default: 20)",
                                "default": 20
                            },
                            "min_score": {
                                "type": "number",
                                "description": "Minimum match score 0-100 (default: 60)",
                                "default": 60
                            }
                        },
                        "required": ["query"]
                    }
                ),
                Tool(
                    name="get_service_definition",
                    description=(
                        "Get complete service definition including all RPC methods with "
                        "their request/response types and comments. "
                        "AUTOMATICALLY resolves all nested types recursively, providing "
                        "the complete structure in a single response. "
                        "Accepts both simple name (e.g., 'UserService') or "
                        "fully qualified name (e.g., 'api.v1.UserService')."
                    ),
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "name": {
                                "type": "string",
                                "description": "Service name (simple or fully qualified)"
                            },
                            "resolve_types": {
                                "type": "boolean",
                                "description": "Recursively resolve all request/response types (default: true)",
                                "default": True
                            },
                            "max_depth": {
                                "type": "number",
                                "description": "Maximum recursion depth (default: 10)",
                                "default": 10
                            }
                        },
                        "required": ["name"]
                    }
                ),
                Tool(
                    name="get_message_definition",
                    description=(
                        "Get complete message definition with all fields, types, and comments. "
                        "AUTOMATICALLY resolves all nested types recursively, providing "
                        "the complete structure in a single response. "
                        "Accepts both simple name (e.g., 'User') or "
                        "fully qualified name (e.g., 'api.v1.User')."
                    ),
                    inputSchema={
                        "type": "object",
                        "properties": {
                            "name": {
                                "type": "string",
                                "description": "Message name (simple or fully qualified)"
                            },
                            "resolve_types": {
                                "type": "boolean",
                                "description": "Recursively resolve all field types (default: true)",
                                "default": True
                            },
                            "max_depth": {
                                "type": "number",
                                "description": "Maximum recursion depth (default: 10)",
                                "default": 10
                            }
                        },
                        "required": ["name"]
                    }
                )
            ]
        
        @self.server.call_tool()
        async def call_tool(name: str, arguments: dict) -> list[TextContent]:
            """Handle tool calls."""
            try:
                if name == "search_proto":
                    return await self._handle_search(arguments)
                elif name == "get_service_definition":
                    return await self._handle_get_service(arguments)
                elif name == "get_message_definition":
                    return await self._handle_get_message(arguments)
                else:
                    raise ValueError(f"Unknown tool: {name}")
            except Exception as e:
                logger.error(f"Error handling tool {name}: {e}", exc_info=True)
                return [TextContent(
                    type="text",
                    text=f"Error: {str(e)}"
                )]
    
    async def _handle_search(self, arguments: dict) -> list[TextContent]:
        """Handle search_proto tool call."""
        query = arguments.get("query", "")
        limit = arguments.get("limit", 20)
        min_score = arguments.get("min_score", 60)
        
        if not query:
            return [TextContent(
                type="text",
                text="Error: query parameter is required"
            )]
        
        results = self.index.search(query, limit=limit, min_score=min_score)
        
        import json
        return [TextContent(
            type="text",
            text=json.dumps({
                "query": query,
                "result_count": len(results),
                "results": results
            }, indent=2)
        )]
    
    async def _handle_get_service(self, arguments: dict) -> list[TextContent]:
        """Handle get_service_definition tool call."""
        name = arguments.get("name", "")
        resolve_types = arguments.get("resolve_types", True)
        max_depth = arguments.get("max_depth", 10)
        
        if not name:
            return [TextContent(
                type="text",
                text="Error: name parameter is required"
            )]
        
        service = self.index.get_service(name, resolve_types=resolve_types, max_depth=max_depth)
        
        if not service:
            return [TextContent(
                type="text",
                text=f"Error: Service '{name}' not found. Try using search_proto to find the correct name."
            )]
        
        import json
        return [TextContent(
            type="text",
            text=json.dumps(service, indent=2)
        )]
    
    async def _handle_get_message(self, arguments: dict) -> list[TextContent]:
        """Handle get_message_definition tool call."""
        name = arguments.get("name", "")
        resolve_types = arguments.get("resolve_types", True)
        max_depth = arguments.get("max_depth", 10)
        
        if not name:
            return [TextContent(
                type="text",
                text="Error: name parameter is required"
            )]
        
        message = self.index.get_message(name, resolve_types=resolve_types, max_depth=max_depth)
        
        if not message:
            # Also try enums
            enum = self.index.get_enum(name)
            if enum:
                import json
                return [TextContent(
                    type="text",
                    text=json.dumps(enum, indent=2)
                )]
            
            return [TextContent(
                type="text",
                text=f"Error: Message or Enum '{name}' not found. Try using search_proto to find the correct name."
            )]
        
        import json
        return [TextContent(
            type="text",
            text=json.dumps(message, indent=2)
        )]
    
    async def initialize(self):
        """Initialize the server by indexing proto files."""
        logger.info(f"Indexing proto files from: {self.proto_root}")
        
        try:
            count = self.index.index_directory(self.proto_root)
            stats = self.index.get_stats()
            
            logger.info(f"Indexing complete: {count} files")
            logger.info(f"Statistics: {stats}")
            
            if self.watch:
                logger.info("File watching enabled")
                # TODO: Implement file watching with watchdog
                
        except Exception as e:
            logger.error(f"Failed to index directory: {e}")
            raise
    
    async def run(self):
        """Run the MCP server."""
        await self.initialize()
        
        # Run stdio server
        async with stdio_server() as (read_stream, write_stream):
            await self.server.run(
                read_stream,
                write_stream,
                self.server.create_initialization_options()
            )


async def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="MCP Proto Server - Index and search Protocol Buffer files"
    )
    parser.add_argument(
        "--root",
        type=str,
        default=os.environ.get("PROTO_ROOT", "."),
        help="Root directory containing .proto files (default: PROTO_ROOT env or current directory)"
    )
    parser.add_argument(
        "--watch",
        action="store_true",
        help="Watch for file changes and re-index automatically"
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Enable verbose logging"
    )
    
    args = parser.parse_args()
    
    if args.verbose:
        logging.getLogger().setLevel(logging.DEBUG)
    
    # Validate root directory (expand env vars and ~ first)
    expanded_root = os.path.expandvars(os.path.expanduser(args.root))
    proto_root = Path(expanded_root).resolve()
    if not proto_root.exists():
        logger.error(f"Root directory does not exist: {proto_root}")
        sys.exit(1)
    
    logger.info(f"Starting MCP Proto Server")
    logger.info(f"Proto root: {proto_root}")
    
    # Create and run server
    server = MCPProtoServer(str(proto_root), watch=args.watch)
    
    try:
        await server.run()
    except KeyboardInterrupt:
        logger.info("Server stopped by user")
    except Exception as e:
        logger.error(f"Server error: {e}", exc_info=True)
        sys.exit(1)


if __name__ == "__main__":
    asyncio.run(main())

