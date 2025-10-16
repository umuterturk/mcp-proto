package proto

import (
	"log/slog"
	"os"
	"testing"
)

// Helper function to create a test logger
func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors in tests
	}))
}

// TestFindMessageByType tests finding messages with different naming strategies
func TestFindMessageByType(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create test messages
	msg1 := &ProtoMessage{
		Name:     "User",
		FullName: "api.v1.User",
		Fields: []ProtoField{
			{Name: "id", Type: "int32", Number: 1},
			{Name: "name", Type: "string", Number: 2},
		},
	}

	msg2 := &ProtoMessage{
		Name:     "Address",
		FullName: "api.v1.Address",
		Fields: []ProtoField{
			{Name: "street", Type: "string", Number: 1},
			{Name: "city", Type: "string", Number: 2},
		},
	}

	msg3 := &ProtoMessage{
		Name:     "User",
		FullName: "api.v2.User",
		Fields: []ProtoField{
			{Name: "id", Type: "int64", Number: 1},
		},
	}

	index.messages["api.v1.User"] = msg1
	index.messages["api.v1.Address"] = msg2
	index.messages["api.v2.User"] = msg3

	tests := []struct {
		name           string
		typeName       string
		contextPackage string
		wantFullName   string
		wantFound      bool
	}{
		{
			name:           "exact match",
			typeName:       "api.v1.User",
			contextPackage: "",
			wantFullName:   "api.v1.User",
			wantFound:      true,
		},
		{
			name:           "match with context package",
			typeName:       "Address",
			contextPackage: "api.v1.User",
			wantFullName:   "api.v1.Address",
			wantFound:      true,
		},
		{
			name:           "match with different context package",
			typeName:       "User",
			contextPackage: "api.v2.UserService",
			wantFullName:   "api.v2.User",
			wantFound:      true,
		},
		{
			name:           "match by simple name (ambiguous, returns first found)",
			typeName:       "User",
			contextPackage: "",
			wantFullName:   "", // Could be either api.v1.User or api.v2.User
			wantFound:      true,
		},
		{
			name:           "not found",
			typeName:       "NotExists",
			contextPackage: "api.v1",
			wantFullName:   "",
			wantFound:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := index.findMessageByType(tt.typeName, tt.contextPackage)
			if (msg != nil) != tt.wantFound {
				t.Errorf("findMessageByType() found = %v, want %v", msg != nil, tt.wantFound)
				return
			}
			if tt.wantFound && tt.wantFullName != "" && msg.FullName != tt.wantFullName {
				t.Errorf("findMessageByType() fullName = %v, want %v", msg.FullName, tt.wantFullName)
			}
		})
	}
}

// TestFindEnumByType tests finding enums with different naming strategies
func TestFindEnumByType(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create test enums
	enum1 := &ProtoEnum{
		Name:     "Status",
		FullName: "api.v1.Status",
		Values: []ProtoField{
			{Name: "ACTIVE", Number: 0},
			{Name: "INACTIVE", Number: 1},
		},
	}

	enum2 := &ProtoEnum{
		Name:     "Role",
		FullName: "api.v1.Role",
		Values: []ProtoField{
			{Name: "ADMIN", Number: 0},
			{Name: "USER", Number: 1},
		},
	}

	index.enums["api.v1.Status"] = enum1
	index.enums["api.v1.Role"] = enum2

	tests := []struct {
		name           string
		typeName       string
		contextPackage string
		wantFullName   string
		wantFound      bool
	}{
		{
			name:           "exact match",
			typeName:       "api.v1.Status",
			contextPackage: "",
			wantFullName:   "api.v1.Status",
			wantFound:      true,
		},
		{
			name:           "match with context package",
			typeName:       "Role",
			contextPackage: "api.v1.User",
			wantFullName:   "api.v1.Role",
			wantFound:      true,
		},
		{
			name:           "match by simple name",
			typeName:       "Status",
			contextPackage: "",
			wantFullName:   "api.v1.Status",
			wantFound:      true,
		},
		{
			name:           "not found",
			typeName:       "NotExists",
			contextPackage: "api.v1",
			wantFullName:   "",
			wantFound:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			enum := index.findEnumByType(tt.typeName, tt.contextPackage)
			if (enum != nil) != tt.wantFound {
				t.Errorf("findEnumByType() found = %v, want %v", enum != nil, tt.wantFound)
				return
			}
			if tt.wantFound && enum.FullName != tt.wantFullName {
				t.Errorf("findEnumByType() fullName = %v, want %v", enum.FullName, tt.wantFullName)
			}
		})
	}
}

// TestIsPrimitiveType tests primitive type detection
func TestIsPrimitiveType(t *testing.T) {
	tests := []struct {
		typeName string
		want     bool
	}{
		{"string", true},
		{"int32", true},
		{"int64", true},
		{"bool", true},
		{"bytes", true},
		{"float", true},
		{"double", true},
		{"User", false},
		{"api.v1.Message", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			if got := isPrimitiveType(tt.typeName); got != tt.want {
				t.Errorf("isPrimitiveType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

// TestResolveMessageTypes tests recursive message type resolution
func TestResolveMessageTypes(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create nested message structure:
	// User -> Address, Status (enum)
	// Address -> Country
	// Country is primitive (string in this case, but let's make it a message for testing)

	country := &ProtoMessage{
		Name:     "Country",
		FullName: "api.v1.Country",
		Fields: []ProtoField{
			{Name: "name", Type: "string", Number: 1},
			{Name: "code", Type: "string", Number: 2},
		},
	}

	address := &ProtoMessage{
		Name:     "Address",
		FullName: "api.v1.Address",
		Fields: []ProtoField{
			{Name: "street", Type: "string", Number: 1},
			{Name: "country", Type: "Country", Number: 2},
		},
	}

	status := &ProtoEnum{
		Name:     "Status",
		FullName: "api.v1.Status",
		Values: []ProtoField{
			{Name: "ACTIVE", Number: 0},
			{Name: "INACTIVE", Number: 1},
		},
	}

	user := &ProtoMessage{
		Name:     "User",
		FullName: "api.v1.User",
		Fields: []ProtoField{
			{Name: "id", Type: "int32", Number: 1},
			{Name: "name", Type: "string", Number: 2},
			{Name: "address", Type: "Address", Number: 3},
			{Name: "status", Type: "Status", Number: 4},
		},
	}

	index.messages["api.v1.User"] = user
	index.messages["api.v1.Address"] = address
	index.messages["api.v1.Country"] = country
	index.enums["api.v1.Status"] = status

	// Test with sufficient depth
	resolved := index.resolveMessageTypes(user, 10, nil)

	// Should resolve Address, Country, and Status
	if len(resolved) != 3 {
		t.Errorf("resolveMessageTypes() resolved %d types, want 3", len(resolved))
	}

	// Check Address is resolved
	if _, ok := resolved["Address"]; !ok {
		t.Error("resolveMessageTypes() did not resolve Address")
	}

	// Check Country is resolved
	if _, ok := resolved["Country"]; !ok {
		t.Error("resolveMessageTypes() did not resolve Country")
	}

	// Check Status enum is resolved
	if _, ok := resolved["Status"]; !ok {
		t.Error("resolveMessageTypes() did not resolve Status enum")
	}

	// Verify Address contains the correct data
	if addrMap, ok := resolved["Address"].(map[string]interface{}); ok {
		if kind := addrMap["kind"]; kind != "message" {
			t.Errorf("Address kind = %v, want 'message'", kind)
		}
		if fullName := addrMap["full_name"]; fullName != "api.v1.Address" {
			t.Errorf("Address full_name = %v, want 'api.v1.Address'", fullName)
		}
	} else {
		t.Error("Address is not a map[string]interface{}")
	}

	// Verify Status enum contains the correct data
	if statusMap, ok := resolved["Status"].(map[string]interface{}); ok {
		if kind := statusMap["kind"]; kind != "enum" {
			t.Errorf("Status kind = %v, want 'enum'", kind)
		}
	} else {
		t.Error("Status is not a map[string]interface{}")
	}
}

// TestResolveMessageTypesMaxDepth tests depth limiting
func TestResolveMessageTypesMaxDepth(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create chain: A -> B -> C
	msgC := &ProtoMessage{
		Name:     "C",
		FullName: "api.v1.C",
		Fields: []ProtoField{
			{Name: "value", Type: "string", Number: 1},
		},
	}

	msgB := &ProtoMessage{
		Name:     "B",
		FullName: "api.v1.B",
		Fields: []ProtoField{
			{Name: "c", Type: "C", Number: 1},
		},
	}

	msgA := &ProtoMessage{
		Name:     "A",
		FullName: "api.v1.A",
		Fields: []ProtoField{
			{Name: "b", Type: "B", Number: 1},
		},
	}

	index.messages["api.v1.A"] = msgA
	index.messages["api.v1.B"] = msgB
	index.messages["api.v1.C"] = msgC

	// Test with depth 1: should only resolve B
	resolved := index.resolveMessageTypes(msgA, 1, nil)
	if len(resolved) != 1 {
		t.Errorf("resolveMessageTypes(depth=1) resolved %d types, want 1", len(resolved))
	}
	if _, ok := resolved["B"]; !ok {
		t.Error("resolveMessageTypes(depth=1) did not resolve B")
	}
	if _, ok := resolved["C"]; ok {
		t.Error("resolveMessageTypes(depth=1) should not resolve C")
	}

	// Test with depth 2: should resolve B and C
	resolved = index.resolveMessageTypes(msgA, 2, nil)
	if len(resolved) != 2 {
		t.Errorf("resolveMessageTypes(depth=2) resolved %d types, want 2", len(resolved))
	}
	if _, ok := resolved["B"]; !ok {
		t.Error("resolveMessageTypes(depth=2) did not resolve B")
	}
	if _, ok := resolved["C"]; !ok {
		t.Error("resolveMessageTypes(depth=2) did not resolve C")
	}

	// Test with depth 0: should resolve nothing
	resolved = index.resolveMessageTypes(msgA, 0, nil)
	if len(resolved) != 0 {
		t.Errorf("resolveMessageTypes(depth=0) resolved %d types, want 0", len(resolved))
	}
}

// TestResolveMessageTypesCircular tests circular reference handling
func TestResolveMessageTypesCircular(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create circular reference: A -> B -> A
	msgA := &ProtoMessage{
		Name:     "A",
		FullName: "api.v1.A",
		Fields: []ProtoField{
			{Name: "b", Type: "B", Number: 1},
		},
	}

	msgB := &ProtoMessage{
		Name:     "B",
		FullName: "api.v1.B",
		Fields: []ProtoField{
			{Name: "a", Type: "A", Number: 1},
		},
	}

	index.messages["api.v1.A"] = msgA
	index.messages["api.v1.B"] = msgB

	// Should not infinite loop - visited map should prevent it
	resolved := index.resolveMessageTypes(msgA, 10, nil)

	// Should resolve both A and B (visited map prevents infinite recursion, not duplicate entries)
	// When resolving A: finds B, resolves B
	// When resolving B: finds A, resolves A (but A's reference to B is skipped as visited)
	if len(resolved) != 2 {
		t.Errorf("resolveMessageTypes() with circular ref resolved %d types, want 2", len(resolved))
	}

	if _, ok := resolved["B"]; !ok {
		t.Error("resolveMessageTypes() did not resolve B")
	}

	if _, ok := resolved["A"]; !ok {
		t.Error("resolveMessageTypes() did not resolve A")
	}
}

// TestResolveServiceTypes tests service type resolution
func TestResolveServiceTypes(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create request and response messages
	request := &ProtoMessage{
		Name:     "GetUserRequest",
		FullName: "api.v1.GetUserRequest",
		Fields: []ProtoField{
			{Name: "id", Type: "int32", Number: 1},
		},
	}

	response := &ProtoMessage{
		Name:     "GetUserResponse",
		FullName: "api.v1.GetUserResponse",
		Fields: []ProtoField{
			{Name: "user", Type: "User", Number: 1},
		},
	}

	user := &ProtoMessage{
		Name:     "User",
		FullName: "api.v1.User",
		Fields: []ProtoField{
			{Name: "id", Type: "int32", Number: 1},
			{Name: "name", Type: "string", Number: 2},
		},
	}

	service := &ProtoService{
		Name:     "UserService",
		FullName: "api.v1.UserService",
		RPCs: []ProtoRPC{
			{
				Name:         "GetUser",
				RequestType:  "GetUserRequest",
				ResponseType: "GetUserResponse",
			},
		},
	}

	index.messages["api.v1.GetUserRequest"] = request
	index.messages["api.v1.GetUserResponse"] = response
	index.messages["api.v1.User"] = user
	index.services["api.v1.UserService"] = service

	// Resolve service types
	resolved := index.resolveServiceTypes(service, 10)

	// Should resolve GetUserRequest, GetUserResponse, and User
	if len(resolved) != 3 {
		t.Errorf("resolveServiceTypes() resolved %d types, want 3", len(resolved))
	}

	// Check all types are resolved
	for _, typeName := range []string{"GetUserRequest", "GetUserResponse", "User"} {
		if _, ok := resolved[typeName]; !ok {
			t.Errorf("resolveServiceTypes() did not resolve %s", typeName)
		}
	}
}

// TestGetMessageWithResolution tests the integrated GetMessage with type resolution
func TestGetMessageWithResolution(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create test data
	address := &ProtoMessage{
		Name:     "Address",
		FullName: "api.v1.Address",
		Fields: []ProtoField{
			{Name: "street", Type: "string", Number: 1},
		},
	}

	user := &ProtoMessage{
		Name:     "User",
		FullName: "api.v1.User",
		Fields: []ProtoField{
			{Name: "id", Type: "int32", Number: 1},
			{Name: "address", Type: "Address", Number: 2},
		},
	}

	index.messages["api.v1.User"] = user
	index.messages["api.v1.Address"] = address

	// Get message without resolution
	result, err := index.GetMessage("User", false, 10)
	if err != nil {
		t.Fatalf("GetMessage() error = %v", err)
	}
	if _, ok := result["resolved_types"]; ok {
		t.Error("GetMessage(resolveTypes=false) should not have resolved_types")
	}

	// Get message with resolution
	result, err = index.GetMessage("User", true, 10)
	if err != nil {
		t.Fatalf("GetMessage() error = %v", err)
	}
	resolvedTypes, ok := result["resolved_types"]
	if !ok {
		t.Fatal("GetMessage(resolveTypes=true) should have resolved_types")
	}

	resolvedMap := resolvedTypes.(map[string]interface{})
	if len(resolvedMap) != 1 {
		t.Errorf("GetMessage() resolved %d types, want 1", len(resolvedMap))
	}

	if _, ok := resolvedMap["Address"]; !ok {
		t.Error("GetMessage() did not resolve Address")
	}
}

// TestGetServiceWithResolution tests the integrated GetService with type resolution
func TestGetServiceWithResolution(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create test data
	request := &ProtoMessage{
		Name:     "GetUserRequest",
		FullName: "api.v1.GetUserRequest",
		Fields: []ProtoField{
			{Name: "id", Type: "int32", Number: 1},
		},
	}

	response := &ProtoMessage{
		Name:     "GetUserResponse",
		FullName: "api.v1.GetUserResponse",
		Fields: []ProtoField{
			{Name: "name", Type: "string", Number: 1},
		},
	}

	service := &ProtoService{
		Name:     "UserService",
		FullName: "api.v1.UserService",
		RPCs: []ProtoRPC{
			{
				Name:         "GetUser",
				RequestType:  "GetUserRequest",
				ResponseType: "GetUserResponse",
			},
		},
	}

	index.messages["api.v1.GetUserRequest"] = request
	index.messages["api.v1.GetUserResponse"] = response
	index.services["api.v1.UserService"] = service

	// Get service without resolution
	result, err := index.GetService("UserService", false, 10)
	if err != nil {
		t.Fatalf("GetService() error = %v", err)
	}
	if _, ok := result["resolved_types"]; ok {
		t.Error("GetService(resolveTypes=false) should not have resolved_types")
	}

	// Get service with resolution
	result, err = index.GetService("UserService", true, 10)
	if err != nil {
		t.Fatalf("GetService() error = %v", err)
	}
	resolvedTypes, ok := result["resolved_types"]
	if !ok {
		t.Fatal("GetService(resolveTypes=true) should have resolved_types")
	}

	resolvedMap := resolvedTypes.(map[string]interface{})
	if len(resolvedMap) != 2 {
		t.Errorf("GetService() resolved %d types, want 2", len(resolvedMap))
	}

	for _, typeName := range []string{"GetUserRequest", "GetUserResponse"} {
		if _, ok := resolvedMap[typeName]; !ok {
			t.Errorf("GetService() did not resolve %s", typeName)
		}
	}
}
