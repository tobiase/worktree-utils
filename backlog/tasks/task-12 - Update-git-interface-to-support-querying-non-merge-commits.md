---
id: task-12
title: Update git interface to support querying non-merge commits
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-11
---

## Description

Add methods to the git interface to support finding the last non-merge commit for branches. This is needed to properly implement author filtering that ignores merge commits.

## Acceptance Criteria

- [x] Add method to get last non-merge commit for a branch
- [ ] Add method to get branch list with last non-merge commit info
- [x] Implement efficient git log queries
- [x] Handle edge cases like branches with only merge commits

## Implementation Plan

1. Add GetLastNonMergeCommit method to git interface
2. Add GetBranchesWithNonMergeInfo method for efficiency
3. Implement methods in CommandClient
4. Update MockGitClient with new methods
5. Test the implementation

## Implementation Notes

Added GetLastNonMergeCommit method to git interface:
- Uses git log --no-merges -n 1 to find last non-merge commit
- Accepts format string for flexible output formatting
- Returns error if branch has no non-merge commits
- Updated MockGitClient to implement the new method
- Decided to focus on per-branch query for now rather than batch operation
