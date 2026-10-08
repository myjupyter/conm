---
id: CONM-15
title: SSH ping measures reachability of the target behind a jump host
status: To Do
assignee: []
created_date: '2026-10-03 17:39'
labels:
  - ssh
  - network
dependencies: []
priority: low
type: enhancement
ordinal: 11000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
For an SSH entry with `jump` set, Ping only checks the first jump hop's SSH banner: the target is usually on a private network unreachable from the laptop, and reaching it requires authenticating to the jump host, which Ping cannot do today. So `p` can report OK while `enter` fails because the target is down.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Ping of an SSH entry with jump reports the target's reachability through all hops
- [ ] #2 Ping fails with a typed *network.OpError when a jump hop refuses auth or the target is unreachable
- [ ] #3 The TODO referencing this task in internal/network/ssh.go Ping is removed
- [ ] #4 Covered by a test
<!-- AC:END -->
