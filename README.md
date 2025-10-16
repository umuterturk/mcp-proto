# MCP Proto Server

A high-performance Protocol Buffer indexing and search server that integrates with AI coding assistants through the Model Context Protocol (MCP).

## 🎯 What It Does

MCP Proto Server makes your Protocol Buffer definitions instantly searchable and accessible to AI assistants like Claude and Cursor. Instead of manually looking through `.proto` files, you can ask questions like:

- "Show me all authentication-related services"
- "What fields does the User message have?"
- "Find RPCs that handle payment processing"

## 🚀 Key Features

- **Instant Search**: Find services, messages, and fields across all your proto files
- **Smart Type Resolution**: Automatically resolves nested types and dependencies
- **AI Integration**: Works seamlessly with Claude Desktop and Cursor
- **Lightning Fast**: Sub-millisecond search performance
- **Zero Dependencies**: Single executable, no runtime requirements

## 📦 Installation

### Download Latest Release

1. Go to the [Releases page](https://github.com/your-username/mcp-proto/releases)
2. Download the appropriate package for your system:
   - **Linux**: `mcp-proto-server-linux-amd64.tar.gz`
   - **macOS Intel**: `mcp-proto-server-macos-amd64.tar.gz`
   - **macOS Apple Silicon**: `mcp-proto-server-macos-arm64.tar.gz`
   - **Windows**: `mcp-proto-server-windows-amd64.zip`

### Install

**Linux/macOS:**
```bash
tar -xzf mcp-proto-server-*.tar.gz
./install.sh
```

**Windows:**
```cmd
# Extract the zip file, then run:
install-windows.bat
```

### Verify Installation
```bash
mcp-proto-server --help
```

## 🔧 Setup with AI Assistants

### Cursor Configuration

Add to your Cursor MCP config (`~/.cursor/mcp.json`):

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "mcp-proto-server",
      "args": ["-root", "/path/to/your/proto/files"],
      "capabilities": [
        "search_proto: Fuzzy search across all proto definitions",
        "get_service_definition: Get service with ALL nested types in one call",
        "get_message_definition: Get message with ALL nested types in one call"
      ]
    }
  }
}
```

### Claude Desktop Configuration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "mcp-proto-server",
      "args": ["-root", "/path/to/your/proto/files"]
    }
  }
}
```

## 🎯 How to Use

Once configured, simply ask your AI assistant questions about your Protocol Buffer definitions:

- "What services are available in the auth package?"
- "Show me the complete structure of the Product message"
- "Find all RPCs that return user data"

The server automatically indexes your proto files and provides instant, intelligent search results.

## 📁 Project Structure

- `go-version/` - High-performance Go implementation (recommended)
- `python-version/` - Python implementation (legacy)

## 📄 License

Same as parent project
