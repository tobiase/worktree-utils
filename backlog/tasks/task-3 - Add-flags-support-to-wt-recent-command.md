---
id: task-3
title: Add flags support to wt recent command
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies:
  - task-2
---

## Description

Implement flag parsing for --me, --others, and -n flags to filter and limit the recent branches output.

## Acceptance Criteria

- [x] Parse --me flag to show only current user branches
- [x] Parse --others flag to show only other users branches
- [x] Parse -n flag to limit number of results
- [x] Default limit of 10 when -n not specified
- [x] Flags work correctly in combination

## Implementation Plan

1. Parse --me, --others, and -n flags from args
2. Get current user name from git config when --me or --others is used
3. Add filter to for-each-ref command based on author when needed
4. Use -n value to control --count parameter (default 10)
5. Handle flag combinations properly

## Implementation Notes

Added full flags support to the wt recent command:
- Implemented flag parsing loop that handles --me, --others, and -n flags
- Get current user from git config user.name when filtering by author
- Filter branches after fetching based on author name comparison
- -n flag controls the --count parameter passed to git for-each-ref
- Validates that --me and --others are not used together
- Default limit remains 10 when -n is not specified
- Maintains display index for consistent numbering after filtering
