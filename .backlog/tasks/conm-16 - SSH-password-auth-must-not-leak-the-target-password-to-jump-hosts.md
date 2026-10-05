---
id: CONM-16
title: SSH password auth must not leak the target password to jump hosts
status: To Do
assignee: []
created_date: '2026-10-05 18:51'
labels:
  - network
  - ssh
  - security
  - issue
dependencies: []
priority: high
type: spike
ordinal: 12000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
With `auth=password` and `jump` set, `SSH_ASKPASS_REQUIRE=force` and `CONM_SSH_PASSWORD` (internal/cli/ssh.go) are inherited by the `-J` jump process, so every jump-host prompt is answered with the target's password: a bastion password prompt, a key passphrase, or the known_hosts yes/no (an unknown bastion then never connects).

Possible fixes:
1. Refuse password auth together with jump at validation.
2. Askpass script answers only the target's `user@host's password:` prompt and fails the rest.
3. Replace `-J` with `ProxyCommand="env -u SSH_ASKPASS -u SSH_ASKPASS_REQUIRE -u CONM_SSH_PASSWORD ssh [-J earlier hops] -W %h:%p <last hop>"`, so jump hosts prompt on the terminal as with plain `ssh -J`.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A fix is chosen and recorded, with the reasons against the alternatives
- [ ] #2 With password auth and a jump, the target password reaches only the target host
<!-- AC:END -->
