---
id: CONM-24
title: >-
  Import connections from and export them to the config files of standard
  database and SSH tooling
status: In Progress
assignee: []
created_date: '2026-10-11 15:46'
updated_date: '2026-11-21 13:54'
labels:
  - release
dependencies: []
references:
  - CONM-17
  - CONM-8
documentation:
  - ROADMAP.md
  - docs/config.md
  - docs/params.md
priority: medium
type: feature
ordinal: 20000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Users already describe their endpoints in the config files of the tools conm launches. Re-entering them by hand is the main barrier to adopting conm, and the reverse is the exit path: a user (or a teammate without conm) should be able to get conm's connections back into the tools' own formats. Part of the scope the author set for the first release (SSH transport, workspaces, import/export).

Candidate formats, to confirm in the plan: PostgreSQL service file and passfile (pg_service.conf, .pgpass), MySQL option files (.my.cnf [client] groups), ~/.ssh/config. Importing SSH hosts from ~/.ssh/config is already CONM-17 and stays there; this task covers the remaining formats and the export direction for all of them.

Decisions already taken:
- One-shot conversion in either direction, no sync and no watching of the foreign files.
- Rejection is the whole strategy: an imported entry that clashes with an existing conm connection is refused and reported, never merged, renamed or overwritten. Export never overwrites an existing foreign file or entry silently.
- Password material is sensitive: a plaintext password found while importing goes into a secret store, not into conm's TOML; export writes no password material unless the user explicitly asks for it.
- Reading and writing these files is file I/O and belongs to the config layer, like every other file conm touches.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 Each supported foreign format can be imported into conm and each supported conm type can be exported to its tool's format, via a command reachable from the CLI
- [ ] #2 Entries that cannot be represented (unsupported options, wildcard or include constructs, unknown types) are skipped with a per-entry warning, not a whole-run failure
- [ ] #3 An imported entry whose name clashes with an existing connection is refused and reported; nothing existing is changed
- [ ] #4 A plaintext password found on import is stored in a secret store and the connection holds a reference to it; conm's TOML never receives the material
- [ ] #5 Export does not write password material by default and refuses to overwrite an existing file or entry without explicit confirmation
- [ ] #6 A dry-run shows what would be imported or exported without writing anything
- [ ] #7 Round-trip tests per format: import of a fixture, then export, yields an equivalent file for every representable field
- [ ] #8 Documentation lists the supported formats, the fields mapped and the known gaps
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Export, on feat/add-export-config (2026-11-21). Import is merged (#36).

Decisions confirmed by the author:
- `conm <type> export [name...]` merges into the client's own file (`--path` overrides), creating it when absent. An entry whose name is already in the file is refused and reported, never replaced. `--show` changes nothing.
- `--with-passwords` (off by default) writes the resolved password into the format's own field: `password=` in pg_service.conf and the MySQL option file, `<password>` for ClickHouse, base64 basic-auth in sqlconfig. SSH has none; the identity file path is always exported.
- Order of work: the mergers for the default config files first.

Steps:
1. `internal/config` — mergers, the inverse of the foreign readers, each taking `[]ForeignSection`, refusing a name the file already holds, writing through `writeFileAtomic`:
   - INI (pg_service.conf, MySQL option file) and ssh_config: text appended, the existing bytes untouched.
   - ClickHouse client config: XML by text insertion into `connections_credentials`, YAML through `yaml.Node`.
   - sqlconfig: `yaml.Node` merge of endpoints / users / contexts.
2. `internal/exporter` — `config.Connection` (+ optional resolved password) to the `ForeignSection`s of its format, with warnings for what the format cannot hold (transport, schema, …); `Supported`, `DefaultPath`, `Write`.
3. `internal/ui/export.go` — `ExportConnections`: pick the connections, resolve passwords when asked, plan, write; rendered like import.
4. `cmd/export.go` — `export [name...]` with `--path`, `--show`, `--skip-errors`, `--with-passwords`; `export` added to `config.ReservedNames`.
5. Round-trip tests per format (export then import yields the same connection), AGENTS.md of config / exporter / ui / cmd / internal, user docs (`add-connection.md` or the owning page, `commands.md`).
6. make tools / build / lint / test.

Layout changed on the author's instruction (2026-11-21), replacing steps 1-2 above:
- The mergers are per type in `internal/exporter/<type>.go` (bytes in, bytes out), with `ini.go` and `yaml.go` for what two types share; `internal/config` keeps only `WriteForeign`, the atomic write.
- `config/foreign.go` and `config/ini.go` are split per type into `internal/importer/<type>.go` (paths, readers, `Section`/`Param`); `SplitSSHOption` stays in `config/ssh.go` because `parser/params` reads with it too.
- `exporter` imports `importer` for the paths, the section shape and the readers of what a file already holds.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Decisions taken while building import on feat/add-config-import (2026-11-21), replacing 'rejection is the whole strategy' for import:
- `conm <type> import [name...]` with `--path`, `--show`, `--skip-errors`, `-i`; `--workspace` is the global flag. `--show` replaces the dry-run of AC #6.
- Default is all-or-nothing: if any entry cannot be imported nothing is written and the exit status is 1. `--skip-errors` imports the rest. `-i` opens each refused entry in the add form to be fixed, so a name clash can be renamed there instead of only being refused (AC #3).
- A connection conm already has (same target, any name) is not an error in any mode; it is left alone and marked as a duplicate.
- Postgres reads pg_service.conf for connections and uses .pgpass only to look up their passwords.
- Types with a native connections file: postgres, mysql, ssh, clickhouse (connections_credentials), mssql (go-sqlcmd sqlconfig). Redis and MongoDB have none and get no import command.
- Linking to the foreign file instead of copying is a separate task.

Export built on feat/add-export-config (2026-11-21), uncommitted, awaiting the author's review:
- `conm <type> export [name...]` with `--path`, `--show`, `--skip-errors`, `--with-passwords` for postgres, mysql, clickhouse, mssql, ssh (the types import has).
- Mergers in `internal/config/foreign_merge.go`: INI and ssh_config appended as text; ClickHouse XML merged as text and re-parsed as a check, ClickHouse YAML and sqlconfig through yaml.Node (comments kept, re-indented). A name already in the file is refused, never replaced; the written file is 0600.
- New package `internal/exporter` (`Plan`/`Write`), `ui.ExportConnections` reporting through `ImportResult`/`RenderImport`, `repository.Session.Password`, `export` added to `config.ReservedNames`.
- An entry already in the file is a refusal even when it is the same target, unlike import's 'already there': exporting twice needs `--skip-errors`.
- A conm name is exportable only when made of letters, digits and `_ . @ -`.
- Round-trip tests go export -> import (`internal/exporter/exporter_test.go`), not fixture -> import -> export as AC #7 words it.
- make build / lint / test pass. Not done: the demo GIF for export (needs Docker + vhs).

After the layout change: make build / lint / test pass; a client's own file is read only in `importer` and changed only by `exporter` (hard rule reworded in internal/AGENTS.md). Still uncommitted, still no export demo GIF.
<!-- SECTION:NOTES:END -->
