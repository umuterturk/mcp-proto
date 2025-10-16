.PHONY: build test bench clean install run help build-all

# Default target
.DEFAULT_GOAL := build

# Build variables
BINARY_NAME=mcp-proto-server
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "2.0.0-dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DIR=bin
DIST_DIR=dist

# Go build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the server binary
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/mcp-proto-server
	@echo "Built: ./$(BINARY_NAME)"

build-all: ## Build for all platforms (Linux, macOS, Windows)
	@echo "Building for all platforms..."
	@mkdir -p $(DIST_DIR)
	
	# Linux amd64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/mcp-proto-server
	
	# Linux arm64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/mcp-proto-server
	
	# macOS amd64 (Intel)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/mcp-proto-server
	
	# macOS arm64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/mcp-proto-server
	
	# Windows amd64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe ./cmd/mcp-proto-server
	
	@echo "Built all platforms in $(DIST_DIR)/"
	@ls -lh $(DIST_DIR)/

test: ## Run all tests
	go test -v ./...

test-coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

bench: ## Run benchmarks
	go test -bench=. -benchmem ./internal/proto

clean: ## Clean build artifacts
	rm -rf $(BINARY_NAME) $(BUILD_DIR) $(DIST_DIR) coverage.out coverage.html

install: build ## Install binary to GOPATH/bin
	go install $(LDFLAGS) ./cmd/mcp-proto-server

run: build ## Build and run with example protos
	./$(BINARY_NAME) -root ./examples -verbose

run-test: build ## Build and run test script
	./scripts/test_server.sh

lint: ## Run linter
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

mod-tidy: ## Tidy go modules
	go mod tidy

check: fmt vet test ## Run fmt, vet, and tests

release: clean build-all ## Create release builds
	@echo "Creating checksums..."
	@cd $(DIST_DIR) && shasum -a 256 * > checksums.txt
	@echo "Release builds ready in $(DIST_DIR)/"
	@cat $(DIST_DIR)/checksums.txt

version: ## Show version information
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
