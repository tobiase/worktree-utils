---
id: task-4
title: Implement numeric navigation for wt recent
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-2
  - task-8
---

## Description

Allow users to quickly navigate to a recent branch by specifying its index number, similar to how wt list works with numeric selection.

## Acceptance Criteria

- [x] wt recent 1 navigates to most recent branch
- [x] wt recent N navigates to Nth recent branch
- [x] If branch has worktree navigates to worktree directory (CD:)
- [x] If branch has no worktree switches branch in current directory
- [x] Numeric argument properly parsed and validated
- [x] Error handling for invalid indices

## Implementation Plan

1. Check if first non-flag argument is numeric
2. If numeric, run the query to get branches
3. Find the branch at the specified index
4. Check if branch has a worktree
5. If has worktree, navigate to it (CD:)
6. If no worktree, checkout the branch
7. Handle invalid indices gracefully

## Implementation Notes

Implemented numeric navigation for wt recent command:
- Parse numeric arguments during flag parsing
- Store branches in a slice for indexed access
- Handle navigation before display when numeric index provided
- If branch has worktree, output CD: command to navigate
- If no worktree, use git Checkout to switch branches
- Validate index is within bounds of available branches
- Works correctly with filtering flags (--me, --others)
