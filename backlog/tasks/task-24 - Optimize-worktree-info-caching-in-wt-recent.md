---
id: task-24
title: Optimize worktree info caching in wt recent
status: To Do
assignee: []
created_date: '2025-07-11'
labels: []
dependencies:
  - task-22
---

## Description

The updateWorktreeInfo function in wt recent could be expensive with many worktrees. Consider caching worktree list results and only calling when necessary for display to improve performance.

## Acceptance Criteria

- [ ] Cache worktree list to avoid repeated git worktree list calls
- [ ] Only update worktree info when displaying branches with worktrees
- [ ] Measure performance improvement with many worktrees
- [ ] Ensure cache invalidation when worktrees change
