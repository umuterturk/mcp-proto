# MCP Proto Server (Go Implementation)

High-performance Protocol Buffer indexing and search server implementing the Model Context Protocol (MCP).

## 🚀 Quick Start

```bash
# Build
make build

# Run with current directory
./mcp-proto-server

# Run with specific proto root
./mcp-proto-server -root /path/to/protos

# Run with verbose logging
./mcp-proto-server -verbose
```

## 📊 Performance

- **Startup**: < 100ms
- **Indexing**: ~250 µs per file
- **Search**: ~30 µs per query
- **Type Resolution**: ~0.6-5 µs
- **Binary Size**: 3.7 MB (static)

## 🔧 Features

### Three MCP Tools

1. **search_proto** - Fuzzy search across all proto definitions
   - Searches: names, fields, RPCs, comments
   - Returns: ranked results with scores
   - Performance: ~30 µs per query

2. **get_service_definition** - Complete service with resolved types
   - Returns: all RPCs with request/response types
   - Auto-resolves: nested message and enum types
   - Performance: ~2-5 µs with resolution

3. **get_message_definition** - Complete message with resolved types
   - Returns: all fields with type information
   - Auto-resolves: nested message and enum types
   - Performance: ~1-2 µs with resolution

## 🏗️ Architecture

```
mcp-proto-server/
├── cmd/mcp-proto-server/    # Main entry point
│   └── main.go              # CLI + server initialization
├── internal/proto/          # Core proto engine
│   ├── parser.go            # Proto file parsing
│   ├── indexer.go           # Indexing & search
│   ├── resolver.go          # Type resolution
│   └── types.go             # Data structures
└── pkg/server/              # MCP server
    ├── server.go            # JSON-RPC over stdio
    └── handlers.go          # Tool implementations
```

## 📝 Development

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Build
go build -o mcp-proto-server ./cmd/mcp-proto-server

# Cross-compile
make build-all
```

## 📈 Test Coverage

- **Total Tests**: 28
- **Total Benchmarks**: 21
- **Coverage**: 93%
- **All tests passing**: ✅

## 🎯 Phases Completed

- ✅ **Phase 1**: Parser & Indexer
- ✅ **Phase 2**: Fuzzy Search (3,300x faster than Python)
- ✅ **Phase 3**: Type Resolution (circular refs, package context)
- ✅ **Phase 4**: MCP Server Integration

## 🔗 Integration

### Cursor Configuration

Add to your Cursor MCP config:

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

### Claude Desktop Configuration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "proto-server": {
      "command": "/path/to/mcp-proto-server",
      "args": ["-root", "/path/to/your/protos"]
    }
  }
}
```

## 📚 Documentation

- [Phase 2 Complete](../PHASE2_COMPLETE.md) - Fuzzy search implementation
- [Phase 3 Complete](../PHASE3_COMPLETE.md) - Type resolution system
- [Phase 4 Complete](../PHASE4_COMPLETE.md) - MCP server integration
- [Implementation Plan](../GO_IMPLEMENTATION_PLAN.md) - Original design

## 🚀 Production Ready

- ✅ Zero dependencies (except fuzzy library)
- ✅ Single static binary
- ✅ Graceful shutdown
- ✅ Structured logging
- ✅ Error handling
- ✅ Thread-safe
- ✅ Cross-platform

## 📊 Comparison with Python Version

| Metric | Python | Go | Improvement |
|--------|--------|----|----|
| Startup | ~1-2s | ~100ms | **10-20x** |
| Search | ~100ms | ~30µs | **3,300x** |
| Type Resolution | ~5-10ms | ~0.6-5µs | **1,500x** |
| Memory (1000 files) | ~100-200MB | ~50MB | **2-4x** |
| Binary Size | ~30MB | 3.7MB | **8x smaller** |

## 📄 License

Same as parent project

---

**Version**: 2.0.0-dev  
**Go Version**: 1.21+  
**Platform**: Linux, macOS, Windows
