// Package proto provides parsing and indexing for Protocol Buffer files.
//
// It supports both proto2 and proto3 syntax and can recursively scan
// directories to build a searchable index of services, messages, and enums.
package proto

// ProtoField represents a field in a message or enum value
type ProtoField struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Number  int               `json:"number"`
	Label   string            `json:"label,omitempty"` // optional, repeated, required
	Comment string            `json:"comment,omitempty"`
	Options map[string]string `json:"options,omitempty"`
}

// ProtoMessage represents a message definition
type ProtoMessage struct {
	Name           string         `json:"name"`
	FullName       string         `json:"full_name"`
	Fields         []ProtoField   `json:"fields"`
	NestedMessages []ProtoMessage `json:"nested_messages,omitempty"`
	NestedEnums    []ProtoEnum    `json:"nested_enums,omitempty"`
	Comment        string         `json:"comment,omitempty"`
}

// ProtoEnum represents an enum definition
type ProtoEnum struct {
	Name     string       `json:"name"`
	FullName string       `json:"full_name"`
	Values   []ProtoField `json:"values"`
	Comment  string       `json:"comment,omitempty"`
}

// ProtoRPC represents an RPC method in a service
type ProtoRPC struct {
	Name              string `json:"name"`
	RequestType       string `json:"request_type"`
	ResponseType      string `json:"response_type"`
	RequestStreaming  bool   `json:"request_streaming"`
	ResponseStreaming bool   `json:"response_streaming"`
	Comment           string `json:"comment,omitempty"`
}

// ProtoService represents a service definition
type ProtoService struct {
	Name     string     `json:"name"`
	FullName string     `json:"full_name"`
	RPCs     []ProtoRPC `json:"rpcs"`
	Comment  string     `json:"comment,omitempty"`
}

// ProtoFile represents a complete parsed proto file
type ProtoFile struct {
	Path     string         `json:"path"`
	Package  string         `json:"package"`
	Syntax   string         `json:"syntax"`
	Services []ProtoService `json:"services"`
	Messages []ProtoMessage `json:"messages"`
	Enums    []ProtoEnum    `json:"enums"`
	Imports  []string       `json:"imports"`
}













