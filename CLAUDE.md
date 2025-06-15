# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Session Continuity

**IMPORTANT**: Before starting work, check `docs/SESSION_LOG.md` for:
- Recent work completed
- Open questions and next steps
- Design decisions made

After completing work, update SESSION_LOG.md with:
- What was accomplished
- Key decisions and rationale
- Next steps for future sessions

When adding session log entries, use `date +%Y-%m-%d` to get the current date in YYYY-MM-DD format.

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

# Run the same tests as GitHub Actions
make test-ci

# Run linting (auto-installs golangci-lint if needed)
make lint

# Install golangci-lint separately
make install-golangci-lint

# Build for all platforms
make build-all

# Clean build artifacts
make clean
```

## Fresh Shell Testing Commands

**IMPORTANT**: Use these commands for testing completion to avoid shell corruption issues:

```bash
# Quick completion test in fresh shell
make test-completion

# Interactive shell with completion loaded (test TAB completion manually)
make test-completion-interactive

# Debug completion script generation
make debug-completion

# Test entire setup process in clean environment
make test-setup

# Completely fresh shell environment for manual testing
make test-fresh

# Standalone test script
./scripts/test-completion.sh
```

These tools solve the common problem where testing completion in the current shell fails due to debugging artifacts and cache corruption.

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

### Shell Completion System

The completion system generates dynamic bash and zsh completion scripts with intelligent context awareness:

**Architecture:**
- Core completion logic in `internal/completion/completion.go`
- Shell-specific generators in `internal/completion/bash.go` and `zsh.go`
- Integration with project configuration for custom commands
- Automatic branch discovery using git commands

**Key Components:**
- `CompletionData` struct holds all completion information (commands, aliases, branches, project commands)
- Dynamic branch completion using `git worktree list` and `git branch` commands
- Context-aware completion based on command position and argument types
- Project command integration when available

**Installation Integration:**
- Completion scripts installed to `~/.config/wt/completion.{bash,zsh}`
- Shell init script (`~/.config/wt/init.sh`) loads appropriate completion based on shell type
- Setup command includes completion installation with `--completion` flags

**Testing completion:**
```bash
# Test completion generation
./wt-bin completion bash | head -20
./wt-bin completion zsh | head -20

# Test completion functionality (requires installation)
wt <TAB>
wt go <TAB>
```

## Key Implementation Details

1. **Self-Installing**: The binary can install itself with `wt-bin setup`, which:
   - Copies itself to `~/.local/bin/wt-bin`
   - Creates shell integration in `~/.config/wt/init.sh`
   - Generates and installs shell completion scripts
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
   - `wt completion <shell>`: Generates shell completion scripts for bash or zsh

4. **Setup Command Handling**: The `setup` command bypasses normal initialization since it needs to work before configuration exists. It's handled specially at the start of main().

5. **Export Requirements**: Several functions in the worktree package (like `GetWorktreeBase`) are exported because they're used by other packages, not just internally.

## Working with GitHub CLI

### Creating Issues
When creating issues with `gh issue create`, be careful with shell escaping:
- Use single quotes for the body to avoid shell interpretation
- Backticks in markdown need to be escaped or avoided in direct commands
- Package paths like `internal/config` can be interpreted as shell paths

Example:
```bash
# Good - using single quotes
gh issue create --title "Title" --body 'Content with `backticks`'

# Problematic - double quotes with backticks
gh issue create --title "Title" --body "Content with \`backticks\`"
```

### Release Workflow

#### Semantic Versioning Guidelines

Follow [SemVer](https://semver.org) strictly for version numbers (MAJOR.MINOR.PATCH):

**MAJOR (X.0.0)** - Breaking changes:
- Changes to command-line interface that break existing usage
- Removal of commands or options
- Changes to shell wrapper behavior
- Changes to config file format that require migration

**MINOR (0.X.0)** - New features (backward compatible):
- New commands or subcommands
- New command-line options
- New project configuration features
- New shell integration features

**PATCH (0.0.X)** - Bug fixes and internal improvements:
- Bug fixes that don't change user-facing behavior
- Performance improvements
- Code quality improvements (linting, refactoring)
- Test additions and improvements
- Documentation updates
- CI/CD improvements

#### Release Process
- Tag pushes trigger the release workflow: `git tag v0.1.0 && git push origin v0.1.0`
- If a release fails, delete and recreate the tag to retrigger
- GoReleaser v2 requires different config syntax than v1
- The `universal_binaries` section must be at the top level, not inside builds

#### Examples of Version Bumps
```bash
# Bug fix: Fixed worktree path handling
git tag v0.5.1

# New feature: Added `wt status` command
git tag v0.6.0

# Breaking change: Changed config file format
git tag v1.0.0
```

## Development Workflow

### Testing Changes
1. Build locally: `make build`
2. Test with local binary: `./wt-bin <command>`
3. For shell integration testing: `export WT_BIN="$PWD/wt-bin" && source <(./wt-bin shell-init)`

## Testing Strategy

### Philosophy: Edge Cases Over Coverage

**Coverage metrics are misleading for CLI tools.** What matters is testing scenarios that could actually break for users. Focus on edge cases where bugs hide and users get frustrated, not on achieving percentage thresholds.

### Critical Edge Case Categories

**Git Repository Edge Cases:**
- Repositories with no commits yet
- Detached HEAD states
- Corrupted .git directories
- Very long branch names with special characters
- Repositories with unusual remote configurations
- Permission issues (read-only repositories)
- Empty repositories
- Repositories with broken symlinks

**Worktree Edge Cases:**
- Worktrees with uncommitted changes
- Worktrees with merge conflicts
- Missing worktree directories (deleted manually)
- Worktrees pointing to non-existent branches
- Nested worktree scenarios
- Very deep directory paths
- Worktrees on different filesystems
- Case-sensitive vs case-insensitive filesystems

**Config Edge Cases:**
- Malformed YAML files
- Missing config directories
- Circular project path references
- Projects with broken remote URLs
- Config files with special characters/unicode
- Very large config files
- Permission issues on config files
- Concurrent config file access

**System Edge Cases:**
- Network timeouts during updates
- Disk full scenarios
- Binary permission issues
- Shell integration failures
- Different Git versions (old/new)
- Unusual $HOME directory setups
- Different shells (bash, zsh, fish)
- Systems without required tools

**User Input Edge Cases:**
- Commands with trailing/leading spaces
- Branch names that look like options (--help, -v)
- Very long command lines
- Special characters in paths/names (unicode, spaces, symbols)
- Ctrl+C during operations
- Invalid command combinations
- Missing required arguments

### Test Implementation Guidelines

1. **Start with what breaks in practice** - Test scenarios users actually encounter
2. **Test failure modes** - What happens when things go wrong?
3. **Test boundaries** - Empty inputs, very large inputs, edge values
4. **Test integration points** - Where our code meets Git, filesystem, network
5. **Test recovery** - Can users recover from error states?

### Example Edge Case Tests

```go
// Test corrupted git repository
func TestWorktreeListCorruptedRepo(t *testing.T) {
    repo := createTestRepo(t)
    // Corrupt the .git directory
    os.Remove(filepath.Join(repo, ".git", "HEAD"))

    _, err := worktree.List(repo)
    // Should handle gracefully, not panic
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "repository")
}

// Test very long branch names
func TestAddWorktreeVeryLongBranchName(t *testing.T) {
    longName := strings.Repeat("very-long-branch-name-", 20) // 400+ chars
    // Test system limits gracefully
}

// Test network timeouts
func TestUpdateWithNetworkTimeout(t *testing.T) {
    // Simulate slow/timeout server
    // Verify graceful failure, not hang
}
```

### Happy Path vs Edge Case Balance

- **Happy path tests**: Ensure core functionality works
- **Edge case tests**: Ensure robustness when things go wrong
- **Integration tests**: Ensure components work together under stress

**Priority order**: Critical user journeys → Common error scenarios → Boundary conditions → Coverage gaps

### Managing Tasks Across Sessions
- GitHub Issues are used to track outstanding work
- Issues are labeled with `enhancement`, `good first issue`, `ux`, etc.
- Reference issues in commits: `feat: add feature (#123)`

## Documentation Structure

The project maintains comprehensive documentation in the `docs/` directory:

1. **SESSION_LOG.md** - Work completed each session (CHECK THIS FIRST!)
   - Tracks what was done, decisions made, and next steps
   - Essential for continuing work between sessions

2. **DESIGN_DECISIONS.md** - Architectural choices and patterns
   - Documents why certain approaches were chosen
   - Reference when implementing new features

3. **DEVELOPMENT.md** - Development workflow and practices
   - How to build, test, and release
   - Code style and conventions

4. **CLI_ERGONOMICS.md** - CLI usability assessment
   - Analysis of command structure
   - Improvement ideas and user experience considerations

5. **GIT_COMMANDS.md** - Git command reference
   - Quick reference for git worktree commands

## Commit Message Guidelines

When creating commits, follow these guidelines:

1. **Format**: Use conventional commit format
   ```
   type: description

   - Bullet points for details
   - Keep it concise and clear
   ```

2. **Types**:
   - `feat`: New feature
   - `fix`: Bug fix
   - `docs`: Documentation changes
   - `style`: Code style changes (formatting, missing semicolons, etc.)
   - `refactor`: Code changes that neither fix a bug nor add a feature
   - `test`: Adding or modifying tests
   - `chore`: Maintenance tasks, dependency updates, etc.

3. **Do NOT include**:
   - "Generated with Claude Code" attribution
   - Co-Authored-By lines

4. **Good examples**:
   ```
   feat: add virtualenv management commands

   - Add support for EXEC: prefix in shell wrapper
   - Implement venv, mkvenv, and rmvenv commands
   - Add VirtualenvConfig to project configuration
   ```

5. **Reference issues when applicable**: `fix: correct worktree path handling (#42)`

## Environment Management Architecture Pattern

The `wt env` command demonstrates the preferred pattern for unified subcommand systems:

### Evolution from Single-Purpose to Unified Commands
```bash
# Old approach: Single-purpose commands
wt env-copy feature --recursive

# New approach: Unified subcommand system
wt env sync feature --recursive
wt env diff feature
wt env list
wt env sync --all
```

### Implementation Pattern
1. **Subcommand Routing**: Main command dispatcher routes to subcommand handlers
2. **Shared Functionality**: Common logic (fuzzy matching, validation) shared across subcommands
3. **Progressive Enhancement**: Add new subcommands without breaking existing usage
4. **Help Integration**: Each subcommand gets comprehensive help documentation

### When to Use This Pattern
- Multiple related operations on the same data type (env files, branches, etc.)
- Commands that share common flags and validation logic
- Operations that benefit from unified help and completion

## Pre-commit Hook Management

### Handling Hook Conflicts During Commits
When pre-commit hooks modify files during commit:

```bash
# 1. Create commit normally
git commit -m "feat: implement feature"

# 2. If hooks modify files, they become unstaged changes
git status  # Shows modified files

# 3. Stage the hook fixes and amend the commit
git add .
git commit --amend --no-edit
```

### Common Hook Issues
- **Trailing whitespace**: Hooks automatically fix, creating unstaged changes
- **Import formatting**: Go hooks may reorganize imports
- **Line endings**: Hooks may normalize line endings

### Best Practice
Always run `git status` after a commit to check if hooks created additional changes that need to be staged and amended.

## Release Workflow Enhancement

### Complete Release Process
```bash
# 1. Push changes and get run ID
git push origin main
gh run list --limit 1  # Get the run ID

# 2. Monitor CI completion
gh run watch <run-id>

# 3. Create and push version tag
git tag v0.x.0
git push origin v0.x.0

# 4. Monitor release workflow
gh run list --limit 1  # Get release run ID
gh run watch <release-run-id>

# 5. Verify release creation
gh release view v0.x.0
```

### Semantic Versioning Decision Tree
- **PATCH (0.0.X)**: Bug fixes, docs, tests, refactoring
- **MINOR (0.X.0)**: New commands, new flags, new features
- **MAJOR (X.0.0)**: Breaking CLI changes, removed commands

### Monitoring Best Practices
- Always watch CI runs to catch failures early
- Verify release assets are generated correctly
- Check release notes include all intended commits

## Universal Help System Implementation

### Help Integration Pattern
Every command handler should start with:
```go
if help.HasHelpFlag(args, "commandName") {
    return
}
```

### Help Content Standards
1. **NAME**: Command and brief description
2. **USAGE**: Syntax with optional/required parameters
3. **ALIASES**: Alternative command names
4. **OPTIONS**: Detailed flag descriptions with examples
5. **EXAMPLES**: Real-world usage scenarios
6. **SEE ALSO**: Related commands

### Implementation Requirements
- All commands MUST support `--help` and `-h`
- Help content MUST include practical examples
- Cross-references MUST link related commands
- Flag examples MUST show realistic usage

## Session Continuation Best Practices

### Essential Pre-Work Checklist
1. **ALWAYS read `docs/SESSION_LOG.md` first** - Contains recent work and context
2. **Check git status** - Understand current state before starting
3. **Review recent commits** - Understand what was accomplished previously
4. **Read any TODO lists** - Check for pending tasks

### Handling Interrupted Work
- Pre-commit hook failures often leave work in progress
- Always check for staged vs unstaged changes
- Use `git status` frequently to understand current state
- Complete interrupted commits before starting new work

### Context Recovery Pattern
```bash
# 1. Check current state
git status
git log --oneline -5

# 2. Read session history
cat docs/SESSION_LOG.md | head -50

# 3. Check for TODOs or pending work
rg "TODO|FIXME|XXX" --type go
```

## Command Evolution Patterns

### Progressive Enhancement Strategy
1. **Maintain Backward Compatibility**: Old commands continue working
2. **Add New Unified Interface**: Implement improved command structure
3. **Update Documentation**: Show both old and new patterns
4. **Gradual Migration**: Users can adopt new patterns over time

### Example: env-copy → wt env Evolution
```bash
# Phase 1: Both commands work
wt env-copy feature    # Old command (still works)
wt env sync feature    # New unified interface

# Phase 2: Deprecation warnings (future)
wt env-copy feature    # Shows: "Consider using 'wt env sync' instead"

# Phase 3: Legacy support (long-term)
# Keep old command for backward compatibility
```

### When NOT to Break Compatibility
- Established user workflows depend on current syntax
- Scripts and automation use existing commands
- No compelling ergonomic improvement justifies breaking changes

### Safe Evolution Patterns
- Add new subcommands alongside existing commands
- Extend existing commands with new flags
- Provide multiple interfaces to the same functionality
- Use deprecation warnings before removal

## Claude Code Integration

### Implemented Custom Commands

Based on repetitive patterns in development sessions, these Claude Code custom commands have been implemented as project-level slash commands in `.claude/commands/`:

#### Available Commands

**Session Management:**
- `/project:session-start` - Comprehensive context recovery from SESSION_LOG.md, git status, and pending work
- `/project:session-end` - Document completed work and update session logs

**Release & CI:**
- `/project:release-workflow` - Execute complete push → CI → tag → release → verify workflow
- `/project:test-ci` - Run comprehensive testing (`make test-ci`, linting, build verification)

**Development Workflow:**
- `/project:fix-hooks` - Handle pre-commit hook modifications with stage → amend pattern
- `/project:help-integration` - Add help system integration to new commands
- `/project:update-docs` - Synchronize documentation after feature implementation

#### Usage Examples

```bash
# Start a new development session
/project:session-start

# Run comprehensive testing before release
/project:test-ci

# Execute full release workflow
/project:release-workflow

# Fix pre-commit hook conflicts
/project:fix-hooks

# End session and document work
/project:session-end
```

These commands eliminate repetitive manual steps and ensure consistent patterns across development sessions. They are implemented as simple Markdown files containing detailed prompts for common workflows.

## Future wt Command Ideas

### High-Priority Commands (Next Implementation Phase)

#### Status and Information Commands
```bash
wt status                  # Cross-worktree status overview
  --detail                 # Show uncommitted changes, ahead/behind status
  --since <date>           # Activity since date

wt info [branch]           # Detailed worktree information
  --files                  # File count and sizes
  --activity               # Recent commits and changes

wt activity                # Recent activity across all worktrees
  --author <name>          # Filter by author
  --since <date>           # Time-based filtering
```

#### Maintenance and Cleanup Commands
```bash
wt clean                   # Remove worktrees for merged/deleted branches
  --merged                 # Only merged branches
  --dry-run                # Show what would be removed
  --force                  # Skip confirmation

wt sync                    # Pull latest changes across all worktrees
  --parallel               # Sync in parallel
  --branch <name>          # Sync specific worktree

wt prune                   # Clean up stale worktree references
  --broken                 # Remove broken symlinks
  --orphaned               # Remove orphaned worktree dirs
```

### Medium-Priority Commands (Future Enhancement)

#### Advanced Workflow Commands
```bash
wt compare <branch1> <branch2>  # Compare files/commits between worktrees
  --files                       # File-level differences
  --commits                     # Commit differences

wt backup [branch]              # Create backup/snapshot of worktree state
  --name <backup-name>          # Named backup
  --all                         # Backup all worktrees

wt restore <backup>             # Restore from backup
  --list                        # List available backups

wt config                       # Interactive project configuration
  --edit                        # Edit current project config
  --global                      # Global wt settings
```

#### Integration and Automation Commands
```bash
wt watch [branch]               # Watch for changes and auto-sync
  --command <cmd>               # Run command on changes

wt template                     # Project template management
  --create <name>               # Create template from current setup
  --apply <template>            # Apply template to new project

wt metrics                      # Worktree usage analytics
  --usage                       # Show usage patterns
  --performance                 # Performance metrics
```

### Implementation Guidelines for Future Commands

#### Command Design Principles
1. **Follow Established Patterns**: Use fuzzy matching, help integration, smart defaults
2. **Unified Subcommands**: Group related operations (like `wt env`)
3. **Progressive Enhancement**: Add new functionality without breaking existing workflows
4. **Cross-Platform Compatibility**: Ensure commands work on all supported platforms

#### Priority Ranking Criteria
- **User Impact**: How many daily workflows does this improve?
- **Complexity**: Implementation effort vs. benefit ratio
- **Integration**: How well does it fit with existing command structure?
- **Maintenance**: Long-term support and testing requirements

#### Next Implementation Target
**`wt status`** should be the next major feature - it addresses the most common need (understanding current state across all worktrees) and provides high value with moderate implementation complexity.
