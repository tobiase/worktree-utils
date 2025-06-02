#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() { echo -e "${GREEN}==>${NC} $1"; }
print_warn() { echo -e "${YELLOW}Warning:${NC} $1"; }
print_error() { echo -e "${RED}Error:${NC} $1"; }

# Function to add line to file only if not present
add_to_shell_config() {
  local file="$1"
  local line="$2"
  local marker="worktree-utils"
  
  # Check if already configured
  if grep -q "$marker" "$file" 2>/dev/null; then
    print_info "wt already configured in $file"
  else
    echo "" >> "$file"
    echo "# $marker" >> "$file"
    echo "$line" >> "$file"
    print_info "Added wt to $file"
  fi
}

# Parse arguments
MODE=${1:-user}
VERSION=${2:-latest}

case "$MODE" in
  local)
    print_info "Installing wt for local development..."
    
    # Build the binary
    if ! command -v go &> /dev/null; then
      print_error "Go is not installed. Please install Go first."
      exit 1
    fi
    
    print_info "Building wt-bin..."
    go build -o wt-bin ./cmd/wt
    
    # Make it executable
    chmod +x wt-bin
    
    print_info "Local wt installed!"
    echo ""
    echo "To use in current shell:"
    echo "  export WT_BIN=\"$PWD/wt-bin\""
    echo "  source <(./wt-bin shell-init)"
    echo ""
    echo "To make permanent, add to your shell config:"
    echo "  export WT_BIN=\"$PWD/wt-bin\""
    echo "  [ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh"
    ;;
    
  user)
    print_info "Installing wt for user..."
    
    # Setup directories
    PREFIX="${PREFIX:-$HOME/.local}"
    CONFIG_DIR="$HOME/.config/wt"
    mkdir -p "$PREFIX/bin" "$CONFIG_DIR"
    
    # Install binary
    if [ "$VERSION" = "latest" ] && [ ! -f "./cmd/wt/main.go" ]; then
      print_info "Downloading latest release from GitHub..."
      # TODO: Replace with actual release URL when available
      print_error "GitHub releases not yet configured. Please build from source."
      exit 1
    else
      print_info "Building from source..."
      if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go first."
        exit 1
      fi
      go build -o "$PREFIX/bin/wt-bin" ./cmd/wt
    fi
    
    chmod +x "$PREFIX/bin/wt-bin"
    
    # Create init script
    print_info "Creating shell initialization script..."
    cat > "$CONFIG_DIR/init.sh" << 'EOF'
# worktree-utils shell initialization
if command -v wt-bin &> /dev/null; then
  source <(wt-bin shell-init)
fi
EOF
    
    # Add to shell configs
    INIT_LINE="[ -f ~/.config/wt/init.sh ] && source ~/.config/wt/init.sh"
    
    # Detect shell and add initialization
    if [ -n "$BASH_VERSION" ]; then
      add_to_shell_config "$HOME/.bashrc" "$INIT_LINE"
    fi
    
    if [ -n "$ZSH_VERSION" ] || [ -f "$HOME/.zshrc" ]; then
      add_to_shell_config "$HOME/.zshrc" "$INIT_LINE"
    fi
    
    print_info "Installation complete!"
    echo ""
    echo "Please restart your shell or run:"
    echo "  source ~/.config/wt/init.sh"
    
    # Check if ~/.local/bin is in PATH
    if [[ ":$PATH:" != *":$PREFIX/bin:"* ]]; then
      print_warn "$PREFIX/bin is not in your PATH"
      echo "Add this to your shell config:"
      echo "  export PATH=\"\$PATH:$PREFIX/bin\""
    fi
    ;;
    
  uninstall)
    print_info "Uninstalling wt..."
    
    # Remove binary
    rm -f "$HOME/.local/bin/wt-bin" /usr/local/bin/wt-bin
    
    # Remove config directory
    rm -rf "$HOME/.config/wt"
    
    print_info "Removed wt files"
    print_warn "Please manually remove the wt initialization line from your shell config files"
    ;;
    
  *)
    echo "Usage: $0 [local|user|uninstall] [version]"
    echo ""
    echo "Modes:"
    echo "  local     - Install for development (uses current directory)"
    echo "  user      - Install to ~/.local/bin (default)"
    echo "  uninstall - Remove wt"
    echo ""
    echo "Examples:"
    echo "  $0              # Install to ~/.local/bin"
    echo "  $0 local        # Install for development"
    echo "  $0 user latest  # Install latest release"
    exit 1
    ;;
esac