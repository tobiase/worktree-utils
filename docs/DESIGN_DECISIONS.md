# Design Decisions

This document captures key architectural and design decisions made in the worktree-utils project to ensure consistency across development sessions.

## Table of Contents
- [Dependency Passing Pattern](#dependency-passing-pattern)
- [Shell Integration Pattern](#shell-integration-pattern)
- [Project Configuration](#project-configuration)
- [Error Handling](#error-handling)
- [Code Organization](#code-organization)

## Dependency Passing Pattern

### Decision: Explicit Parameter Passing
When functions need access to configuration or other dependencies, we pass them explicitly as parameters rather than using global state or other patterns.

**Example:**
```go
// Good - explicit dependency
func Add(branch string, cfg *config.Manager) error

// Not used - global state
var globalConfig *config.Manager
func Add(branch string) error  // accesses globalConfig
```

**Rationale:**
- Makes dependencies explicit and visible
- Easier to test (can pass mock configs)
- Follows Go idioms of explicit over implicit
- No hidden coupling between packages

**Applied in:**
- `worktree.Add()` - accepts config.Manager to access worktree_base setting
- `worktree.NewWorktree()` - accepts config.Manager for same reason

## Shell Integration Pattern

### Decision: CD: and EXEC: Prefixes
The Go binary communicates directory changes and commands to execute via special prefixes that the shell wrapper interprets.

**Pattern:**
- `CD:/path/to/dir` - Shell wrapper performs `cd` to this directory
- `EXEC:command` - Shell wrapper executes this command (e.g., `source .venv/bin/activate`)

**Rationale:**
- Go binary runs as subprocess and cannot change parent shell's directory
- Clean separation between binary logic and shell operations
- Extensible for future shell operations

## Project Configuration

### Decision: YAML-based Per-Project Configs
Project configurations are stored as individual YAML files in `~/.config/wt/projects/<name>.yaml`.

**Structure:**
```yaml
name: projectname
match:
  paths:
    - /path/to/repo
    - /path/to/repo-worktrees/*
  remotes:
    - git@github.com:user/repo.git
settings:
  worktree_base: /custom/path/to/worktrees
commands:
  api:
    description: "Go to API"
    target: "services/api"
```

**Rationale:**
- Each project isolated in its own file
- Easy to share project configs between machines
- Simple to edit manually
- Match patterns allow flexible project detection

## Error Handling

### Decision: Early Exit with Descriptive Messages
Commands exit immediately on error with status 1 and descriptive error messages to stderr.

**Pattern:**
```go
if err != nil {
    fmt.Fprintf(os.Stderr, "wt: failed to do X: %v\n", err)
    os.Exit(1)
}
```

**Rationale:**
- Clear error messages help debugging
- Consistent "wt: " prefix identifies source
- Non-zero exit codes allow shell scripting
- Fail fast prevents cascading errors

## Code Organization

### Decision: Package by Feature
Code is organized into packages by feature/domain rather than by technical layers.

**Structure:**
```
internal/
  config/     # Configuration management
  worktree/   # Git worktree operations
  setup/      # Installation/setup logic
  update/     # Self-update functionality
```

**Rationale:**
- Each package has clear responsibility
- Easier to understand and modify features
- Natural boundaries for testing
- Follows Go conventions for internal packages

## Binary Naming Convention

### Decision: Binary Named wt-bin, Shell Function Named wt
The actual binary is named `wt-bin` while users interact with a shell function named `wt`.

**Rationale:**
- Avoids naming conflicts
- Shell function can intercept special commands (CD:, EXEC:)
- Clear distinction between binary and user interface
- Binary name indicates it's not meant to be called directly

## Worktree Organization Convention

### Decision: Sibling Directory Pattern
Default worktree organization uses a sibling directory pattern:
- Main repo: `/path/to/myproject`
- Worktrees: `/path/to/myproject-worktrees/<branch-name>`

**Rationale:**
- Keeps worktrees organized together
- Easy to identify which worktrees belong to which project
- Simple to clean up all worktrees for a project
- Can be overridden per-project via worktree_base setting
