---
id: CONM-23
title: Mark a connection read-only so its client cannot write
status: To Do
assignee: []
created_date: '2026-10-11 15:37'
labels: []
dependencies: []
references:
  - 'https://github.com/LeoAlecksey/opsdeck/blob/main/src-tauri/src/db.rs'
documentation:
  - docs/cli.md
  - docs/forms.md
  - ROADMAP.md
priority: medium
type: feature
ordinal: 19000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Production databases sit in the same list as dev and staging, one keypress away from an interactive client with full write access. An accidental UPDATE/DROP in the wrong tab is the classic failure. A per-connection read-only flag lets the user open prod with the client refusing writes.

Constraint from conm's principle ("conm never implements the client"): the protection must be enforced by the server or by the client binary itself, through what conm passes when launching it — never by conm inspecting or filtering statements. A type whose clients offer no such mechanism (likely Redis and MongoDB) must not pretend: the flag is refused for it, in the spirit of "rejection is the whole strategy".

Prior art: OpsDeck (github.com/LeoAlecksey/opsdeck, src-tauri/src/db.rs) does server-side read-only for PostgreSQL and ClickHouse and a session-level READ ONLY for MySQL; its Redis/Mongo variant is a client-side command allowlist, which is exactly what conm should not copy.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A connection record carries an optional read-only flag; absent means writable, and existing files load unchanged
- [ ] #2 Opening a read-only connection launches its client so that a write statement is rejected by the server or client, for every client of a type that supports it
- [ ] #3 A read-only connection whose configured client cannot enforce read-only is refused at launch with a typed error, not opened writable
- [ ] #4 Setting the flag on a type with no enforceable mechanism is refused at the write point, and the form stays open with the message in the status line
- [ ] #5 The add/edit form exposes the flag only for types that support it
- [ ] #6 Read-only connections are visibly marked in the connections table
- [ ] #7 Tests cover the launch arguments/environment per supporting client and the refusals for unsupported types and clients
- [ ] #8 docs/cli.md lists which clients enforce read-only and how; the add-new-database-type skill asks a new type to declare its read-only support
<!-- AC:END -->
