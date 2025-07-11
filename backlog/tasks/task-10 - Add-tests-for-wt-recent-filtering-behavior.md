---
id: task-10
title: Add tests for wt recent filtering behavior
status: To Do
assignee: []
created_date: '2025-07-11'
labels: []
dependencies:
  - task-9
---

## Description

Add tests to verify that --me and --others flags return the requested number of branches after filtering, not before. This ensures the fix for the filtering issue works correctly.

## Acceptance Criteria

- [ ] Test --me returns requested count of user's branches
- [ ] Test --others returns requested count of other users' branches
- [ ] Test behavior when fewer branches exist than requested
- [ ] Test edge cases with mixed authorship
