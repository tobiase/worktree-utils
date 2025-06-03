#!/bin/sh
# Download and install script for wt - Git Worktree Manager
# This script detects the OS/architecture, downloads the appropriate release,
# and runs the setup process.
#
# Usage: curl -fsSL https://raw.githubusercontent.com/tobiase/worktree-utils/main/get.sh | sh

set -e

# Colors for output (using printf for better compatibility)
if [ -t 1 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

# Helper functions
info() {
    printf "${GREEN}▶${NC} %s\n" "$1"
}

error() {
    printf "${RED}✗${NC} %s\n" "$1" >&2
    exit 1
}

warn() {
    printf "${YELLOW}⚠${NC} %s\n" "$1"
}

# Check for required tools
check_requirements() {
    if ! command -v curl >/dev/null 2>&1; then
        error "curl is required but not installed"
    fi
    
    if ! command -v tar >/dev/null 2>&1; then
        error "tar is required but not installed"
    fi
}

# Detect OS and architecture
detect_platform() {
    OS=$(uname -s)
    ARCH=$(uname -m)
    
    case "$OS" in
        Darwin)
            PLATFORM="Darwin_all"
            ;;
        Linux)
            case "$ARCH" in
                x86_64)
                    PLATFORM="Linux_x86_64"
                    ;;
                aarch64|arm64)
                    PLATFORM="Linux_arm64"
                    ;;
                *)
                    error "Unsupported Linux architecture: $ARCH"
                    ;;
            esac
            ;;
        *)
            error "Unsupported operating system: $OS"
            ;;
    esac
    
    info "Detected platform: $PLATFORM"
}

# Download and install
install_wt() {
    # Create a temporary directory
    TEMP_DIR=$(mktemp -d 2>/dev/null || mktemp -d -t 'wt-install')
    
    # Ensure cleanup on exit
    trap 'rm -rf "$TEMP_DIR"' EXIT INT TERM
    
    info "Downloading wt..."
    
    # Get latest version from GitHub API
    VERSION=$(curl -fsSL https://api.github.com/repos/tobiase/worktree-utils/releases/latest | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        error "Failed to determine latest version"
    fi
    
    info "Latest version: $VERSION"
    
    # Construct download URL with version (strip 'v' prefix for asset name)
    VERSION_NO_V=${VERSION#v}
    URL="https://github.com/tobiase/worktree-utils/releases/download/${VERSION}/wt_${VERSION_NO_V}_${PLATFORM}.tar.gz"
    
    cd "$TEMP_DIR"
    
    # Download with progress indicator
    if ! curl -fL# "$URL" -o wt.tar.gz; then
        error "Failed to download wt from: $URL"
    fi
    
    info "Extracting..."
    # List what will be extracted to get the directory name
    EXTRACT_DIR=$(tar tzf wt.tar.gz | head -1 | cut -d'/' -f1)
    
    if [ -z "$EXTRACT_DIR" ]; then
        error "Could not determine archive structure"
    fi
    
    if ! tar xzf wt.tar.gz; then
        error "Failed to extract archive"
    fi
    
    if [ ! -d "$EXTRACT_DIR" ]; then
        error "Expected directory '$EXTRACT_DIR' not found after extraction"
    fi
    
    info "Running setup..."
    # The binary might be named 'worktree-utils' or 'wt-bin'
    if [ -f "$EXTRACT_DIR/wt-bin" ]; then
        BINARY="$EXTRACT_DIR/wt-bin"
    elif [ -f "$EXTRACT_DIR/worktree-utils" ]; then
        BINARY="$EXTRACT_DIR/worktree-utils"
    else
        error "Could not find binary in $EXTRACT_DIR"
    fi
    
    if ! "$BINARY" setup; then
        error "Setup failed"
    fi
    
    cd - >/dev/null 2>&1 || true
    
    printf "\n${GREEN}✓${NC} wt has been successfully installed!\n\n"
    printf "To get started, either:\n"
    printf "  • Restart your shell, or\n"
    printf "  • Run: ${GREEN}source ~/.config/wt/init.sh${NC}\n\n"
    printf "Then try: ${GREEN}wt${NC}\n\n"
}

# Main
main() {
    printf "\n${GREEN}wt${NC} - Git Worktree Manager installer\n\n"
    
    check_requirements
    detect_platform
    install_wt
}

# Run main function
main "$@"