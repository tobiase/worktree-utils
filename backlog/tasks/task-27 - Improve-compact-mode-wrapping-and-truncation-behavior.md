---
id: task-27
title: Improve compact mode wrapping and truncation behavior
status: To Do
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels:
  - low-priority
  - formatting
dependencies: []
---

## Description

The current truncation in compact mode doesn't make sense when terminals already wrap text naturally. Improve the behavior to either wrap properly or make truncation configurable, especially for narrow terminal windows.

## Acceptance Criteria

- [ ] Remove or make truncation optional in compact mode
- [ ] Consider terminal width for smart wrapping
- [ ] Add --no-truncate flag for scripts
- [ ] Make truncation width configurable
