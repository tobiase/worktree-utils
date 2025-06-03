.PHONY: build install install-local test clean

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