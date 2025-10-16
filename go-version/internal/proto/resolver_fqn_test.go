package proto

import (
	"os"
	"path/filepath"
	"testing"
)

// TestResolveFullyQualifiedTypes tests that the resolver correctly handles
// fully qualified type names and resolves to the correct message when multiple
// messages with the same simple name exist
func TestResolveFullyQualifiedTypes(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create two different Price messages in different packages
	// This simulates the real issue where there are multiple Price messages
	priceV1 := &ProtoMessage{
		Name:     "Price",
		FullName: "acme.dto.payments.checkout_orchestrator.v1beta1.Price",
		Fields: []ProtoField{
			{Name: "amount", Type: "string", Number: 1},
			{Name: "currency_code", Type: "string", Number: 2},
		},
	}

	priceV2 := &ProtoMessage{
		Name:     "Price",
		FullName: "com.payments.common.Price",
		Fields: []ProtoField{
			{Name: "iso_currency_code", Type: "string", Number: 1},
			{Name: "amount", Type: "uint64", Number: 2},
			{Name: "precision", Type: "uint64", Number: 3},
			{Name: "iso_country_code", Type: "string", Number: 4},
		},
	}

	productRef := &ProtoMessage{
		Name:     "ProductReference",
		FullName: "acme.dto.payments.checkout_orchestrator.v1beta1.ProductReference",
		Fields: []ProtoField{
			{Name: "product_id", Type: "int64", Number: 1},
			{Name: "product_type", Type: "string", Number: 2},
		},
	}

	// TaxableLine uses the fully qualified type name
	taxableLine := &ProtoMessage{
		Name:     "TaxableLine",
		FullName: "acme.rpc.payments.checkout_orchestrator.info.v1beta1.TaxableLine",
		Fields: []ProtoField{
			{Name: "product_reference", Type: "acme.dto.payments.checkout_orchestrator.v1beta1.ProductReference", Number: 1},
			{Name: "unit_net_price", Type: "acme.dto.payments.checkout_orchestrator.v1beta1.Price", Number: 2},
			{Name: "quantity", Type: "int64", Number: 3},
		},
	}

	// Index all messages
	index.messages["acme.dto.payments.checkout_orchestrator.v1beta1.Price"] = priceV1
	index.messages["com.payments.common.Price"] = priceV2
	index.messages["acme.dto.payments.checkout_orchestrator.v1beta1.ProductReference"] = productRef
	index.messages["acme.rpc.payments.checkout_orchestrator.info.v1beta1.TaxableLine"] = taxableLine

	// Resolve types for TaxableLine
	resolved := index.resolveMessageTypes(taxableLine, 10, nil)

	// Should resolve ProductReference and the correct Price (v1beta1, not common)
	if len(resolved) != 2 {
		t.Errorf("resolveMessageTypes() resolved %d types, want 2", len(resolved))
		for k := range resolved {
			t.Logf("  Resolved: %s", k)
		}
	}

	// Check ProductReference is resolved
	productRefKey := "acme.dto.payments.checkout_orchestrator.v1beta1.ProductReference"
	if _, ok := resolved[productRefKey]; !ok {
		t.Errorf("resolveMessageTypes() did not resolve %s", productRefKey)
	}

	// Check the correct Price is resolved (v1beta1, not common)
	priceKey := "acme.dto.payments.checkout_orchestrator.v1beta1.Price"
	priceResolved, ok := resolved[priceKey]
	if !ok {
		t.Errorf("resolveMessageTypes() did not resolve %s", priceKey)
		t.Logf("Available keys:")
		for k := range resolved {
			t.Logf("  %s", k)
		}
		t.FailNow()
	}

	// Verify it's the correct Price message with amount (string) and currency_code
	priceMap, ok := priceResolved.(map[string]interface{})
	if !ok {
		t.Fatal("Price is not a map[string]interface{}")
	}

	if fullName := priceMap["full_name"]; fullName != "acme.dto.payments.checkout_orchestrator.v1beta1.Price" {
		t.Errorf("Price full_name = %v, want 'acme.dto.payments.checkout_orchestrator.v1beta1.Price'", fullName)
	}

	// Check fields - should have amount (string) and currency_code, NOT iso_currency_code, precision, etc.
	fields, ok := priceMap["fields"].([]map[string]interface{})
	if !ok {
		t.Fatal("Price fields is not a []map[string]interface{}")
	}

	if len(fields) != 2 {
		t.Errorf("Price has %d fields, want 2 (amount, currency_code)", len(fields))
	}

	fieldNames := make(map[string]string)
	for _, field := range fields {
		name := field["name"].(string)
		fieldType := field["type"].(string)
		fieldNames[name] = fieldType
	}

	// Verify correct fields
	if fieldType, ok := fieldNames["amount"]; !ok {
		t.Error("Price is missing 'amount' field")
	} else if fieldType != "string" {
		t.Errorf("Price.amount type = %s, want 'string'", fieldType)
	}

	if _, ok := fieldNames["currency_code"]; !ok {
		t.Error("Price is missing 'currency_code' field")
	}

	// Verify wrong fields are NOT present
	if _, ok := fieldNames["iso_currency_code"]; ok {
		t.Error("Price should NOT have 'iso_currency_code' field (wrong Price message resolved)")
	}

	if _, ok := fieldNames["precision"]; ok {
		t.Error("Price should NOT have 'precision' field (wrong Price message resolved)")
	}

	if _, ok := fieldNames["iso_country_code"]; ok {
		t.Error("Price should NOT have 'iso_country_code' field (wrong Price message resolved)")
	}
}

// TestResolveFullyQualifiedTypesWithParser tests the full end-to-end flow
// from parsing a proto file with fully qualified types to resolving them
func TestResolveFullyQualifiedTypesWithParser(t *testing.T) {
	index := NewProtoIndex(testLogger())

	// Create temporary proto files
	tempDir := t.TempDir()

	// Create the correct Price message
	priceProto := filepath.Join(tempDir, "price_v1.proto")
	priceContent := `syntax = "proto3";

package acme.dto.payments.checkout_orchestrator.v1beta1;

message Price {
	string amount = 1;
	string currency_code = 2;
}

message ProductReference {
	int64 product_id = 1;
	string product_type = 2;
}
`
	if err := os.WriteFile(priceProto, []byte(priceContent), 0644); err != nil {
		t.Fatalf("Failed to write price proto file: %v", err)
	}

	// Create the wrong Price message (different package)
	wrongPriceProto := filepath.Join(tempDir, "price_common.proto")
	wrongPriceContent := `syntax = "proto3";

package com.payments.common;

message Price {
	string iso_currency_code = 1;
	uint64 amount = 2;
	uint64 precision = 3;
	string iso_country_code = 4;
}
`
	if err := os.WriteFile(wrongPriceProto, []byte(wrongPriceContent), 0644); err != nil {
		t.Fatalf("Failed to write wrong price proto file: %v", err)
	}

	// Create TaxableLine that references the correct Price with fully qualified name
	taxableProto := filepath.Join(tempDir, "taxable.proto")
	taxableContent := `syntax = "proto3";

package acme.rpc.payments.checkout_orchestrator.info.v1beta1;

message TaxableLine {
	acme.dto.payments.checkout_orchestrator.v1beta1.ProductReference product_reference = 1;
	acme.dto.payments.checkout_orchestrator.v1beta1.Price unit_net_price = 2;
	int64 quantity = 3;
}
`
	if err := os.WriteFile(taxableProto, []byte(taxableContent), 0644); err != nil {
		t.Fatalf("Failed to write taxable proto file: %v", err)
	}

	// Index all proto files
	count, err := index.IndexDirectory(tempDir)
	if err != nil {
		t.Fatalf("IndexDirectory() error = %v", err)
	}
	if count != 3 {
		t.Logf("Warning: indexed %d files, expected 3", count)
	}

	// Get TaxableLine message with type resolution
	result, err := index.GetMessage("TaxableLine", true, 10)
	if err != nil {
		t.Fatalf("GetMessage() error = %v", err)
	}

	resolvedTypes, ok := result["resolved_types"]
	if !ok {
		t.Fatal("GetMessage() should have resolved_types")
	}

	resolvedMap := resolvedTypes.(map[string]interface{})

	// Check that the correct Price is resolved
	priceKey := "acme.dto.payments.checkout_orchestrator.v1beta1.Price"
	priceResolved, ok := resolvedMap[priceKey]
	if !ok {
		t.Errorf("GetMessage() did not resolve %s", priceKey)
		t.Logf("Available keys:")
		for k := range resolvedMap {
			t.Logf("  %s", k)
		}
		t.FailNow()
	}

	priceMap := priceResolved.(map[string]interface{})

	// Verify it's the correct Price
	if fullName := priceMap["full_name"]; fullName != "acme.dto.payments.checkout_orchestrator.v1beta1.Price" {
		t.Errorf("Price full_name = %v, want 'acme.dto.payments.checkout_orchestrator.v1beta1.Price'", fullName)
	}

	fields := priceMap["fields"].([]map[string]interface{})
	if len(fields) != 2 {
		t.Errorf("Price has %d fields, want 2", len(fields))
	}

	// Verify we got the correct Price with string amount and currency_code
	fieldNames := make(map[string]string)
	for _, field := range fields {
		name := field["name"].(string)
		fieldType := field["type"].(string)
		fieldNames[name] = fieldType
	}

	if fieldType, ok := fieldNames["amount"]; !ok || fieldType != "string" {
		t.Errorf("Price.amount type = %s, want 'string'", fieldType)
	}

	if _, ok := fieldNames["currency_code"]; !ok {
		t.Error("Price is missing 'currency_code' field")
	}

	// Verify we didn't get the wrong Price
	if _, ok := fieldNames["iso_currency_code"]; ok {
		t.Error("Resolved wrong Price message (has iso_currency_code field)")
	}
}
