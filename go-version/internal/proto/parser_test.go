package proto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFile(t *testing.T) {
	// Create a temporary test proto file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.proto")

	testContent := `syntax = "proto3";

package api.v1;

import "google/protobuf/timestamp.proto";

// User service handles user operations
service UserService {
    // Get user by ID
    rpc GetUser(GetUserRequest) returns (GetUserResponse);
    // Create new user
    rpc CreateUser(CreateUserRequest) returns (CreateUserResponse);
}

// User message
message User {
    string id = 1;          // User ID
    string name = 2;        // User name
    string email = 3;       // User email
    UserRole role = 4;      // User role
}

// User role enum
enum UserRole {
    USER_ROLE_UNSPECIFIED = 0;
    USER_ROLE_ADMIN = 1;
    USER_ROLE_USER = 2;
}

message GetUserRequest {
    string id = 1;
}

message GetUserResponse {
    User user = 1;
}

message CreateUserRequest {
    string name = 1;
    string email = 2;
}

message CreateUserResponse {
    User user = 1;
}
`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Parse the file
	parser := NewParser()
	protoFile, err := parser.ParseFile(testFile)
	if err != nil {
		t.Fatalf("Failed to parse file: %v", err)
	}

	// Verify syntax
	if protoFile.Syntax != "proto3" {
		t.Errorf("Expected syntax 'proto3', got '%s'", protoFile.Syntax)
	}

	// Verify package
	if protoFile.Package != "api.v1" {
		t.Errorf("Expected package 'api.v1', got '%s'", protoFile.Package)
	}

	// Verify imports
	if len(protoFile.Imports) != 1 {
		t.Errorf("Expected 1 import, got %d", len(protoFile.Imports))
	}
	if len(protoFile.Imports) > 0 && protoFile.Imports[0] != "google/protobuf/timestamp.proto" {
		t.Errorf("Expected import 'google/protobuf/timestamp.proto', got '%s'", protoFile.Imports[0])
	}

	// Verify services
	if len(protoFile.Services) != 1 {
		t.Fatalf("Expected 1 service, got %d", len(protoFile.Services))
	}
	service := protoFile.Services[0]
	if service.Name != "UserService" {
		t.Errorf("Expected service name 'UserService', got '%s'", service.Name)
	}
	if service.FullName != "api.v1.UserService" {
		t.Errorf("Expected full name 'api.v1.UserService', got '%s'", service.FullName)
	}
	if len(service.RPCs) != 2 {
		t.Errorf("Expected 2 RPCs, got %d", len(service.RPCs))
	}

	// Verify messages
	if len(protoFile.Messages) < 1 {
		t.Fatalf("Expected at least 1 message, got %d", len(protoFile.Messages))
	}

	// Find User message
	var userMsg *ProtoMessage
	for i := range protoFile.Messages {
		if protoFile.Messages[i].Name == "User" {
			userMsg = &protoFile.Messages[i]
			break
		}
	}
	if userMsg == nil {
		t.Fatal("User message not found")
	}
	if len(userMsg.Fields) != 4 {
		t.Errorf("Expected 4 fields in User message, got %d", len(userMsg.Fields))
	}

	// Verify enums
	if len(protoFile.Enums) != 1 {
		t.Fatalf("Expected 1 enum, got %d", len(protoFile.Enums))
	}
	enum := protoFile.Enums[0]
	if enum.Name != "UserRole" {
		t.Errorf("Expected enum name 'UserRole', got '%s'", enum.Name)
	}
	if len(enum.Values) != 3 {
		t.Errorf("Expected 3 enum values, got %d", len(enum.Values))
	}
}

func TestExtractSyntax(t *testing.T) {
	tests := []struct {
		content  string
		expected string
	}{
		{`syntax = "proto3";`, "proto3"},
		{`syntax = 'proto2';`, "proto2"},
		{`// no syntax declaration`, "proto2"},
	}

	parser := NewParser()
	for _, tt := range tests {
		result := parser.extractSyntax(tt.content)
		if result != tt.expected {
			t.Errorf("extractSyntax(%q) = %q, want %q", tt.content, result, tt.expected)
		}
	}
}

func TestExtractPackage(t *testing.T) {
	tests := []struct {
		content  string
		expected string
	}{
		{`package api.v1;`, "api.v1"},
		{`package com.example;`, "com.example"},
		{`// no package`, ""},
	}

	parser := NewParser()
	for _, tt := range tests {
		result := parser.extractPackage(tt.content)
		if result != tt.expected {
			t.Errorf("extractPackage(%q) = %q, want %q", tt.content, result, tt.expected)
		}
	}
}













