# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

```bash
# Build the binary
make build

# Install for local development (uses current directory)
make install-local

# Install to ~/.local/bin
make install

# Run tests (none exist yet)
make test

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

## Architecture Overview

This is a Git worktree management CLI tool written in Go. The key architectural pattern is the separation between a Go binary that handles logic and a minimal shell wrapper that enables directory changes.

### Shell Integration Pattern

The tool uses a `CD:` prefix pattern to signal directory changes:
- Go binary outputs `CD:/path/to/dir` when a directory change is needed
- Shell wrapper function intercepts this output and performs the actual `cd`
- This allows the tool to change the shell's directory without subprocess limitations

Example from the shell wrapper:
```bash
if [[ "$output" == "CD:"* ]]; then
  cd "${output#CD:}"
```

### Project-Specific Commands

The tool supports project-specific navigation commands that are only available when in that project:
- Project detection uses git repository paths and remote URLs
- Configurations stored in `~/.config/wt/projects/<project>.yaml`
- Commands are dynamically loaded based on current directory

Example project config:
```yaml
name: myproject
match:
  paths:
    - /path/to/repo
    - /path/to/repo-worktrees/*
commands:
  api:
    description: "Go to API"
    target: "services/api"
```

### Worktree Convention

Worktrees are organized in a predictable structure:
- Main repo: `/path/to/myproject`
- Worktrees: `/path/to/myproject-worktrees/<branch-name>`

This convention is enforced throughout the codebase.

## Key Implementation Details

1. **Self-Installing**: The binary can install itself with `wt-bin setup`, which:
   - Copies itself to `~/.local/bin/wt-bin`
   - Creates shell integration in `~/.config/wt/init.sh`
   - Modifies shell configs (.bashrc/.zshrc)

2. **Config Loading**: When any command runs (except shell-init):
   - Detects current git repository
   - Loads matching project config
   - Makes project commands available

3. **Error Handling**: Commands exit with status 1 on error, printing to stderr

4. **Dependencies**: Minimal - only `gopkg.in/yaml.v3` for YAML parsing

## Important Design Decisions

1. **Binary Name**: The binary is named `wt-bin` (not `wt`) because the shell function is named `wt`. This avoids conflicts.

2. **GitHub Releases**: Uses GoReleaser v2 with GitHub Actions. The workflow triggers on version tags (`v*`).

3. **Special Commands**:
   - `wt new <branch>`: Creates a worktree AND switches to it (combines add + go)
   - `wt env-copy <branch>`: Copies .env files from current location to same relative path in target worktree
   - `wt project init <name>`: Registers current repo as a project

4. **Setup Command Handling**: The `setup` command bypasses normal initialization since it needs to work before configuration exists. It's handled specially at the start of main().

5. **Export Requirements**: Several functions in the worktree package (like `GetWorktreeBase`) are exported because they're used by other packages, not just internally.