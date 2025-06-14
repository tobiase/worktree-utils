.PHONY: build install install-local test test-ci clean lint install-golangci-lint test-completion test-completion-interactive test-setup test-fresh debug-completion

# Build variables
BINARY_NAME=wt-bin
MAIN_PATH=./cmd/wt
BUILD_DIR=./build

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE ?= $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')
LDFLAGS = -X 'main.version=$(VERSION)' -X 'main.commit=$(COMMIT)' -X 'main.date=$(DATE)'

# Default target
all: build

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) $(MAIN_PATH)

# Install for local development
install-local:
	./install.sh local

# Install to ~/.local/bin
install:
	./install.sh user

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run the same tests that GitHub Actions runs
test-ci:
	@echo "🚀 Running exactly what GitHub Actions runs..."
	@echo "1. Running linting..."
	make lint
	@echo "2. Running tests with race detection and coverage..."
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	@echo "3. Building binary..."
	make build
	@echo "4. Testing binary functionality..."
	./$(BINARY_NAME) version
	./$(BINARY_NAME) help
	@echo "✅ All GitHub Actions checks passed! Safe to push."

# Install golangci-lint
install-golangci-lint:
	@if ! which golangci-lint > /dev/null && ! test -f $(shell go env GOPATH)/bin/golangci-lint; then \
		echo "Installing golangci-lint v1.64.2..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.64.2; \
		echo "✅ golangci-lint installed to $(shell go env GOPATH)/bin/golangci-lint"; \
	else \
		echo "✅ golangci-lint already installed"; \
	fi

# Run linting
lint: install-golangci-lint
	@if which golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		$(shell go env GOPATH)/bin/golangci-lint run; \
	fi

# Setup pre-commit hooks
setup-hooks:
	@echo "Installing pre-commit..."
	@which pre-commit > /dev/null || pip install --user pre-commit
	pre-commit install
	pre-commit install --hook-type commit-msg
	pre-commit install --hook-type pre-push
	@echo "Pre-commit hooks installed successfully!"

# Run pre-commit on all files
lint-all:
	pre-commit run --all-files

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

# Build for multiple platforms
build-all:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

# Fresh shell testing targets
test-completion: build
	@echo "🧪 Testing completion in fresh zsh shell..."
	@echo 'source <(./$(BINARY_NAME) completion zsh); echo "✓ Completion loaded"; type _wt; echo "Ready for testing: try typing \"wt \" and press TAB"' | zsh

test-completion-interactive: build
	@echo "🚀 Starting fresh zsh with completion loaded..."
	@echo "Type 'wt ' and press TAB to test completion. Type 'exit' to return."
	@echo 'source <(./$(BINARY_NAME) completion zsh); echo "✓ Completion loaded successfully! Try: wt <TAB>"; exec zsh -i' | zsh

test-setup: build
	@echo "🔧 Testing setup process in fresh environment..."
	@echo 'echo "=== Setup Test ==="; ./$(BINARY_NAME) setup --check; echo "=== Testing binary ==="; ./$(BINARY_NAME) --help | head -5; echo "=== Testing completion ==="; source <(./$(BINARY_NAME) completion zsh); type _wt' | env -i HOME=$$HOME PATH=$$PATH zsh

test-fresh: build
	@echo "🆕 Starting completely fresh shell environment..."
	@echo "Binary built. Starting clean zsh shell..."
	@env -i HOME=$$HOME PATH=$$PATH:/Users/tobias/Projects/worktree-utils zsh -c 'echo "Fresh shell ready! Binary at: ./$(BINARY_NAME)"; exec zsh -i'

debug-completion: build
	@echo "🐛 Debugging completion generation..."
	@echo "=== Completion script output ==="
	@./$(BINARY_NAME) completion zsh | head -20
	@echo ""
	@echo "=== Testing in fresh shell ==="
	@echo 'source <(./$(BINARY_NAME) completion zsh) && echo "Completion loaded" && type _wt' | zsh
