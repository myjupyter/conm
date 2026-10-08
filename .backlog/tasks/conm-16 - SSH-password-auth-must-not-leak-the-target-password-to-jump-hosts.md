---
id: CONM-16
title: SSH password auth must not leak the target password to jump hosts
status: Done
assignee: []
created_date: '2026-10-05 18:51'
updated_date: '2026-10-08 17:21'
labels:
  - network
  - ssh
  - security
  - issue
dependencies: []
modified_files:
  - internal/cli/ssh.go
  - internal/cli/ssh_test.go
  - internal/cli/command_test.go
  - docs/cli.md
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
- [x] #1 A fix is chosen and recorded, with the reasons against the alternatives
- [x] #2 With password auth and a jump, the target password reaches only the target host
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Option 3: when the askpass is on and a jump is set, spell the jump as the ProxyCommand -J stands for, behind `env -u SSH_ASKPASS -u SSH_ASKPASS_REQUIRE`. Also take the password out of the environment entirely: askpass reads it from a per-run 0600 file.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Chose option 3 (ProxyCommand with a cleaned env). Option 1 (refuse password + jump) drops a real use case; option 2 (askpass answering only the target prompt) would swallow every bastion prompt under SSH_ASKPASS_REQUIRE=force, so an unknown bastion could never be accepted. Verified on OpenSSH 10.3: `ssh -G` parses the built ProxyCommand, and a ProxyCommand printing its env saw no SSH_ASKPASS* (leaked=0). Not exercised against a live bastion.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
With password auth and a jump, the target password no longer reaches jump hosts.

- `internal/cli/ssh.go`: when the askpass is on and `Jump` is set, `-J` is replaced by `-o ProxyCommand=env -u SSH_ASKPASS -u SSH_ASKPASS_REQUIRE ssh [-J earlier hops] -W %h:%p ssh://<last hop>`; jump hosts prompt on the terminal like a plain `ssh -J`. Without the askpass `-J` stays.
- The password no longer travels in the environment (`CONM_SSH_PASSWORD` is gone): `sshAskpass` writes it to a 0600 file in a fresh 0700 dir next to a fixed `cat` script, removed by the command's cleanup when ssh exits.
- Tests in `internal/cli/ssh_test.go`: exact argv for one- and multi-hop jumps, `-J` kept without a password, the password absent from every env var, the askpass script printing the password byte for byte and its dir removed after cleanup.
- Docs: `docs/cli.md`.

Risk: the password sits on disk (0600, own user) for the session's lifetime; a unix-socket askpass would avoid that at more code.
<!-- SECTION:FINAL_SUMMARY:END -->
