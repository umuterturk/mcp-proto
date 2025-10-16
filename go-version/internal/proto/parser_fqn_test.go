package proto

import (
	"os"
	"path/filepath"
	"testing"
)

// TestParseFullyQualifiedFieldTypes tests that the parser correctly handles
// fully qualified type names in field definitions (e.g., udemy.dto.Price)
func TestParseFullyQualifiedFieldTypes(t *testing.T) {
	// Create a temporary proto file with fully qualified types
	tempDir := t.TempDir()
	protoFile := filepath.Join(tempDir, "test.proto")

	content := `syntax = "proto3";

package test.v1;

message TaxableLine {
	udemy.dto.payments.checkout_orchestrator.v1beta1.ProductReference product_reference = 1;
	udemy.dto.payments.checkout_orchestrator.v1beta1.Price unit_net_price = 2;
	int64 quantity = 3;
}

message SimpleMessage {
	string name = 1;
	Price local_price = 2;
}

service TestService {
	rpc Calculate(udemy.rpc.payments.v1.CalculateRequest) returns (udemy.rpc.payments.v1.CalculateResponse);
	rpc SimpleCall(SimpleRequest) returns (SimpleResponse);
}
`

	if err := os.WriteFile(protoFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test proto file: %v", err)
	}

	// Parse the file
	parser := NewParser()
	parsed, err := parser.ParseFile(protoFile)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	// Test message with fully qualified field types
	if len(parsed.Messages) < 1 {
		t.Fatal("Expected at least 1 message")
	}

	taxableLine := parsed.Messages[0]
	if taxableLine.Name != "TaxableLine" {
		t.Errorf("Expected message name TaxableLine, got %s", taxableLine.Name)
	}

	if len(taxableLine.Fields) != 3 {
		t.Fatalf("Expected 3 fields in TaxableLine, got %d", len(taxableLine.Fields))
	}

	// Check fully qualified field types are preserved
	tests := []struct {
		fieldName string
		wantType  string
	}{
		{
			fieldName: "product_reference",
			wantType:  "udemy.dto.payments.checkout_orchestrator.v1beta1.ProductReference",
		},
		{
			fieldName: "unit_net_price",
			wantType:  "udemy.dto.payments.checkout_orchestrator.v1beta1.Price",
		},
		{
			fieldName: "quantity",
			wantType:  "int64",
		},
	}

	for _, tt := range tests {
		var found bool
		for _, field := range taxableLine.Fields {
			if field.Name == tt.fieldName {
				found = true
				if field.Type != tt.wantType {
					t.Errorf("Field %s: type = %q, want %q", tt.fieldName, field.Type, tt.wantType)
				}
				break
			}
		}
		if !found {
			t.Errorf("Field %s not found in TaxableLine", tt.fieldName)
		}
	}

	// Test message with simple (non-qualified) field type
	simpleMsg := parsed.Messages[1]
	if simpleMsg.Name != "SimpleMessage" {
		t.Errorf("Expected message name SimpleMessage, got %s", simpleMsg.Name)
	}

	if len(simpleMsg.Fields) != 2 {
		t.Fatalf("Expected 2 fields in SimpleMessage, got %d", len(simpleMsg.Fields))
	}

	// Check that simple type names also work
	for _, field := range simpleMsg.Fields {
		if field.Name == "local_price" {
			if field.Type != "Price" {
				t.Errorf("Field local_price: type = %q, want %q", field.Type, "Price")
			}
		}
	}

	// Test service with fully qualified request/response types
	if len(parsed.Services) < 1 {
		t.Fatal("Expected at least 1 service")
	}

	service := parsed.Services[0]
	if service.Name != "TestService" {
		t.Errorf("Expected service name TestService, got %s", service.Name)
	}

	if len(service.RPCs) != 2 {
		t.Fatalf("Expected 2 RPCs in TestService, got %d", len(service.RPCs))
	}

	// Check fully qualified RPC types
	rpcTests := []struct {
		rpcName      string
		wantRequest  string
		wantResponse string
	}{
		{
			rpcName:      "Calculate",
			wantRequest:  "udemy.rpc.payments.v1.CalculateRequest",
			wantResponse: "udemy.rpc.payments.v1.CalculateResponse",
		},
		{
			rpcName:      "SimpleCall",
			wantRequest:  "SimpleRequest",
			wantResponse: "SimpleResponse",
		},
	}

	for _, tt := range rpcTests {
		var found bool
		for _, rpc := range service.RPCs {
			if rpc.Name == tt.rpcName {
				found = true
				if rpc.RequestType != tt.wantRequest {
					t.Errorf("RPC %s: RequestType = %q, want %q", tt.rpcName, rpc.RequestType, tt.wantRequest)
				}
				if rpc.ResponseType != tt.wantResponse {
					t.Errorf("RPC %s: ResponseType = %q, want %q", tt.rpcName, rpc.ResponseType, tt.wantResponse)
				}
				break
			}
		}
		if !found {
			t.Errorf("RPC %s not found in TestService", tt.rpcName)
		}
	}
}
