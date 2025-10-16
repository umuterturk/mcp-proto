package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/uerturk/mcp-proto-server/internal/proto"
)

const (
	mcpVersion        = "2024-11-05"
	serverName        = "mcp-proto-server"
	serverVersion     = "2.0.0"
	protocolVersion   = "2024-11-05"
	capabilitiesTools = true
)

// MCPServer is the main MCP server implementation
type MCPServer struct {
	index  *proto.ProtoIndex
	logger *slog.Logger
	reader *bufio.Reader
	writer *bufio.Writer
}

// NewMCPServer creates a new MCP server instance
func NewMCPServer(index *proto.ProtoIndex, logger *slog.Logger) *MCPServer {
	return &MCPServer{
		index:  index,
		logger: logger,
		reader: bufio.NewReader(os.Stdin),
		writer: bufio.NewWriter(os.Stdout),
	}
}

// Run starts the MCP server and processes requests via stdio
func (s *MCPServer) Run(ctx context.Context) error {
	s.logger.Info("MCP server starting", "protocol", protocolVersion)

	requestCount := 0

	defer func() {
		if r := recover(); r != nil {
			s.logger.Error("PANIC recovered in server", "panic", r, "requests_processed", requestCount)
		}
		s.logger.Info("server exiting", "requests_processed", requestCount)
	}()

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("context cancelled, shutting down gracefully", "reason", ctx.Err(), "requests_processed", requestCount)
			return ctx.Err()
		default:
			// Read a line from stdin
			s.logger.Debug("waiting for input on stdin...")
			line, err := s.reader.ReadBytes('\n')
			if err != nil {
				if err == io.EOF {
					s.logger.Warn("EOF received on stdin - client disconnected", "requests_processed", requestCount)
					return nil
				}
				s.logger.Error("failed to read from stdin", "error", err, "error_type", fmt.Sprintf("%T", err), "requests_processed", requestCount)
				return err
			}

			requestCount++
			s.logger.Debug("received data from stdin", "length", len(line), "request_number", requestCount)

			// Parse and handle the request
			if err := s.handleRequest(line); err != nil {
				s.logger.Error("failed to handle request",
					"error", err,
					"request_number", requestCount,
					"request_data", string(line))
				// Don't return error, continue processing
			}
		}
	}
}

// handleRequest processes a single JSON-RPC request
func (s *MCPServer) handleRequest(data []byte) error {
	var req JSONRPCRequest
	if err := json.Unmarshal(data, &req); err != nil {
		s.logger.Error("JSON parse error", "error", err, "data", string(data))
		return s.sendError(0, -32700, "Parse error", map[string]interface{}{"details": err.Error()})
	}

	s.logger.Info("processing request", "method", req.Method, "id", req.ID)

	// Handle notifications (no response needed)
	if req.ID == nil {
		s.logger.Info("received notification (no response required)", "method", req.Method)
		switch req.Method {
		case "notifications/initialized":
			s.logger.Info("client initialization complete - server is ready")
		case "cancelled":
			s.logger.Info("request cancelled notification")
		default:
			s.logger.Debug("unknown notification", "method", req.Method)
		}
		return nil
	}

	// Route to appropriate handler
	var result interface{}
	var err error

	switch req.Method {
	case "initialize":
		s.logger.Info("handling initialize request")
		result, err = s.handleInitialize(req.Params)
	case "tools/list":
		s.logger.Info("handling tools/list request")
		result, err = s.handleListTools()
	case "tools/call":
		s.logger.Info("handling tools/call request")
		result, err = s.handleToolCall(req.Params)
	case "ping":
		s.logger.Debug("handling ping request")
		result = map[string]interface{}{}
	default:
		s.logger.Warn("unknown method requested", "method", req.Method)
		return s.sendError(req.ID, -32601, "Method not found", map[string]interface{}{"method": req.Method})
	}

	if err != nil {
		s.logger.Error("handler error", "method", req.Method, "error", err)
		return s.sendError(req.ID, -32603, err.Error(), map[string]interface{}{"method": req.Method})
	}

	s.logger.Info("request completed successfully", "method", req.Method, "id", req.ID)
	return s.sendResponse(req.ID, result)
}

// handleInitialize handles the initialize request
func (s *MCPServer) handleInitialize(params json.RawMessage) (interface{}, error) {
	return map[string]interface{}{
		"protocolVersion": protocolVersion,
		"serverInfo": map[string]interface{}{
			"name":    serverName,
			"version": serverVersion,
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{},
		},
	}, nil
}

// handleListTools returns the list of available tools
func (s *MCPServer) handleListTools() (interface{}, error) {
	tools := []map[string]interface{}{
		{
			"name": "search_proto",
			"description": "Fuzzy search across all proto definitions (services, messages, enums). " +
				"Searches in names, fields, RPC methods, and comments. " +
				"Returns structured results with match scores.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"query": map[string]interface{}{
						"type":        "string",
						"description": "Search query - can be partial name, keyword, or field name",
					},
					"limit": map[string]interface{}{
						"type":        "number",
						"description": "Maximum number of results (default: 20)",
						"default":     20,
					},
					"min_score": map[string]interface{}{
						"type":        "number",
						"description": "Minimum match score 0-100 (default: 60)",
						"default":     60,
					},
				},
				"required": []string{"query"},
			},
		},
		{
			"name": "get_service_definition",
			"description": "Get complete service definition including all RPC methods with " +
				"their request/response types and comments. " +
				"AUTOMATICALLY resolves all nested types recursively, providing " +
				"the complete structure in a single response. " +
				"Accepts both simple name (e.g., 'UserService') or " +
				"fully qualified name (e.g., 'api.v1.UserService').",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "gRPC service name (simple like 'UserService' or fully qualified like 'api.v1.UserService')",
					},
					"resolve_types": map[string]interface{}{
						"type":        "boolean",
						"description": "Recursively resolve all request/response types (default: true)",
						"default":     true,
					},
					"max_depth": map[string]interface{}{
						"type":        "number",
						"description": "Maximum recursion depth for type resolution (default: 10)",
						"default":     10,
					},
				},
				"required": []string{"name"},
			},
		},
		{
			"name": "get_message_definition",
			"description": "Get complete message definition with all fields, types, and comments. " +
				"AUTOMATICALLY resolves all nested types recursively, providing " +
				"the complete structure in a single response. " +
				"Accepts both simple name (e.g., 'User') or " +
				"fully qualified name (e.g., 'api.v1.User').",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Proto message or enum name (simple like 'User' or fully qualified like 'api.v1.User')",
					},
					"resolve_types": map[string]interface{}{
						"type":        "boolean",
						"description": "Recursively resolve all field types (default: true)",
						"default":     true,
					},
					"max_depth": map[string]interface{}{
						"type":        "number",
						"description": "Maximum recursion depth for type resolution (default: 10)",
						"default":     10,
					},
				},
				"required": []string{"name"},
			},
		},
		{
			"name": "find_type_usages",
			"description": "Find all services and RPC methods that use a given type (message or enum). " +
				"Searches recursively through nested types to find deep dependencies. " +
				"Returns information about which services, RPCs, and field paths use the type. " +
				"Useful for impact analysis and understanding type dependencies.",
			"inputSchema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"type_name": map[string]interface{}{
						"type":        "string",
						"description": "Proto message or enum name to find usages for (simple like 'User' or fully qualified like 'api.v1.User')",
					},
				},
				"required": []string{"type_name"},
			},
		},
	}

	return map[string]interface{}{
		"tools": tools,
	}, nil
}

// handleToolCall handles a tool call request
func (s *MCPServer) handleToolCall(params json.RawMessage) (interface{}, error) {
	var toolCall struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}

	if err := json.Unmarshal(params, &toolCall); err != nil {
		return nil, fmt.Errorf("invalid tool call params: %w", err)
	}

	s.logger.Info("executing tool", "tool_name", toolCall.Name, "arguments", toolCall.Arguments)

	var content string
	var err error

	switch toolCall.Name {
	case "search_proto":
		content, err = s.handleSearchProto(toolCall.Arguments)
	case "get_service_definition":
		content, err = s.handleGetService(toolCall.Arguments)
	case "get_message_definition":
		content, err = s.handleGetMessage(toolCall.Arguments)
	case "find_type_usages":
		content, err = s.handleFindTypeUsages(toolCall.Arguments)
	default:
		s.logger.Error("unknown tool requested", "tool_name", toolCall.Name, "available_tools", []string{"search_proto", "get_service_definition", "get_message_definition", "find_type_usages"})
		return nil, fmt.Errorf("unknown tool: %s (available tools: search_proto, get_service_definition, get_message_definition, find_type_usages)", toolCall.Name)
	}

	if err != nil {
		s.logger.Error("tool execution failed", "tool_name", toolCall.Name, "error", err)
		return nil, err
	}

	s.logger.Info("tool execution successful", "tool_name", toolCall.Name, "result_length", len(content))

	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": content,
			},
		},
	}, nil
}

// sendResponse sends a JSON-RPC response
func (s *MCPServer) sendResponse(id interface{}, result interface{}) error {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		s.logger.Error("failed to marshal response", "error", err, "id", id)
		return err
	}

	s.logger.Debug("sending response", "id", id, "length", len(data))

	data = append(data, '\n')
	if _, err := s.writer.Write(data); err != nil {
		s.logger.Error("failed to write response to stdout", "error", err, "id", id)
		return err
	}

	if err := s.writer.Flush(); err != nil {
		s.logger.Error("failed to flush response to stdout", "error", err, "id", id)
		return err
	}

	s.logger.Debug("response sent successfully", "id", id)
	return nil
}

// sendError sends a JSON-RPC error response
func (s *MCPServer) sendError(id interface{}, code int, message string, data interface{}) error {
	s.logger.Warn("sending error response", "id", id, "code", code, "message", message, "data", data)

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}

	respData, err := json.Marshal(resp)
	if err != nil {
		s.logger.Error("failed to marshal error response", "error", err)
		return err
	}

	respData = append(respData, '\n')
	if _, err := s.writer.Write(respData); err != nil {
		s.logger.Error("failed to write error response", "error", err)
		return err
	}

	if err := s.writer.Flush(); err != nil {
		s.logger.Error("failed to flush error response", "error", err)
		return err
	}

	return nil
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      interface{}   `json:"id"`
	Result  interface{}   `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error
type JSONRPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
