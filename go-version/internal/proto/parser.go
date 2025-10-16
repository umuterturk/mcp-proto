package proto

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Compiled regex patterns for performance
var (
	syntaxRegex  = regexp.MustCompile(`syntax\s*=\s*["'](\w+)["']`)
	packageRegex = regexp.MustCompile(`package\s+([\w.]+)\s*;`)
	importRegex  = regexp.MustCompile(`import\s+(?:public\s+|weak\s+)?["']([^"']+)["']`)
	serviceRegex = regexp.MustCompile(`service\s+(\w+)\s*\{([^}]*)\}`)
	rpcRegex     = regexp.MustCompile(`rpc\s+(\w+)\s*\(\s*(stream\s+)?([\w.]+)\s*\)\s*returns\s*\(\s*(stream\s+)?([\w.]+)\s*\)`)
	messageRegex = regexp.MustCompile(`message\s+(\w+)\s*\{([^}]*(?:\{[^}]*\}[^}]*)*)\}`)
	fieldRegex   = regexp.MustCompile(`(optional|required|repeated)?\s*([\w.]+)\s+(\w+)\s*=\s*(\d+)`)
	enumRegex    = regexp.MustCompile(`enum\s+(\w+)\s*\{([^}]*)\}`)
	enumValRegex = regexp.MustCompile(`(\w+)\s*=\s*(\d+)`)
	commentRegex = regexp.MustCompile(`//(.*)$`)
)

// Parser handles parsing of .proto files
type Parser struct {
	currentPackage string
}

// NewParser creates a new proto parser
func NewParser() *Parser {
	return &Parser{}
}

// ParseFile parses a .proto file and extracts all definitions
func (p *Parser) ParseFile(filePath string) (*ProtoFile, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	contentStr := string(content)
	protoFile := &ProtoFile{
		Path: filePath,
	}

	// Preprocess content to extract lines with comments
	lines := p.preprocessContent(contentStr)

	// Parse top-level constructs
	protoFile.Syntax = p.extractSyntax(contentStr)
	protoFile.Package = p.extractPackage(contentStr)
	p.currentPackage = protoFile.Package
	protoFile.Imports = p.extractImports(contentStr)

	// Parse services, messages, and enums
	protoFile.Services = p.extractServices(lines, contentStr)
	protoFile.Messages = p.extractMessages(lines, contentStr, "")
	protoFile.Enums = p.extractEnums(lines, contentStr, "")

	return protoFile, nil
}

type lineWithComment struct {
	line    string
	comment string
}

func (p *Parser) preprocessContent(content string) []lineWithComment {
	lines := strings.Split(content, "\n")
	result := make([]lineWithComment, 0, len(lines))
	var currentComment []string

	for _, line := range lines {
		// Extract trailing comment
		var comment string
		if match := commentRegex.FindStringSubmatch(line); match != nil {
			comment = strings.TrimSpace(match[1])
		}

		// Remove comment from line
		lineWithoutComment := strings.TrimSpace(commentRegex.ReplaceAllString(line, ""))

		// Check if this is a standalone comment line
		if lineWithoutComment == "" && comment != "" {
			currentComment = append(currentComment, comment)
		} else {
			// Attach accumulated comments to this line
			fullComment := ""
			if len(currentComment) > 0 {
				fullComment = strings.Join(currentComment, " ")
			}
			if comment != "" {
				if fullComment != "" {
					fullComment = fullComment + " " + comment
				} else {
					fullComment = comment
				}
			}

			result = append(result, lineWithComment{
				line:    lineWithoutComment,
				comment: fullComment,
			})
			currentComment = nil
		}
	}

	return result
}

func (p *Parser) extractSyntax(content string) string {
	if match := syntaxRegex.FindStringSubmatch(content); match != nil {
		return match[1]
	}
	return "proto2"
}

func (p *Parser) extractPackage(content string) string {
	if match := packageRegex.FindStringSubmatch(content); match != nil {
		return match[1]
	}
	return ""
}

func (p *Parser) extractImports(content string) []string {
	matches := importRegex.FindAllStringSubmatch(content, -1)
	imports := make([]string, 0, len(matches))
	for _, match := range matches {
		imports = append(imports, match[1])
	}
	return imports
}

func (p *Parser) extractServices(lines []lineWithComment, content string) []ProtoService {
	var services []ProtoService

	matches := serviceRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		serviceName := match[1]
		serviceBody := match[2]

		serviceComment := p.findCommentForConstruct(lines, serviceName, "service")

		fullName := serviceName
		if p.currentPackage != "" {
			fullName = p.currentPackage + "." + serviceName
		}

		service := ProtoService{
			Name:     serviceName,
			FullName: fullName,
			Comment:  serviceComment,
			RPCs:     p.extractRPCs(serviceBody),
		}

		services = append(services, service)
	}

	return services
}

func (p *Parser) extractRPCs(serviceBody string) []ProtoRPC {
	var rpcs []ProtoRPC

	matches := rpcRegex.FindAllStringSubmatch(serviceBody, -1)
	for _, match := range matches {
		rpcName := match[1]
		requestStreaming := match[2] != ""
		requestType := match[3]
		responseStreaming := match[4] != ""
		responseType := match[5]

		rpcComment := p.findCommentInBody(serviceBody, rpcName)

		rpc := ProtoRPC{
			Name:              rpcName,
			RequestType:       requestType,
			ResponseType:      responseType,
			RequestStreaming:  requestStreaming,
			ResponseStreaming: responseStreaming,
			Comment:           rpcComment,
		}

		rpcs = append(rpcs, rpc)
	}

	return rpcs
}

func (p *Parser) extractMessages(lines []lineWithComment, content, prefix string) []ProtoMessage {
	var messages []ProtoMessage

	matches := messageRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		messageName := match[1]
		messageBody := match[2]

		messageComment := p.findCommentForConstruct(lines, messageName, "message")

		fullName := prefix + messageName
		if p.currentPackage != "" {
			fullName = p.currentPackage + "." + fullName
		}

		message := ProtoMessage{
			Name:     messageName,
			FullName: fullName,
			Comment:  messageComment,
			Fields:   p.extractFields(messageBody),
		}

		messages = append(messages, message)
	}

	return messages
}

func (p *Parser) extractFields(messageBody string) []ProtoField {
	var fields []ProtoField

	matches := fieldRegex.FindAllStringSubmatch(messageBody, -1)
	for _, match := range matches {
		label := match[1]
		fieldType := match[2]
		fieldName := match[3]
		fieldNumberStr := match[4]

		// Skip nested message/enum definitions
		if fieldType == "message" || fieldType == "enum" || fieldType == "service" {
			continue
		}

		fieldNumber, _ := strconv.Atoi(fieldNumberStr)
		fieldComment := p.findCommentInBody(messageBody, fieldName)

		field := ProtoField{
			Name:    fieldName,
			Type:    fieldType,
			Number:  fieldNumber,
			Label:   label,
			Comment: fieldComment,
			Options: make(map[string]string),
		}

		fields = append(fields, field)
	}

	return fields
}

func (p *Parser) extractEnums(lines []lineWithComment, content, prefix string) []ProtoEnum {
	var enums []ProtoEnum

	matches := enumRegex.FindAllStringSubmatch(content, -1)
	for _, match := range matches {
		enumName := match[1]
		enumBody := match[2]

		enumComment := p.findCommentForConstruct(lines, enumName, "enum")

		fullName := prefix + enumName
		if p.currentPackage != "" {
			fullName = p.currentPackage + "." + fullName
		}

		protoEnum := ProtoEnum{
			Name:     enumName,
			FullName: fullName,
			Comment:  enumComment,
			Values:   p.extractEnumValues(enumBody),
		}

		enums = append(enums, protoEnum)
	}

	return enums
}

func (p *Parser) extractEnumValues(enumBody string) []ProtoField {
	var values []ProtoField

	matches := enumValRegex.FindAllStringSubmatch(enumBody, -1)
	for _, match := range matches {
		valueName := match[1]
		valueNumberStr := match[2]

		valueNumber, _ := strconv.Atoi(valueNumberStr)
		valueComment := p.findCommentInBody(enumBody, valueName)

		value := ProtoField{
			Name:    valueName,
			Type:    "enum_value",
			Number:  valueNumber,
			Comment: valueComment,
		}

		values = append(values, value)
	}

	return values
}

func (p *Parser) findCommentForConstruct(lines []lineWithComment, name, keyword string) string {
	for i, lwc := range lines {
		if strings.Contains(lwc.line, keyword) && strings.Contains(lwc.line, name) {
			var comments []string

			// Check previous lines for comments
			for j := i - 1; j >= 0; j-- {
				if lines[j].line == "" || lines[j].comment != "" {
					if lines[j].comment != "" {
						comments = append([]string{lines[j].comment}, comments...)
					}
				} else {
					break
				}
			}

			// Also include inline comment
			if lwc.comment != "" {
				comments = append(comments, lwc.comment)
			}

			return strings.Join(comments, " ")
		}
	}
	return ""
}

func (p *Parser) findCommentInBody(body, name string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.Contains(line, name) {
			// Check for inline comment
			if match := commentRegex.FindStringSubmatch(line); match != nil {
				return strings.TrimSpace(match[1])
			}

			// Check previous line
			if i > 0 {
				if match := commentRegex.FindStringSubmatch(lines[i-1]); match != nil {
					return strings.TrimSpace(match[1])
				}
			}
		}
	}
	return ""
}
