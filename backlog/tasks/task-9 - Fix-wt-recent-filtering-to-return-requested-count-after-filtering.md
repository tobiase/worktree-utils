---
id: task-9
title: Fix wt recent filtering to return requested count after filtering
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies: []
---

## Description

The current implementation fetches N branches and then filters, resulting in fewer than N branches when using --me or --others. Instead, it should fetch enough branches to ensure N results after filtering.

## Acceptance Criteria

- [x] Fetch more branches when --me or --others is used
- [x] Filter branches by author
- [x] Limit to requested count after filtering
- [x] --me returns N branches by current user
- [x] --others returns N branches by other users
- [x] Works correctly when there are fewer matching branches than requested

## Implementation Plan

1. Modify handleRecentCommand to fetch more branches when filtering
2. Apply author filtering before limiting count
3. Ensure correct count is returned after filtering
4. Handle edge case when fewer branches exist than requested
5. Test the fix manually

## Implementation Notes

Fixed the filtering behavior to return the requested count after filtering:
- When --me or --others is used, fetch count * 10 branches (minimum 100)
- Apply author filtering as before
- Limit display to requested count after filtering
- This ensures users get N branches of their chosen type, not N branches total
- Handles cases where fewer matching branches exist gracefully
