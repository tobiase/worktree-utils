---
id: task-21
title: Improve flag parsing in wt commands
status: To Do
assignee: []
created_date: '2025-07-11'
labels: []
dependencies: []
---

## Description

The manual flag parsing in handleRecentCommand and other commands works but could be more robust. Consider using a flag parsing library or creating a reusable flag parsing utility to handle complex commands more reliably.

## Acceptance Criteria

- [ ] Evaluate flag parsing libraries (e.g. pflag)
- [ ] Create consistent flag parsing utility if needed
- [ ] Handle edge cases like flags after positional args
- [ ] Improve error messages for invalid flag usage
- [ ] Apply to all commands for consistency
