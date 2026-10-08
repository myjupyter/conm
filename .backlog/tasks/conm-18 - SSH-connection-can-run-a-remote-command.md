---
id: CONM-18
title: SSH connection can run a remote command
status: To Do
assignee: []
created_date: '2026-10-06 19:21'
labels:
  - ssh
dependencies: []
type: feature
ordinal: 14000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Connect to what runs on a machine rather than the machine itself, e.g. `docker exec -it api sh` or `kubectl exec -it pod -- sh`, while a plain entry for the same machine still exists.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 An SSH connection stores an optional remote command and whether to force a TTY
- [ ] #2 The command is part of the connection's Identity, so a machine entry and a command entry on it do not collide
- [ ] #3 Pasting `ssh host <command>` keeps the command instead of dropping it
- [ ] #4 Covered by tests
<!-- AC:END -->
