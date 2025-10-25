---
id: task-31
title: Add integrate command and branch cleanup
status: To Do
assignee: []
created_date: '2025-10-25'
updated_date: '2025-10-25'
labels:
  - enhancement
dependencies: []
---

## Description

Implement the missing integrations between `wt` and our disciplined worktree workflow:

1. Extend `wt rm` with a flag that removes the associated Git branch once the worktree goes away, so operators can tear down a task in one step.
2. Add a first-class `wt integrate <worktree>` command that automates the documented “rebase, fast-forward merge, clean up” checklist for finishing a task.

## Acceptance Criteria

- [ ] `wt rm --branch` deletes the branch only when it is fully merged; it refuses with a clear message otherwise unless `--force` is provided.
- [ ] Attempting to delete a branch that still has other worktrees attached or is the current branch results in a helpful error without removing anything.
- [ ] `wt integrate <worktree>` fetches, rebases the worktree branch onto `origin/main`, fast-forward merges into `main`, and on success removes both the worktree and branch.
- [ ] Integration aborts cleanly with actionable output when the rebase conflicts or the merge cannot fast-forward.
- [ ] CLI help + docs (`CLAUDE.md`, `docs/WORKTREE_WORKFLOW.md`, and command help) describe the new flag and integrate flow.
- [ ] Unit/integration tests cover the new flag behavior and integration command happy/error paths.
