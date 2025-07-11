---
id: task-13
title: Make wt recent default to showing only current user's branches
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies: []
---

## Description

The most common use case is viewing your own recent branches. Make this the default behavior with no flags required. This makes the tool more intuitive and requires less typing for the primary use case.

## Acceptance Criteria

- [x] Default behavior shows only current user's branches
- [x] No flag needed for filtering to current user
- [x] Remove --me flag entirely
- [x] Existing -n flag continues to work with default behavior
- [x] Help documentation updated to reflect new default

## Implementation Plan

1. Change default behavior to filter by current user
2. Remove --me flag parsing
3. Update flag validation logic
4. Update help documentation
5. Update usage message
6. Test the new default behavior

## Implementation Notes

Successfully changed default behavior:
- Removed --me flag entirely from parsing and help
- Default now shows only current user's branches
- Always get current user unless --all flag is used
- Updated filtering logic to default to current user
- Updated help documentation to reflect new behavior
- Updated main usage message
- Added placeholder for --all flag (to be implemented in task-14)
