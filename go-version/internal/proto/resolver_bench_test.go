package proto

import (
	"testing"
)

// BenchmarkFindMessageByType benchmarks message type lookup
func BenchmarkFindMessageByType(b *testing.B) {
	index := NewProtoIndex(testLogger())

	// Create test messages
	for i := 0; i < 100; i++ {
		msg := &ProtoMessage{
			Name:     "Message" + string(rune(i)),
			FullName: "api.v1.Message" + string(rune(i)),
			Fields: []ProtoField{
				{Name: "id", Type: "int32", Number: 1},
			},
		}
		index.messages[msg.FullName] = msg
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.findMessageByType("Message50", "api.v1")
	}
}

// BenchmarkFindEnumByType benchmarks enum type lookup
func BenchmarkFindEnumByType(b *testing.B) {
	index := NewProtoIndex(testLogger())

	// Create test enums
	for i := 0; i < 100; i++ {
		enum := &ProtoEnum{
			Name:     "Enum" + string(rune(i)),
			FullName: "api.v1.Enum" + string(rune(i)),
			Values: []ProtoField{
				{Name: "VALUE_0", Number: 0},
			},
		}
		index.enums[enum.FullName] = enum
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.findEnumByType("Enum50", "api.v1")
	}
}

// BenchmarkResolveMessageTypesSimple benchmarks simple message resolution
func BenchmarkResolveMessageTypesSimple(b *testing.B) {
	index := NewProtoIndex(testLogger())

	// Create simple nested structure: A -> B
	msgB := &ProtoMessage{
		Name:     "B",
		FullName: "api.v1.B",
		Fields: []ProtoField{
			{Name: "value", Type: "string", Number: 1},
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.resolveMessageTypes(msgA, 10, nil)
	}
}

// BenchmarkResolveMessageTypesDeep benchmarks deep nested resolution
func BenchmarkResolveMessageTypesDeep(b *testing.B) {
	index := NewProtoIndex(testLogger())

	// Create deep nested structure: A -> B -> C -> D -> E
	msgE := &ProtoMessage{
		Name:     "E",
		FullName: "api.v1.E",
		Fields: []ProtoField{
			{Name: "value", Type: "string", Number: 1},
		},
	}

	msgD := &ProtoMessage{
		Name:     "D",
		FullName: "api.v1.D",
		Fields: []ProtoField{
			{Name: "e", Type: "E", Number: 1},
		},
	}

	msgC := &ProtoMessage{
		Name:     "C",
		FullName: "api.v1.C",
		Fields: []ProtoField{
			{Name: "d", Type: "D", Number: 1},
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
	index.messages["api.v1.D"] = msgD
	index.messages["api.v1.E"] = msgE

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.resolveMessageTypes(msgA, 10, nil)
	}
}

// BenchmarkResolveMessageTypesWide benchmarks wide message resolution (many fields)
func BenchmarkResolveMessageTypesWide(b *testing.B) {
	index := NewProtoIndex(testLogger())

	// Create wide structure: A has 10 fields, each pointing to a different message
	msgA := &ProtoMessage{
		Name:     "A",
		FullName: "api.v1.A",
		Fields:   make([]ProtoField, 10),
	}

	for i := 0; i < 10; i++ {
		typeName := "Type" + string(rune('A'+i))
		msgA.Fields[i] = ProtoField{
			Name:   "field" + string(rune('0'+i)),
			Type:   typeName,
			Number: i + 1,
		}

		// Create the referenced message
		msg := &ProtoMessage{
			Name:     typeName,
			FullName: "api.v1." + typeName,
			Fields: []ProtoField{
				{Name: "value", Type: "string", Number: 1},
			},
		}
		index.messages["api.v1."+typeName] = msg
	}

	index.messages["api.v1.A"] = msgA

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.resolveMessageTypes(msgA, 10, nil)
	}
}

// BenchmarkResolveMessageTypesCircular benchmarks circular reference resolution
func BenchmarkResolveMessageTypesCircular(b *testing.B) {
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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.resolveMessageTypes(msgA, 10, nil)
	}
}

// BenchmarkResolveServiceTypes benchmarks service type resolution
func BenchmarkResolveServiceTypes(b *testing.B) {
	index := NewProtoIndex(testLogger())

	// Create service with multiple RPCs
	service := &ProtoService{
		Name:     "UserService",
		FullName: "api.v1.UserService",
		RPCs: []ProtoRPC{
			{Name: "GetUser", RequestType: "GetUserRequest", ResponseType: "GetUserResponse"},
			{Name: "CreateUser", RequestType: "CreateUserRequest", ResponseType: "CreateUserResponse"},
			{Name: "UpdateUser", RequestType: "UpdateUserRequest", ResponseType: "UpdateUserResponse"},
			{Name: "DeleteUser", RequestType: "DeleteUserRequest", ResponseType: "DeleteUserResponse"},
		},
	}

	// Create request and response messages
	for _, rpc := range service.RPCs {
		reqMsg := &ProtoMessage{
			Name:     rpc.RequestType,
			FullName: "api.v1." + rpc.RequestType,
			Fields: []ProtoField{
				{Name: "id", Type: "int32", Number: 1},
			},
		}
		index.messages["api.v1."+rpc.RequestType] = reqMsg

		respMsg := &ProtoMessage{
			Name:     rpc.ResponseType,
			FullName: "api.v1." + rpc.ResponseType,
			Fields: []ProtoField{
				{Name: "user", Type: "User", Number: 1},
			},
		}
		index.messages["api.v1."+rpc.ResponseType] = respMsg
	}

	// Create User message
	user := &ProtoMessage{
		Name:     "User",
		FullName: "api.v1.User",
		Fields: []ProtoField{
			{Name: "id", Type: "int32", Number: 1},
			{Name: "name", Type: "string", Number: 2},
		},
	}
	index.messages["api.v1.User"] = user

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.resolveServiceTypes(service, 10)
	}
}

// BenchmarkGetMessageWithResolution benchmarks GetMessage with type resolution
func BenchmarkGetMessageWithResolution(b *testing.B) {
	index := NewProtoIndex(testLogger())

	// Create nested messages
	address := &ProtoMessage{
		Name:     "Address",
		FullName: "api.v1.Address",
		Fields: []ProtoField{
			{Name: "street", Type: "string", Number: 1},
			{Name: "city", Type: "string", Number: 2},
		},
	}

	user := &ProtoMessage{
		Name:     "User",
		FullName: "api.v1.User",
		Fields: []ProtoField{
			{Name: "id", Type: "int32", Number: 1},
			{Name: "name", Type: "string", Number: 2},
			{Name: "address", Type: "Address", Number: 3},
		},
	}

	index.messages["api.v1.User"] = user
	index.messages["api.v1.Address"] = address

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.GetMessage("User", true, 10)
	}
}

// BenchmarkGetServiceWithResolution benchmarks GetService with type resolution
func BenchmarkGetServiceWithResolution(b *testing.B) {
	index := NewProtoIndex(testLogger())

	// Create service
	service := &ProtoService{
		Name:     "UserService",
		FullName: "api.v1.UserService",
		RPCs: []ProtoRPC{
			{Name: "GetUser", RequestType: "GetUserRequest", ResponseType: "GetUserResponse"},
		},
	}

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

	index.services["api.v1.UserService"] = service
	index.messages["api.v1.GetUserRequest"] = request
	index.messages["api.v1.GetUserResponse"] = response
	index.messages["api.v1.User"] = user

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.GetService("UserService", true, 10)
	}
}

// BenchmarkMessageToMap benchmarks message serialization
func BenchmarkMessageToMap(b *testing.B) {
	index := NewProtoIndex(testLogger())

	msg := &ProtoMessage{
		Name:     "User",
		FullName: "api.v1.User",
		Fields: []ProtoField{
			{Name: "id", Type: "int32", Number: 1},
			{Name: "name", Type: "string", Number: 2},
			{Name: "email", Type: "string", Number: 3},
			{Name: "age", Type: "int32", Number: 4},
			{Name: "active", Type: "bool", Number: 5},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.messageToMap(msg)
	}
}

// BenchmarkEnumToMap benchmarks enum serialization
func BenchmarkEnumToMap(b *testing.B) {
	index := NewProtoIndex(testLogger())

	enum := &ProtoEnum{
		Name:     "Status",
		FullName: "api.v1.Status",
		Values: []ProtoField{
			{Name: "UNKNOWN", Number: 0},
			{Name: "ACTIVE", Number: 1},
			{Name: "INACTIVE", Number: 2},
			{Name: "PENDING", Number: 3},
			{Name: "DELETED", Number: 4},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.enumToMap(enum)
	}
}
















