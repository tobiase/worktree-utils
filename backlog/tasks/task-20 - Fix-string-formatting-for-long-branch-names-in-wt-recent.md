---
id: task-20
title: Fix string formatting for long branch names in wt recent
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
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

## Implementation Plan

1. Analyze current displayBranches function formatting
2. Calculate maximum widths dynamically based on actual branch data
3. Implement truncation with ellipsis for very long branch names
4. Ensure proper alignment with variable-width content
5. Add tests for long branch name formatting
6. Test with real repositories containing long branch names

## Implementation Notes

Implemented dynamic width calculation for branch name display:

- Calculate column widths based on actual content (using rune count for proper Unicode handling)
- Set reasonable maximum widths (40 chars for branch, 50 for subject, 20 for date)
- Implement truncation with ellipsis for content exceeding max width
- Ensure proper alignment with variable-width content
- Added comprehensive tests for truncation and formatting

Technical decisions:
- Used rune-based string handling for proper Unicode support
- Dynamic width calculation adapts to content while maintaining readability
- Reasonable max widths prevent overly wide output on large screens

Modified files:
- cmd/wt/main.go: Enhanced displayBranches with dynamic formatting, added truncateWithEllipsis
- cmd/wt/recent_test.go: Added TestTruncateWithEllipsis and TestDisplayBranchesFormatting
