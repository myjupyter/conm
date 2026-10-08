---
id: CONM-2.1
title: Reach a database through an SSH tunnel
status: To Do
assignee: []
created_date: '2026-09-13 17:28'
labels: []
dependencies:
  - CONM-13
documentation:
  - docs/network.md
  - docs/repository.md
  - docs/forms.md
  - CLAUDE.md
parent_task_id: CONM-2
priority: high
type: feature
ordinal: 10000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Every client in internal/network dials the host in the connection directly over TCP, which assumes the database is routable from the machine conm runs on. For the most common deployment it is not: the database listens on a private network and only a bastion is reachable, so the host saved in the config is meaningful only on the far side of a jump host. Today a user has to stand the tunnel up by hand in another terminal and then save a connection pointing at localhost — which loses the real address, breaks when the local port changes, and makes the saved entry a lie about where the database actually lives.

This is the first concrete transport of CONM-2 and the one that proves the seam: if a tunnel can be brought up before Ping/Run, shared by both the driver and the CLI child process, and torn down after, the remaining transports are variations on the same lifecycle.

The bastion is not re-described per connection: a database connection references an SSH connection saved by CONM-13, the way a password references a secret record. One edit of the jump host covers every database behind it.

Two things the design has to keep straight. The user must keep seeing the real host — in the table, in the info screen, in an error's target — while the driver and the client binary are pointed at the local forwarded endpoint. And the tunnel is a new and independent failure source: the bastion can refuse a key or be unreachable before the database is ever contacted, and the user must be able to tell that apart from the database being down.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A database connection can name an SSH connection as its tunnel, stored as a reference to a saved SSH entry rather than an inline copy of the jump host settings
- [ ] #2 A connection whose tunnel reference does not resolve to an existing SSH entry fails validation with a typed error and is shown in the TUI as unusable rather than silently dialled direct
- [ ] #3 The tunnel is established before Ping and before Run, and is torn down when the operation ends, including when the CLI child process exits or the ping fails
- [ ] #4 Both the Go driver and the spawned CLI client reach the database through the same tunnel instance; the CLI is not handed a raw remote address
- [ ] #5 The local forwarded endpoint is chosen automatically without the user configuring a port, and two connections tunnelling at the same time do not collide
- [ ] #6 Every user-visible surface — table row, info screen, connection string shown to the user, OpError target — reports the configured remote host and port, never the local forwarded address
- [ ] #7 A tunnel failure (bastion unreachable, auth rejected, forward refused) surfaces as a *network.OpError that is distinguishable from a database-side failure and carries a hint naming the bastion
- [ ] #8 Removing or editing an SSH connection that connections tunnel through is refused or reported, not left to produce dangling references
- [ ] #9 The tunnel is configurable from the add/edit form through the existing spec mechanism, with no bespoke screen code
- [ ] #10 Table-driven testify tests cover reference resolution, validation of a dangling reference, and the mapping of tunnel failures to OpError codes
- [ ] #11 docs/network.md documents the tunnel lifecycle and the dialled-vs-configured address split, and docs/repository.md documents cross-entity reference resolution
- [ ] #12 CLAUDE.md's layer rules state which layer owns transport setup and that no layer above network learns the local forwarded address
<!-- AC:END -->
