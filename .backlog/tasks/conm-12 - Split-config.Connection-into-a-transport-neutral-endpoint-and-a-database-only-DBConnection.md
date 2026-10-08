---
id: CONM-12
title: >-
  Split config.Connection into a transport-neutral endpoint and a database-only
  DBConnection
status: In Progress
assignee: []
created_date: '2026-09-13 17:25'
updated_date: '2026-09-18 15:45'
labels: []
dependencies: []
documentation:
  - docs/config.md
  - docs/interfaces.md
  - CLAUDE.md
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
- [ ] #1 config.Connection declares only Meta, Username, Host, Port, ConnType, ConnectionString, IsValid, Validate — no Database() or Schema()
- [ ] #2 Exactly one narrowing interface embeds Connection and adds Database() and Schema(); no duplicate in config.go and database.go
- [ ] #3 The conm.toml settings row is config.ConnectionSettings{Kind, Type, CLI, Enabled} under [[conm.connection]]; no type named Database or DatabaseConnection remains in internal/config
- [ ] #4 ConnKind round-trips through TOML as text ("database", "ssh"), and a row saved without a kind loads as a database row
- [ ] #5 Consumers needing Database()/Schema() — network URL builders, ui/view/connection.go, ui/table_view.go — get them through the narrowing interface or a concrete type
- [ ] #6 A type with no database satisfies config.Connection without stub accessors, shown by a compile-time assertion
- [ ] #7 go vet, go test and make lint pass clean
- [ ] #8 docs/config.md, docs/interfaces.md, docs/ui.md, docs/forms.md, CLAUDE.md and the add-new-database-type skill references describe both interfaces and the [[conm.connection]] row
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Implementation plan

Order: config first (everything compiles against it), then the consumers, then docs. One commit.

### 1. `internal/config/conm.go` — the settings row
- Rename `DatabaseConnection` → `ConnectionSettings`. Field names follow: `Conm.Databases` → `Conm.Connections` (`toml:"connection"`), `Conm.Database(t)` → `Conm.Connection(t)`, `SetDatabase` → `SetConnection`, `withDatabaseTypes` → `withDatabaseKinds`.
- `withDatabaseKinds` still fills a row per `config.Databases` type; new rows get `Kind: DatabaseConnKind`, and a stored row whose `Kind` is zero (a file written before the field existed) is set to `DatabaseConnKind`.
- Give `ConnKind` `String`/`MarshalText`/`UnmarshalText` mirroring `ConnType` (`"database"`, `"ssh"`), so the file says `kind = "database"` rather than `kind = 1`. `SSHConnKind` is declared but nothing produces it until CONM-13.
- `CLI(t)` stays; it reads the row through `Connection(t)`.

### 2. `internal/config/database.go` / `config.go`
- Already done: `Connection` is narrowed, `DBConnection` is the one narrowing interface. Leave as is.
- Delete the empty `internal/config/ssh.go` — CONM-13 creates it with content.
- Compile-time assertion for AC #6: a minimal test-only type in `config_test.go` implementing `Connection` without `Database()`/`Schema()`, `var _ config.Connection = noDatabase{}`.

### 3. Consumers of the settings row (mechanical rename)
- `internal/ui/spec/database.go`: `FormSpec[config.ConnectionSettings]`; `BuildFunc` sets `Kind: config.DatabaseConnKind`; `SeedFunc` unchanged.
- `internal/ui/init_form.go`, `init_form_commands.go`, `init_table.go`, `init_table_commands.go`, `init_table_view.go`: `config.Database` → `config.ConnectionSettings`, `conm.SetDatabase` → `conm.SetConnection`, `conm.Databases` → `conm.Connections`.
- `internal/ui/ui.go` `enabledDatabases`, `internal/repository/workspace.go`: iterate `cfg.Connections`.
- Tests: `internal/repository/unique_test.go`, `internal/cli/launcher_test.go` literals.
- `gopls rename` for the type and the `Conm` field/methods; grep `config.Database\b`, `.Databases\b`, `SetDatabase`, `\.Database(` afterwards to catch string-typed spots.

### 4. Consumers of `Database()`/`Schema()` (AC #5)
- Already done on the branch: `network/{mssql,clickhouse,mongodb,redis}.go`, `ui/view/connection.go`, `ui/table_view.go` take `config.DBConnection` or the concrete type. Re-check `network/postgres.go:114` and `network/mysql.go` after the rename compiles.

### 5. Tests
- `internal/config`: `ConnKind` text round-trip; `ReadConm` on a file with `[[conm.connection]]` rows lacking `kind` yields `DatabaseConnKind`; a file with only the old `[[conm.database]]` section yields all types disabled.
- `go vet ./... && go test ./... && make lint`.

### 6. Docs
- `CLAUDE.md`: runtime files line (`[[conm.connection]]`), interfaces mention in the `config` row.
- `docs/config.md`, `docs/interfaces.md`: `Connection` vs `DBConnection`, `ConnectionSettings` + `ConnKind`.
- `docs/ui.md:159`, `docs/forms.md:28`, `.claude/skills/add-new-database-type/references/{checklist.md:124,spec.md:189}`: rename the row and the spec type.

### Out of scope / flagged
- No migration of `[[conm.database]]` rows (see description).
- CONM-13 AC #2 says SSH is excluded from the `[[conm.database]]` rows; with `Kind` on the row that is now the opposite — SSH gets a `kind = "ssh"` row. Update CONM-13 before starting it.
<!-- SECTION:PLAN:END -->
