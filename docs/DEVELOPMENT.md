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

# Run tests (when implemented)
make test

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
- Mock external dependencies (git commands, file system)

## Adding New Commands

1. **Define command in main.go** - Add case to switch statement
2. **Implement logic** - Preferably in appropriate package (worktree, config, etc.)
3. **Update usage** - Add to showUsage() function
4. **Document** - Update README and relevant docs

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

## Future Considerations

- Keep backward compatibility in mind
- Document breaking changes clearly
- Consider user workflows when adding features
- Keep the tool focused on worktree management
