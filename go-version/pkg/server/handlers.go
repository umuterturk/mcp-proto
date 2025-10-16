package server

import (
	"encoding/json"
	"fmt"
	"strings"
)

// handleSearchProto handles the search_proto tool
func (s *MCPServer) handleSearchProto(args map[string]interface{}) (string, error) {
	// Extract parameters
	query, ok := args["query"].(string)
	if !ok || query == "" {
		return "", fmt.Errorf("query parameter is required")
	}

	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	minScore := 60
	if ms, ok := args["min_score"].(float64); ok {
		minScore = int(ms)
	}

	s.logger.Debug("search_proto", "query", query, "limit", limit, "min_score", minScore)

	// Perform search
	results := s.index.Search(query, limit, minScore)

	// Format results as JSON
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal results: %w", err)
	}

	// Add summary
	summary := fmt.Sprintf("Found %d results for query '%s':\n\n", len(results), query)
	return summary + string(data), nil
}

// handleGetService handles the get_service_definition tool
func (s *MCPServer) handleGetService(args map[string]interface{}) (string, error) {
	// Extract parameters
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name parameter is required")
	}

	resolveTypes := true
	if rt, ok := args["resolve_types"].(bool); ok {
		resolveTypes = rt
	}

	maxDepth := 10
	if md, ok := args["max_depth"].(float64); ok {
		maxDepth = int(md)
	}

	s.logger.Debug("get_service_definition", "name", name, "resolve_types", resolveTypes, "max_depth", maxDepth)

	// Get service definition
	service, err := s.index.GetService(name, resolveTypes, maxDepth)
	if err != nil {
		return "", fmt.Errorf("service not found: %s", name)
	}

	// Format as JSON
	data, err := json.MarshalIndent(service, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal service: %w", err)
	}

	// Add summary
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Service: %s\n", service["full_name"]))
	summary.WriteString(fmt.Sprintf("File: %s\n", service["file"]))

	if rpcs, ok := service["rpcs"].([]map[string]interface{}); ok {
		summary.WriteString(fmt.Sprintf("RPCs: %d\n", len(rpcs)))
	}

	if resolvedTypes, ok := service["resolved_types"].(map[string]interface{}); ok && len(resolvedTypes) > 0 {
		summary.WriteString(fmt.Sprintf("Resolved Types: %d\n", len(resolvedTypes)))
	}

	summary.WriteString("\nFull Definition:\n\n")

	return summary.String() + string(data), nil
}

// handleGetMessage handles the get_message_definition tool
func (s *MCPServer) handleGetMessage(args map[string]interface{}) (string, error) {
	// Extract parameters
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return "", fmt.Errorf("name parameter is required")
	}

	resolveTypes := true
	if rt, ok := args["resolve_types"].(bool); ok {
		resolveTypes = rt
	}

	maxDepth := 10
	if md, ok := args["max_depth"].(float64); ok {
		maxDepth = int(md)
	}

	s.logger.Debug("get_message_definition", "name", name, "resolve_types", resolveTypes, "max_depth", maxDepth)

	// Get message definition
	message, err := s.index.GetMessage(name, resolveTypes, maxDepth)
	if err != nil {
		return "", fmt.Errorf("message not found: %s", name)
	}

	// Format as JSON
	data, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	// Add summary
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Message: %s\n", message["full_name"]))
	summary.WriteString(fmt.Sprintf("File: %s\n", message["file"]))

	if fields, ok := message["fields"].([]map[string]interface{}); ok {
		summary.WriteString(fmt.Sprintf("Fields: %d\n", len(fields)))
	}

	if resolvedTypes, ok := message["resolved_types"].(map[string]interface{}); ok && len(resolvedTypes) > 0 {
		summary.WriteString(fmt.Sprintf("Resolved Types: %d\n", len(resolvedTypes)))
	}

	summary.WriteString("\nFull Definition:\n\n")

	return summary.String() + string(data), nil
}

// handleFindTypeUsages handles the find_type_usages tool
func (s *MCPServer) handleFindTypeUsages(args map[string]interface{}) (string, error) {
	// Extract parameters
	typeName, ok := args["type_name"].(string)
	if !ok || typeName == "" {
		return "", fmt.Errorf("type_name parameter is required")
	}

	s.logger.Debug("find_type_usages", "type_name", typeName)

	// Find usages
	usages, err := s.index.FindTypeUsages(typeName)
	if err != nil {
		return "", fmt.Errorf("failed to find usages: %w", err)
	}

	// Format as JSON
	data, err := json.MarshalIndent(usages, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal usages: %w", err)
	}

	// Add summary
	var summary strings.Builder
	summary.WriteString(fmt.Sprintf("Found %d usage(s) of type '%s':\n\n", len(usages), typeName))

	if len(usages) > 0 {
		// Group by service
		serviceMap := make(map[string][]string)
		for _, usage := range usages {
			serviceName := usage.ServiceName
			rpcInfo := fmt.Sprintf("  - RPC: %s (%s)", usage.RPCName, usage.UsageContext)
			if len(usage.FieldPath) > 0 {
				rpcInfo += fmt.Sprintf(" → %s", strings.Join(usage.FieldPath, "."))
			}
			serviceMap[serviceName] = append(serviceMap[serviceName], rpcInfo)
		}

		summary.WriteString("Services using this type:\n")
		for serviceName, rpcs := range serviceMap {
			summary.WriteString(fmt.Sprintf("- %s:\n", serviceName))
			for _, rpc := range rpcs {
				summary.WriteString(rpc + "\n")
			}
		}
		summary.WriteString("\nDetailed Results:\n\n")
	}

	return summary.String() + string(data), nil
}
