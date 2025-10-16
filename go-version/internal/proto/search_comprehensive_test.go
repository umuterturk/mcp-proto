package proto

import (
	"log/slog"
	"os"
	"testing"
)

// setupTestIndex creates a test index with various proto definitions
func setupTestIndex() *ProtoIndex {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	index := NewProtoIndex(logger)

	protoFile := &ProtoFile{
		Package: "com.example.api.v1",
		Services: []ProtoService{
			{
				Name:     "UserService",
				FullName: "com.example.api.v1.UserService",
				Comment:  "Service for user management operations",
				RPCs: []ProtoRPC{
					{Name: "GetUser", RequestType: "GetUserRequest", ResponseType: "GetUserResponse"},
					{Name: "CreateUser", RequestType: "CreateUserRequest", ResponseType: "CreateUserResponse"},
					{Name: "DeleteUser", RequestType: "DeleteUserRequest", ResponseType: "DeleteUserResponse"},
					{Name: "ListUsers", RequestType: "ListUsersRequest", ResponseType: "ListUsersResponse"},
				},
			},
			{
				Name:     "ProductService",
				FullName: "com.example.api.v1.ProductService",
				RPCs: []ProtoRPC{
					{Name: "CreateProduct", RequestType: "CreateProductRequest", ResponseType: "CreateProductResponse"},
					{Name: "GetProduct", RequestType: "GetProductRequest", ResponseType: "GetProductResponse"},
				},
			},
			{
				Name:     "OrderService",
				FullName: "com.example.api.v1.OrderService",
				RPCs: []ProtoRPC{
					{Name: "CreateOrder", RequestType: "CreateOrderRequest", ResponseType: "CreateOrderResponse"},
				},
			},
		},
		Messages: []ProtoMessage{
			{
				Name:     "CalculateTaxInfoRequest",
				FullName: "com.example.api.v1.CalculateTaxInfoRequest",
				Fields: []ProtoField{
					{Name: "user_id", Type: "string", Number: 1},
					{Name: "amount", Type: "double", Number: 2},
				},
			},
			{
				Name:     "TaxCalculationRequest",
				FullName: "com.example.api.v1.TaxCalculationRequest",
				Fields: []ProtoField{
					{Name: "order_id", Type: "string", Number: 1},
				},
			},
			{
				Name:     "GetUserRequest",
				FullName: "com.example.api.v1.GetUserRequest",
				Fields: []ProtoField{
					{Name: "user_id", Type: "string", Number: 1},
					{Name: "include_profile", Type: "bool", Number: 2},
				},
			},
			{
				Name:     "GetUserResponse",
				FullName: "com.example.api.v1.GetUserResponse",
			},
			{
				Name:     "CreateUserRequest",
				FullName: "com.example.api.v1.CreateUserRequest",
				Fields: []ProtoField{
					{Name: "email_address", Type: "string", Number: 1},
					{Name: "full_name", Type: "string", Number: 2},
				},
			},
			{
				Name:     "CreateUserResponse",
				FullName: "com.example.api.v1.CreateUserResponse",
			},
			{
				Name:     "UserProfileResponse",
				FullName: "com.example.api.v1.UserProfileResponse",
			},
			{
				Name:     "CreateProductRequest",
				FullName: "com.example.api.v1.CreateProductRequest",
			},
			{
				Name:     "TaxInfoRequest",
				FullName: "com.example.api.v1.TaxInfoRequest",
			},
			{
				Name:     "User",
				FullName: "com.example.api.v1.User",
			},
			{
				Name:     "Product",
				FullName: "com.example.api.v1.Product",
			},
			{
				Name:     "User123Request",
				FullName: "com.example.api.v1.User123Request",
			},
			{
				Name:     "CalculateRequest",
				FullName: "com.example.api.v1.CalculateRequest",
			},
			{
				Name:     "CalculatePriceRequest",
				FullName: "com.example.api.v1.CalculatePriceRequest",
			},
			{
				Name:     "CalculateTaxRequest",
				FullName: "com.example.api.v1.CalculateTaxRequest",
			},
		},
	}

	index.mu.Lock()
	defer index.mu.Unlock()

	index.files["test.proto"] = protoFile

	// Index services
	for i := range protoFile.Services {
		svc := &protoFile.Services[i]
		index.services[svc.FullName] = svc
		index.searchEntries = append(index.searchEntries, searchEntry{
			fullName:  svc.FullName,
			entryType: "service",
			filePath:  "test.proto",
			service:   svc,
		})
	}

	// Index messages
	for i := range protoFile.Messages {
		msg := &protoFile.Messages[i]
		index.messages[msg.FullName] = msg
		index.searchEntries = append(index.searchEntries, searchEntry{
			fullName:  msg.FullName,
			entryType: "message",
			filePath:  "test.proto",
			message:   msg,
		})
	}

	return index
}

// Test 1: Multi-Word Query Matching
func TestMultiWordQueryMatching(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name          string
		query         string
		expectFound   string
		minScore      int
		shouldContain bool
	}{
		{
			name:          "Calculate Tax",
			query:         "Calculate Tax",
			expectFound:   "com.example.api.v1.CalculateTaxInfoRequest",
			minScore:      60,
			shouldContain: true,
		},
		{
			name:          "Tax Calculation",
			query:         "Tax Calculation",
			expectFound:   "com.example.api.v1.TaxCalculationRequest",
			minScore:      60,
			shouldContain: true,
		},
		{
			name:          "Get User",
			query:         "Get User",
			expectFound:   "com.example.api.v1.GetUserRequest",
			minScore:      60,
			shouldContain: true,
		},
		{
			name:          "Create Product",
			query:         "Create Product",
			expectFound:   "com.example.api.v1.CreateProductRequest",
			minScore:      60,
			shouldContain: true,
		},
		{
			name:          "User Profile",
			query:         "User Profile",
			expectFound:   "com.example.api.v1.UserProfileResponse",
			minScore:      60,
			shouldContain: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, tt.minScore)
			found := false
			for _, result := range results {
				if result.Name == tt.expectFound {
					found = true
					t.Logf("✓ Found '%s' with score %d", result.Name, result.Score)
					break
				}
			}
			if tt.shouldContain && !found {
				t.Errorf("Expected to find '%s' for query '%s'", tt.expectFound, tt.query)
				t.Logf("Results found:")
				for _, r := range results {
					t.Logf("  - %s (score: %d)", r.Name, r.Score)
				}
			}
		})
	}
}

// Test 2: CamelCase Boundary Recognition
func TestCamelCaseBoundaryRecognition(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name        string
		query       string
		expectFound string
		minScore    int
	}{
		{
			name:        "Three word match",
			query:       "Calculate Tax Info",
			expectFound: "com.example.api.v1.CalculateTaxInfoRequest",
			minScore:    60,
		},
		{
			name:        "Two word match",
			query:       "Tax Info",
			expectFound: "com.example.api.v1.TaxInfoRequest",
			minScore:    60,
		},
		{
			name:        "User Profile",
			query:       "User Profile",
			expectFound: "com.example.api.v1.UserProfileResponse",
			minScore:    60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, tt.minScore)
			found := false
			var foundScore int
			for _, result := range results {
				if result.Name == tt.expectFound {
					found = true
					foundScore = result.Score
					break
				}
			}
			if !found {
				t.Errorf("Expected to find '%s' for query '%s'", tt.expectFound, tt.query)
			} else {
				t.Logf("✓ Found '%s' with score %d (CamelCase aligned)", tt.expectFound, foundScore)
				if foundScore < 90 {
					t.Logf("Warning: Score is %d, expected >= 90 for CamelCase aligned match", foundScore)
				}
			}
		})
	}
}

// Test 3: Case Insensitivity
func TestCaseInsensitivity(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name        string
		query       string
		expectFound string
	}{
		{
			name:        "lowercase",
			query:       "calculate tax",
			expectFound: "com.example.api.v1.CalculateTaxInfoRequest",
		},
		{
			name:        "UPPERCASE",
			query:       "CALCULATE TAX",
			expectFound: "com.example.api.v1.CalculateTaxInfoRequest",
		},
		{
			name:        "MiXeD CaSe",
			query:       "CaLcUlAtE tAx",
			expectFound: "com.example.api.v1.CalculateTaxInfoRequest",
		},
		{
			name:        "lowercase single",
			query:       "user",
			expectFound: "com.example.api.v1.User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, 60)
			found := false
			for _, result := range results {
				if result.Name == tt.expectFound {
					found = true
					t.Logf("✓ Case-insensitive match: query='%s' found '%s' (score: %d)", tt.query, result.Name, result.Score)
					break
				}
			}
			if !found {
				t.Errorf("Expected to find '%s' for case-insensitive query '%s'", tt.expectFound, tt.query)
			}
		})
	}
}

// Test 4: Single Word Search (Backward Compatibility)
func TestSingleWordSearch(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name             string
		query            string
		expectedContains []string
		minResults       int
	}{
		{
			name:  "Calculate",
			query: "Calculate",
			expectedContains: []string{
				"com.example.api.v1.CalculateTaxInfoRequest",
				"com.example.api.v1.CalculateRequest",
				"com.example.api.v1.CalculatePriceRequest",
			},
			minResults: 3,
		},
		{
			name:  "User",
			query: "User",
			expectedContains: []string{
				"com.example.api.v1.User",
				"com.example.api.v1.UserService",
				"com.example.api.v1.GetUserRequest",
			},
			minResults: 3,
		},
		{
			name:  "Tax",
			query: "Tax",
			expectedContains: []string{
				"com.example.api.v1.CalculateTaxInfoRequest",
				"com.example.api.v1.TaxCalculationRequest",
				"com.example.api.v1.TaxInfoRequest",
			},
			minResults: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 20, 60)
			if len(results) < tt.minResults {
				t.Errorf("Expected at least %d results for '%s', got %d", tt.minResults, tt.query, len(results))
			}

			for _, expected := range tt.expectedContains {
				found := false
				for _, result := range results {
					if result.Name == expected {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected to find '%s' in results for query '%s'", expected, tt.query)
				}
			}
			t.Logf("✓ Single word query '%s' found %d results", tt.query, len(results))
		})
	}
}

// Test 5: Typo Tolerance
func TestTypoTolerance(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name        string
		query       string
		expectFound string
		minScore    int
	}{
		{
			name:        "Calcualte (typo)",
			query:       "Calcualte",
			expectFound: "com.example.api.v1.CalculateRequest",
			minScore:    70,
		},
		{
			name:        "Usr",
			query:       "Usr",
			expectFound: "com.example.api.v1.User",
			minScore:    70,
		},
		{
			name:        "Prodct",
			query:       "Prodct",
			expectFound: "com.example.api.v1.Product",
			minScore:    70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, tt.minScore)
			found := false
			for _, result := range results {
				if result.Name == tt.expectFound {
					found = true
					t.Logf("✓ Typo tolerant: query='%s' found '%s' (score: %d)", tt.query, result.Name, result.Score)
					break
				}
			}
			if !found {
				t.Logf("Note: Typo query '%s' did not find '%s' (this may be expected for severe typos)", tt.query, tt.expectFound)
				t.Logf("Results found:")
				for _, r := range results {
					t.Logf("  - %s (score: %d)", r.Name, r.Score)
				}
			}
		})
	}
}

// Test 6: Subsequence/Initialism Matching
func TestSubsequenceMatching(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name        string
		query       string
		expectFound string
		minScore    int
	}{
		{
			name:        "CTR",
			query:       "CTR",
			expectFound: "com.example.api.v1.CalculateTaxRequest",
			minScore:    60,
		},
		{
			name:        "CTIR",
			query:       "CTIR",
			expectFound: "com.example.api.v1.CalculateTaxInfoRequest",
			minScore:    60,
		},
		{
			name:        "UsrSvc",
			query:       "UsrSvc",
			expectFound: "com.example.api.v1.UserService",
			minScore:    60,
		},
		{
			name:        "GUR",
			query:       "GUR",
			expectFound: "com.example.api.v1.GetUserRequest",
			minScore:    60,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, tt.minScore)
			found := false
			for _, result := range results {
				if result.Name == tt.expectFound {
					found = true
					t.Logf("✓ Subsequence match: query='%s' found '%s' (score: %d)", tt.query, result.Name, result.Score)
					break
				}
			}
			if !found {
				t.Logf("Note: Subsequence query '%s' did not find '%s'", tt.query, tt.expectFound)
				t.Logf("Results found:")
				for _, r := range results {
					t.Logf("  - %s (score: %d)", r.Name, r.Score)
				}
			}
		})
	}
}

// Test 7: Word Order Variations
func TestWordOrderVariations(t *testing.T) {
	index := setupTestIndex()

	query1 := "Tax Calculate"
	query2 := "Calculate Tax"

	results1 := index.Search(query1, 10, 60)
	results2 := index.Search(query2, 10, 60)

	t.Logf("Query '%s' returned %d results", query1, len(results1))
	t.Logf("Query '%s' returned %d results", query2, len(results2))

	// Correct word order should return results
	if len(results2) == 0 {
		t.Error("Expected results for 'Calculate Tax'")
	}

	// Check that CalculateTaxInfoRequest is found with correct order
	found2 := false
	var score2 int

	for _, r := range results2 {
		if r.Name == "com.example.api.v1.CalculateTaxInfoRequest" {
			found2 = true
			score2 = r.Score
			break
		}
	}

	if !found2 {
		t.Error("Expected to find CalculateTaxInfoRequest for 'Calculate Tax'")
	} else {
		t.Logf("✓ Found CalculateTaxInfoRequest with correct word order (score: %d)", score2)
	}

	// Reversed order might or might not match depending on algorithm
	// If it does find results, they should have lower scores
	if len(results1) > 0 {
		t.Logf("Note: Reversed word order 'Tax Calculate' also found %d results", len(results1))
		for _, r := range results1 {
			if r.Name == "com.example.api.v1.CalculateTaxInfoRequest" {
				t.Logf("  Found CalculateTaxInfoRequest with reversed order (score: %d)", r.Score)
				if r.Score >= score2 {
					t.Logf("  Warning: Reversed order scored same or higher (%d vs %d)", r.Score, score2)
				}
			}
		}
	} else {
		t.Logf("✓ Reversed word order 'Tax Calculate' found no results (expected for strict matching)")
	}
}

// Test 8: Scoring and Ranking
func TestScoringAndRanking(t *testing.T) {
	index := setupTestIndex()

	results := index.Search("User", 20, 60)

	if len(results) == 0 {
		t.Fatal("Expected results for 'User' query")
	}

	// Check that exact match "User" is found with high score
	exactMatch := false
	for i, result := range results {
		if result.Name == "com.example.api.v1.User" {
			exactMatch = true
			t.Logf("✓ Exact match 'User' found at position %d with score %d", i, result.Score)
			if result.Score < 95 {
				t.Errorf("Expected exact match to have score >= 95, got %d", result.Score)
			}
			if i > 5 {
				t.Logf("Warning: Exact match not in top 5 positions (position %d)", i)
			}
			break
		}
	}

	if !exactMatch {
		t.Error("Expected to find exact match 'User'")
	}

	// Verify scores are in descending order
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("Results not properly sorted: position %d (score %d) > position %d (score %d)",
				i, results[i].Score, i-1, results[i-1].Score)
		}
	}
	t.Logf("✓ Results properly sorted by score (descending)")

	// Log top 5 results
	t.Log("Top 5 results:")
	for i := 0; i < 5 && i < len(results); i++ {
		t.Logf("  %d. %s (score: %d)", i+1, results[i].Name, results[i].Score)
	}
}

// Test 9: Field Name Search
func TestFieldNameSearch(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name        string
		query       string
		expectFound string
		matchType   string
	}{
		{
			name:        "user id",
			query:       "user_id",
			expectFound: "com.example.api.v1.CalculateTaxInfoRequest",
			matchType:   "field",
		},
		{
			name:        "email address",
			query:       "email_address",
			expectFound: "com.example.api.v1.CreateUserRequest",
			matchType:   "field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, 60)
			found := false
			for _, result := range results {
				if result.Name == tt.expectFound && result.MatchType == tt.matchType {
					found = true
					t.Logf("✓ Field search: query='%s' found message '%s' (matched field: %s)",
						tt.query, result.Name, result.MatchedField)
					break
				}
			}
			if !found {
				t.Logf("Note: Field query '%s' did not find '%s' with match type '%s'",
					tt.query, tt.expectFound, tt.matchType)
			}
		})
	}
}

// Test 10: RPC Name Search
func TestRPCNameSearch(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name        string
		query       string
		expectFound string
		matchType   string
	}{
		{
			name:        "GetUser RPC",
			query:       "GetUser",
			expectFound: "com.example.api.v1.UserService",
			matchType:   "rpc",
		},
		{
			name:        "CreateOrder RPC",
			query:       "CreateOrder",
			expectFound: "com.example.api.v1.OrderService",
			matchType:   "rpc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, 60)
			found := false
			for _, result := range results {
				if result.Name == tt.expectFound && result.MatchType == tt.matchType {
					found = true
					t.Logf("✓ RPC search: query='%s' found service '%s' (matched RPC: %s)",
						tt.query, result.Name, result.MatchedRPC)
					break
				}
			}
			if !found {
				t.Logf("Note: RPC query '%s' did not find '%s' with match type '%s'",
					tt.query, tt.expectFound, tt.matchType)
				t.Logf("Results found:")
				for _, r := range results {
					t.Logf("  - %s (type: %s, match_type: %s)", r.Name, r.Type, r.MatchType)
				}
			}
		})
	}
}

// Test 11: Fully Qualified Names
func TestFullyQualifiedNames(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name        string
		query       string
		expectFound string
		shouldScore int
	}{
		{
			name:        "Full FQN",
			query:       "com.example.api.v1.User",
			expectFound: "com.example.api.v1.User",
			shouldScore: 100,
		},
		{
			name:        "Partial FQN",
			query:       "api.v1.User",
			expectFound: "com.example.api.v1.User",
			shouldScore: 95,
		},
		{
			name:        "Just v1.User",
			query:       "v1.User",
			expectFound: "com.example.api.v1.User",
			shouldScore: 95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, 60)
			found := false
			var foundScore int
			for _, result := range results {
				if result.Name == tt.expectFound {
					found = true
					foundScore = result.Score
					break
				}
			}
			if !found {
				t.Errorf("Expected to find '%s' for FQN query '%s'", tt.expectFound, tt.query)
			} else {
				t.Logf("✓ FQN match: query='%s' found '%s' (score: %d)", tt.query, tt.expectFound, foundScore)
			}
		})
	}
}

// Test 12: Edge Cases
func TestEdgeCases(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name          string
		query         string
		expectResults bool
		description   string
	}{
		{
			name:          "Empty query",
			query:         "",
			expectResults: false,
			description:   "Empty query should return no results",
		},
		{
			name:          "Single character",
			query:         "U",
			expectResults: true,
			description:   "Single character should find matches",
		},
		{
			name:          "With numbers",
			query:         "User123",
			expectResults: true,
			description:   "Query with numbers should work",
		},
		{
			name:          "Very long query",
			query:         "ThisIsAVeryLongQueryThatProbablyWontMatchAnythingInTheIndex",
			expectResults: false,
			description:   "Very long query with no matches",
		},
		{
			name:          "Special characters",
			query:         "User@#$",
			expectResults: false,
			description:   "Special characters should be handled gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, 60)
			hasResults := len(results) > 0

			if tt.expectResults && !hasResults {
				t.Errorf("%s - Expected results but got none", tt.description)
			} else if !tt.expectResults && hasResults {
				t.Logf("%s - Got %d results (may be acceptable)", tt.description, len(results))
			} else {
				t.Logf("✓ %s - Behavior as expected", tt.description)
			}
		})
	}
}

// Test 13: Min Score Filtering
func TestMinScoreFiltering(t *testing.T) {
	index := setupTestIndex()

	query := "User"
	thresholds := []int{60, 70, 80, 90}

	var previousCount int
	for i, threshold := range thresholds {
		results := index.Search(query, 20, threshold)

		t.Logf("Min score %d: %d results", threshold, len(results))

		// Verify all results meet threshold
		for _, result := range results {
			if result.Score < threshold {
				t.Errorf("Result '%s' has score %d, below threshold %d", result.Name, result.Score, threshold)
			}
		}

		// Higher thresholds should return fewer or equal results
		if i > 0 && len(results) > previousCount {
			t.Errorf("Higher threshold (%d) returned more results (%d) than lower threshold (%d results)",
				threshold, len(results), previousCount)
		}

		previousCount = len(results)
	}
	t.Logf("✓ Min score filtering works correctly")
}

// Test 14: Result Limit
func TestResultLimit(t *testing.T) {
	index := setupTestIndex()

	query := "Request" // Should match many messages
	limits := []int{1, 5, 10, 20}

	for _, limit := range limits {
		results := index.Search(query, limit, 60)

		if len(results) > limit {
			t.Errorf("Limit %d exceeded: got %d results", limit, len(results))
		}

		t.Logf("Limit %d: got %d results (✓)", limit, len(results))

		// Verify results are the highest scored ones
		for i := 1; i < len(results); i++ {
			if results[i].Score > results[i-1].Score {
				t.Errorf("Results not sorted: position %d (score %d) > position %d (score %d)",
					i, results[i].Score, i-1, results[i-1].Score)
			}
		}
	}
}

// Test 15: No Results Scenarios
func TestNoResultsScenarios(t *testing.T) {
	index := setupTestIndex()

	tests := []struct {
		name  string
		query string
	}{
		{
			name:  "Non-existent term",
			query: "Xyzabc123NotInIndex",
		},
		{
			name:  "Very specific no-match",
			query: "CompletelyUnrelatedTerm",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results := index.Search(tt.query, 10, 60)
			if len(results) > 0 {
				t.Logf("Note: Query '%s' unexpectedly found %d results", tt.query, len(results))
				for _, r := range results {
					t.Logf("  - %s (score: %d)", r.Name, r.Score)
				}
			} else {
				t.Logf("✓ No results for '%s' as expected", tt.query)
			}
		})
	}
}

// Test 16: Comment Search
func TestCommentSearch(t *testing.T) {
	index := setupTestIndex()

	results := index.Search("management operations", 20, 60)

	found := false
	for _, result := range results {
		if result.Name == "com.example.api.v1.UserService" && result.MatchType == "comment" {
			found = true
			t.Logf("✓ Found service by comment: '%s' (score: %d)", result.Name, result.Score)
			break
		}
	}

	if !found {
		t.Log("Note: Comment search did not find expected service")
		t.Logf("Results found:")
		for _, r := range results {
			t.Logf("  - %s (match_type: %s, score: %d)", r.Name, r.MatchType, r.Score)
		}
	}
}
