package proto

import (
	"fmt"
	"io/fs"
	"log/slog"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/lithammer/fuzzysearch/fuzzy"
	sahilfuzzy "github.com/sahilm/fuzzy"
)

// SearchResult represents a search result with metadata
type SearchResult struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	File         string   `json:"file"`
	Score        int      `json:"score"`
	MatchType    string   `json:"match_type"`
	Comment      string   `json:"comment,omitempty"`
	RPCs         []string `json:"rpcs,omitempty"`
	RPCCount     int      `json:"rpc_count,omitempty"`
	Fields       []string `json:"fields,omitempty"`
	FieldCount   int      `json:"field_count,omitempty"`
	Values       []string `json:"values,omitempty"`
	ValueCount   int      `json:"value_count,omitempty"`
	MatchedRPC   string   `json:"matched_rpc,omitempty"`
	MatchedField string   `json:"matched_field,omitempty"`
}

// Stats represents indexing statistics
type Stats struct {
	TotalFiles             int `json:"total_files"`
	TotalServices          int `json:"total_services"`
	TotalMessages          int `json:"total_messages"`
	TotalEnums             int `json:"total_enums"`
	TotalSearchableEntries int `json:"total_searchable_entries"`
}

type searchEntry struct {
	fullName  string
	entryType string
	filePath  string
	service   *ProtoService
	message   *ProtoMessage
	enum      *ProtoEnum
}

// ProtoIndex is an in-memory index of proto files with search capabilities
type ProtoIndex struct {
	mu            sync.RWMutex
	files         map[string]*ProtoFile
	services      map[string]*ProtoService
	messages      map[string]*ProtoMessage
	enums         map[string]*ProtoEnum
	searchEntries []searchEntry
	logger        *slog.Logger
}

// NewProtoIndex creates a new proto index
func NewProtoIndex(logger *slog.Logger) *ProtoIndex {
	if logger == nil {
		logger = slog.Default()
	}
	return &ProtoIndex{
		files:         make(map[string]*ProtoFile),
		services:      make(map[string]*ProtoService),
		messages:      make(map[string]*ProtoMessage),
		enums:         make(map[string]*ProtoEnum),
		searchEntries: make([]searchEntry, 0),
		logger:        logger,
	}
}

// IndexDirectory recursively scans directory for .proto files and indexes them
func (pi *ProtoIndex) IndexDirectory(rootPath string) (int, error) {
	matches, err := filepath.Glob(filepath.Join(rootPath, "**/*.proto"))
	if err != nil {
		return 0, fmt.Errorf("failed to glob proto files: %w", err)
	}

	// Also try direct scan
	count := 0
	err = filepath.Walk(rootPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Ext(path) == ".proto" {
			if err := pi.IndexFile(path); err != nil {
				pi.logger.Error("failed to index file", "path", path, "error", err)
			} else {
				count++
			}
		}
		return nil
	})

	if err != nil {
		return count, fmt.Errorf("failed to walk directory: %w", err)
	}

	pi.logger.Info("indexed proto files", "count", count)

	// Also index matches from glob if any
	for _, match := range matches {
		if err := pi.IndexFile(match); err != nil {
			pi.logger.Error("failed to index file", "path", match, "error", err)
		}
	}

	return count, nil
}

// IndexFile parses and indexes a single proto file
func (pi *ProtoIndex) IndexFile(filePath string) error {
	parser := NewParser()
	protoFile, err := parser.ParseFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	pi.mu.Lock()
	defer pi.mu.Unlock()

	pi.files[filePath] = protoFile

	// Index services
	for i := range protoFile.Services {
		service := &protoFile.Services[i]
		pi.services[service.FullName] = service
		pi.searchEntries = append(pi.searchEntries, searchEntry{
			fullName:  service.FullName,
			entryType: "service",
			filePath:  filePath,
			service:   service,
		})
	}

	// Index messages
	for i := range protoFile.Messages {
		message := &protoFile.Messages[i]
		pi.messages[message.FullName] = message
		pi.searchEntries = append(pi.searchEntries, searchEntry{
			fullName:  message.FullName,
			entryType: "message",
			filePath:  filePath,
			message:   message,
		})
	}

	// Index enums
	for i := range protoFile.Enums {
		enum := &protoFile.Enums[i]
		pi.enums[enum.FullName] = enum
		pi.searchEntries = append(pi.searchEntries, searchEntry{
			fullName:  enum.FullName,
			entryType: "enum",
			filePath:  filePath,
			enum:      enum,
		})
	}

	pi.logger.Debug("indexed file",
		"path", filePath,
		"services", len(protoFile.Services),
		"messages", len(protoFile.Messages),
		"enums", len(protoFile.Enums),
	)

	return nil
}

// RemoveFile removes a file from the index
func (pi *ProtoIndex) RemoveFile(filePath string) {
	pi.mu.Lock()
	defer pi.mu.Unlock()

	protoFile, exists := pi.files[filePath]
	if !exists {
		return
	}

	// Remove services
	for _, service := range protoFile.Services {
		delete(pi.services, service.FullName)
	}

	// Remove messages
	for _, message := range protoFile.Messages {
		delete(pi.messages, message.FullName)
	}

	// Remove enums
	for _, enum := range protoFile.Enums {
		delete(pi.enums, enum.FullName)
	}

	// Remove from search entries
	newEntries := make([]searchEntry, 0, len(pi.searchEntries))
	for _, entry := range pi.searchEntries {
		if entry.filePath != filePath {
			newEntries = append(newEntries, entry)
		}
	}
	pi.searchEntries = newEntries

	delete(pi.files, filePath)
	pi.logger.Debug("removed file from index", "path", filePath)
}

// Search performs fuzzy search across all proto definitions
// Searches in: names, field names, RPC names, and comments
func (pi *ProtoIndex) Search(query string, limit, minScore int) []SearchResult {
	if query == "" {
		return nil
	}

	pi.mu.RLock()
	defer pi.mu.RUnlock()

	var results []SearchResult
	seen := make(map[string]bool)
	queryLower := strings.ToLower(query)

	// 1. Search in definition names (highest priority)
	nameMatches := pi.searchInNames(query, minScore)
	for _, result := range nameMatches {
		if !seen[result.Name] {
			results = append(results, result)
			seen[result.Name] = true
		}
	}

	// 2. Search in field names (for messages)
	if len(results) < limit {
		fieldMatches := pi.searchInFields(queryLower, minScore, seen)
		results = append(results, fieldMatches...)
	}

	// 3. Search in RPC names (for services)
	if len(results) < limit {
		rpcMatches := pi.searchInRPCs(queryLower, minScore, seen)
		results = append(results, rpcMatches...)
	}

	// 4. Search in comments
	if len(results) < limit {
		commentMatches := pi.searchInComments(queryLower, minScore, seen)
		results = append(results, commentMatches...)
	}

	// Sort by score (descending) and limit results
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

// searchInNames performs fuzzy search on definition names
func (pi *ProtoIndex) searchInNames(query string, minScore int) []SearchResult {
	// Build list of searchable names
	names := make([]string, len(pi.searchEntries))
	for i, entry := range pi.searchEntries {
		names[i] = entry.fullName
	}

	queryLower := strings.ToLower(query)
	var results []SearchResult
	seen := make(map[int]bool)

	// Strategy 1: Exact substring matches (case-insensitive) - highest priority
	for i, name := range names {
		nameLower := strings.ToLower(name)

		if idx := strings.Index(nameLower, queryLower); idx >= 0 {
			score := 100

			// Adjust based on position
			if strings.HasSuffix(nameLower, queryLower) {
				score = 100 // Perfect suffix match (simple name)
			} else if idx == 0 {
				score = 98 // Match at beginning
			} else if idx > 0 && nameLower[idx-1] == '.' {
				score = 97 // Match after package separator
			} else {
				score = 95 // Match in middle
			}

			// Adjust for length ratio
			lengthRatio := float64(len(name)) / float64(len(query))
			if lengthRatio > 5.0 {
				score -= 3 // Penalize very long FQNs
			}

			if score >= minScore {
				entry := pi.searchEntries[i]
				result := pi.createSearchResult(entry, score, "name")
				results = append(results, result)
				seen[i] = true
			}
		}
	}

	// Strategy 2: Levenshtein distance for typo tolerance
	// Check each name's simple name (last part after final dot) against query
	for i, name := range names {
		if seen[i] {
			continue
		}

		// Extract simple name (last component)
		simpleName := name
		if lastDot := strings.LastIndex(name, "."); lastDot >= 0 {
			simpleName = name[lastDot+1:]
		}

		simpleNameLower := strings.ToLower(simpleName)

		// Calculate Levenshtein distance
		distance := fuzzy.LevenshteinDistance(queryLower, simpleNameLower)

		// Convert distance to score (0-100)
		// For similar lengths, small distances should score high
		maxLen := len(queryLower)
		if len(simpleNameLower) > maxLen {
			maxLen = len(simpleNameLower)
		}

		if maxLen == 0 {
			continue
		}

		// Score based on how many characters are correct
		similarity := float64(maxLen-distance) / float64(maxLen)
		score := int(similarity * 100)

		// Require high similarity for Levenshtein matches (at least 70%)
		if score >= 70 && score >= minScore {
			entry := pi.searchEntries[i]
			result := pi.createSearchResult(entry, score, "name")
			results = append(results, result)
			seen[i] = true
		}
	}

	// Strategy 3: Subsequence matching with sahilm/fuzzy (like VSCode)
	// This catches cases like "UsrSvc" matching "UserService"
	matches := sahilfuzzy.Find(query, names)

	for _, match := range matches {
		if seen[match.Index] {
			continue
		}

		score := calculateSubsequenceScore(match.Score, len(query), len(match.Str))

		if score >= minScore {
			entry := pi.searchEntries[match.Index]
			result := pi.createSearchResult(entry, score, "name")
			results = append(results, result)
			seen[match.Index] = true
		}
	}

	return results
}

// searchInFields searches for query in message field names
func (pi *ProtoIndex) searchInFields(query string, minScore int, seen map[string]bool) []SearchResult {
	var results []SearchResult
	queryLower := strings.ToLower(query)

	for _, entry := range pi.searchEntries {
		if seen[entry.fullName] || entry.entryType != "message" || entry.message == nil {
			continue
		}

		// Check each field for matches
		var bestScore int
		var bestField string

		for _, field := range entry.message.Fields {
			fieldLower := strings.ToLower(field.Name)

			// Try exact match first
			if fieldLower == queryLower {
				bestScore = 100
				bestField = field.Name
				break
			}

			// Try substring match
			if strings.Contains(fieldLower, queryLower) {
				score := 95
				if score > bestScore {
					bestScore = score
					bestField = field.Name
				}
				continue
			}

			// Try Levenshtein distance for typo tolerance
			distance := fuzzy.LevenshteinDistance(queryLower, fieldLower)
			maxLen := len(queryLower)
			if len(fieldLower) > maxLen {
				maxLen = len(fieldLower)
			}

			if maxLen > 0 {
				similarity := float64(maxLen-distance) / float64(maxLen)
				score := int(similarity * 100)

				if score >= 70 && score > bestScore {
					bestScore = score
					bestField = field.Name
				}
			}
		}

		if bestScore >= minScore && bestField != "" {
			result := pi.createSearchResult(entry, bestScore, "field")
			result.MatchedField = bestField
			results = append(results, result)
			seen[entry.fullName] = true
		}
	}

	return results
}

// searchInRPCs searches for query in service RPC names
func (pi *ProtoIndex) searchInRPCs(query string, minScore int, seen map[string]bool) []SearchResult {
	var results []SearchResult
	queryLower := strings.ToLower(query)

	for _, entry := range pi.searchEntries {
		if seen[entry.fullName] || entry.entryType != "service" || entry.service == nil {
			continue
		}

		// Check each RPC for matches
		var bestScore int
		var bestRPC string

		for _, rpc := range entry.service.RPCs {
			rpcLower := strings.ToLower(rpc.Name)

			// Try exact match first
			if rpcLower == queryLower {
				bestScore = 100
				bestRPC = rpc.Name
				break
			}

			// Try substring match
			if strings.Contains(rpcLower, queryLower) {
				score := 95
				if score > bestScore {
					bestScore = score
					bestRPC = rpc.Name
				}
				continue
			}

			// Try Levenshtein distance for typo tolerance
			distance := fuzzy.LevenshteinDistance(queryLower, rpcLower)
			maxLen := len(queryLower)
			if len(rpcLower) > maxLen {
				maxLen = len(rpcLower)
			}

			if maxLen > 0 {
				similarity := float64(maxLen-distance) / float64(maxLen)
				score := int(similarity * 100)

				if score >= 70 && score > bestScore {
					bestScore = score
					bestRPC = rpc.Name
				}
			}
		}

		if bestScore >= minScore && bestRPC != "" {
			result := pi.createSearchResult(entry, bestScore, "rpc")
			result.MatchedRPC = bestRPC
			results = append(results, result)
			seen[entry.fullName] = true
		}
	}

	return results
}

// searchInComments searches for query in comments
func (pi *ProtoIndex) searchInComments(query string, minScore int, seen map[string]bool) []SearchResult {
	var results []SearchResult

	for _, entry := range pi.searchEntries {
		if seen[entry.fullName] {
			continue
		}

		var comment string
		switch entry.entryType {
		case "service":
			if entry.service != nil {
				comment = entry.service.Comment
			}
		case "message":
			if entry.message != nil {
				comment = entry.message.Comment
			}
		case "enum":
			if entry.enum != nil {
				comment = entry.enum.Comment
			}
		}

		if comment == "" {
			continue
		}

		// Simple substring match for comments (case-insensitive)
		commentLower := strings.ToLower(comment)
		if strings.Contains(commentLower, query) {
			// Score based on position and length
			score := calculateCommentScore(query, commentLower)

			if score >= minScore {
				result := pi.createSearchResult(entry, score, "comment")
				results = append(results, result)
				seen[entry.fullName] = true
			}
		}
	}

	return results
}

// createSearchResult creates a SearchResult from a search entry
func (pi *ProtoIndex) createSearchResult(entry searchEntry, score int, matchType string) SearchResult {
	result := SearchResult{
		Name:      entry.fullName,
		Type:      entry.entryType,
		File:      entry.filePath,
		Score:     score,
		MatchType: matchType,
	}

	// Add type-specific metadata
	switch entry.entryType {
	case "service":
		if entry.service != nil {
			result.RPCCount = len(entry.service.RPCs)
			result.RPCs = make([]string, len(entry.service.RPCs))
			for i, rpc := range entry.service.RPCs {
				result.RPCs[i] = rpc.Name
			}
			result.Comment = entry.service.Comment
		}
	case "message":
		if entry.message != nil {
			result.FieldCount = len(entry.message.Fields)
			result.Fields = make([]string, len(entry.message.Fields))
			for i, field := range entry.message.Fields {
				result.Fields[i] = field.Name
			}
			result.Comment = entry.message.Comment
		}
	case "enum":
		if entry.enum != nil {
			result.ValueCount = len(entry.enum.Values)
			result.Values = make([]string, len(entry.enum.Values))
			for i, value := range entry.enum.Values {
				result.Values[i] = value.Name
			}
			result.Comment = entry.enum.Comment
		}
	}

	return result
}

// calculateSubsequenceScore converts sahilm/fuzzy library score to 0-100 scale
// sahilm/fuzzy: lower score = better match, but scores can be very large for long strings with gaps
// we want: higher score = better match (100 = exact)
func calculateSubsequenceScore(fuzzyScore, queryLen, targetLen int) int {
	// For exact matches
	if fuzzyScore == 0 {
		return 100
	}

	// The fuzzy library gives very large scores for distant matches.
	// We need a better approach based on the characteristics of the match.

	// Calculate a score based on the density of the match
	// Lower fuzzy scores relative to target length indicate better matches

	// Base score calculation:
	// Good matches have low fuzzyScore relative to targetLen
	// The score represents penalties for gaps and distance

	// Normalize the fuzzy score by target length to get a penalty ratio
	penaltyRatio := float64(fuzzyScore) / float64(targetLen)

	// Convert penalty ratio to a score (0-100)
	// penaltyRatio < 1.0 = very good match (95-100)
	// penaltyRatio 1-10 = good match (80-95)
	// penaltyRatio 10-100 = moderate match (60-80)
	// penaltyRatio > 100 = poor match (< 60)

	var baseScore int
	if penaltyRatio < 1.0 {
		baseScore = 95 + int((1.0-penaltyRatio)*5.0)
	} else if penaltyRatio < 10.0 {
		baseScore = 80 + int((10.0-penaltyRatio)*1.5)
	} else if penaltyRatio < 100.0 {
		baseScore = 60 + int((100.0-penaltyRatio)*0.2)
	} else {
		baseScore = int(60.0 * (1000.0 / (penaltyRatio + 900.0)))
	}

	// Bonus for targets close in length to query (more precise match)
	lengthRatio := float64(targetLen) / float64(queryLen)
	if lengthRatio >= 1.0 && lengthRatio <= 3.0 {
		// Target is 1-3x the query length - good precision
		baseScore += 5
	} else if lengthRatio > 10.0 {
		// Very long target compared to query - less precise
		baseScore -= 5
	}

	// Cap the score
	if baseScore > 100 {
		baseScore = 100
	}
	if baseScore < 0 {
		baseScore = 0
	}

	return baseScore
}

// calculateCommentScore scores comment matches
func calculateCommentScore(query, commentLower string) int {
	// Base score for containing the query
	score := 70

	// Bonus if query is at the start
	if strings.HasPrefix(commentLower, query) {
		score += 15
	} else {
		// Check if it's at word boundary
		idx := strings.Index(commentLower, query)
		if idx > 0 && (commentLower[idx-1] == ' ' || commentLower[idx-1] == '\t') {
			score += 10
		}
	}

	// Bonus for exact word match
	words := strings.Fields(commentLower)
	for _, word := range words {
		if word == query {
			score += 10
			break
		}
	}

	// Penalty for very long comments (less precise match)
	if len(commentLower) > len(query)*10 {
		score -= 5
	}

	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}

	return score
}

// GetStats returns statistics about the indexed proto files
func (pi *ProtoIndex) GetStats() Stats {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	return Stats{
		TotalFiles:             len(pi.files),
		TotalServices:          len(pi.services),
		TotalMessages:          len(pi.messages),
		TotalEnums:             len(pi.enums),
		TotalSearchableEntries: len(pi.searchEntries),
	}
}

// GetService retrieves a service by name
func (pi *ProtoIndex) GetService(name string, resolveTypes bool, maxDepth int) (map[string]interface{}, error) {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	// Try exact match first
	service, exists := pi.services[name]
	if !exists {
		// Try fuzzy match
		for fullName, svc := range pi.services {
			if endsWith(fullName, "."+name) || svc.Name == name {
				service = svc
				break
			}
		}
	}

	if service == nil {
		return nil, fmt.Errorf("service not found: %s", name)
	}

	// Build result
	result := map[string]interface{}{
		"name":      service.Name,
		"full_name": service.FullName,
		"comment":   service.Comment,
		"file":      pi.findFileForDefinition(service.FullName, "service"),
	}

	// Add RPCs
	rpcs := make([]map[string]interface{}, len(service.RPCs))
	for i, rpc := range service.RPCs {
		rpcs[i] = map[string]interface{}{
			"name":               rpc.Name,
			"request_type":       rpc.RequestType,
			"response_type":      rpc.ResponseType,
			"request_streaming":  rpc.RequestStreaming,
			"response_streaming": rpc.ResponseStreaming,
			"comment":            rpc.Comment,
		}
	}
	result["rpcs"] = rpcs

	// Recursively resolve request/response types
	if resolveTypes && maxDepth > 0 {
		resolvedTypes := pi.resolveServiceTypes(service, maxDepth)
		if len(resolvedTypes) > 0 {
			result["resolved_types"] = resolvedTypes
		}
	}

	return result, nil
}

// GetMessage retrieves a message by name
func (pi *ProtoIndex) GetMessage(name string, resolveTypes bool, maxDepth int) (map[string]interface{}, error) {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	// Try exact match first
	message, exists := pi.messages[name]
	if !exists {
		// Try fuzzy match
		for fullName, msg := range pi.messages {
			if endsWith(fullName, "."+name) || msg.Name == name {
				message = msg
				break
			}
		}
	}

	if message == nil {
		return nil, fmt.Errorf("message not found: %s", name)
	}

	// Build result
	result := map[string]interface{}{
		"name":      message.Name,
		"full_name": message.FullName,
		"comment":   message.Comment,
		"file":      pi.findFileForDefinition(message.FullName, "message"),
	}

	// Add fields
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
	result["fields"] = fields

	// Recursively resolve field types
	if resolveTypes && maxDepth > 0 {
		resolvedTypes := pi.resolveMessageTypes(message, maxDepth, nil)
		if len(resolvedTypes) > 0 {
			result["resolved_types"] = resolvedTypes
		}
	}

	return result, nil
}

// GetEnum retrieves an enum by name
func (pi *ProtoIndex) GetEnum(name string) (map[string]interface{}, error) {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	// Try exact match first
	enum, exists := pi.enums[name]
	if !exists {
		// Try fuzzy match
		for fullName, e := range pi.enums {
			if endsWith(fullName, "."+name) || e.Name == name {
				enum = e
				break
			}
		}
	}

	if enum == nil {
		return nil, fmt.Errorf("enum not found: %s", name)
	}

	// Build result
	result := map[string]interface{}{
		"name":      enum.Name,
		"full_name": enum.FullName,
		"comment":   enum.Comment,
		"file":      pi.findFileForDefinition(enum.FullName, "enum"),
	}

	// Add values
	values := make([]map[string]interface{}, len(enum.Values))
	for i, value := range enum.Values {
		values[i] = map[string]interface{}{
			"name":    value.Name,
			"number":  value.Number,
			"comment": value.Comment,
		}
	}
	result["values"] = values

	return result, nil
}

func (pi *ProtoIndex) findFileForDefinition(fullName, defType string) string {
	for filePath, protoFile := range pi.files {
		switch defType {
		case "service":
			for _, s := range protoFile.Services {
				if s.FullName == fullName {
					return filePath
				}
			}
		case "message":
			for _, m := range protoFile.Messages {
				if m.FullName == fullName {
					return filePath
				}
			}
		case "enum":
			for _, e := range protoFile.Enums {
				if e.FullName == fullName {
					return filePath
				}
			}
		}
	}
	return ""
}

func endsWith(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}
