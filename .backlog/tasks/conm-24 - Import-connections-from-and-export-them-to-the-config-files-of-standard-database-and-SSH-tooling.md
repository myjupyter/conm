---
id: CONM-24
title: >-
  Import connections from and export them to the config files of standard
  database and SSH tooling
status: To Do
assignee: []
created_date: '2026-10-11 15:46'
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
