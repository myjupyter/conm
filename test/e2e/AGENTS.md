# `test/e2e` — end-to-end tests of the shell commands

## Purpose

`conm <type> import` and `conm <type> export` run as a user runs them, against every format
they support: the real command tree from `cmd.NewRootCmd`, real files, a throwaway home. No
database or client is started; `ssh -G` is the one client a test calls, skipped when ssh is
not installed. The package holds tests only.

May import (internal): `cmd`, `config`, `importer`, `secret`, `ui` — what a user can observe,
plus the readers of conm's own files to check what it stored.

## Files

| File | Holds |
|---|---|
| `e2e_test.go` | `e2eEnv` — the throwaway environment and the helpers that run a command and read back what conm stored; `golden`, `fixture` |
| `import_export_test.go` | `e2eFormats`, the format table, and the cases every format runs: the round trip, import, export, files, workspaces |
| `import_export_formats_test.go` | the cases one format alone can show, every type's default file, the types without the commands |
| `testdata/<type>/` | one fixture per format and its golden exports, `exported-<file>` and `exported-with-passwords-<file>` |

## Rules

- **Every test starts from a fresh `e2eEnv`**: `HOME`, `XDG_CONFIG_HOME` and `XDG_DATA_HOME`
  in a temporary directory, `PGSERVICEFILE` unset and `PGPASSFILE` under that home, an empty
  working directory, go-keyring's in-memory mock as the keyring. Nothing reaches the real
  config, the real keyring or the real `~/.ssh/config`. A second environment in one test
  replaces the mock keyring, so read what the first one stored before opening it.
- **A command runs on a fresh `cmd.NewRootCmd()`**: flag values live in the tree, so reusing
  one would carry a flag into the next run. Its error is not in the output — the root
  command silences errors — so a test looking for a message checks both.
- **What conm stored is read back the way it reads it**: connections through
  `ui.Connections`, passwords from the mock keyring under the service
  `config.SecretWorkspace()` returns — the workspace's id, not its name.
- **A connection a file cannot describe is seeded** through `ui.ImportConnections`
  (`e2eEnv.seed`), which fails the test on a refusal. A transport is `transport.toml`
  written beside an SSH connection, and a connection names it `ssh:<name>`.
- **The round trip** imports a fixture, exports it, imports the export into a fresh
  environment and requires the same connections and passwords, then a byte-identical second
  export. A per-format case round-trips the same way unless it says `noRoundTrip`.
- Fixtures hold passwords no user, host or database spells (`pg-s3cret`, …), so a test can
  tell that a password is absent. Two entries of one fixture never share a host, port, user
  and database: conm would take the second for the first.
- Tests follow the gotests table shape; a stateful scenario is a case with a `prepare` hook.

## Golden files

`go test ./test/e2e -run TestExport_roundTrip -update` rewrites `testdata/<type>/exported-*`
from what export writes now. The diff is the review: an unexpected change there is a change
to what users get in their files.

## Adding a format

1. A fixture under `testdata/<type>/` with importable and refused entries.
2. A row in `e2eFormats`: the names it imports and refuses, its passwords, a file of other
   entries, one it cannot parse, and the constructors of a plain and a routed connection.
3. `-update` for its golden files, then the cases only it can show in
   `import_export_formats_test.go`, and its default file in `TestImportExport_defaultPath`.
