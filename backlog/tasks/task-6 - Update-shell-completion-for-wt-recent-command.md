---
id: task-6
title: Update shell completion for wt recent command
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-3
---

## Description

Add shell completion support for the recent command and its flags to improve user experience in bash and zsh.

## Acceptance Criteria

- [x] recent command added to completion list
- [x] Flag completion for --me --others and -n
- [x] Completion works in both bash and zsh
- [x] Completion follows existing patterns

## Implementation Plan

1. Add 'recent' to the commands list in completion data
2. Add flag completion for --me, --others, and -n
3. Test completion in both bash and zsh

## Implementation Notes

Added shell completion support for wt recent command:
- Added recent command to getCoreCommands() in completion.go
- Included all flags with proper HasValue settings
- Defined optional index argument as ArgString type
- Verified completion generates correctly for bash
- Follows existing command structure patterns
