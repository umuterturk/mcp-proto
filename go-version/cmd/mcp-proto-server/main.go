package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/uerturk/mcp-proto-server/internal/proto"
	"github.com/uerturk/mcp-proto-server/pkg/server"
)

var (
	version = "2.0.0-dev"
	commit  = "dev"
)

func main() {
	// Parse command line flags
	protoRoot := flag.String("root", getEnvOrDefault("PROTO_ROOT", "."), "Root directory containing .proto files")
	watch := flag.Bool("watch", false, "Watch for file changes and re-index automatically")
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	versionFlag := flag.Bool("version", false, "Print version and exit")
	noLogFile := flag.Bool("no-log-file", false, "Disable automatic file logging")

	flag.Parse()

	if *versionFlag {
		fmt.Printf("mcp-proto-server version %s (commit: %s)\n", version, commit)
		os.Exit(0)
	}

	// Setup logging
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}

	// Create log writers
	var logWriter io.Writer = os.Stderr
	var logFileHandle *os.File

	// Automatically create timestamped log file unless disabled
	if !*noLogFile {
		timestamp := time.Now().Format("20060102-150405")
		logFilePath := fmt.Sprintf("/tmp/mcp-proto-server-%s.log", timestamp)

		var err error
		logFileHandle, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: failed to open log file %s: %v\n", logFilePath, err)
			fmt.Fprintf(os.Stderr, "Continuing with stderr logging only\n")
		} else {
			// Write to both stderr and file
			logWriter = io.MultiWriter(os.Stderr, logFileHandle)
			defer logFileHandle.Close()

			// Print to stderr so user knows where logs are
			fmt.Fprintf(os.Stderr, "Logs: %s\n", logFilePath)
		}
	}

	logger := slog.New(slog.NewTextHandler(logWriter, &slog.HandlerOptions{
		Level: logLevel,
	}))
	slog.SetDefault(logger)

	// Validate and resolve proto root
	expandedRoot := os.ExpandEnv(*protoRoot)

	// Handle tilde expansion
	if len(expandedRoot) > 0 && expandedRoot[0] == '~' {
		if home, err := os.UserHomeDir(); err == nil {
			if len(expandedRoot) == 1 {
				expandedRoot = home
			} else {
				expandedRoot = filepath.Join(home, expandedRoot[1:])
			}
		}
	}

	absProtoRoot, err := filepath.Abs(expandedRoot)
	if err != nil {
		logger.Error("failed to resolve proto root", "error", err)
		os.Exit(1)
	}

	if _, err := os.Stat(absProtoRoot); os.IsNotExist(err) {
		logger.Error("proto root directory does not exist", "path", absProtoRoot)
		os.Exit(1)
	}

	logger.Info("starting MCP Proto Server",
		"version", version,
		"proto_root", absProtoRoot,
		"watch", *watch,
	)

	// Create index and scan directory
	index := proto.NewProtoIndex(logger)

	logger.Info("indexing proto files", "root", absProtoRoot)
	count, err := index.IndexDirectory(absProtoRoot)
	if err != nil {
		logger.Error("failed to index directory", "error", err)
		os.Exit(1)
	}

	stats := index.GetStats()
	logger.Info("indexing complete",
		"files", count,
		"services", stats.TotalServices,
		"messages", stats.TotalMessages,
		"enums", stats.TotalEnums,
	)

	if *watch {
		logger.Info("file watching enabled")
		// TODO: Implement file watching in a future phase
	}

	// Create and start MCP server
	mcpServer := server.NewMCPServer(index, logger)
	logger.Info("MCP server ready, waiting for requests on stdio")

	// Setup signal handling for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("shutdown signal received")
		cancel()
	}()

	// Run the server
	logger.Info("entering server run loop")
	if err := mcpServer.Run(ctx); err != nil && err != context.Canceled {
		logger.Error("server exited with error", "error", err, "error_type", fmt.Sprintf("%T", err))
		os.Exit(1)
	}

	logger.Info("server shutdown complete - exiting normally")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
