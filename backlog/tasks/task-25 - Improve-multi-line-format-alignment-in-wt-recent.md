---
id: task-25
title: Improve multi-line format alignment in wt recent
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels:
  - high-priority
  - formatting
dependencies: []
---

## Description

Fix alignment issues in the multi-line display format. The star should be directly attached to the branch name, and subsequent lines should align with the branch name, not the index. Add blank lines between entries for better readability.

## Acceptance Criteria

- [x] Star directly next to branch name (0: *branch)
- [x] Second/third lines align with branch name
- [x] Blank line between entries
- [x] No truncation in multi-line mode

## Implementation Plan

1. Modify displayBranches function to show star directly after colon
2. Ensure proper indentation for second and third lines
3. Add blank lines between all entries (except after last)
4. Update tests to match new format

## Implementation Notes

Successfully improved the multi-line format alignment:

- Star is now directly attached to branch name: `0: *branch-name`
- Second and third lines are indented with 3 spaces to align with branch name
- Blank lines added between all entries for better readability
- No truncation in multi-line mode (full branch names shown)
- Tests updated to verify the new format

The format is now much cleaner and easier to scan, especially with long branch names.
