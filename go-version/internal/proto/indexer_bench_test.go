package proto

import (
	"log/slog"
	"os"
	"testing"
)

func BenchmarkSearchNames(b *testing.B) {
	index := createBenchIndex(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = index.Search("UserService", 20, 60)
	}
}

func BenchmarkSearchFields(b *testing.B) {
	index := createBenchIndex(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = index.Search("email", 20, 60)
	}
}

func BenchmarkSearchComments(b *testing.B) {
	index := createBenchIndex(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = index.Search("user operations", 20, 60)
	}
}

func BenchmarkSearchPartial(b *testing.B) {
	index := createBenchIndex(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = index.Search("User", 20, 60)
	}
}

func BenchmarkSearchNoMatch(b *testing.B) {
	index := createBenchIndex(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = index.Search("nonexistent", 20, 60)
	}
}

func BenchmarkIndexFile(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	// Find a proto file to benchmark
	exampleDir := "../../../python-version/examples"
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		b.Skip("Example directory not found")
	}

	testFile := "../../../python-version/examples/api/v1/user.proto"
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		b.Skip("Test file not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index := NewProtoIndex(logger)
		_ = index.IndexFile(testFile)
	}
}

func BenchmarkIndexDirectory(b *testing.B) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))

	exampleDir := "../../../python-version/examples"
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		b.Skip("Example directory not found")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index := NewProtoIndex(logger)
		_, _ = index.IndexDirectory(exampleDir)
	}
}

func BenchmarkGetService(b *testing.B) {
	index := createBenchIndex(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = index.GetService("UserService", false, 0)
	}
}

func BenchmarkGetMessage(b *testing.B) {
	index := createBenchIndex(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = index.GetMessage("User", false, 0)
	}
}

func BenchmarkConcurrentSearch(b *testing.B) {
	index := createBenchIndex(b)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		queries := []string{"User", "Service", "email", "auth"}
		i := 0
		for pb.Next() {
			query := queries[i%len(queries)]
			_ = index.Search(query, 20, 60)
			i++
		}
	})
}

// Helper function to create a benchmark index
func createBenchIndex(b *testing.B) *ProtoIndex {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelError,
	}))
	index := NewProtoIndex(logger)

	exampleDir := "../../../python-version/examples"
	if _, err := os.Stat(exampleDir); os.IsNotExist(err) {
		b.Skip("Example directory not found")
	}

	_, err := index.IndexDirectory(exampleDir)
	if err != nil {
		b.Fatalf("Failed to create benchmark index: %v", err)
	}

	return index
}













