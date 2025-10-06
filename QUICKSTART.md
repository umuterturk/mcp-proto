# Quick Start Guide

Get up and running with MCP Proto Server in 5 minutes.

## Prerequisites

- Python 3.11 or later
- pip package manager
- A directory with .proto files (or use our examples)

## Installation

```bash
# Clone or download the repository
cd /path/to/mcp-proto

# Install dependencies
pip install -r requirements.txt
```

That's it! No protoc compiler, no additional tools needed.

## Basic Usage

### 1. Test the Server

Run the test suite to verify everything works:

```bash
python test_server.py
```

You should see output like:
```
✓ Indexed 3 proto files
✓ Indexing: PASSED
✓ Search: PASSED
✓ Get Service: PASSED
✓ Get Message: PASSED
✓ Fuzzy Matching: PASSED
```

### 2. Run the MCP Server

Start the server pointing to your proto files:

```bash
# Using the included examples
python mcp_proto_server.py --root examples/

# Or point to your own proto files
python mcp_proto_server.py --root /path/to/your/protos
```

The server will:
1. Scan the directory for all .proto files
2. Parse and index them
3. Start listening for MCP requests via stdio

### 3. Connect an AI Agent

#### Option A: Cursor / Claude Desktop

Add to your MCP settings file:

**MacOS:** `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows:** `%APPDATA%\Claude\claude_desktop_config.json`

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "python",
      "args": [
        "/absolute/path/to/mcp_proto_server.py",
        "--root",
        "/absolute/path/to/your/protos"
      ]
    }
  }
}
```

Restart Claude Desktop, and you'll see the proto-server available.

#### Option B: Custom MCP Client

```python
from mcp import ClientSession, StdioServerParameters
from mcp.client.stdio import stdio_client

server_params = StdioServerParameters(
    command="python",
    args=["mcp_proto_server.py", "--root", "examples/"]
)

async with stdio_client(server_params) as (read, write):
    async with ClientSession(read, write) as session:
        await session.initialize()
        
        # Use tools
        result = await session.call_tool(
            "search_proto",
            {"query": "User"}
        )
        print(result)
```

## Example Interactions

### Search for Services

**Ask the AI:**
> "What authentication services are available?"

**AI will:**
1. Call `search_proto("authentication")`
2. Find AuthService
3. Call `get_service_definition("AuthService")`
4. Show you: Login, Logout, RefreshToken, VerifyToken methods

### Explore Message Structure

**Ask the AI:**
> "What fields does the User message have?"

**AI will:**
1. Call `search_proto("User")` or directly `get_message_definition("User")`
2. Show you all fields: id, email, name, role, created_at, updated_at, is_active

### Find RPC Methods

**Ask the AI:**
> "How do I create a new user?"

**AI will:**
1. Call `search_proto("create user")`
2. Find CreateUser RPC in UserService
3. Show request type: CreateUserRequest
4. Call `get_message_definition("CreateUserRequest")`
5. Show required fields: email, name, password, role

## Common Commands

```bash
# Basic usage
python mcp_proto_server.py --root examples/

# With environment variable
export PROTO_ROOT=/path/to/protos
python mcp_proto_server.py

# Verbose logging
python mcp_proto_server.py --root examples/ --verbose

# Run tests
python test_server.py

# Test on your own protos
python mcp_proto_server.py --root ~/my-project/proto
```

## Troubleshooting

### "No proto files found"
- Check that your path contains .proto files
- Try: `find /path/to/protos -name "*.proto"` to verify

### "Failed to parse file"
- Check proto file syntax
- The parser supports standard proto2/proto3
- Run with `--verbose` to see detailed errors

### "Module not found"
- Ensure dependencies are installed: `pip install -r requirements.txt`
- Check Python version: `python --version` (need 3.11+)

### "Service not found"
- First use `search_proto` to find the correct name
- Try both simple name ("UserService") and qualified ("api.v1.UserService")

## What's Next?

1. **Read USAGE.md** for detailed examples and JSON responses
2. **Read ARCHITECTURE.md** to understand how it works
3. **Try with your own protos** by changing the --root path
4. **Integrate with your AI workflow** using the MCP tools

## Example Session

```bash
$ python test_server.py

# Output:
✓ Indexed 3 proto files

Statistics:
  - Total services: 3
  - Total messages: 28
  - Total enums: 3

Search for 'auth':
  1. api.v1.AuthService (score: 67.5)
     Comment: Authentication service for user login...

Get UserService definition:
  ✓ Service: api.v1.UserService
  RPCs (5):
    - CreateUser: CreateUserRequest → CreateUserResponse
    - GetUser: GetUserRequest → GetUserResponse
    - UpdateUser: UpdateUserRequest → UpdateUserResponse
    - DeleteUser: DeleteUserRequest → DeleteUserResponse
    - ListUsers: ListUsersRequest → ListUsersResponse (streaming)
```

## Performance Tips

- **First run**: Indexing is fast (~1000 files/sec)
- **Searches**: Sub-millisecond for most queries
- **Large repos**: Works great with 1000+ proto files
- **Memory**: ~1-2 KB per proto definition

## Support

For issues or questions:
1. Check USAGE.md for examples
2. Run with --verbose flag for debugging
3. Verify proto files are valid proto2/proto3 syntax

---

**Ready to explore your proto files with AI? Start the server and ask away!**

