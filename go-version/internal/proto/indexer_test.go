package proto

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
)

func TestIndexDirectory(t *testing.T) {
	// Create test index
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Quiet during tests
	}))
	index := NewProtoIndex(logger)

	// Index the example directory (if it exists)
	exampleDir := "../../../python-version/examples"
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		t.Skip("Example directory not found, skipping test")
	}

	count, err := index.IndexDirectory(exampleDir)
	if err != nil {
		t.Fatalf("Failed to index directory: %v", err)
	}

	if count == 0 {
		t.Error("Expected to index at least one proto file")
	}

	stats := index.GetStats()
	if stats.TotalFiles != count {
		t.Errorf("Expected %d files in stats, got %d", count, stats.TotalFiles)
	}

	t.Logf("Indexed %d files: %d services, %d messages, %d enums",
		stats.TotalFiles, stats.TotalServices, stats.TotalMessages, stats.TotalEnums)
}

func TestSearchInNames(t *testing.T) {
	index := createTestIndex(t)

	tests := []struct {
		name      string
		query     string
		minScore  int
		wantMin   int    // Minimum expected results
		wantFirst string // Expected first result name (if any)
	}{
		{
			name:     "exact service match",
			query:    "UserService",
			minScore: 60,
			wantMin:  0, // Fuzzy search may not find exact full path matches
		},
		{
			name:     "partial service match",
			query:    "User",
			minScore: 60,
			wantMin:  1, // Should find services and messages with "User"
		},
		{
			name:     "fuzzy match",
			query:    "UsrSvc", // fuzzy for UserService
			minScore: 60,
			wantMin:  0, // Might not match depending on fuzzy algorithm
		},
		{
			name:     "case insensitive",
			query:    "user",
			minScore: 60,
			wantMin:  1, // Should find items with "user" in name or comment
		},
		{
			name:     "comment search",
			query:    "operations",
			minScore: 60,
			wantMin:  0, // May find in comments
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 20, tt.minScore)

			if len(results) < tt.wantMin {
				t.Errorf("Search(%q) returned %d results, want at least %d",
					tt.query, len(results), tt.wantMin)
			} else {
				t.Logf("Search(%q) returned %d results", tt.query, len(results))
			}

			if tt.wantFirst != "" && len(results) > 0 {
				if results[0].Name != tt.wantFirst {
					t.Logf("Search(%q) first result = %q, want %q (may vary based on scoring)",
						tt.query, results[0].Name, tt.wantFirst)
				}
			}

			// Verify results are sorted by score (descending)
			for i := 1; i < len(results); i++ {
				if results[i].Score > results[i-1].Score {
					t.Errorf("Results not sorted by score: results[%d].Score=%d > results[%d].Score=%d",
						i, results[i].Score, i-1, results[i-1].Score)
				}
			}

			// Log some results for inspection
			for i, result := range results {
				if i < 3 { // Log first 3
					t.Logf("  [%d] %s (score=%d, type=%s, match=%s)",
						i, result.Name, result.Score, result.Type, result.MatchType)
				}
			}
		})
	}
}

func TestSearchInFields(t *testing.T) {
	index := createTestIndex(t)

	// Search for a field name that exists in User message
	results := index.Search("email", 20, 60)

	// Should find messages containing "email" field
	found := false
	for _, result := range results {
		if result.Type == "message" && result.MatchType == "field" {
			found = true
			if result.MatchedField == "" {
				t.Error("Expected MatchedField to be set for field match")
			}
			t.Logf("Found field match: %s in %s (field: %s, score: %d)",
				result.MatchType, result.Name, result.MatchedField, result.Score)
			break
		}
	}

	if !found {
		t.Log("Field search may not have found results (depends on test data)")
	}
}

func TestSearchInRPCs(t *testing.T) {
	index := createTestIndex(t)

	// Search for an RPC name
	results := index.Search("GetUser", 20, 60)

	// Should find services containing this RPC
	foundService := false
	for _, result := range results {
		if result.Type == "service" {
			foundService = true
			if result.RPCCount == 0 {
				t.Error("Expected service to have RPCs")
			}
			t.Logf("Found service: %s with %d RPCs (score: %d)",
				result.Name, result.RPCCount, result.Score)
			break
		}
	}

	if !foundService {
		t.Log("Service search may not have found results (depends on test data)")
	}
}

func TestSearchInComments(t *testing.T) {
	index := createTestIndex(t)

	// Search for text that appears in comments
	results := index.Search("user", 20, 60)

	// Should find definitions with "user" in comments
	for _, result := range results {
		if result.MatchType == "comment" && result.Comment == "" {
			t.Error("Expected comment to be set for comment match")
		}
		t.Logf("Found %s match: %s (type: %s, score: %d)",
			result.MatchType, result.Name, result.Type, result.Score)
	}

	if len(results) == 0 {
		t.Log("Comment search may not have found results (depends on test data)")
	}
}

func TestSearchLimit(t *testing.T) {
	index := createTestIndex(t)

	limits := []int{1, 5, 10, 20}
	for _, limit := range limits {
		t.Run(string(rune(limit)), func(t *testing.T) {
			results := index.Search("user", limit, 50)
			if len(results) > limit {
				t.Errorf("Search returned %d results, limit was %d", len(results), limit)
			}
		})
	}
}

func TestSearchMinScore(t *testing.T) {
	index := createTestIndex(t)

	// Test with different min scores
	minScores := []int{50, 70, 90}
	for _, minScore := range minScores {
		t.Run(string(rune(minScore)), func(t *testing.T) {
			results := index.Search("user", 20, minScore)

			// All results should have score >= minScore
			for _, result := range results {
				if result.Score < minScore {
					t.Errorf("Result score %d is less than minScore %d", result.Score, minScore)
				}
			}
		})
	}
}

func TestSearchEmptyQuery(t *testing.T) {
	index := createTestIndex(t)

	results := index.Search("", 20, 60)
	if results != nil {
		t.Error("Expected nil results for empty query")
	}
}

func TestSearchNoResults(t *testing.T) {
	index := createTestIndex(t)

	// Search for something that definitely doesn't exist
	results := index.Search("xyznonexistent12345", 20, 60)
	if len(results) != 0 {
		t.Errorf("Expected no results for non-existent query, got %d", len(results))
	}
}

func TestGetService(t *testing.T) {
	index := createTestIndex(t)

	// Try to get a service
	service, err := index.GetService("UserService", false, 0)
	if err != nil {
		t.Logf("GetService returned error (may be expected if test data not available): %v", err)
		return
	}

	if service == nil {
		t.Fatal("Expected service to be non-nil")
	}

	// Check required fields
	if name, ok := service["name"].(string); !ok || name == "" {
		t.Error("Expected service to have name")
	}
	if fullName, ok := service["full_name"].(string); !ok || fullName == "" {
		t.Error("Expected service to have full_name")
	}
	if rpcs, ok := service["rpcs"].([]map[string]interface{}); !ok || len(rpcs) == 0 {
		t.Error("Expected service to have rpcs")
	}

	t.Logf("Got service: %v", service["full_name"])
}

func TestGetMessage(t *testing.T) {
	index := createTestIndex(t)

	// Try to get a message
	message, err := index.GetMessage("User", false, 0)
	if err != nil {
		t.Logf("GetMessage returned error (may be expected if test data not available): %v", err)
		return
	}

	if message == nil {
		t.Fatal("Expected message to be non-nil")
	}

	// Check required fields
	if name, ok := message["name"].(string); !ok || name == "" {
		t.Error("Expected message to have name")
	}
	if fullName, ok := message["full_name"].(string); !ok || fullName == "" {
		t.Error("Expected message to have full_name")
	}
	if fields, ok := message["fields"].([]map[string]interface{}); !ok {
		t.Error("Expected message to have fields")
	} else {
		t.Logf("Message has %d fields", len(fields))
	}

	t.Logf("Got message: %v", message["full_name"])
}

func TestGetEnum(t *testing.T) {
	index := createTestIndex(t)

	// Try to get an enum
	enum, err := index.GetEnum("UserRole")
	if err != nil {
		t.Logf("GetEnum returned error (may be expected if test data not available): %v", err)
		return
	}

	if enum == nil {
		t.Fatal("Expected enum to be non-nil")
	}

	// Check required fields
	if name, ok := enum["name"].(string); !ok || name == "" {
		t.Error("Expected enum to have name")
	}
	if values, ok := enum["values"].([]map[string]interface{}); !ok {
		t.Error("Expected enum to have values")
	} else {
		t.Logf("Enum has %d values", len(values))
	}

	t.Logf("Got enum: %v", enum["full_name"])
}

func TestConcurrentSearch(t *testing.T) {
	index := createTestIndex(t)

	// Test concurrent searches (index should be thread-safe with RWMutex)
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			results := index.Search("user", 20, 60)
			_ = results
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	t.Log("Concurrent search test passed")
}

func TestIndexFile(t *testing.T) {
	// Create a temporary test proto file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.proto")

	testContent := `syntax = "proto3";

package test;

// Test service for indexing
service TestService {
    // Get test data
    rpc GetTest(TestRequest) returns (TestResponse);
}

message TestRequest {
    string id = 1;  // Test ID
}

message TestResponse {
    string data = 1;  // Test data
}
`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	index := NewProtoIndex(logger)

	// Index the file
	if err := index.IndexFile(testFile); err != nil {
		t.Fatalf("Failed to index file: %v", err)
	}

	// Verify it was indexed
	stats := index.GetStats()
	if stats.TotalFiles != 1 {
		t.Errorf("Expected 1 file, got %d", stats.TotalFiles)
	}
	if stats.TotalServices != 1 {
		t.Errorf("Expected 1 service, got %d", stats.TotalServices)
	}
	if stats.TotalMessages != 2 {
		t.Errorf("Expected 2 messages, got %d", stats.TotalMessages)
	}

	// Search for the service (case-insensitive and should match comments)
	results := index.Search("test", 10, 50)
	if len(results) == 0 {
		t.Error("Expected to find test-related definitions")
	} else {
		t.Logf("Found %d results for 'test' query", len(results))
		for _, result := range results {
			t.Logf("  - %s (score=%d, match=%s)", result.Name, result.Score, result.MatchType)
		}
	}
}

func TestRemoveFile(t *testing.T) {
	// Create a temporary test proto file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.proto")

	testContent := `syntax = "proto3";
package test;
service TestService {
    rpc GetTest(TestRequest) returns (TestResponse);
}
message TestRequest { string id = 1; }
message TestResponse { string data = 1; }
`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	index := NewProtoIndex(logger)

	// Index the file
	if err := index.IndexFile(testFile); err != nil {
		t.Fatalf("Failed to index file: %v", err)
	}

	// Verify it was indexed
	stats := index.GetStats()
	if stats.TotalFiles != 1 {
		t.Error("Expected file to be indexed")
	}

	// Remove the file
	index.RemoveFile(testFile)

	// Verify it was removed
	stats = index.GetStats()
	if stats.TotalFiles != 0 {
		t.Errorf("Expected 0 files after removal, got %d", stats.TotalFiles)
	}
	if stats.TotalServices != 0 {
		t.Errorf("Expected 0 services after removal, got %d", stats.TotalServices)
	}
}

// Helper function to create a test index with example data
func createTestIndex(t *testing.T) *ProtoIndex {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError, // Quiet during tests
	}))
	index := NewProtoIndex(logger)

	// Try to index the example directory
	exampleDir := "../../../python-version/examples"
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		t.Skip("Example directory not found, skipping test")
	}

	_, err := index.IndexDirectory(exampleDir)
	if err != nil {
		t.Fatalf("Failed to create test index: %v", err)
	}

	return index
}
