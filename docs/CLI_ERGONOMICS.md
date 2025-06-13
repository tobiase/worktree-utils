# CLI Ergonomics Assessment

This document evaluates the usability and ergonomics of the `wt` command-line interface, identifying what works well and areas for potential improvement.

## Current Command Structure

### Core Commands (Strengths ✓)
- `wt list` - Clear and standard
- `wt add <branch>` - Follows git convention
- `wt rm <branch>` - Short alias for remove (good!)
- `wt go [branch/index]` - Unique and short
- `wt new <branch>` - Intuitive combo command

### Ergonomic Wins ✅

1. **Short command name** - `wt` is quick to type
2. **Index-based navigation** - `wt go 0` is faster than typing branch names
3. **Smart defaults** - `wt go` (no args) returns to repo root
4. **Combo commands** - `wt new` creates AND switches (saves a step)
5. **Project-specific shortcuts** - Custom navigation commands per project

### Pain Points & Considerations 🤔

1. **Command Naming Inconsistency**
   - `rm` is abbreviated but `list` is not (`ls` might be better?)
   - `env-copy` uses hyphen while other commands don't

2. **Missing Conveniences**
   - No `wt switch` alias for `wt go` (git users might expect this)
   - No `-b` flag for `wt add` to match git's branch creation pattern
   - Can't create worktree from specific commit/tag

3. **Subcommand Organization**
   - `wt project init` makes sense
   - But `wt setup` is top-level (inconsistent?)
   - Should utility commands be grouped? (`wt util env-copy`?)

4. **Discovery Issues**
   - Project commands only visible when in that project
   - No `wt help <command>` for detailed help
   - No command completion hints

## Proposed Improvements

### High-Impact, Low-Effort
1. **Add aliases:**
   ```bash
   wt ls          # alias for list
   wt switch      # alias for go
   wt s           # shorter alias for switch/go
   ```

2. **Standardize naming:**
   ```bash
   wt env copy    # instead of env-copy
   wt env sync    # future: sync all .env files
   ```

3. **Add `-b` flag to add:**
   ```bash
   wt add -b new-feature origin/main  # create from base
   ```

### Medium-Effort Improvements
1. **Interactive mode for common tasks:**
   ```bash
   wt                    # no args shows menu
   > 1. main             # numbered list for quick selection
   > 2. feature-branch
   > 3. bugfix-123
   ```

2. **Better discovery:**
   ```bash
   wt commands           # list ALL available commands
   wt help go           # detailed help for specific command
   ```

3. **Smart suggestions:**
   ```bash
   $ wt go feat
   Did you mean 'feature-branch'? [Y/n]
   ```

### Usage Patterns to Optimize For

1. **Most Common Flow:**
   ```bash
   wt new feature-x     # Start new feature
   # ... work ...
   wt go main          # Back to main
   wt rm feature-x     # Cleanup
   ```

2. **Quick Switching:**
   ```bash
   wt go 0             # By index (fastest)
   wt go feat<TAB>     # With completion
   wt s                # Even shorter
   ```

3. **Project Navigation:**
   ```bash
   wt api              # Project shortcut
   wt docs             # Another shortcut
   ```

## Command Length Analysis

| Command | Keystrokes | Could Be |
|---------|------------|----------|
| wt list | 7 | wt ls (5) |
| wt go 0 | 7 | wt 0 (4)? |
| wt add branch | 13 | - |
| wt new branch | 13 | - |
| wt env-copy br | 14 | wt env cp br (12) |

## Comparison with Similar Tools

### git worktree
- Verbose: `git worktree add ../project-branch branch`
- Our win: `wt add branch` (automatic path management)

### tmux/screen
- Pattern: short commands with subcommands
- We follow this well

### Modern CLIs (gh, cargo, etc.)
- Use subcommands extensively
- Have good help systems
- We could improve here

## Recommendations

### Immediate (No Breaking Changes)
1. Add `wt ls` alias
2. Add `wt switch` and `wt s` aliases
3. Implement `wt help <cmd>`
4. Add shell completions

### Future Considerations
1. Restructure into command groups:
   ```
   wt work new/add/rm/list    # worktree management
   wt go/switch/s             # navigation
   wt env copy/sync           # environment management
   wt project init/add        # project management
   ```

2. Interactive picker for branch selection
3. Fuzzy matching for branch names
4. Integration with fzf/skim for selection

## Summary

The CLI is already quite ergonomic with good defaults and short commands. Main improvements would be:
- Adding a few aliases for muscle memory from other tools
- Better command discovery and help
- Shell completions
- Keeping the core simplicity while adding power-user features
