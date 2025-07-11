---
id: task-11
title: Change wt recent to use last non-merge commit for author filtering
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies: []
---

## Description

Currently wt recent uses the last commit author, which can be a merge commit. This causes branches to incorrectly appear under different authors when someone merges main into a feature branch. Change to use the last non-merge commit author for filtering.

## Acceptance Criteria

- [x] Replace git for-each-ref with logic that finds last non-merge commit per branch
- [x] Use last non-merge commit author for --me and --others filtering
- [x] Display information from last non-merge commit not last commit
- [x] Maintain performance with efficient git commands
- [x] Handle branches that only have merge commits gracefully

## Implementation Plan

1. Replace for-each-ref approach with branch list + per-branch queries
2. For each branch, get last non-merge commit info
3. Use that commit's author for filtering
4. Display that commit's info instead of latest commit
5. Handle performance implications
6. Test with branches that have merge commits

## Implementation Notes

Completely rewrote the handleRecentCommand to use last non-merge commits:
- First get all branch names using for-each-ref
- For each branch, query the last non-merge commit using git log --no-merges
- Parse commit info including unix timestamp for proper sorting
- Sort branches by commit timestamp (most recent first)
- Filter by author based on non-merge commit author
- Display shows non-merge commit info, not latest commit
- Branches with only merge commits are skipped gracefully
- Performance impact: one git command per branch, but provides accurate ownership
