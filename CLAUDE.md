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
- Tag pushes trigger the release workflow: `git tag v0.1.0 && git push origin v0.1.0`
- If a release fails, delete and recreate the tag to retrigger
- GoReleaser v2 requires different config syntax than v1
- The `universal_binaries` section must be at the top level, not inside builds

## Development Workflow

### Testing Changes
1. Build locally: `make build`
2. Test with local binary: `./wt-bin <command>`
3. For shell integration testing: `export WT_BIN="$PWD/wt-bin" && source <(./wt-bin shell-init)`

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