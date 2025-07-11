---
id: task-26
title: Add configuration management commands to wt
status: To Do
assignee: []
created_date: '2025-07-11'
updated_date: '2025-07-11'
labels:
  - high-priority
  - config
dependencies: []
---

## Description

Implement a set of commands to manage wt configuration without manually editing config files. This improves usability by providing quick access to configuration values and the ability to edit them from the command line.

## Acceptance Criteria

- [ ] wt config list - show all config values
- [ ] wt config get <key> - get specific value
- [ ] wt config set <key> <value> - set value
- [ ] wt config edit - open in editor
- [ ] wt config path - show file path
- [ ] wt config dir - navigate to config dir
- [ ] Editor selection (wt.editor then  then )
