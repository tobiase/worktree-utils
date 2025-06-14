# wt - Git Worktree Manager

![Tests](https://github.com/tobiase/worktree-utils/workflows/Tests/badge.svg)
![contributions not accepted](https://img.shields.io/badge/contributions-not%20accepted-red.svg)

A fast, project-aware Git worktree manager that simplifies working with multiple branches.

## Status

This is a personal utility project. While public for ease of access, I'm not accepting external contributions at this time.

## Features

- 🚀 **Quick Navigation** - Switch between worktrees instantly with `wt go`
- 📁 **Project-Specific Commands** - Define custom navigation shortcuts per project
- 🔄 **Environment Sync** - Copy `.env` files between worktrees
- 🛠️ **Self-Installing** - Single binary that sets itself up
- 🎯 **Smart Detection** - Automatically loads commands based on current project
- ⌨️ **Shell Completion** - Intelligent tab completion for commands, branches, and flags

## Installation

### Quick Install

```bash
curl -fsSL https://raw.githubusercontent.com/tobiase/worktree-utils/main/get.sh | sh
```

### Prebuilt Binaries

Download the latest releases from the [GitHub releases page](https://github.com/tobiase/worktree-utils/releases).

### Build from Source

```bash
git clone https://github.com/tobiase/worktree-utils.git
cd worktree-utils
make build
./wt-bin setup
```

## Usage

### Core Commands

```bash
# List all worktrees
wt list

# Create a new worktree
wt add feature-branch

# Create and switch to a new worktree
wt new feature-branch --base main

# Switch to a worktree (by index or name)
wt go 1
wt go feature-branch

# Remove a worktree
wt rm feature-branch
```

### Utility Commands

```bash
# Copy .env files to another worktree
wt env-copy feature-branch
wt env-copy feature-branch --recursive

# Initialize project configuration
wt project init myproject
```

### Project-Specific Commands

Create project-specific navigation commands that only appear when you're in that project:

```yaml
# ~/.config/wt/projects/myproject.yaml
name: myproject
match:
  paths:
    - /Users/you/projects/myproject
    - /Users/you/projects/myproject-worktrees/*
commands:
  dash:
    description: "Go to dashboard"
    target: "apps/dashboard"
  api:
    description: "Go to API"
    target: "services/api"
```

Now `wt dash` and `wt api` are available only in the myproject repository.

## Shell Completion

wt provides intelligent shell completion for commands, branches, and flags to enhance your workflow.

### Installation

Completion is automatically installed when you run the setup command:

```bash
wt setup
```

### Manual Installation

For existing installations or custom setups:

```bash
# Bash users
wt completion bash >> ~/.bashrc
source ~/.bashrc

# Zsh users
wt completion zsh >> ~/.zshrc
source ~/.zshrc

# Or use with eval for temporary testing
eval "$(wt completion bash)"
eval "$(wt completion zsh)"
```

### Features

- **Command completion**: Tab-complete all wt commands and aliases (`list`, `ls`, `go`, `switch`, etc.)
- **Branch completion**: Intelligent branch name suggestions for relevant commands
- **Flag completion**: Complete command flags with descriptions (e.g., `--base`, `--recursive`)
- **Project commands**: Auto-complete project-specific commands when available
- **Context-aware**: Different completions based on command position and context

### Setup Options

Control completion installation during setup:

```bash
# Install with auto-detected shell completion (default)
wt setup

# Install with specific shell completion
wt setup --completion bash
wt setup --completion zsh

# Install without completion
wt setup --no-completion
```

## Configuration

Configuration files are stored in `~/.config/wt/`:

```
~/.config/wt/
├── init.sh              # Shell integration
├── completion.bash      # Bash completion script
├── completion.zsh       # Zsh completion script
└── projects/            # Project-specific configs
    ├── project1.yaml
    └── project2.yaml
```

## Documentation

- [Git Commands Reference](docs/GIT_COMMANDS.md) - Detailed documentation of the underlying git commands

## Uninstall

```bash
wt setup --uninstall
```

Then remove the initialization line from your shell config (`.bashrc`, `.zshrc`, etc.).

## License

MIT License - see [LICENSE](LICENSE) file for details.
