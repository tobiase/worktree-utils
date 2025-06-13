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

## Configuration

Configuration files are stored in `~/.config/wt/`:

```
~/.config/wt/
├── init.sh              # Shell integration
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