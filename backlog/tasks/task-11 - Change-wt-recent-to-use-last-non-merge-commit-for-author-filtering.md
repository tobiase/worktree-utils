---
id: task-11
title: Change wt recent to use last non-merge commit for author filtering
status: To Do
assignee: []
created_date: '2025-07-11'
labels: []
dependencies: []
---

## Description

Currently wt recent uses the last commit author, which can be a merge commit. This causes branches to incorrectly appear under different authors when someone merges main into a feature branch. Change to use the last non-merge commit author for filtering.

## Acceptance Criteria

- [ ] Replace git for-each-ref with logic that finds last non-merge commit per branch
- [ ] Use last non-merge commit author for --me and --others filtering
- [ ] Display information from last non-merge commit not last commit
- [ ] Maintain performance with efficient git commands
- [ ] Handle branches that only have merge commits gracefully
