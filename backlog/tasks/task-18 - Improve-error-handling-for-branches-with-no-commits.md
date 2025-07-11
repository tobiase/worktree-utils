---
id: task-18
title: Improve error handling for branches with no commits
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-16
---

## Description

Currently when GetLastNonMergeCommit fails for a branch, it's silently skipped. This could confuse users who expect to see all their branches. Add debug logging or a way to inform users about skipped branches.

## Acceptance Criteria

- [x] Debug logging added for skipped branches
- [x] Option to show count of skipped branches
- [x] Clear indication when branches are filtered due to errors
- [x] No silent failures

## Implementation Plan

1. Analyze current error handling in collectBranchInfo function
2. Add verbose/debug flag to wt recent command
3. Track and count skipped branches during collection
4. Display summary of skipped branches at the end
5. Add debug logging for specific failure reasons
6. Update tests to verify error handling behavior

## Implementation Notes

Implemented comprehensive error handling for branches with no commits:

- Added verbose flag (-v, --verbose) to wt recent command
- Created branchCollectionResult struct to track skipped branches with reasons
- Modified collectBranchInfo to return detailed tracking of skipped branches
- Updated handleRecentCommand to display count and details of skipped branches
- Added verbose flag documentation to help system
- Users now get clear feedback when branches are skipped instead of silent failures

Key implementation decisions:
- Used structured error tracking rather than just logging
- Provided both summary (count) and detailed (with --verbose) feedback
- Maintained backward compatibility with existing behavior

Modified files:
- cmd/wt/main.go: Added verbose flag, branchCollectionResult struct, enhanced error display
- internal/help/help.go: Added --verbose flag documentation
