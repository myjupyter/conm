---
id: CONM-25
title: 'Share conm connections across a team, each member binding their own secrets'
status: To Do
assignee: []
created_date: '2026-10-11 15:46'
labels:
  - release
dependencies:
  - CONM-8
references:
  - CONM-8
  - CONM-14
documentation:
  - ROADMAP.md
  - docs/repository.md
  - docs/config.md
priority: medium
type: feature
ordinal: 21000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A team works against the same resources — the same databases, bastions and transports — but every member authenticates with their own credentials. Today each member has to recreate every connection by hand. Exporting conm's own configuration and importing it on another machine lets the team share the topology once while keeping secrets personal. Part of the scope the author set for the first release (SSH transport, workspaces, import/export).

Decisions already taken:
- What travels is the shape of the connections — endpoints, users where relevant, clients, transports and the edges between connections — never secret material and never a reference into the exporter's own secret store, which means nothing on another machine.
- On import, connections arrive without a bound secret and the importing user binds their own; until then the connection is clearly marked and refuses to ping or run instead of failing obscurely.
- Rejection is the whole strategy: a clash with an existing connection, transport or name is refused and reported, never merged or overwritten.
- The unit of export is expected to be a workspace (CONM-8), which is why this task waits on it; exporting a subset of connections is an option to weigh in the plan.
- A transport or bastion a connection depends on travels with it; an export that would leave a dangling edge is refused or includes the dependency.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A workspace can be exported to a single file and imported on another machine from the CLI
- [ ] #2 The exported file contains no password material and no secret-store reference; tests assert this for every connection type and secret scheme
- [ ] #3 Imported connections without a bound secret are marked in the table and refuse ping/run with a typed error naming the missing secret
- [ ] #4 The importing user can bind their own secret to each imported connection from the TUI
- [ ] #5 Connections reached through a transport or bastion import with that dependency intact, or the import is refused with a message naming what is missing
- [ ] #6 A clash with an existing connection, transport or name is refused and reported; nothing existing is changed
- [ ] #7 Importing a file produced by the same conm version round-trips every non-secret field
- [ ] #8 Documentation describes the file format, what is and is not exported, and the bind-your-own-secret flow
<!-- AC:END -->
