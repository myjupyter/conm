# `internal/` — the architecture

What concerns more than one package: the layers, which may import which, the rules that
cross module boundaries, the life of a connection and the seams new things plug into. `cmd`
sits outside this directory but is the outermost layer. The detail of one module is in that
module's `AGENTS.md`.

## Layers and dependency rules

Dependencies point **inward toward the leaves, `config` and `secret`**. Do not add an import
that reverses an arrow.

```
cmd ──► ui ──► ui/view ──► repository ──► network ──► cli ───────┐
        │         │              │           │                   ├──► config   (leaf)
        └──► ui/spec ────────────┴───────────┴──► transport ─────┴──► secret   (leaf)
                                                                 │
cmd ──► parser/params ───────────────────────────────────────────┘
cmd ──► ui/spec
parser/params ──► cli
cmd, ui ──► importer ──► parser/params
cmd, ui ──► exporter ──► importer, config, secret

repository ──► utils        (leaf — domain-free generics)
config, ui ──► utils/keys   (leaf — keystroke parsing)
config     ──► utils/pathname (leaf — portable folder-name rules)
ui/spec    ──► utils/names  (leaf — random adjective-noun names)
```

| Package | May import (internal) | Responsibility | Detail |
|---|---|---|---|
| `internal/utils` (+ `keys`, `pathname`, `names`) | *(nothing)* | Domain-free helpers: a duplicate finder, keystroke parsing, portable folder names, random names. | `internal/utils/AGENTS.md` |
| `internal/config` | `utils/keys`, `utils/pathname` | Data model — connections, secrets, transports, settings, workspaces — validation, **all TOML file I/O**, config paths. | `internal/config/AGENTS.md` |
| `internal/secret` | *(nothing)* | Resolve/store/remove password material behind a scheme (`literal:`, `none:`, `filepath:`, `keyring:`) via pluggable `Provider`s. | `internal/secret/AGENTS.md` |
| `internal/transport` | `config`, `secret` | The route to a connection that is not directly reachable: open an SSH tunnel, hand back its local end. | `internal/transport/AGENTS.md` |
| `internal/cli` | `config` | The binaries conm shells out to: clients per type, detection, argv and env, the `exec`, the transport process, the clipboard. | `internal/cli/AGENTS.md` |
| `internal/network` | `config`, `secret`, `cli`, `transport` | Turn a `config.Connection` into a live one: `Ping`, `Run`, `Close`; resolves the secret lazily, per call. | `internal/network/AGENTS.md` |
| `internal/repository` | `config`, `network`, `cli`, `secret`, `transport`, `utils` | Composition root: opens the config files, builds live connections, secret stores and transports; the only mutation API the runtime UI has. | `internal/repository/AGENTS.md` |
| `internal/ui/view` | `repository`, `config`, `network`, `cli`, `secret`, `transport` | What a screen is looking at and where: one view per entity, a search index across them, and the cursor. | `internal/ui/view/AGENTS.md` |
| `internal/ui/spec` | `config`, `secret`, `transport`, `utils/names` | Declarative form definitions, one named spec per entity, and the values-map formats they are read and written through. | `internal/ui/spec/AGENTS.md` |
| `internal/ui` | `ui/view`, `ui/spec`, `repository`, `importer`, `exporter`, `config`, `network`, `cli`, `secret`, `transport`, `utils/keys` | Bubble Tea TUI. | `internal/ui/AGENTS.md` |
| `internal/parser/params` | `config`, `cli`, `secret` | Read a client's command line back into a connection. | `internal/parser/params/AGENTS.md` |
| `internal/importer` | `config`, `parser/params` | Find and read a client's own config file; hand its connections over as `config.Connection`s. | `internal/importer/AGENTS.md` |
| `internal/exporter` | `config`, `secret`, `importer` | Turn `config.Connection`s into the entries of a client's own config file and merge them into it. | `internal/exporter/AGENTS.md` |
| `cmd` | `config`, `ui`, `ui/spec`, `parser/params`, `importer`, `exporter` | Cobra command tree. | `cmd/AGENTS.md` |

Hard rules — the ones that cross modules. Each module's own rules are in its file.

- **`config` is the only package that reads or writes conm's own files.** Everyone else goes
  through `config.OpenConfig` / `config.File` / `CreateConm` / `ReadConm`; never open a TOML
  file by hand from `ui`, `repository` or `cmd`.
- **A client's own file (`pg_service.conf`, an option file, `~/.ssh/config`) is read only in
  `importer` and changed only by `exporter`**, which adds entries, never replaces one, and
  writes through `config.WriteForeign`.
- **`cli` is the only package that names a client binary.** The resolved client of a
  connection travels up as a `cli.Info` value — `cli.Launcher.Info(ctx)` →
  `repository.Connections.ClientInfo(ctx)` → `view.Connections.ClientInfo(ctx)` — never as a
  name a screen spells for itself. A client name written to `conm.toml` is checked through
  `cli.Validate` at the write point.
- **`secret` knows schemes and material, not connections.** A new secret source is one new
  `Provider` and nothing else in that package.
- **`network` never imports `repository` or `ui`**, and **launches *a* client, never a named
  one**: no binary name, argv, subcommand or password environment variable appears above
  `internal/cli`.
- **A connection's password field holds a *reference*, not necessarily a password**: a bare
  literal or `"<scheme>:<location>"`. The form layer assembles and takes it apart only
  through `spec.JoinSecret`/`SplitSecret`. Below the form layer it is built and read only
  through `secret.Ref`/`secret.ParseRef`: `parser/params` (`params`) writes `literal:` for a
  parsed password and `filepath:` for an ssh identity (`params/ssh.go`); `repository` builds
  one from a secret record and parses them for the usage index (`repository/secret.go`,
  `secret_usage.go`). `config.ValidateSSHCredential`, which cannot import `secret`, reads the
  scheme itself. `secret.IsStore` — never `!= secret.Literal` — is the one test for "this
  points into a store".
- **The password never lives in a displayed string.** `ConnectionString(pw)` injects the
  resolved secret; views build password-free targets, and the search index cannot reach it.
- **The runtime TUI touches entities only through `ui/view`.** A screen holds a
  `*view.Connections`/`*view.Secrets`/…, not a repository: no file I/O, no
  `network.Connection` built by hand, no own row index or filtered slice.
- **Deliberate exceptions to the line above**, and the only ones. New runtime UI goes
  through `ui/view`.
  - The settings screen reads and writes `conm.toml` through `config` itself: it edits the
    file a session is built *from*.
  - A form's preview ping (`ui/form.go`, `ui/transport_table_commands.go`) builds a throwaway
    `network.Connection` from values that are not saved yet.
  - The shell entry points in `internal/ui` — `Connect`, `Ping`, `Connections`,
    `RemoveConnection` in `ui/ui.go`, `ImportConnections` in `ui/import.go`,
    `ExportConnections` in `ui/export.go` — work on repositories directly; import opens the
    add form from there.
- **No layer above `repository` names a connection type in its control flow.** A session
  hands out repositories by `config.ConnType`, never through a per-type field. Two
  exceptions: `builtinTypeColor` in `ui/theme.go`, one default colour per type; and `config.SSHConnType`
  where the form and the transports screen list the SSH connections a transport can go
  through (`ui/form_commands.go`, `ui/transport_table.go`, `ui/transport_table_commands.go`).
- **Every typed refusal lives in its layer's error file** — `repository/errors.go`,
  `network/error.go`, `cli/errors.go`, `transport/errors.go`. Nothing above `network` matches
  on error text.
- **Uniqueness is a property of the store, not of the record**, enforced at the write point.
  Rejection is the whole strategy: nothing is merged, renamed or overwritten behind the
  user's back.
- **A refused write never leaves the screen it came from.** `ui/form_loop.go` re-opens the
  form with its own values and the message in the status line.
- **Compose screens from the shared primitives**, never a private copy — `ui/frame.go`,
  `ui/theme.go`, `ui/keybinding.go`, `ui/keyhint_table.go`, `ui/form_loop.go`.

## Connection lifecycle

1. `cmd/root.go` selects the workspace, opens `conm.toml` → `ui.Run(conm)`, which loops: per
   pass it builds a `repository.Session`, wraps its repositories in `view.NewConnections` /
   `view.NewSecrets` / `view.NewTransports` and hands them to the connections table. A pass
   that ends asking for a reload closes the session and starts the next from a re-read
   `conm.toml`.
2. Each connection repository opens its file and, per record, builds a `network.Connection`
   if the entry validates, else a `network.NoClient` — a stub that displays but errors on
   `Ping`/`Run`. A record that names a transport becomes a `network.Transported`.
3. Add/Edit: form → `spec.BuildFunc(values)` → a `config.Connection` → `repo.Add`/`Edit`,
   which validates, checks uniqueness, rebuilds the live connection and saves the file.

## Extension seams

- **New connection type** — follow the `add-new-database-type` skill
  (`.claude/skills/add-new-database-type/`), which owns the fit check, the order of work and
  every registration table. Nothing in `internal/ui` changes except the type's default colour in
  `builtinTypeColor` (`ui/theme.go`). **A change that moves one of those seams is not finished until
  the skill says so too.**
- **New client for an existing type** — `internal/cli/AGENTS.md`.
- **New secret store** — a `secret.Provider` registered in `secret.DefaultFor`, a
  `config.Secret` implementation if its records differ from `Keyring`, a repository
  constructor, a named spec in `secretFormSpecs`, the scheme in `spec.SecretProvidersOrder`,
  the repository passed to `view.NewSecrets`. No form or screen code changes.
- **New transport kind** — `internal/transport/AGENTS.md`.
- **New import format** — `internal/importer/AGENTS.md`.
- **New export format** — `internal/exporter/AGENTS.md`.
- **New entity table** — `internal/ui/view/AGENTS.md`.
