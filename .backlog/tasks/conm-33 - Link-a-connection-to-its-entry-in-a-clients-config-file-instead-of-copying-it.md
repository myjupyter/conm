---
id: CONM-33
title: Link a connection to its entry in a client's config file instead of copying it
status: To Do
assignee: []
created_date: '2026-11-21 11:34'
labels:
  - import
dependencies: []
references:
  - CONM-24
  - CONM-17
priority: low
type: feature
ordinal: 25000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Import (CONM-24) copies a foreign entry into conm's TOML once. A linked connection would instead keep only conm's own metadata (name, description, tags, links) plus a reference to the entry in the client's file (for example a service of pg_service.conf or a Host of ~/.ssh/config), so the client's file stays the single source of truth and options conm has no field for (TLS certificate files, extra ssh options) keep working because the client resolves the entry itself.

Deliberately not part of the import branch: a live link changes what a connection is and reaches the table, uniqueness checks, ping, transports, the form and the non-native clients.

Open questions for the plan:
- Every feature reads host, port, user and database, so a linked entry has to be re-read from the foreign file on start. What a connection looks like when the file moved, the entry was renamed, or two linked entries became duplicates.
- Which clients can open an entry by name (psql service=, ssh alias, clickhouse --connection, mysql --defaults-group-suffix) and what happens with the ones that cannot (usql, possibly pgcli).
- How SSH transports rewrite host and port of a linked database connection.
- The form: linked fields read-only, metadata editable.
- Passwords: a linked connection would resolve them the client's way (.pgpass) instead of the keyring.
- This reverses the CONM-24 decision "one-shot conversion, no sync and no watching of the foreign files"; copying stays the default, linking is opt-in.

Cheaper first step to consider: record the origin (file and entry name) of an imported connection as metadata, which allows an explicit re-sync (import --update) without a live link.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Copy stays the default way of importing; linking is chosen explicitly
- [ ] #2 A linked connection opens its client through the entry of the foreign file, so options conm has no field for keep working
- [ ] #3 A linked connection whose file or entry is gone is shown as broken with the reason, and nothing else in the workspace is affected
- [ ] #4 Only conm's own metadata of a linked connection can be edited in the form
- [ ] #5 Documentation says what a linked connection can and cannot do compared to a copied one
<!-- AC:END -->
