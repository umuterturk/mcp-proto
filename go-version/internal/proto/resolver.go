package proto

import (
	"strings"
)

// Primitive proto types that don't need resolution
var primitiveTypes = map[string]bool{
	"string":   true,
	"int32":    true,
	"int64":    true,
	"uint32":   true,
	"uint64":   true,
	"sint32":   true,
	"sint64":   true,
	"fixed32":  true,
	"fixed64":  true,
	"sfixed32": true,
	"sfixed64": true,
	"bool":     true,
	"bytes":    true,
	"float":    true,
	"double":   true,
}

// isPrimitiveType checks if a type is a primitive proto type
func isPrimitiveType(typeName string) bool {
	return primitiveTypes[typeName]
}

// resolveServiceTypes recursively resolves all request and response types for a service
// This is called by GetService when resolveTypes=true
func (pi *ProtoIndex) resolveServiceTypes(service *ProtoService, maxDepth int) map[string]interface{} {
	if maxDepth <= 0 {
		return nil
	}

	resolved := make(map[string]interface{})
	visited := make(map[string]bool)

	// Get the package context from the service full name
	contextPackage := service.FullName

	// Resolve all request and response types
	for _, rpc := range service.RPCs {
		// Resolve request type
		if !visited[rpc.RequestType] {
			if msg := pi.findMessageByType(rpc.RequestType, contextPackage); msg != nil {
				visited[rpc.RequestType] = true
				resolved[rpc.RequestType] = pi.messageToMap(msg)

				// Recursively resolve nested types
				nested := pi.resolveMessageTypes(msg, maxDepth-1, visited)
				for k, v := range nested {
					resolved[k] = v
				}
			}
		}

		// Resolve response type
		if !visited[rpc.ResponseType] {
			if msg := pi.findMessageByType(rpc.ResponseType, contextPackage); msg != nil {
				visited[rpc.ResponseType] = true
				resolved[rpc.ResponseType] = pi.messageToMap(msg)

				// Recursively resolve nested types
				nested := pi.resolveMessageTypes(msg, maxDepth-1, visited)
				for k, v := range nested {
					resolved[k] = v
				}
			}
		}
	}

	return resolved
}

// resolveMessageTypes recursively resolves all field types in a message
// This is called by GetMessage when resolveTypes=true or by resolveServiceTypes
func (pi *ProtoIndex) resolveMessageTypes(message *ProtoMessage, maxDepth int, visited map[string]bool) map[string]interface{} {
	if maxDepth <= 0 {
		return nil
	}

	if visited == nil {
		visited = make(map[string]bool)
	}

	resolved := make(map[string]interface{})
	contextPackage := message.FullName

	for _, field := range message.Fields {
		fieldType := field.Type

		// Skip primitive types
		if isPrimitiveType(fieldType) {
			continue
		}

		// Skip if already visited (circular reference)
		if visited[fieldType] {
			continue
		}

		visited[fieldType] = true

		// Try to resolve as message
		if msg := pi.findMessageByType(fieldType, contextPackage); msg != nil {
			resolved[fieldType] = pi.messageToMap(msg)

			// Recursively resolve nested types
			nested := pi.resolveMessageTypes(msg, maxDepth-1, visited)
			for k, v := range nested {
				resolved[k] = v
			}
			continue
		}

		// Try to resolve as enum
		if enum := pi.findEnumByType(fieldType, contextPackage); enum != nil {
			resolved[fieldType] = pi.enumToMap(enum)
		}
	}

	return resolved
}

// findMessageByType finds a message by type name, considering package context
// It tries multiple resolution strategies:
// 1. Exact match with full name
// 2. Match with context package prefix
// 3. Match by simple name
func (pi *ProtoIndex) findMessageByType(typeName, contextPackage string) *ProtoMessage {
	// Try exact match first
	if msg, exists := pi.messages[typeName]; exists {
		return msg
	}

	// Try with context package prefix
	// For context "api.v1.UserService", we try "api.v1.TypeName"
	if contextPackage != "" {
		packagePrefix := contextPackage
		// Remove the last component (the service/message name)
		if lastDot := strings.LastIndex(contextPackage, "."); lastDot != -1 {
			packagePrefix = contextPackage[:lastDot]
		}

		if packagePrefix != "" {
			qualifiedName := packagePrefix + "." + typeName
			if msg, exists := pi.messages[qualifiedName]; exists {
				return msg
			}
		}
	}

	// Try matching by simple name or suffix
	for fullName, msg := range pi.messages {
		if msg.Name == typeName || strings.HasSuffix(fullName, "."+typeName) {
			return msg
		}
	}

	return nil
}

// findEnumByType finds an enum by type name, considering package context
// It tries multiple resolution strategies:
// 1. Exact match with full name
// 2. Match with context package prefix
// 3. Match by simple name
func (pi *ProtoIndex) findEnumByType(typeName, contextPackage string) *ProtoEnum {
	// Try exact match first
	if enum, exists := pi.enums[typeName]; exists {
		return enum
	}

	// Try with context package prefix
	if contextPackage != "" {
		packagePrefix := contextPackage
		// Remove the last component (the service/message name)
		if lastDot := strings.LastIndex(contextPackage, "."); lastDot != -1 {
			packagePrefix = contextPackage[:lastDot]
		}

		if packagePrefix != "" {
			qualifiedName := packagePrefix + "." + typeName
			if enum, exists := pi.enums[qualifiedName]; exists {
				return enum
			}
		}
	}

	// Try matching by simple name or suffix
	for fullName, enum := range pi.enums {
		if enum.Name == typeName || strings.HasSuffix(fullName, "."+typeName) {
			return enum
		}
	}

	return nil
}

// messageToMap converts a ProtoMessage to a map for JSON serialization
func (pi *ProtoIndex) messageToMap(message *ProtoMessage) map[string]interface{} {
	fields := make([]map[string]interface{}, len(message.Fields))
	for i, field := range message.Fields {
		fields[i] = map[string]interface{}{
			"name":    field.Name,
			"type":    field.Type,
			"number":  field.Number,
			"label":   field.Label,
			"comment": field.Comment,
		}
	}

	return map[string]interface{}{
		"kind":      "message",
		"name":      message.Name,
		"full_name": message.FullName,
		"comment":   message.Comment,
		"fields":    fields,
		"file":      pi.findFileForDefinition(message.FullName, "message"),
	}
}

// enumToMap converts a ProtoEnum to a map for JSON serialization
func (pi *ProtoIndex) enumToMap(enum *ProtoEnum) map[string]interface{} {
	values := make([]map[string]interface{}, len(enum.Values))
	for i, value := range enum.Values {
		values[i] = map[string]interface{}{
			"name":    value.Name,
			"number":  value.Number,
			"comment": value.Comment,
		}
	}

	return map[string]interface{}{
		"kind":      "enum",
		"name":      enum.Name,
		"full_name": enum.FullName,
		"comment":   enum.Comment,
		"values":    values,
		"file":      pi.findFileForDefinition(enum.FullName, "enum"),
	}
}
















