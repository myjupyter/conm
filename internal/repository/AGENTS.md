# `internal/repository` — the composition root

## Purpose

Opens the config files of the active workspace, builds the live connections, the secret
store and the transports over them, and is **the only mutation API the runtime UI has**.
Owns the cross-record rules — uniqueness and "still in use" — and their typed refusals.

May import (internal): `config`, `network`, `cli`, `secret`, `transport`, `utils`.

## Files

| File | Holds |
|---|---|
| `session.go` | `Session`, `NewSession`, `connectionRepositories` (`config.ConnType` → constructor) |
| `connection.go` | `Connections`, `ConnectionRepository` |
| `postgres.go`, `mysql.go`, `mssql.go`, `clickhouse.go`, `redis.go`, `mongodb.go`, `ssh.go` | one constructor each (`NewPostgresRepository`, …) |
| `secret.go`, `keyring.go` | `Secrets`, `SecretRepository`; `NewKeyringRepository` |
| `transport.go` | `Transports`, `TransportRepository`, `ConnectionFinder` |
| `workspace.go` | `Workspaces`, `WorkspaceRepository` |
| `unique.go` | the uniqueness rule over `Identity()` and the name |
| `secret_usage.go`, `transport_usage.go`, `connection_usage.go` | `Usage` and the indexes answering "who references this": `Index`, `TransportIndex`, `ConnectionIndex`, fed by `SecretUser` / `TransportUser` / `ConnectionUser` |
| `doctor.go` | `CheckClients`, `CheckFiles` — the facts behind `conm doctor` |
| `errors.go` | every typed refusal of this layer |

## Main entities

```go
type Connections interface {
    Kind() config.ConnType
    ClientInfo(ctx context.Context) cli.Info
    Len() int
    IndexOf(name string) (int, bool)
    ConnectionAt(int) (network.Connection, bool)
    ConfigAt(int) (config.Connection, bool)
    Add(cfg config.Connection) error
    AddAll(cfgs ...config.Connection) error
    Check(cfg config.Connection, pending ...config.Connection) error
    Edit(int, config.Connection) error
    Remove(int) error
}

type Secrets interface {
    Kind() secret.Scheme
    Len() int
    Get(int) (config.Secret, bool)
    UsagesAt(int) []Usage
    Add(ctx context.Context, s config.Secret, password string) error
    PutAll(ctx context.Context, writes ...SecretWrite) ([]error, error)
    Edit(ctx context.Context, i int, s config.Secret, password string) error
    Remove(ctx context.Context, i int) error
    Save() error
}

type Transports interface {
    transport.Provider
    Len() int
    Get(int) (config.Transport, bool)
    UsagesAt(int) []Usage
    Add(config.Transport) error
    Edit(int, config.Transport) error
    Remove(int) error
}

type Workspaces interface {
    Len() int
    Get(int) (string, bool)
    Active() string
    Use(int) error
    Add(name string) error
    Rename(i int, name string) error
    Remove(ctx context.Context, i int) error
}
```

- `Session` is one open workspace: the connection repositories (`Connections()` /
  `ConnectionsOf(t)`), `Keyring` and `Transports`. `Password(ctx, conn)` resolves a
  connection's secret reference to its material for `conm <type> export --with-passwords`;
  a `filepath:` reference is a path, not material, and resolves to nothing. Built by `NewSession(conm)`, closed with
  `Close()`.
- `Connections` is not generic: `Kind()` says which `config.ConnType` every entry has — what
  `ConfigAt` asserts to — and `Add`/`Edit` reject a foreign type.
- `AddAll` saves several connections with one write or none: each is admitted against the
  stored ones and the earlier ones of the same call, so a batch cannot hold duplicates of
  itself. A failure — at admission, at building a live connection or at the save — leaves
  nothing behind: what was built is closed and the records are taken back out. `Add` is
  `AddAll` of one. `Check` is the same admission without the write, `pending` standing for
  connections that would be added first; an import uses it to refuse a whole file before
  touching it.
- `Secrets`: the password travels beside the record, never inside it — from the form into
  `Add`/`Edit`, and from an import into `PutAll` as `SecretWrite.Password`.
- `Secrets.PutAll` is the write an import makes: each `SecretWrite` stores its password at
  the record's location and adds the record when the location has none, all under one save.
  A location no connection uses is overwritten (what a removed connection left behind); one
  in use is a `*SecretInUseError` for that write alone. Returns one refusal per write, in
  order, and an error only when the file could not be saved.
- `Transports` is also the `transport.Provider` connections resolve their route through; it
  looks the `Via` connection up through a `ConnectionFinder` (the `Session`).
- `Workspaces` lists, switches, creates, renames and removes workspaces through `config`.
- Refusals in `errors.go`: `*SecretInUseError`, `*TransportInUseError`,
  `*ConnectionInUseError`, `*DuplicateConnectionError`, `*DuplicateNameError`,
  `*DuplicateTransportError`, `*TransportNotFoundError`, `*ConnectionNotFoundError`,
  `*ActiveWorkspaceError`.

## Rules

`repository.Session` is the composition root. It opens the repository of every type in
`cfg.Connections` through `connectionRepositories`, in that order — `withKnownTypes` keeps
it in `ConnTypes` order, so the tab order is `config`'s to decide — plus the transport and
keyring repositories, and wires a `repository.Index` across all of them: re-pointing or
deleting a secret in use fails with `*SecretInUseError` instead of orphaning a connection.
A repository that owns secrets feeds the index by implementing `SecretUser`
(`SecretRefs() iter.Seq2[string, secret.Reference]`).

**Every typed failure the repository raises lives in `repository/errors.go`**, as
`network/error.go` holds that layer's — one file for a caller using `errors.AsType`. The
files a rule is *computed* in (`secret_usage.go`, `transport_usage.go`,
`connection_usage.go`, `unique.go`) hold the rule, never the error it returns.

**Uniqueness is a property of the store, not of the record.** Only the collection can say
whether a `config.Connection` already exists, so the rule is in `repository/unique.go` —
like `*SecretInUseError`, a typed error enforced at the write point. The key is
`config.Connection.Identity()`, which each type builds through `config.identity(...)`
(unambiguous joining, no per-type separator — see `internal/config/AGENTS.md`). The
repository never recomputes or inspects it; it only compares. `ConnectionRepository.check`
supplies the scope: `others(self)`, an iterator over every row but the one being edited —
`Edit` passes its own index, `Add` passes `-1`.

What a type puts in its identity — every field that changes *where* you land:

| Part | Types |
|---|---|
| type, host, port, username | all |
| database | Postgres, MySQL, MSSQL, ClickHouse, MongoDB |
| schema | Postgres |
| db index | Redis |
| transport reference — two rows differing only by it are not duplicates | every database type, not SSH |
| normalised jump chain, local forward | SSH |

Nothing about *how* you get there — TLS, auth, keepalive — is in it. The deliberate choices:

- **the password is not part of it** — two rows differing only by the secret they resolve
  are the same endpoint twice, the mistake being caught;
- **the schema is** — the same server under another `search_path` is a working context of
  its own, and refusing it would be a false positive on a hard error;
- **an empty name never collides**, since the name is optional on every form; two rows with
  the *same* name are ambiguous for real, because `SecretRefs()` yields `Meta().Name` as the
  key the usage index is built on.

Rejection is the whole strategy: nothing is merged, renamed or overwritten behind the user's
back. The walk carries no domain: `utils.FirstDuplicate`, called once per key —
`config.Connection.Identity` for the endpoint, a closure over `Meta().Name` for the label.

**A refused write never leaves the screen it came from.** `formLoop[T]` in
`ui/form_loop.go` is the one shape every form takes: its `run` opens, calls `save`, and on
failure opens again with the message in the status line as `kindErr`. Nothing is retyped and
`conm` does not exit; cancelling is the way out. Every form shares it — `connectionForm`
behind `RunAddForm` (`conm <type> add`), `runAddConnection` and `runEditConnection` (the
connections table), `secretForm`, `transportForm` and `workspaceForm` (their screens) —
differing only in the title, `isEdit` and the `save` func (`repo.Add`, or the `Add`/`Edit`
of the view the screen holds). **A screen never calls `runForm` and then saves for itself.**
Detail: `internal/ui/AGENTS.md`.

- The retry re-opens with **the form's own values**, `formResult.values`, not a seed rebuilt
  from the record: `SeedFunc` projects what is *stored*, and `KeyringFormSpec` seeds the
  password blank because conm never writes it to disk — a rebuilt seed would blank it on
  every retry and make the loop know which field is the password.
- `DuplicateConnectionError` names the *existing* record rather than the endpoint when it
  can: the status line is `formInner` wide, the endpoint is already on screen in the form,
  and a long host would push the one new fact — which connection it collided with — past
  the truncation.

**A secret's material is probed outside the lock, and only when it has to be.**

- `SecretRepository.Add` and `Edit` call the provider before taking the repository's lock: a
  provider call is arbitrarily slow — it may prompt, unlock a store or cross the network —
  and the render path reads this repository meanwhile.
- `Edit` does not probe when the reference is unchanged, so a metadata-only edit never
  reaches the provider.
- `Add` writes the password only when nothing is stored at that location. The two mismatches
  — material already there with a password typed, or none there and none typed — come back
  as errors *after* the record is saved: information for the status line, not a refused
  write.
- Re-pointing a secret that connections still reference is refused (`*SecretInUseError`):
  it would orphan them.

**Known gaps.** A failed `Close` of a replaced or removed live connection is dropped, as is
a record that is not a `secret.Reference` when the usage index is fed; neither is logged,
because conm has no log. `config`'s wrappers do not check for collisions on `Add`/`Put` —
uniqueness is enforced here, at the write point, and nowhere below.

**A repository answers which client it launches.** `Connections.ClientInfo(ctx)` returns the
`cli.Info` of the launcher the repository was opened with — name, path, installed flag and
version — so a screen never builds a launcher or names a binary. It is the only accessor on
the interface that can start a subprocess (the version probe), hence the context.

**A session hands out repositories, never a named one.** `Connections()` returns them in
`ConnTypes` order (what `view.NewConnections` is built from); `ConnectionsOf(t)` looks one
up by `config.ConnType`, reporting `false` for a type it holds no repository for (what a
form saves through). There is no `Session.Postgres` field, so **no layer above `repository`
names a database type in its control flow** — adding a type is one entry in
`connectionRepositories`.
