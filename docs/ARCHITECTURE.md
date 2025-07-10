# Architecture Overview

This document describes the internal architecture of the worktree-utils project, including design patterns, key abstractions, and testing strategies.

## Design Principles

1. **Testability First** - Use dependency injection and interfaces for easy testing
2. **Separation of Concerns** - Clear boundaries between packages
3. **No External Dependencies** - Minimal dependencies for fast builds and easy distribution
4. **Edge Case Resilience** - Handle failures gracefully with clear error messages

## Package Structure

### Core Packages

#### `cmd/wt/`
The main binary entry point and command routing.
- `main.go` - Command dispatcher and shell integration
- `main_test.go` - Integration tests for command handlers

#### `internal/worktree/`
Core worktree management functionality.
- `worktree.go` - Basic worktree operations (list, add, remove, go)
- `project.go` - Project-specific operations (env sync, smart new)
- `service.go` - Dependency-injected service for testable operations
- `interactive.go` - Interactive branch selection
- `command.go` - Command execution utilities

#### `internal/config/`
Project configuration management.
- `config.go` - Configuration loading and management
- `types.go` - Configuration data structures

#### `internal/cli/`
Reusable CLI utilities for consistent command handling.
- `flags.go` - Unified flag parsing with FlagSet struct
- `branch.go` - Branch resolution and fuzzy matching
- `errors.go` - Consistent error handling
- `handler.go` - Command handler wrappers
- `router.go` - Subcommand routing

#### `internal/help/`
Comprehensive help system.
- `help.go` - Help display and flag detection
- `topics/` - Individual help topics as markdown files

### Abstraction Layers

#### `internal/git/`
Git operations abstraction.
```go
type Client interface {
    RevParse(args ...string) (string, error)
    WorktreeList() ([]GitWorktree, error)
    WorktreeAdd(path, branch string, options ...string) error
    // ... more operations
}
```

#### `internal/filesystem/`
File system operations abstraction.
```go
type Filesystem interface {
    ReadFile(path string) ([]byte, error)
    WriteFile(path string, data []byte, perm os.FileMode) error
    MkdirAll(path string, perm os.FileMode) error
    // ... more operations
}
```

#### `internal/shell/`
Shell command execution abstraction.
```go
type Shell interface {
    Run(name string, args ...string) error
    RunWithOutput(name string, args ...string) (string, error)
    RunInDir(dir, name string, args ...string) error
    // ... more operations
}
```

## Design Patterns

### Dependency Injection

The `worktree.Service` demonstrates the dependency injection pattern:

```go
type Service struct {
    git   git.Client
    fs    filesystem.Filesystem
    shell shell.Shell
}

func NewService(git git.Client, fs filesystem.Filesystem, shell shell.Shell) *Service {
    return &Service{git: git, fs: fs, shell: shell}
}
```

This allows for:
- Easy unit testing with mocks
- Clear dependencies
- Flexible implementations

### Command Pattern

Commands follow a consistent pattern:
1. Check for help flag first
2. Parse and validate arguments
3. Execute business logic
4. Handle errors consistently

```go
func handleCommand(args []string) {
    if help.HasHelpFlag(args, "command") {
        return
    }

    flags := cli.ParseFlags(args)
    // ... command logic

    if err != nil {
        cli.ExitWithError("%v", err)
    }
}
```

### Shell Integration Pattern

The tool uses special prefixes for shell integration:
- `CD:` - Change directory
- `EXEC:` - Execute command in shell

This allows the Go binary to control the shell environment.

## Testing Strategy

### Unit Tests

Each package has comprehensive unit tests using:
- Table-driven tests for multiple scenarios
- Mock implementations for external dependencies
- Edge case coverage

Example:
```go
func TestServiceGetRepoRoot(t *testing.T) {
    mockGit := &MockGitClient{
        RevParseFunc: func(args ...string) (string, error) {
            return "/test/repo", nil
        },
    }

    service := NewService(mockGit, nil, nil)
    got, err := service.GetRepoRoot()
    // ... assertions
}
```

### Integration Tests

The `main_test.go` file contains integration tests that:
- Test complete command flows
- Use temporary directories and git repositories
- Verify shell integration behavior

### Edge Case Testing

Focus on scenarios that break in practice:
- Repositories with no commits
- Detached HEAD states
- Missing worktree directories
- Permission issues
- Very long branch names
- Special characters in paths

## Future Considerations

### CLI Framework Migration

The codebase is structured to allow easy migration to a CLI framework like Cobra:
- Commands are already separated
- Flag parsing is centralized
- Help system is modular

### Performance Optimizations

Current areas for potential optimization:
- Parallel git operations for multiple worktrees
- Caching of git command outputs
- Lazy loading of project configurations

### Extended Functionality

The architecture supports adding:
- Remote worktree operations
- Worktree templates
- Advanced filtering and search
- Integration with other git tools
