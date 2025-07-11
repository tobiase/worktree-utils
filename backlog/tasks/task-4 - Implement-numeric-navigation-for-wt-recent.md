---
id: task-4
title: Implement numeric navigation for wt recent
status: To Do
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

- [ ] wt recent 1 navigates to most recent branch
- [ ] wt recent N navigates to Nth recent branch
- [ ] If branch has worktree navigates to worktree directory (CD:)
- [ ] If branch has no worktree switches branch in current directory
- [ ] Numeric argument properly parsed and validated
- [ ] Error handling for invalid indices
