---
id: task-23
title: Add interactive mode for wt recent branch selection
status: Won't Do
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies: []
---

## Description

Add an interactive mode to wt recent that allows users to select a branch from the list using arrow keys, similar to the fuzzy matching feature. This would provide a better UX when users want to browse through their recent branches.

## Acceptance Criteria

- [ ] Interactive selection with arrow keys
- [ ] Search/filter within the list
- [ ] Show branch details on hover/selection
- [ ] Integrate with existing interactive utilities
- [ ] Work with --all and --others flags

## Implementation Notes

**Decision: Won't implement**

Based on previous experience with interactive mode, it didn't work well for this use case. The current implementation with numeric navigation (e.g., `wt recent 2`) provides a fast and efficient workflow that better fits the command-line nature of the tool.

Users who need interactive branch selection can use the existing fuzzy matching functionality in other commands or pipe the output to external tools like fzf for interactive selection.
