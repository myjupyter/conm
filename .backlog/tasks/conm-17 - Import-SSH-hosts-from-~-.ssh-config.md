---
id: CONM-17
title: Import SSH hosts from ~/.ssh/config
status: To Do
assignee: []
created_date: '2026-10-06 19:21'
updated_date: '2026-11-21 12:00'
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Built on feat/add-config-import (2026-11-21) as `conm ssh import`, the same command every type with a native config file has (see CONM-24 notes). Differences from the criteria above:
- AC #2: wildcard Host patterns are not connections and are not reported, because a reported entry would fail every strict import of a config that has `Host *`. Their options are applied to the concrete hosts they match, first value winning, as ssh does. Match blocks are dropped whole, without a warning.
- AC #3: a host conm already has (same target) is marked as a duplicate and left alone; a taken name with another target is an error, fixable with `-i`.
- A host with no User (and none from a matching pattern) fails validation: the OS-user default of ssh is not guessed.
<!-- SECTION:NOTES:END -->
