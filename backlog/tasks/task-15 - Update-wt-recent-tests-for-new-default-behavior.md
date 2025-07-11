---
id: task-15
title: Update wt recent tests for new default behavior
status: To Do
assignee: []
created_date: '2025-07-11'
labels: []
dependencies:
  - task-13
  - task-14
---

## Description

With the default behavior changing to show only current user's branches and the removal of --me flag, tests need to be updated to reflect the new behavior and flag structure.

## Acceptance Criteria

- [ ] Update existing tests to expect default filtering
- [ ] Remove tests for --me flag
- [ ] Add tests for --all flag
- [ ] Test that numeric navigation respects active filtering
- [ ] Ensure flag combinations work correctly
