# wt - Git Worktree Manager

![Tests](https://github.com/tobiase/worktree-utils/workflows/Tests/badge.svg)
![contributions not accepted](https://img.shields.io/badge/contributions-not%20accepted-red.svg)

A fast, project-aware Git worktree manager that simplifies working with multiple branches.

## Status

This is a personal utility project. While public for ease of access, I'm not accepting external contributions at this time.

## Features

- 🚀 **Quick Navigation** - Switch between worktrees instantly with `wt go` or `wt 0`, `wt 1`
- 🧠 **Smart Commands** - `wt new feature` works regardless of branch state (new/existing/has worktree)
- 🔍 **Fuzzy Matching** - `wt go mai` automatically switches to `main`, with smart suggestions
- ✅ **One-Command Integration** - `wt integrate <branch>` rebases onto `main`, fast-forward merges, then removes the worktree/branch
- 📖 **Universal Help** - All commands support `--help`/`-h` with detailed documentation
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
wt list                    # or: wt ls

# Smart worktree creation (handles any branch state)
wt new feature-branch      # Creates branch + worktree OR switches if exists
wt new feature --base main # Create from specific base branch

# Quick navigation with fuzzy matching
wt go 1                    # Switch by index
wt go feature-branch       # Switch by exact name
wt go feat                 # Fuzzy match to 'feature-branch'
wt go mai                  # Auto-switches to 'main'
wt 0                       # Direct shortcut to first worktree
wt 1                       # Direct shortcut to second worktree

# Smart removal with suggestions
wt rm feature-branch       # Remove by exact name
wt rm feat                 # Fuzzy match for removal
wt rm feature --branch     # Remove worktree AND delete branch (safe by default)
wt rm feature --branch --force # Force branch deletion if it's not merged

# Integrate and clean up in one step
wt integrate feature-branch     # Rebase onto main, fast-forward merge, remove worktree/branch

# Get help for any command
wt go --help               # Detailed help for 'go' command
wt new -h                  # Short help flag also works
```

### Intelligent Behavior

`wt` uses "Do What I Mean" design - commands understand your intent and provide helpful guidance:

```bash
# Fuzzy matching with auto-resolution
$ wt go mai
# → Automatically switches to 'main' (unambiguous match)

# Smart suggestions for ambiguous input
$ wt go te
# → Shows interactive picker: [test-branch, test-feature, temp-fix]

# Helpful error messages
$ wt go xyz
# → "branch 'xyz' not found. Did you mean:
#     1. main
#     2. fix-xyz-bug
#     3. feature-xyz"

# Smart worktree creation
$ wt new existing-branch
# → "Switched to existing worktree 'existing-branch'" (no error!)

$ wt new new-branch
# → Creates branch + worktree + switches (handles everything)
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
