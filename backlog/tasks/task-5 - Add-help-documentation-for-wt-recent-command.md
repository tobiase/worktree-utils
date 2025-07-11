---
id: task-5
title: Add help documentation for wt recent command
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-3
  - task-4
---

## Description

Integrate the recent command with the existing help system to provide users with usage information and examples.

## Acceptance Criteria

- [x] Help content created for recent command
- [x] Help flag support added to handleRecentCommand
- [x] Documentation includes usage examples
- [x] Documentation covers all flags and options

## Implementation Plan

1. Add help documentation in help/help.go for recent command
2. Include all flags and options
3. Add usage examples
4. Ensure handleRecentCommand already has help flag support

## Implementation Notes

Added comprehensive help documentation for wt recent command:
- Added entry to commandHelpMap in help/help.go
- Included all flags (--me, --others, -n) with descriptions and examples
- Added multiple usage examples showing different scenarios
- Verified handleRecentCommand already has help flag support
- Updated main usage message to include recent command
- Cross-referenced with related commands (list, go, new)
