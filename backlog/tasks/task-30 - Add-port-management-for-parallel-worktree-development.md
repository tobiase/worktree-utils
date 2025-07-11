---
id: task-30
title: Add port management for parallel worktree development
status: To Do
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels:
  - future
  - enhancement
dependencies: []
---

## Description

When working with multiple worktrees in parallel, port conflicts are common. Implement a system to track and allocate unique ports for each worktree to enable true parallel development.

## Acceptance Criteria

- [ ] Track ports in use per worktree
- [ ] Allocate unique ports on setup
- [ ] Configure services with allocated ports
- [ ] Show port allocation with wt ports
- [ ] Free ports when worktree removed
