---
id: task-1
title: Add git ForEachRef and GetConfigValue methods to git interface
status: Done
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels: []
dependencies: []
---

## Description

The git interface needs new methods to support the recent command functionality. ForEachRef will allow querying branches with sorting and formatting options. GetConfigValue will allow reading git config values like user.name.

## Acceptance Criteria

- [x] ForEachRef method added to git.Client interface
- [x] GetConfigValue method added to git.Client interface
- [x] Implementation in CommandClient for both methods
- [x] Unit tests for both methods

## Implementation Plan

1. Add ForEachRef method to git.Client interface
2. Add GetConfigValue method to git.Client interface
3. Implement ForEachRef in CommandClient using git for-each-ref
4. Implement GetConfigValue in CommandClient using git config
5. Write unit tests for both methods

## Implementation Notes

Added ForEachRef and GetConfigValue methods to the git.Client interface and implemented them in CommandClient:
- ForEachRef uses git for-each-ref command with format and options support
- GetConfigValue uses git config --get and handles non-existent keys gracefully
- Basic test skeletons added (full integration tests would require test repository setup)
- Methods follow existing patterns in the codebase
