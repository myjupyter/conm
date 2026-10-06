---
id: CONM-19
title: SSH tunnel-only sessions and more forward kinds
status: To Do
assignee: []
created_date: '2026-10-06 19:22'
labels:
  - ssh
dependencies: []
type: feature
ordinal: 15000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Today one local forward (-L) lives only as long as the interactive shell. Support a tunnel-only session (-N), remote (-R) and dynamic/SOCKS (-D) forwards, and more than one forward per entry. Depends on the decision whether LocalForward stays or SSH becomes a transport.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A decision on LocalForward vs SSH-as-transport is recorded before implementation
- [ ] #2 A connection can open a tunnel without a shell
- [ ] #3 Local, remote and dynamic forwards are supported, several per connection
- [ ] #4 Covered by tests
<!-- AC:END -->
