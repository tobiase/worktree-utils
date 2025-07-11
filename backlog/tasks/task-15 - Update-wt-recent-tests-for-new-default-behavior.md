---
id: task-15
title: Update wt recent tests for new default behavior
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-13
  - task-14
---

## Description

With the default behavior changing to show only current user's branches and the removal of --me flag, tests need to be updated to reflect the new behavior and flag structure.

## Acceptance Criteria

- [x] Update existing tests to expect default filtering
- [x] Remove tests for --me flag
- [x] Add tests for --all flag
- [x] Test that numeric navigation respects active filtering
- [x] Ensure flag combinations work correctly

## Implementation Notes

Successfully updated tests to reflect new default behavior:
- Updated TestHandleRecentCommand to test default filtering and --all flag
- Removed references to deprecated --me flag
- Updated TestParseRecentFlags to include --all flag and conflicting flags test
- Updated TestBranchFiltering to test default behavior, --all flag, and --others flag
- Fixed linting issues (formatting and string constant)
