---
id: CONM-12
title: >-
  Split config.Connection into a transport-neutral endpoint and a database-only
  DBConnection
status: Done
assignee: []
created_date: '2026-09-13 17:25'
updated_date: '2026-10-08 19:09'
labels: []
dependencies: []
modified_files:
  - internal/config/config.go
  - internal/config/database.go
  - internal/network/mssql.go
  - internal/network/clickhouse.go
  - internal/network/mongodb.go
  - internal/network/redis.go
  - internal/ui/view/connection.go
  - internal/ui/table_view.go
priority: high
type: enhancement
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`config.Connection` requires `Database()` and `Schema()` from every implementation, which assumes every saved connection is a database. An SSH endpoint (CONM-13) has a host, port, user and client but no database, so it cannot implement the interface without stub accessors that lie.

Narrow `Connection` to what every dialable endpoint has, and keep the database fields in a second interface only database types satisfy, so "is this a database?" becomes an explicit assertion.

The `[[conm.database]]` settings row in conm.go (`Database{Type, CLI, Enabled}`) is no longer a database row either: once SSH is a connection type, the same row — which type, which client, enabled or not — describes any connection conm manages, and a `Kind` tells database rows from SSH rows. The struct is renamed `ConnectionSettings` (the interim name `DatabaseConnection` was wrong for the same reason `Database` was) and the section becomes `[[conm.connection]]`. Existing `[[conm.database]]` rows are not migrated: they are silently dropped and the database types come back as disabled rows, which is accepted for a pre-release tool.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 config.Connection declares only Meta, Username, Host, Port, ConnType, ConnectionString, IsValid, Validate — no Database() or Schema()
- [x] #2 Exactly one narrowing interface embeds Connection and adds Database() and Schema(); no duplicate in config.go and database.go
- [x] #3 The conm.toml settings row is config.ConnectionSettings{Type, CLI, Enabled} under [[conm.connection]]; no type named Database or DatabaseConnection remains in internal/config
- [x] #4 ConnKind is derived from the type via ConnType.Kind() and is not stored in conm.toml; a file holding only old [[conm.database]] rows loads every type as a disabled row
- [x] #5 Consumers needing Database()/Schema() — network URL builders, ui/view/connection.go, ui/table_view.go — get them through the narrowing interface or a concrete type
- [x] #6 A type with no database satisfies config.Connection without stub accessors, shown by a compile-time assertion
- [x] #7 go vet, go test and make lint pass clean
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Executed as planned, with one deviation: `Kind` is not a field on `ConnectionSettings`. It is fully determined by the type, so it is derived through `ConnType.Kind()` instead of stored — storing it would be a second source of truth that the reader overwrites anyway, plus `ConnKind` text marshalling for nothing. Documentation updates are moved out of this task; docs/ is reworked separately.
<!-- SECTION:PLAN:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Split the connection model so a non-database endpoint (SSH) is a first-class connection.

- `config.Connection` (`internal/config/config.go`) now carries only what every dialable endpoint has; `config.DBConnection` (`internal/config/database.go`) embeds it and adds `Database()`/`Schema()`.
- Network drivers take `DBConnection`; `ui/view/connection.go` and `ui/table_view.go` assert it before reading a database name.
- The `conm.toml` settings row is `ConnectionSettings{Type, CLI, Enabled}` under `[[conm.connection]]`. `ConnKind` is derived via `ConnType.Kind()` rather than stored — it is fully determined by the type, so storing it would only add a second source of truth.
- Old `[[conm.database]]` rows are dropped without migration (pre-release tool); every type comes back as a disabled row.

Verification: `go vet ./...`, `go test ./...` and `make lint` (0 issues) pass. `TestReadConm` covers the new section and the dropped old section; `hostOnly` in `conm_test.go` and `SSH` in `ssh.go` are compile-time proofs that a type without a database satisfies `Connection`.

Follow-ups: docs/ (config.md, ui.md, CLAUDE.md still describe a stored `Kind`) are handled in the separate docs rework; CONM-13 AC #2 still references `[[conm.database]]` and needs rewording.
<!-- SECTION:FINAL_SUMMARY:END -->
