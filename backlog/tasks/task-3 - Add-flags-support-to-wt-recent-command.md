---
id: task-3
title: Add flags support to wt recent command
status: To Do
assignee: []
created_date: '2025-07-11'
labels: []
dependencies:
  - task-2
---

## Description

Implement flag parsing for --me, --others, and -n flags to filter and limit the recent branches output.

## Acceptance Criteria

- [ ] Parse --me flag to show only current user branches
- [ ] Parse --others flag to show only other users branches
- [ ] Parse -n flag to limit number of results
- [ ] Default limit of 10 when -n not specified
- [ ] Flags work correctly in combination
