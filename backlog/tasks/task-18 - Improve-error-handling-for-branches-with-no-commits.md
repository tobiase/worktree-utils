---
id: task-18
title: Improve error handling for branches with no commits
status: To Do
assignee: []
created_date: '2025-07-11'
labels: []
dependencies:
  - task-16
---

## Description

Currently when GetLastNonMergeCommit fails for a branch, it's silently skipped. This could confuse users who expect to see all their branches. Add debug logging or a way to inform users about skipped branches.

## Acceptance Criteria

- [ ] Debug logging added for skipped branches
- [ ] Option to show count of skipped branches
- [ ] Clear indication when branches are filtered due to errors
- [ ] No silent failures
