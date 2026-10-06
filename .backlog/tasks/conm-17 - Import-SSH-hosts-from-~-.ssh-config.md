---
id: CONM-17
title: Import SSH hosts from ~/.ssh/config
status: To Do
assignee: []
created_date: '2026-10-06 19:21'
labels:
  - ssh
dependencies: []
type: feature
ordinal: 13000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Most users already describe their hosts in ~/.ssh/config; re-entering them by hand is the main barrier to adopting conm. One-time import into conm's TOML, no sync.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Each concrete Host block becomes an SSH connection (HostName, User, Port, IdentityFile, ProxyJump, ForwardAgent, ServerAliveInterval)
- [ ] #2 Wildcard Host patterns and Match blocks are skipped with a warning
- [ ] #3 Hosts already in conm are refused as duplicates, not overwritten
- [ ] #4 ~/.ssh/config is never written
- [ ] #5 Covered by tests
<!-- AC:END -->
