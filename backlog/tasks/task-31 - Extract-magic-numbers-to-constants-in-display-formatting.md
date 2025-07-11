---
id: task-31
title: Extract magic numbers to constants in display formatting
status: To Do
assignee: []
created_date: '2025-07-11'
labels:
  - code-quality
  - low-priority
dependencies: []
---

## Description

The displayBranchesCompact function uses magic numbers for default and maximum widths. These should be extracted as named constants for better maintainability and clarity.

## Acceptance Criteria

- [ ] Extract default widths as constants (15 10 30)
- [ ] Extract maximum widths as constants (40 50 20)
- [ ] Apply to displayBranchesCompact function
- [ ] Consider other magic numbers in codebase
