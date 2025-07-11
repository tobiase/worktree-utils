---
id: task-20
title: Fix string formatting for long branch names in wt recent
status: To Do
assignee: []
created_date: '2025-07-11'
labels: []
dependencies: []
---

## Description

The fixed-width formatting in the recent command output might break with very long branch names. The current format uses hard-coded widths that could cause misalignment or truncation issues.

## Acceptance Criteria

- [ ] Dynamic width calculation implemented
- [ ] Long branch names handled gracefully
- [ ] Output remains readable and aligned
- [ ] Consider truncation with ellipsis for very long names
