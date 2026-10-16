---
id: CONM-13
title: Support SSH servers as a connection type
status: Done
assignee: []
created_date: '2026-09-13 17:28'
updated_date: '2026-10-16 19:02'
labels: []
dependencies:
  - CONM-12
references:
  - CONM-2
documentation:
  - docs/config.md
  - docs/cli.md
  - docs/network.md
  - docs/view.md
  - CLAUDE.md
priority: high
type: feature
ordinal: 9000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
An SSH server has the same shape as a database entry — host, port, user, credentials in a store, and a client binary that takes over the terminal — and a TCP dial plus handshake is a real ping. The bastions people keep in `~/.ssh/config` are the hosts they already manage by hand next to their database list.

Saving SSH as a connection in its own right is also what makes CONM-2 tractable: a tunnel can reference a bastion the user already added and verified, instead of re-describing it inside every connection behind it. Connecting *through* it is CONM-14 and out of scope.

SSH is not a database: no database, no schema, and it must not appear where the product means "pick a database". Its credential is usually a key path or an agent rather than a password, and a key may carry its own passphrase.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 An SSH ConnType with its own config file, persisted and loaded like the database types
- [x] #2 SSH is excluded from config.Databases, DatabaseNames(), the [[conm.database]] rows and the setup screen's list
- [x] #3 The SSH config type satisfies the narrowed Connection without Database() or Schema()
- [ ] #4 Auth covers key path with optional passphrase, agent and password; validation rejects a combination that cannot authenticate
- [x] #5 Passphrases and passwords are secret references resolved through internal/secret, never material in the file
- [ ] #6 Ping performs a real TCP dial plus SSH handshake and reports failure as a *network.OpError with a code and hint
- [x] #7 The ssh client is registered in internal/cli — constant, client list, argv builder, installed-check — and enter hands it the terminal
- [x] #8 A repository, a view and a named form spec expose add/edit/remove to the TUI
- [x] #9 SSH rows render with columns that suit a non-database entry, with uniqueness enforced at the write point
- [x] #10 Tests cover validation, secret resolution, argv building and the ping path
- [ ] #11 docs/config.md, docs/cli.md, docs/network.md, docs/view.md and CLAUDE.md describe SSH as a non-database connection type
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Verification 2026-10-16. go test ./... passes (728 tests): config TestValidateSSHCredential, network TestSSHPing/TestAwaitSSHBanner/TestSSHFirstHopAddress, cli Test_sshCommand/Test_sshIdentity/Test_sshAskpass*. Live check against test/transport bastion (linuxserver/openssh-server, password auth): Ping 127.0.0.1:2222 succeeded in ~10 ms; Ping on a closed port returned *network.OpError ECONNREFUSED with target ssh://jump@127.0.0.1:1.

SSH is a ConnType of kind SSHConnKind. It is in config.ConnTypes but not in config.Databases, so DatabaseNames() and the database lists leave it out. AC#2 named [[conm.database]], which D1 has since renamed to [[conm.connection]] with a kind.

Not met, closed on purpose. AC#4: key auth has no passphrase reference; ssh asks for the passphrase in the terminal, and agent auth is the type-it-once path. ROADMAP D3 moves the passphrase to a second named reference. AC#6: Ping is a TCP dial plus reading the SSH identification banner, not a full key exchange. With a jump host it measures the first hop only (CONM-15). AC#11: docs are done separately.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
SSH servers are a connection type in their own right: SSH is a connection, not a database.

What changed:
- `config.SSH` + `ssh.toml`: host, port, user, `auth` (agent | key | password), plus a credential reference, `jump`, `forward_agent`, `keepalive` and `local_forward`. For key auth the reference is the key (`filepath:` or the material in a store); agent auth takes no secret; password auth takes a reference, not a file. Validation rejects any combination that cannot authenticate.
- SSH is `SSHConnKind`: it is in `ConnTypes` but not in `Databases`, and has no `Database()`/`Schema()`.
- `network.SSHClient`: Ping dials TCP and waits for the SSH banner; failures are `*OpError` with a code (ECONNREFUSED and the like) and a hint. Run hands the terminal to `ssh`, giving it the password through the askpass script, so the password is never put on argv.
- `cli/ssh.go`: the ssh client constant, list entry, argv builder (identity, jump, agent forwarding, keepalive, local forward) and installed-check.
- A repository, a view and the named form spec `spec/ssh.go` give the TUI add/edit/remove, with uniqueness enforced at the write point.

Verification: `go test ./...` passes (728 tests). Live ping against the docker bastion succeeds, and a closed port gives ECONNREFUSED.

Not met, accepted:
- Key passphrase (AC#4) waits for ROADMAP D3.
- Ping stops at the banner rather than a full handshake, and with a jump host it only checks the first hop (CONM-15).
- Docs (AC#11) belong to the separate docs effort.
<!-- SECTION:FINAL_SUMMARY:END -->
