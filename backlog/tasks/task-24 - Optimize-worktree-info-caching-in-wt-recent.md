---
id: task-24
title: Optimize worktree info caching in wt recent
status: Won't Do
updated_date: '2025-07-11'
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

## Implementation Notes

**Decision: Won't implement**

Based on performance benchmarks from task-22, caching is not needed:

- Branch filtering performance is excellent (< 0.2ms for 10,000 branches)
- The bottleneck is git operations (`git for-each-ref`, `git worktree list`), not our Go code
- Caching worktree info in memory wouldn't help since:
  - We still need to call git commands each time
  - The Go processing is already extremely fast
  - Most users won't have hundreds of worktrees (typical usage is < 10)
- Adding caching would increase complexity without meaningful performance gains

The real bottleneck is git command execution time, which caching cannot improve.
