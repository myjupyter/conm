---
id: CONM-27
title: Rethink how a connection of a new database type is added
status: To Do
assignee: []
created_date: '2026-10-16 17:47'
labels: []
dependencies: []
priority: low
type: enhancement
ordinal: 23000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Adding a connection of a type not yet in use takes two steps: enable the type on the setup screen (ctrl+a, a plain table), then add on its tab. The `enabled` flag in conm.toml also decides which repositories the workspace opens, so a disabled type drops out of usage checks and disabling ssh breaks every transport that goes through it (the user gets an error, which is acceptable for now). Decide whether the enable step should stay, and redesign the add flow accordingly.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A connection of any supported type can be added from the connections table without visiting a separate setup step first, or the decision to keep the setup step is recorded with its reason
- [ ] #2 Transports and usage checks behave the same regardless of which database tabs are shown
- [ ] #3 Tests cover the new add flow
<!-- AC:END -->
