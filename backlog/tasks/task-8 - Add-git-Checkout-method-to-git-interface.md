---
id: task-8
title: Add git Checkout method to git interface
status: To Do
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-1
---

## Description

To support switching branches when no worktree exists, we need a Checkout method in the git interface that can switch the current branch.

## Acceptance Criteria

- [ ] Checkout method added to git.Client interface
- [ ] Implementation in CommandClient
- [ ] Method handles branch switching safely
- [ ] Unit tests for Checkout method
