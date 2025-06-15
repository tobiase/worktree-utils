# CLI Ergonomics Roadmap: "Do What I Mean" Design

## Executive Summary

This document outlines a comprehensive plan to transform `wt` from a tool that requires precise knowledge into a tool that reads user intent. The core philosophy is **"Do What I Mean" (DWIM)** - users should express their intent, and the tool should figure out the implementation details.

### The Transformation Vision

**From:** Cognitive overhead and precise syntax requirements
**To:** Intent-based commands that "just work"

**Key Principle:** Every command should accept fuzzy input, provide smart defaults, and fall back to interactive selection when ambiguous.

## Current State Analysis

### What Works Well ✅
- Short command name (`wt`)
- Index-based navigation (`wt go 0`)
- Interactive command selection (`wt` with no args)
- Comprehensive fuzzy finding with `--fuzzy` flag
- Good alias support (`ls`, `switch`, `s`)
- Smart repo root default (`wt go` with no args)

### Pain Points Identified 🔴

1. **Command Duplication**: `add` vs `new` requires users to know branch state
2. **Exact Name Requirements**: `wt go mai` fails instead of matching `main`
3. **No Command-Specific Help**: `wt go --help` tries to switch to "--help" branch
4. **Inconsistent Error Messages**: No suggestions when commands fail
5. **Environment Management**: `env-copy` is awkward and limited
6. **Help System Gaps**: No per-command help or better discovery

## Smart Command Patterns

### The Template: Smart `wt new`

**Current Problem:**
- `wt add <existing-branch>` - Create worktree for existing branch
- `wt new <new-branch>` - Create branch + worktree + switch

**Smart Solution:**
```bash
wt new <branch>  # Always works, regardless of branch state:
                 # 1. Branch doesn't exist → Create branch + worktree + switch
                 # 2. Branch exists, no worktree → Create worktree + switch
                 # 3. Branch + worktree exist → Just switch
```

**Benefits:**
- Idempotent (safe to run multiple times)
- Single command for one intent ("I want to work on this branch")
- Zero cognitive overhead about current state

### Applying This Pattern

This template should be applied to:
1. **Branch name resolution** - All commands accepting branch names
2. **Help system** - All commands should accept `--help`
3. **Environment management** - Unified approach to env operations
4. **Removal operations** - Smart safety and suggestions

## Detailed Improvement Plan

### Phase 1: Smart Branch Matching (Highest Impact)

#### 1.1 Fuzzy Branch Name Resolution
**Implementation:** Add fuzzy matching to all branch-accepting commands

```bash
# Current behavior
wt go mai → Error: "branch 'mai' not found"

# New behavior
wt go mai → Auto-switches to 'main' (unambiguous match)
wt go te  → Interactive picker: [test-branch, test-feature]
wt go fea → Auto-switches to 'feature-branch' (only match)
```

**Technical Approach:**
- Add `resolveBranchName(input string) (string, error)` function
- Use Levenshtein distance or similar fuzzy matching
- Auto-select if only one match within threshold
- Fall back to interactive picker for multiple matches
- Apply to: `go`, `rm`, `env-copy`, `new`

#### 1.2 Smart Index Shortcuts
**Implementation:** Direct index access without `go` command

```bash
wt 0  # Equivalent to wt go 0
wt 1  # Equivalent to wt go 1
wt 2  # Equivalent to wt go 2
```

**Technical Approach:**
- Detect numeric arguments in main command parsing
- Route directly to `handleGoCommand` with index

### Phase 2: Unified Smart Commands

#### 2.1 Smart `wt new` Implementation
**Behavior Matrix:**

| Branch State | Worktree State | Action |
|--------------|----------------|--------|
| Doesn't exist | N/A | Create branch + worktree + switch |
| Exists | Doesn't exist | Create worktree + switch |
| Exists | Exists | Switch to existing worktree |

**Implementation:**
```go
func SmartNewWorktree(branch string, baseBranch string) (string, error) {
    // 1. Check if branch exists
    branchExists := checkBranchExists(branch)

    // 2. Check if worktree exists
    worktreeExists := checkWorktreeExists(branch)

    // 3. Execute appropriate action
    if !branchExists {
        return createBranchAndWorktree(branch, baseBranch)
    } else if !worktreeExists {
        return createWorktreeForExistingBranch(branch)
    } else {
        return switchToExistingWorktree(branch)
    }
}
```

#### 2.2 Smart Help System
**Implementation:** Recognize help flags in all commands

```bash
wt go --help     # Show detailed help for 'go' command
wt new -h        # Show detailed help for 'new' command
wt rm --help     # Show detailed help for 'rm' command
```

**Technical Approach:**
- Add help flag detection before command execution
- Create detailed help text for each command
- Include examples and option explanations

#### 2.3 Smart Environment Management
**Current:** `env-copy` with limited functionality
**New:** Unified `env` subcommand system

```bash
wt env sync <target>     # Copy current .env to target worktree
wt env sync --all        # Sync .env to all worktrees
wt env diff <target>     # Show env differences
wt env                   # Interactive picker for env operations
```

### Phase 3: Enhanced User Experience

#### 3.1 Smart Error Messages with Suggestions
**Current:**
```
wt: branch 'mai' not found among worktrees
```

**Enhanced:**
```
wt: branch 'mai' not found. Did you mean:
  1. main
  2. maintenance
  3. mailbox-feature
```

#### 3.2 Smart `rm` with Safety Features
```bash
wt rm feat           # Fuzzy match + confirmation: "Remove 'feature-branch'? [y/N]"
wt rm                # Interactive picker of removable worktrees
wt rm --merged       # Remove all worktrees for branches merged to main
wt rm --dry-run      # Show what would be removed
```

#### 3.3 Enhanced Discovery
```bash
wt commands          # List all available commands (including project-specific)
wt help <command>    # Detailed help for specific command
wt --help            # Enhanced main help with examples
```

## Implementation Guidelines

### Design Principles

1. **Idempotent Operations:** Commands should be safe to run multiple times
2. **Progressive Enhancement:** Start with basic fuzzy matching, add intelligence over time
3. **Clear Feedback:** Always tell users what action was taken
4. **Graceful Fallbacks:** When automation fails, fall back to interactive selection
5. **Consistent Patterns:** Apply DWIM principles uniformly across all commands

### Error Handling Strategy

1. **Fuzzy Matching:** Try to resolve ambiguous input intelligently
2. **Interactive Fallback:** When multiple options exist, show picker
3. **Helpful Suggestions:** When fuzzy matching fails, suggest similar options
4. **Clear Context:** Error messages should explain what was expected

### Backwards Compatibility

1. **Keep Existing Behavior:** All current command syntax continues to work
2. **Gradual Migration:** Add new smart behaviors alongside existing ones
3. **Deprecation Path:** For replaced commands (like `add`), provide migration warnings

## User Experience Transformation

### Before: Cognitive Overhead
```bash
# User mental process:
# "I want to work on a feature branch..."
# "Does the branch exist? Do I need 'add' or 'new'?"
# "What's the exact branch name?"
# "Does a worktree already exist?"

git branch -a | grep feature  # Check if branch exists
wt list | grep feature        # Check if worktree exists
wt add feature-branch         # Hope I chose the right command
```

### After: Intent-Based Workflow
```bash
# User mental process:
# "I want to work on a feature branch"

wt new feat                   # Tool figures out the rest
# → "Switched to existing worktree 'feature-branch'"
```

### Common Workflow Improvements

#### Starting New Work
**Before:**
```bash
git checkout main
git pull
git checkout -b feature-xyz
wt new feature-xyz
```

**After:**
```bash
wt new feature-xyz --from main  # Creates from main, handles pull automatically
```

#### Quick Switching
**Before:**
```bash
wt list                       # Find the index
wt go 2                       # Switch by index
```

**After:**
```bash
wt 2                         # Direct index access
# OR
wt mai                       # Fuzzy match to 'main'
```

#### Environment Sync
**Before:**
```bash
cp .env ../project-worktrees/feature-branch/.env
```

**After:**
```bash
wt env sync feature          # Smart sync with fuzzy matching
```

## Success Metrics

### Quantitative Goals
- **Reduce command failures** by 80% through fuzzy matching
- **Eliminate `add` vs `new` confusion** through unified smart command
- **Reduce keystrokes** for common operations by 30%

### Qualitative Goals
- **Zero learning curve** for basic operations
- **Predictable behavior** - commands should "just work"
- **Self-discovering** - users can explore functionality without reading docs

## Implementation Phases

### Phase 1: Foundation (Week 1-2)
- [ ] Smart `wt new` implementation
- [ ] Basic fuzzy branch name matching
- [ ] Command-specific help system
- [ ] Direct index shortcuts (`wt 0`, `wt 1`)

### Phase 2: Enhancement (Week 3-4)
- [ ] Unified environment management (`wt env`)
- [ ] Smart error messages with suggestions
- [ ] Enhanced `rm` with safety features
- [ ] Better command discovery

### Phase 3: Polish (Week 5-6)
- [ ] Advanced fuzzy matching algorithms
- [ ] Comprehensive testing of edge cases
- [ ] Documentation updates
- [ ] Migration guides for deprecated features

## Conclusion

This roadmap transforms `wt` from a precise-syntax tool into an intent-based assistant. By applying the "Do What I Mean" philosophy consistently, we eliminate cognitive overhead and make worktree management truly frictionless.

The key insight is that users have goals ("work on this feature", "clean up old branches", "sync environments") rather than specific technical operations. Our CLI should understand these goals and handle the implementation details automatically.
