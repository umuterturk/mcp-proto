package proto

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFindTypeUsages_DirectUsage tests finding a type that's used directly in request/response
func TestFindTypeUsages_DirectUsage(t *testing.T) {
	index := NewProtoIndex(testLogger())
	tempDir := t.TempDir()

	// Create a message type
	userProto := filepath.Join(tempDir, "user.proto")
	userContent := `syntax = "proto3";

package api.v1;

message User {
	int64 id = 1;
	string name = 2;
}

message GetUserRequest {
	int64 user_id = 1;
}
`
	if err := os.WriteFile(userProto, []byte(userContent), 0644); err != nil {
		t.Fatalf("Failed to write user proto: %v", err)
	}

	// Create a service that uses User directly
	serviceProto := filepath.Join(tempDir, "service.proto")
	serviceContent := `syntax = "proto3";

package api.v1;

service UserService {
	rpc GetUser(GetUserRequest) returns (User);
}
`
	if err := os.WriteFile(serviceProto, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("Failed to write service proto: %v", err)
	}

	// Index files
	if _, err := index.IndexDirectory(tempDir); err != nil {
		t.Fatalf("Failed to index directory: %v", err)
	}

	// Find usages of User
	usages, err := index.FindTypeUsages("User")
	if err != nil {
		t.Fatalf("FindTypeUsages() error = %v", err)
	}

	if len(usages) != 1 {
		t.Fatalf("Expected 1 usage, got %d", len(usages))
	}

	usage := usages[0]
	if usage.ServiceName != "UserService" {
		t.Errorf("ServiceName = %s, want UserService", usage.ServiceName)
	}
	if usage.RPCName != "GetUser" {
		t.Errorf("RPCName = %s, want GetUser", usage.RPCName)
	}
	if usage.UsageContext != "Response" {
		t.Errorf("UsageContext = %s, want Response", usage.UsageContext)
	}
	if usage.Depth != 0 {
		t.Errorf("Depth = %d, want 0 (direct usage)", usage.Depth)
	}
	if len(usage.FieldPath) != 0 {
		t.Errorf("FieldPath should be empty for direct usage, got %v", usage.FieldPath)
	}
}

// TestFindTypeUsages_NestedUsage tests finding a type that's nested in another message
func TestFindTypeUsages_NestedUsage(t *testing.T) {
	index := NewProtoIndex(testLogger())
	tempDir := t.TempDir()

	// Create nested message types
	typesProto := filepath.Join(tempDir, "types.proto")
	typesContent := `syntax = "proto3";

package api.v1;

message Price {
	string amount = 1;
	string currency_code = 2;
}

message Product {
	int64 id = 1;
	string name = 2;
	Price price = 3;
}

message GetProductRequest {
	int64 product_id = 1;
}

message GetProductResponse {
	Product product = 1;
}
`
	if err := os.WriteFile(typesProto, []byte(typesContent), 0644); err != nil {
		t.Fatalf("Failed to write types proto: %v", err)
	}

	// Create service
	serviceProto := filepath.Join(tempDir, "service.proto")
	serviceContent := `syntax = "proto3";

package api.v1;

service ProductService {
	rpc GetProduct(GetProductRequest) returns (GetProductResponse);
}
`
	if err := os.WriteFile(serviceProto, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("Failed to write service proto: %v", err)
	}

	// Index files
	if _, err := index.IndexDirectory(tempDir); err != nil {
		t.Fatalf("Failed to index directory: %v", err)
	}

	// Find usages of Price (nested in Product, which is in GetProductResponse)
	usages, err := index.FindTypeUsages("Price")
	if err != nil {
		t.Fatalf("FindTypeUsages() error = %v", err)
	}

	if len(usages) != 1 {
		t.Fatalf("Expected 1 usage, got %d", len(usages))
	}

	usage := usages[0]
	if usage.ServiceName != "ProductService" {
		t.Errorf("ServiceName = %s, want ProductService", usage.ServiceName)
	}
	if usage.RPCName != "GetProduct" {
		t.Errorf("RPCName = %s, want GetProduct", usage.RPCName)
	}
	if usage.UsageContext != "Response" {
		t.Errorf("UsageContext = %s, want Response", usage.UsageContext)
	}
	if usage.Depth != 2 {
		t.Errorf("Depth = %d, want 2 (nested: Response->Product->Price)", usage.Depth)
	}

	// Check field path: product -> price
	expectedPath := []string{"product", "price"}
	if len(usage.FieldPath) != len(expectedPath) {
		t.Errorf("FieldPath length = %d, want %d", len(usage.FieldPath), len(expectedPath))
	} else {
		for i, expected := range expectedPath {
			if usage.FieldPath[i] != expected {
				t.Errorf("FieldPath[%d] = %s, want %s", i, usage.FieldPath[i], expected)
			}
		}
	}
}

// TestFindTypeUsages_MultipleUsages tests finding a type used in multiple places
func TestFindTypeUsages_MultipleUsages(t *testing.T) {
	index := NewProtoIndex(testLogger())
	tempDir := t.TempDir()

	// Create common types
	typesProto := filepath.Join(tempDir, "types.proto")
	typesContent := `syntax = "proto3";

package api.v1;

message Address {
	string street = 1;
	string city = 2;
	string country = 3;
}

message User {
	int64 id = 1;
	string name = 2;
	Address address = 3;
}

message Company {
	int64 id = 1;
	string name = 2;
	Address headquarters = 3;
}

message CreateUserRequest {
	string name = 1;
	Address address = 2;
}
`
	if err := os.WriteFile(typesProto, []byte(typesContent), 0644); err != nil {
		t.Fatalf("Failed to write types proto: %v", err)
	}

	// Create services
	serviceProto := filepath.Join(tempDir, "service.proto")
	serviceContent := `syntax = "proto3";

package api.v1;

service UserService {
	rpc GetUser(GetUserRequest) returns (User);
	rpc CreateUser(CreateUserRequest) returns (User);
}

service CompanyService {
	rpc GetCompany(GetCompanyRequest) returns (Company);
}

message GetUserRequest {
	int64 id = 1;
}

message GetCompanyRequest {
	int64 id = 1;
}
`
	if err := os.WriteFile(serviceProto, []byte(serviceContent), 0644); err != nil {
		t.Fatalf("Failed to write service proto: %v", err)
	}

	// Index files
	if _, err := index.IndexDirectory(tempDir); err != nil {
		t.Fatalf("Failed to index directory: %v", err)
	}

	// Find usages of Address (used in multiple places)
	usages, err := index.FindTypeUsages("Address")
	if err != nil {
		t.Fatalf("FindTypeUsages() error = %v", err)
	}

	// Address is used in:
	// 1. CreateUserRequest.address (direct, request)
	// 2. User.address (via GetUser response and CreateUser response)
	// 3. Company.headquarters (via GetCompany response)
	// Expected: at least 4 usages (2 from UserService, 2 from CompanyService)
	if len(usages) < 3 {
		t.Errorf("Expected at least 3 usages, got %d", len(usages))
		for i, u := range usages {
			t.Logf("Usage %d: Service=%s, RPC=%s, Context=%s, Path=%v",
				i, u.ServiceName, u.RPCName, u.UsageContext, u.FieldPath)
		}
	}

	// Verify we have usages from both services
	serviceUsages := make(map[string]int)
	for _, usage := range usages {
		serviceUsages[usage.ServiceName]++
	}

	if serviceUsages["UserService"] == 0 {
		t.Error("Expected usages from UserService")
	}
	if serviceUsages["CompanyService"] == 0 {
		t.Error("Expected usages from CompanyService")
	}
}

// TestFindTypeUsages_ProductReference tests the real-world ProductReference scenario
func TestFindTypeUsages_ProductReference(t *testing.T) {
	index := NewProtoIndex(testLogger())
	tempDir := t.TempDir()

	// Create shared types (simulating acme.dto.payments.checkout_orchestrator.v1beta1)
	sharedTypesProto := filepath.Join(tempDir, "shared_types.proto")
	sharedTypesContent := `syntax = "proto3";

package acme.dto.payments.checkout_orchestrator.v1beta1;

message ProductReference {
	int64 product_id = 1;
	string product_type = 2;
}

message Price {
	string amount = 1;
	string currency_code = 2;
}
`
	if err := os.WriteFile(sharedTypesProto, []byte(sharedTypesContent), 0644); err != nil {
		t.Fatalf("Failed to write shared types proto: %v", err)
	}

	// Create tax service types and RPCs
	taxProto := filepath.Join(tempDir, "tax.proto")
	taxContent := `syntax = "proto3";

package acme.rpc.payments.checkout_orchestrator.info.v1beta1;

message TaxableLine {
	acme.dto.payments.checkout_orchestrator.v1beta1.ProductReference product_reference = 1;
	acme.dto.payments.checkout_orchestrator.v1beta1.Price unit_net_price = 2;
	int64 quantity = 3;
}

message TaxedLine {
	acme.dto.payments.checkout_orchestrator.v1beta1.ProductReference product_reference = 1;
	PriceBreakdown unit_price_breakdown = 2;
	string line_item_ref = 3;
	int64 quantity = 4;
}

message PriceBreakdown {
	acme.dto.payments.checkout_orchestrator.v1beta1.Price tax_price = 1;
	acme.dto.payments.checkout_orchestrator.v1beta1.Price net_price = 2;
	acme.dto.payments.checkout_orchestrator.v1beta1.Price gross_price = 3;
}

message CalculateTaxInfoRequest {
	repeated TaxableLine taxable_lines = 1;
}

message CalculateTaxInfoResponse {
	repeated TaxedLine taxed_lines = 1;
}

service TaxInfoService {
	rpc CalculateTaxInfo(CalculateTaxInfoRequest) returns (CalculateTaxInfoResponse);
}
`
	if err := os.WriteFile(taxProto, []byte(taxContent), 0644); err != nil {
		t.Fatalf("Failed to write tax proto: %v", err)
	}

	// Index files
	if _, err := index.IndexDirectory(tempDir); err != nil {
		t.Fatalf("Failed to index directory: %v", err)
	}

	// Find usages of ProductReference
	usages, err := index.FindTypeUsages("ProductReference")
	if err != nil {
		t.Fatalf("FindTypeUsages() error = %v", err)
	}

	// ProductReference is used in:
	// 1. CalculateTaxInfoRequest -> taxable_lines (TaxableLine) -> product_reference
	// 2. CalculateTaxInfoResponse -> taxed_lines (TaxedLine) -> product_reference
	if len(usages) != 2 {
		t.Errorf("Expected 2 usages, got %d", len(usages))
		for i, u := range usages {
			t.Logf("Usage %d: Service=%s, RPC=%s, Context=%s, Path=%v, Depth=%d",
				i, u.ServiceName, u.RPCName, u.UsageContext, u.FieldPath, u.Depth)
		}
	}

	// Verify both usages are from TaxInfoService.CalculateTaxInfo
	for _, usage := range usages {
		if usage.ServiceName != "TaxInfoService" {
			t.Errorf("ServiceName = %s, want TaxInfoService", usage.ServiceName)
		}
		if usage.RPCName != "CalculateTaxInfo" {
			t.Errorf("RPCName = %s, want CalculateTaxInfo", usage.RPCName)
		}

		// Check field paths
		if usage.UsageContext == "Request" {
			expectedPath := []string{"taxable_lines", "product_reference"}
			if len(usage.FieldPath) != len(expectedPath) {
				t.Errorf("Request FieldPath length = %d, want %d", len(usage.FieldPath), len(expectedPath))
			} else {
				for i, expected := range expectedPath {
					if usage.FieldPath[i] != expected {
						t.Errorf("Request FieldPath[%d] = %s, want %s", i, usage.FieldPath[i], expected)
					}
				}
			}
		} else if usage.UsageContext == "Response" {
			expectedPath := []string{"taxed_lines", "product_reference"}
			if len(usage.FieldPath) != len(expectedPath) {
				t.Errorf("Response FieldPath length = %d, want %d", len(usage.FieldPath), len(expectedPath))
			} else {
				for i, expected := range expectedPath {
					if usage.FieldPath[i] != expected {
						t.Errorf("Response FieldPath[%d] = %s, want %s", i, usage.FieldPath[i], expected)
					}
				}
			}
		}
	}
}

// TestFindTypeUsages_NonExistent tests finding a type that doesn't exist
func TestFindTypeUsages_NonExistent(t *testing.T) {
	index := NewProtoIndex(testLogger())
	tempDir := t.TempDir()

	// Create a simple proto file
	protoFile := filepath.Join(tempDir, "test.proto")
	content := `syntax = "proto3";

package test.v1;

message User {
	int64 id = 1;
}
`
	if err := os.WriteFile(protoFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write proto: %v", err)
	}

	if _, err := index.IndexDirectory(tempDir); err != nil {
		t.Fatalf("Failed to index directory: %v", err)
	}

	// Try to find usages of a non-existent type
	_, err := index.FindTypeUsages("NonExistentType")
	if err == nil {
		t.Error("Expected error for non-existent type, got nil")
	}
}

// TestFindTypeUsages_EnumType tests finding usages of an enum type
func TestFindTypeUsages_EnumType(t *testing.T) {
	index := NewProtoIndex(testLogger())
	tempDir := t.TempDir()

	// Create proto with enum
	protoFile := filepath.Join(tempDir, "test.proto")
	content := `syntax = "proto3";

package test.v1;

enum Status {
	STATUS_UNSPECIFIED = 0;
	STATUS_ACTIVE = 1;
	STATUS_INACTIVE = 2;
}

message User {
	int64 id = 1;
	string name = 2;
	Status status = 3;
}

message GetUserRequest {
	int64 id = 1;
}

service UserService {
	rpc GetUser(GetUserRequest) returns (User);
}
`
	if err := os.WriteFile(protoFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write proto: %v", err)
	}

	if _, err := index.IndexDirectory(tempDir); err != nil {
		t.Fatalf("Failed to index directory: %v", err)
	}

	// Find usages of Status enum
	usages, err := index.FindTypeUsages("Status")
	if err != nil {
		t.Fatalf("FindTypeUsages() error = %v", err)
	}

	if len(usages) != 1 {
		t.Errorf("Expected 1 usage of Status enum, got %d", len(usages))
	}

	if len(usages) > 0 {
		usage := usages[0]
		if usage.ServiceName != "UserService" {
			t.Errorf("ServiceName = %s, want UserService", usage.ServiceName)
		}

		expectedPath := []string{"status"}
		if len(usage.FieldPath) != len(expectedPath) {
			t.Errorf("FieldPath length = %d, want %d", len(usage.FieldPath), len(expectedPath))
		}
	}
}
