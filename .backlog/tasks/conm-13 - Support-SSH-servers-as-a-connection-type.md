---
id: CONM-13
title: Support SSH servers as a connection type
status: To Do
assignee: []
created_date: '2026-09-13 17:28'
updated_date: '2026-09-13 17:30'
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
- [ ] #1 An SSH ConnType with its own config file, persisted and loaded like the database types
- [ ] #2 SSH is excluded from config.Databases, DatabaseNames(), the [[conm.database]] rows and the setup screen's list
- [ ] #3 The SSH config type satisfies the narrowed Connection without Database() or Schema()
- [ ] #4 Auth covers key path with optional passphrase, agent and password; validation rejects a combination that cannot authenticate
- [ ] #5 Passphrases and passwords are secret references resolved through internal/secret, never material in the file
- [ ] #6 Ping performs a real TCP dial plus SSH handshake and reports failure as a *network.OpError with a code and hint
- [ ] #7 The ssh client is registered in internal/cli — constant, client list, argv builder, installed-check — and enter hands it the terminal
- [ ] #8 A repository, a view and a named form spec expose add/edit/remove to the TUI
- [ ] #9 SSH rows render with columns that suit a non-database entry, with uniqueness enforced at the write point
- [ ] #10 Tests cover validation, secret resolution, argv building and the ping path
- [ ] #11 docs/config.md, docs/cli.md, docs/network.md, docs/view.md and CLAUDE.md describe SSH as a non-database connection type
<!-- AC:END -->
