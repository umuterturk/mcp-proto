# Quick Start Guide

Get up and running with MCP Proto Server in under 5 minutes.

## Two Installation Options

### Option A: Pre-built Binary (Recommended)

**No Python or dependencies required!**

1. Download the latest release for your platform from [GitHub Releases](https://github.com/umuterturk/mcp-proto/releases):
   - **Linux**: `mcp-proto-server-linux-amd64.tar.gz`
   - **macOS Intel**: `mcp-proto-server-macos-amd64.tar.gz`
   - **macOS Apple Silicon**: `mcp-proto-server-macos-arm64.tar.gz`
   - **Windows**: `mcp-proto-server-windows-amd64.zip`

2. Extract and run the installation script:
   
   **macOS:**
   ```bash
   tar -xzf mcp-proto-server-macos-*.tar.gz
   cd mcp-proto-server-macos-*
   ./install.sh
   ```
   
   **Linux:**
   ```bash
   tar -xzf mcp-proto-server-linux-*.tar.gz
   cd mcp-proto-server-linux-*
   ./install.sh
   ```
   
   **Windows:**
   ```cmd
   REM Extract the ZIP file, then run:
   install-windows.bat
   ```
   
   The scripts will automatically:
   - Remove macOS quarantine (macOS only)
   - Set proper permissions
   - Install to `/usr/local/bin` (Unix) or `C:\Program Files\mcp-proto` (Windows)
   - Request sudo/admin privileges only when needed

3. **Skip to Step 3** below to configure Cursor

### Option B: From Source

**Prerequisites:**
- Python 3.11 or later
- pip package manager

#### Step 1: Clone the Repository

```bash
git clone https://github.com/umuterturk/mcp-proto.git
cd mcp-proto
```

#### Step 2: Install Dependencies

```bash
pip install -r requirements.txt
```

This installs:
- `mcp` - Model Context Protocol SDK
- `protobuf` - Proto parsing support
- `rapidfuzz` - Fast fuzzy search
- `watchdog` - File watching (optional)
- `pyinstaller` - For building executables (optional)

### Step 3: Configure Cursor

Add to your Cursor MCP settings file:

**macOS:** `~/Library/Application Support/Cursor/mcp.json`  
**Windows:** `%APPDATA%\Cursor\mcp.json`

#### If using pre-built binary:

**macOS/Linux:**
```json
{
  "mcpServers": {
    "proto-server": {
      "command": "/usr/local/bin/mcp-proto-server",
      "args": [
        "--root",
        "/absolute/path/to/your/proto/files"
      ]
    }
  }
}
```

**Windows:**
```json
{
  "mcpServers": {
    "proto-server": {
      "command": "C:\\Program Files\\mcp-proto\\mcp-proto-server.exe",
      "args": [
        "--root",
        "C:\\absolute\\path\\to\\your\\proto\\files"
      ]
    }
  }
}
```

**Note:** After using the installation scripts, the binary is automatically placed in the correct location.

#### If using Python (from source):

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
- Replace paths with your actual locations
- Or use included example protos for testing

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

**macOS: "cannot be opened because it is from an unidentified developer":**
- This is expected for unsigned executables
- **Solution:** `xattr -cr mcp-proto-server` (removes quarantine)
- Or right-click → Open → Open
- See [MACOS_SECURITY.md](MACOS_SECURITY.md) for details

**"Module not found" error (Python installation):**
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

- **BUILD.md** - Building executables and releases
- **USAGE.md** - Detailed examples and JSON responses
- **ARCHITECTURE.md** - How the system works
- **RECURSIVE_RESOLUTION.md** - Efficiency features

## Building from Source

See **BUILD.md** for instructions on:
- Building your own executables with PyInstaller
- Creating releases with GitHub Actions
- Platform-specific build instructions

**Ready to explore your proto files with AI!**

