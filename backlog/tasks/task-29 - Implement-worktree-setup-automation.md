---
id: task-29
title: Implement worktree setup automation
status: To Do
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels:
  - future
  - automation
dependencies: []
---

## Description

Add the ability to define and run automatic setup steps when creating new worktrees. This could include running npm install, syncing env files, and other project-specific initialization tasks.

## Acceptance Criteria

- [ ] Define setup steps in project config
- [ ] Run setup automatically on wt new
- [ ] Support various command types
- [ ] Handle setup failures gracefully
- [ ] Optional manual trigger with wt setup
