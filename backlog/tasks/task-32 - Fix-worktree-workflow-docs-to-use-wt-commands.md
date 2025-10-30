---
id: task-32
title: Fix worktree workflow docs to use wt commands
status: Done
assignee:
  - '@Claude'
created_date: '2025-10-30 22:37'
updated_date: '2025-10-30 22:45'
labels: []
dependencies: []
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The docs/WORKTREE_WORKFLOW.md file currently shows git commands as primary examples and wt commands as commented alternatives. This is bad promotion for our own tool. We need to flip this: wt commands should be primary, with git commands as fallback/alternatives for understanding what happens under the hood.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 All worktree creation examples use 'wt new' as primary command
- [x] #2 All worktree removal examples use 'wt rm' as primary command
- [x] #3 Git commands shown as alternatives/equivalents for understanding
- [x] #4 Examples are technically correct (wt new takes only branch name, no path)
- [x] #5 Documentation clearly demonstrates wt tool benefits over raw git

- [x] #6 Content moved from docs/WORKTREE_WORKFLOW.md to CLAUDE.md between markers
- [x] #7 Worktree section in CLAUDE.md is OUTSIDE backlog markers (currently inside)
- [x] #8 docs/WORKTREE_WORKFLOW.md file deleted (no longer needed)
- [x] #9 Markers added: <!-- WORKTREE WORKFLOW START/END -->
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. Update docs/WORKTREE_WORKFLOW.md to use wt commands as primary (wt new, wt rm, wt integrate)
2. Rewrite examples to show wt benefits (no paths to remember, convention-based)
3. Add git command equivalents as 'Alternative' or 'Under the hood' sections
4. Find correct insertion point in CLAUDE.md (before backlog markers, after existing content)
5. Copy updated workflow content into CLAUDE.md with <!-- WORKTREE WORKFLOW START/END --> markers
6. Delete docs/WORKTREE_WORKFLOW.md
7. Remove the incorrect worktree summary from inside backlog markers (lines 885-893)
8. Verify markers are properly placed and content is correct
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Successfully updated worktree workflow documentation to promote wt commands.

Changes made:
1. Updated all worktree creation examples to use 'wt new task-<id>-<slug>' as primary command
2. Updated all cleanup examples to use 'wt rm <branch> --branch' as primary
3. Added collapsible <details> sections showing git command equivalents as 'Alternative: Using git directly'
4. Emphasized wt benefits: 'No paths to remember!', 'One command, done.'
5. Moved entire workflow content from docs/WORKTREE_WORKFLOW.md into CLAUDE.md between <!-- WORKTREE WORKFLOW START/END --> markers
6. Placed worktree section BEFORE backlog guidelines (lines 699-806), not inside them
7. Removed incorrect worktree summary that was nested inside backlog markers
8. Deleted docs/WORKTREE_WORKFLOW.md as it's no longer needed

Technical decisions:
- Used HTML <details> tags for git alternatives to keep wt commands prominent while still showing what happens under the hood
- Positioned worktree workflow section before backlog guidelines for better visibility
- Single source of truth: all worktree workflow documentation now lives in CLAUDE.md with clear markers for future updates
<!-- SECTION:NOTES:END -->
