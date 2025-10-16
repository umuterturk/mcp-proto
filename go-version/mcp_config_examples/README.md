# MCP Configuration Examples

## Cursor Configuration

Location: Settings → Features → MCP → Edit Config

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "/path/to/mcp-proto-server",
      "args": ["-root", "/path/to/your/protos"],
      "env": {}
    }
  }
}
```

## Claude Desktop Configuration

### macOS

Location: `~/Library/Application Support/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "/Users/username/.local/bin/mcp-proto-server",
      "args": ["-root", "/path/to/your/protos"]
    }
  }
}
```

### Linux

Location: `~/.config/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "/home/username/.local/bin/mcp-proto-server",
      "args": ["-root", "/path/to/your/protos"]
    }
  }
}
```

### Windows

Location: `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "C:\\Users\\username\\.local\\bin\\mcp-proto-server.exe",
      "args": ["-root", "C:\\path\\to\\your\\protos"]
    }
  }
}
```

## Environment Variables

You can use the `PROTO_ROOT` environment variable instead of the `-root` flag:

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "/path/to/mcp-proto-server",
      "args": [],
      "env": {
        "PROTO_ROOT": "/path/to/your/protos"
      }
    }
  }
}
```

## Multiple Proto Roots

To work with multiple proto directories, create multiple server entries:

```json
{
  "mcpServers": {
    "proto-server-api": {
      "command": "/path/to/mcp-proto-server",
      "args": ["-root", "/path/to/api/protos"]
    },
    "proto-server-internal": {
      "command": "/path/to/mcp-proto-server",
      "args": ["-root", "/path/to/internal/protos"]
    }
  }
}
```

## Verbose Logging

Enable verbose logging for debugging:

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "/path/to/mcp-proto-server",
      "args": ["-root", "/path/to/protos", "-verbose"]
    }
  }
}
```

## Verify Configuration

After setting up, restart Cursor/Claude Desktop and check:

1. The server appears in the available tools
2. You can search for proto definitions
3. You can retrieve service/message definitions

## Troubleshooting

### Server Not Starting

- Check the binary path is correct
- Ensure binary has execute permissions: `chmod +x /path/to/mcp-proto-server`
- Verify proto root directory exists
- Check logs (stderr is captured by the MCP client)

### No Results

- Ensure proto files exist in the specified directory
- Check proto files have valid syntax
- Try with `-verbose` flag to see indexing output

### Permission Errors

```bash
chmod +x /path/to/mcp-proto-server
```

## Testing Configuration

Test the server manually before configuring:

```bash
# Start server
mcp-proto-server -root /path/to/protos -verbose

# Send test request (in another terminal)
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}' | mcp-proto-server -root /path/to/protos
```

















