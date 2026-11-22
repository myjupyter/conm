# `cmd` — the Cobra command tree

## Purpose

The shell surface: parses arguments, picks the workspace, reads `conm.toml` and hands over to
`internal/ui`. No business logic and no rendering of its own — a command is a thin adapter.

May import (internal): `config`, `ui`, `ui/spec`, `parser/params`, `importer`, `exporter`.

## Files

| File | Holds |
|---|---|
| `root.go` | `NewRootCmd` (bare `conm` → `ui.Run`) assembling the whole tree, the `--workspace` flag and the `PersistentPreRunE` that selects the workspace, `Execute` and the exit codes |
| `conntype.go` | `connTypeCmd(t)` — `conm <type> [name]`, registered once per `config.ConnTypes` entry, with its aliases (`pg`, `my`, `ch`, …) |
| `add.go`, `edit.go`, `remove.go`, `ping.go`, `list.go`, `exec.go` | the subcommands every connection type gets: `add [command line]`, `edit <name>`, `remove <name>`, `ping <name>`, `list`, `exec <name> -- <query>` |
| `import.go` | `import [name...]`, added only to the types `importer.Supported` names |
| `export.go` | `export [name...]`, added only to the types `exporter.Supported` names |
| `workspace.go` | `conm workspace`, `workspace add [name]`, `workspace use <name>` |
| `doctor.go` | `conm doctor` |

## Rules

- **Commands are generated per type, never written per type.** A subcommand is a function of
  `config.ConnType`; a new connection type needs no change here beyond an optional alias.
- A command that opens a screen wraps it in `ui.WithScreens`; one that needs a connection by
  name goes through the `ui` entry points (`ui.Connect`, `ui.Ping`, `ui.Connections`,
  `ui.RemoveConnection`), not through `repository`.
- `add` turns a command line into a seeded form: `params.Parse`/`ParseArgs` →
  `spec.FormSpecs[t].SeedFunc` → the form. One argument is a quoted command line, several are
  its tokens.
- `import` reads the entries with `importer.Read`, narrows them to the names given and hands
  them to `ui.ImportConnections` with the mode its flags select; `ui.RenderImport` prints the
  result. The one decision made here is the exit status: a strict import with a refused entry
  is an error. `--path` has no default value: `importer.DefaultPath` is asked in `RunE`,
  because finding a client's file touches the disk and the tree is built on every start.
- `export` hands the names and its flags to `ui.ExportConnections` and prints through
  `ui.RenderImport`. As with `import`, the exit status is decided here — a refusal without
  `--skip-errors` or `--show` is an error — and `--path` is resolved in `RunE` through
  `exporter.DefaultPath`.
- Every subcommand name of a connection type is in `config.ReservedNames`: `conm <type> <name>`
  would otherwise be ambiguous.
- **The tree is built by `NewRootCmd`, never held in a global.** Every flag value lives in a
  closure of the command that owns it, so a fresh tree starts from the defaults; `Execute`
  builds one per run and the end-to-end tests in `test/e2e` one per command they run, which is
  why it is exported. A new command is
  added in `NewRootCmd` or by its parent, not from an `init`.
- Exit codes are decided in `Execute` only: a client's own exit code is passed through,
  `ui.ErrDoctorFound` is `1` without a message, any other error is printed and is `1`.
- A new or changed command is user-visible: update `docs/user/commands.md` and the page of
  the feature it belongs to (see `docs/user/AGENTS.md`).
