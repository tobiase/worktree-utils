---
id: task-2
title: Implement core wt recent command handler
status: To Do
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

- [ ] handleRecentCommand function created in main.go
- [ ] Command added to runCommand switch statement
- [ ] Basic functionality to list recent branches works
- [ ] Output formatted as numbered list similar to wt list
- [ ] Shows all branches including those with existing worktrees
- [ ] Detects which branches have worktrees and which don't
