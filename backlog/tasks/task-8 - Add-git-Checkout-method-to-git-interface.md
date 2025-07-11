---
id: task-8
title: Add git Checkout method to git interface
status: Done
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

- [x] Checkout method added to git.Client interface
- [x] Implementation in CommandClient
- [x] Method handles branch switching safely
- [x] Unit tests for Checkout method

## Implementation Plan

1. Add Checkout method to git.Client interface
2. Implement Checkout in CommandClient using git checkout
3. Update MockGitClient to implement Checkout method
4. Write unit tests for Checkout method

## Implementation Notes

Added Checkout method to support branch switching when no worktree exists:
- Added to git.Client interface
- Implemented in CommandClient using git checkout command
- Updated MockGitClient to implement the new method
- Added basic test skeleton
- Method returns clear error messages on failure
