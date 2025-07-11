---
id: task-2
title: Implement core wt recent command handler
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-1
---

## Description

Create the main handler function for the wt recent command that displays recently active branches. This will use the git interface methods to query and display branches.

## Acceptance Criteria

- [x] handleRecentCommand function created in main.go
- [x] Command added to runCommand switch statement
- [x] Basic functionality to list recent branches works
- [x] Output formatted as numbered list similar to wt list
- [x] Shows all branches including those with existing worktrees
- [x] Detects which branches have worktrees and which don't

## Implementation Plan

1. Create handleRecentCommand function in main.go
2. Add 'recent' case to runCommand switch statement
3. Use git ForEachRef to get branches sorted by committerdate
4. Get list of worktrees to check which branches have worktrees
5. Format output similar to wt list with numbered entries
6. Display branch name, relative date, commit subject, and author

## Implementation Notes

Implemented core functionality for the wt recent command:
- Created handleRecentCommand that uses git ForEachRef to get branches sorted by committer date
- Added import for git package and added "recent" case to runCommand switch
- Shows numbered list with worktree indicator (*), branch name, relative date, commit subject, and author
- Successfully detects which branches have worktrees using git.WorktreeList()
- Default limit of 10 branches (will be configurable in task-3)
- TODOs added for flags support (task-3) and numeric navigation (task-4)
