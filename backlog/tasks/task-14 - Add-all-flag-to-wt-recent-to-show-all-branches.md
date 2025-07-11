---
id: task-14
title: Add --all flag to wt recent to show all branches
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-13
---

## Description

While the default shows only your branches, sometimes you need to see all branches regardless of author. Add an --all flag to show all branches with their last non-merge commit info.

## Acceptance Criteria

- [x] Add --all flag to show all branches
- [x] --all overrides default filtering
- [x] Works with -n flag
- [x] Works with numeric navigation
- [x] Help and completion updated

## Implementation Notes

Successfully implemented --all flag for wt recent command. The flag overrides the default filtering behavior and shows all branches regardless of author. Works correctly with -n flag for count limiting and numeric navigation. Help documentation and shell completion have been updated.
