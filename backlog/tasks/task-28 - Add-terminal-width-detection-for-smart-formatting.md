---
id: task-28
title: Add terminal width detection for smart formatting
status: To Do
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels:
  - low-priority
  - enhancement
dependencies: []
---

## Description

Detect terminal width to make intelligent decisions about when to truncate or wrap text. This would improve the display on various terminal sizes without manual configuration.

## Acceptance Criteria

- [ ] Detect terminal width using appropriate system calls
- [ ] Use width to determine truncation behavior
- [ ] Graceful fallback when width cannot be detected
- [ ] Work correctly with terminal resizing
