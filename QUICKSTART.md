# Quick Start Guide

Get up and running with MCP Proto Server in under 5 minutes.

## Prerequisites

- Python 3.11 or later
- pip package manager

## Setup Steps

### Step 1: Clone the Repository

```bash
git clone https://github.com/umuterturk/mcp-proto.git
cd mcp-proto
```

### Step 2: Install Dependencies

**⚠️ Required:** You must manually install Python dependencies.

```bash
pip install -r requirements.txt
```

This installs:
- `mcp` - Model Context Protocol SDK
- `protobuf` - Proto parsing support
- `rapidfuzz` - Fast fuzzy search
- `watchdog` - File watching (optional)

### Step 3: Configure Cursor

Add to your Cursor MCP settings file:

**macOS:** `~/Library/Application Support/Cursor/mcp.json`  
**Windows:** `%APPDATA%\Cursor\mcp.json`

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "python",
      "args": [
        "/absolute/path/to/mcp-proto/mcp_proto_server.py",
        "--root",
        "/absolute/path/to/your/proto/files"
      ]
    }
  }
}
```

**Important:**
- Use **absolute paths** (not relative)
- Replace `/absolute/path/to/mcp-proto/` with your clone location
- Replace `/absolute/path/to/your/proto/files` with your proto directory
- Or use the included `examples/` folder to test

**Example for macOS:**
```json
{
  "mcpServers": {
    "proto-server": {
      "command": "python",
      "args": [
        "/Users/yourname/mcp-proto/mcp_proto_server.py",
        "--root",
        "/Users/yourname/mcp-proto/examples"
      ]
    }
  }
}
```

### Step 4: Restart Cursor

Close and reopen Cursor. The server will automatically start!

**Note:** Cursor automatically starts/stops the server. You don't need to run it manually.

---

## Usage

Once configured, ask Cursor questions about your proto files:

- "What services are available?"
- "Show me the User message structure"
- "How do I authenticate?"

The AI will use three MCP tools to explore your protos:
- `search_proto` - Fuzzy search across all definitions
- `get_service_definition` - Get complete service with all RPCs
- `get_message_definition` - Get message with all fields

---

## Troubleshooting

**"Module not found" error:**
- Run: `pip install -r requirements.txt`
- Check Python version: `python --version` (need 3.11+)

**"No proto files found":**
- Verify the `--root` path in your config points to a directory with `.proto` files
- Test: `find /path/to/protos -name "*.proto"`

**Server not appearing in Cursor:**
- Ensure absolute paths are used in `mcp.json`
- Check Cursor's MCP logs for errors
- Restart Cursor completely

---

## Optional: Testing Before Configuration

Want to verify everything works before configuring Cursor?

### Step 3: Test the Installation (Optional)

Run the test suite:

```bash
python test_server.py
```

Expected output:
```
✓ Indexed 3 proto files
✓ Indexing: PASSED
✓ Search: PASSED
✓ Get Service: PASSED
✓ Get Message: PASSED
✓ Fuzzy Matching: PASSED
```

Or test the server manually:
```bash
# Test with included examples
python mcp_proto_server.py --root examples/

# Test with your own protos
python mcp_proto_server.py --root /path/to/your/protos
```

Press `Ctrl+C` to stop the server.

---

## What's Next?

- **USAGE.md** - Detailed examples and JSON responses
- **ARCHITECTURE.md** - How the system works
- **RECURSIVE_RESOLUTION.md** - Efficiency features

**Ready to explore your proto files with AI!**

