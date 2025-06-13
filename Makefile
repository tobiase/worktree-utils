.PHONY: build install install-local test test-ci clean

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
	@echo "Running the same tests as GitHub Actions..."
	@echo "1. Building binary..."
	make build
	@echo "2. Testing basic commands..."
	./$(BINARY_NAME) version
	./$(BINARY_NAME) help
	@echo "3. Running full test suite..."
	go test ./...
	@echo "✅ All CI tests passed!"

# Run linting
lint:
	golangci-lint run

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
