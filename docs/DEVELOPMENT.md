# Development Guide

This document outlines development practices and workflows for the worktree-utils project.

## Development Commands

```bash
# Build the binary
make build

# Install for local development (uses current directory)
make install-local

# Install to ~/.local/bin
make install

# Run tests
make test

# Run tests with coverage
make test-ci

# Run linting
make lint

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

## Testing Changes

When developing new features or fixing bugs:

1. **Build locally:**
   ```bash
   make build
   ```

2. **Test with local binary:**
   ```bash
   ./wt-bin <command>
   ```

3. **Test shell integration:**
   ```bash
   export WT_BIN="$PWD/wt-bin"
   source <(./wt-bin shell-init)
   wt <command>
   ```

## Code Style Guidelines

### Go Code

1. **No comments unless requested** - Keep code self-documenting through good naming
2. **Follow existing patterns** - Check neighboring code for conventions
3. **Explicit over implicit** - Pass dependencies, don't hide them
4. **Error handling** - Always check errors and provide context

### Imports

- Group imports: stdlib, external deps, internal packages
- Use goimports or gofmt for formatting

### Testing

- Test files should be in the same package
- Use table-driven tests where appropriate
- Mock external dependencies using interfaces
- Aim for edge case coverage over percentage metrics
- Test actual user scenarios and failure modes

#### Using the Interface Pattern

When testing code that interacts with external systems:

```go
// Use interfaces for dependencies
type Service struct {
    git   git.Client
    fs    filesystem.Filesystem
    shell shell.Shell
}

// In tests, use mocks
mockGit := &MockGitClient{
    RevParseFunc: func(args ...string) (string, error) {
        return "/test/repo", nil
    },
}
service := NewService(mockGit, mockFS, mockShell)
```

## Adding New Commands

1. **Define command in main.go** - Add case to switch statement
2. **Implement logic** - Preferably in appropriate package (worktree, config, etc.)
3. **Add help integration** - Use `help.HasHelpFlag()` at start of handler
4. **Create help content** - Add to internal/help/topics/
5. **Update usage** - Add to showUsage() function
6. **Add tests** - Cover main functionality and edge cases
7. **Document** - Update README and relevant docs

### Using CLI Utilities

For new commands, consider using the CLI utilities:

```go
// Use flag parsing
flags := cli.ParseFlags(args)
if flags.Fuzzy { /* ... */ }

// Use branch resolution
target := cli.ResolveBranchArgument(args, flags.Fuzzy, "Usage: wt cmd <branch>")

// Use error handling
cli.HandleError(err, "failed to add worktree")
```

### Command Guidelines

- Commands should be short and memorable
- Use subcommands for related functionality (e.g., `wt project init`)
- Provide helpful error messages
- Exit with status 1 on error

## Project Configuration

When adding new project settings:

1. **Update config struct** - Add field to ProjectConfig or ProjectSettings
2. **Handle in commands** - Check and use the setting where needed
3. **Document** - Add to example configs and docs

## Release Process

1. **Update version** - Releases are triggered by version tags
2. **Create tag:**
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
3. **GitHub Actions** - Automatically builds and creates release
4. **Test release** - Use `wt update` to test the new release

## Common Patterns

### Shell Integration Commands

For commands that need shell integration:

```go
// Directory change
fmt.Printf("CD:%s", targetPath)

// Execute command (like sourcing scripts)
fmt.Printf("EXEC:%s", command)
```

### Project Detection

```go
configMgr, _ := config.NewManager()
cwd, _ := os.Getwd()
gitRemote, _ := worktree.GetGitRemote()
configMgr.LoadProject(cwd, gitRemote)
```

### Error Messages

```go
fmt.Fprintf(os.Stderr, "wt: action failed: %v\n", err)
os.Exit(1)
```

## Debugging

1. **Verbose git output** - Git commands print to stdout/stderr by default
2. **Check exit codes** - Non-zero indicates error
3. **Test shell wrapper** - Run commands with `bash -x` to see execution

## Architecture Overview

### Dependency Injection

The codebase uses dependency injection for better testability:

- **git.Client** - Interface for all git operations
- **filesystem.Filesystem** - Interface for file system operations
- **shell.Shell** - Interface for command execution
- **worktree.Service** - Main service using injected dependencies

### Package Structure

- **cmd/wt/** - Main binary and command handlers
- **internal/cli/** - Common CLI utilities (flags, routing, errors)
- **internal/config/** - Project configuration management
- **internal/git/** - Git operations interface and implementation
- **internal/filesystem/** - File system interface and implementation
- **internal/help/** - Help system and content
- **internal/shell/** - Shell command execution
- **internal/worktree/** - Core worktree operations

## Future Considerations

- Keep backward compatibility in mind
- Document breaking changes clearly
- Consider user workflows when adding features
- Keep the tool focused on worktree management
- Consider migrating to a CLI framework (e.g., Cobra) for better structure
