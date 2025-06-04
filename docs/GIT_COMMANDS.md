# Git Commands Reference

This document describes the actual git commands executed by each `wt` command.

## Core Commands

### `wt list`
Lists all worktrees in the repository.

**Git command:**
```bash
git -C <repo-root> worktree list --porcelain
```

**Notes:**
- Uses `--porcelain` format for reliable parsing
- Executed from the repository root
- Parses output to extract path and branch information

### `wt add <branch>`
Creates a new worktree for an existing branch.

**Git command:**
```bash
git -C <repo-root> worktree add <worktree-base>/<branch> <branch>
```

**Example:**
```bash
# For branch "feature-x" in repo "/path/to/myproject"
git -C /path/to/myproject worktree add /path/to/myproject-worktrees/feature-x feature-x
```

**Notes:**
- Creates worktree directory following the `<repo>-worktrees/<branch>` convention
- The branch must already exist
- Will fail if worktree already exists

### `wt rm <branch>`
Removes a worktree by branch name or path.

**Git command:**
```bash
git -C <repo-root> worktree remove <worktree-path>
```

**Notes:**
- First tries to find worktree by branch name
- Falls back to treating the argument as a path
- Works with worktrees that don't follow the standard pattern
- The worktree path must be the exact path shown in `git worktree list`

### `wt new <branch> [--base <base-branch>]`
Creates a new branch and its worktree.

**Git commands:**

Without base branch (creates from current HEAD):
```bash
git -C <repo-root> worktree add <worktree-base>/<branch> <branch>
```

With base branch:
```bash
git -C <repo-root> worktree add <worktree-base>/<branch> -b <branch> <base-branch>
```

**Example:**
```bash
# Create new branch "feature-y" from "main"
git -C /path/to/myproject worktree add /path/to/myproject-worktrees/feature-y -b feature-y main
```

**Notes:**
- The `-b` flag creates a new branch
- Without `--base`, creates branch from current HEAD
- Combines branch creation and worktree addition

## Utility Commands

### `wt go [index|branch]`
Changes directory to a worktree. No git commands are executed.

**Notes:**
- Uses `git worktree list --porcelain` to find worktrees
- Returns path for shell wrapper to execute `cd`

### `wt env-copy <branch> [--recursive]`
Copies .env files between worktrees. No git commands are executed.

**Notes:**
- Pure file operations
- Preserves file permissions
- Creates target directories as needed

### Repository Information Commands

These commands are used internally by various `wt` operations:

#### Get Repository Root
```bash
git rev-parse --show-toplevel
```

#### Get Remote URL
```bash
git remote get-url origin
```

#### List Worktrees
```bash
git worktree list --porcelain
```

## Error Handling

All git commands are executed with:
- Working directory set to repository root using `-C <repo-root>`
- stdout and stderr passed through to the user
- Exit codes propagated to the caller

## Worktree Organization

The tool enforces a convention for organizing worktrees:
- Main repository: `/path/to/myproject`
- Worktrees: `/path/to/myproject-worktrees/<branch-name>`

This convention:
- Keeps worktrees organized and discoverable
- Prevents clutter in the repository parent directory
- Makes it easy to identify which worktrees belong to which repository

However, the tool also works with existing worktrees that don't follow this pattern, especially for the `rm` and `go` commands.