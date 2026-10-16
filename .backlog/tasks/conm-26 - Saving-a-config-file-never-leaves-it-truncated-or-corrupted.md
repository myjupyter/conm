---
id: CONM-26
title: Saving a config file never leaves it truncated or corrupted
status: To Do
assignee: []
created_date: '2026-10-11 16:00'
labels:
  - release
dependencies: []
references:
  - internal/config/config.go
  - internal/config/conm.go
priority: medium
type: bug
ordinal: 22000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
conm rewrites its TOML files in place: the connection files are held open for the session and saved by seek-to-start, write, truncate; conm.toml is rewritten whole. A crash, kill or full disk in the middle of a save leaves a half-written file, and the user's connections are lost or unparseable on the next start.

Related, to decide in the plan: two conm instances running at once each keep their own in-memory copy, and the later save silently overwrites the earlier one's changes. Whether this task also guards against that, or only makes a single save atomic, is open.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 An interrupted save leaves either the previous or the new file content on disk, never a partial one
- [ ] #2 File permissions stay 0600 after a save
- [ ] #3 Applies to every file conm writes: conm.toml, the per-type connection files and the secret records
- [ ] #4 A decision on concurrent instances is recorded in the task: either handled, or explicitly out of scope
- [ ] #5 Tests cover a successful save and a save that fails mid-way leaving the old content intact
<!-- AC:END -->
