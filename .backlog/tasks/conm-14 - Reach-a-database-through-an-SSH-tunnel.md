---
id: CONM-14
title: Reach a database through an SSH tunnel
status: To Do
assignee: []
created_date: '2026-09-13 17:30'
labels: []
dependencies:
  - CONM-13
references:
  - CONM-2
documentation:
  - docs/network.md
  - docs/repository.md
  - docs/forms.md
  - CLAUDE.md
priority: high
type: feature
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Clients in internal/network dial the database host directly, which assumes it is routable from this machine. Usually it is not: the database is on a private network behind a bastion. Today the user stands the tunnel up by hand and saves a connection pointing at localhost, which loses the real address and breaks when the local port changes.

First concrete transport of CONM-2, and the one that proves the seam. The bastion is referenced, not copied: a connection names an SSH entry from CONM-13, the way a password names a secret record.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A database connection references a saved SSH entry as its tunnel; the jump host is not copied inline
- [ ] #2 A dangling tunnel reference fails validation and the connection is shown unusable, never dialled direct
- [ ] #3 The tunnel comes up before Ping/Run and down when the operation ends, including on CLI exit and on failure
- [ ] #4 Driver and CLI client share the same tunnel instance
- [ ] #5 The local forwarded port is picked automatically and two live tunnels do not collide
- [ ] #6 Table, info screen, connection string and OpError target show the configured host, never the forwarded address
- [ ] #7 Tunnel failures surface as *network.OpError distinguishable from database failures, with a hint naming the bastion
- [ ] #8 Editing or removing an SSH entry other connections tunnel through is refused, not left dangling
- [ ] #9 The tunnel is configured through the existing form spec, with no bespoke screen code
- [ ] #10 Tests cover reference resolution, dangling references and tunnel-failure error mapping
- [ ] #11 docs/network.md and docs/repository.md cover the tunnel lifecycle, the dialled-vs-configured address split and cross-entity references
<!-- AC:END -->
