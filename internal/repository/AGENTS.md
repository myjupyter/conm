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

- `Session` is one open workspace: `Connections()` / `ConnectionsOf(t)` hand out the
  connection repositories by `config.ConnType`, `Keyring` and `Transports` are the two other
  stores. Built by `NewSession(conm)`, closed with `Close()`.
- `Connections` is not generic: `Kind()` says which `config.ConnType` every entry has, so a
  caller knows what `ConfigAt` asserts to, and `Add`/`Edit` reject a foreign type.
- `AddAll` saves several connections with one write or none of them: each is admitted
  against the stored ones and the ones before it in the same call, so two entries of one
  batch cannot be duplicates of each other. A batch that fails — at admission, at building a
  live connection or at the save — leaves nothing behind: what was built is closed and the
  records are taken back out. `Add` is `AddAll` of one. `Check` is the same
  admission without the write, with `pending` standing for connections that would be added
  first; it is what lets an import refuse a whole file before touching it.
- `Secrets`: the password travels beside the record, never inside it, and only from the form
  into `Add`/`Edit`.
- `Transports` is also the `transport.Provider` connections resolve their route through; it
  looks the `Via` connection up through a `ConnectionFinder` (the `Session`).
- `Workspaces` lists, switches, creates, renames and removes workspaces through `config`.
- Refusals in `errors.go`: `*SecretInUseError`, `*TransportInUseError`,
  `*ConnectionInUseError`, `*DuplicateConnectionError`, `*DuplicateNameError`,
  `*DuplicateTransportError`, `*TransportNotFoundError`, `*ConnectionNotFoundError`,
  `*ActiveWorkspaceError`.

## Rules

`repository.Session` is the composition root: it walks `cfg.Connections` and opens the
repository of every connection type in it — `connectionRepositories` maps a `config.ConnType`
to its constructor. Order comes from `cfg.Connections`, which `withKnownTypes` keeps
in `ConnTypes` order, so the tab order is `config`'s to decide. It also opens the transport
repository and the keyring repository, and wires a `repository.Index` across all of them, so a secret knows
which connections reference it. That index is what makes editing a re-pointed or deleting a
used entry fail with `*SecretInUseError` instead of orphaning a connection. A repository
that owns secrets implements `SecretUser` (`SecretRefs() iter.Seq2[string, secret.Reference]`)
to feed the index.

**Every typed failure the repository raises lives in `repository/errors.go`**, the way
`network/error.go` holds that layer's vocabulary — the list is under *Main entities* — and
the file each rule is *computed* in (`secret_usage.go`, `transport_usage.go`,
`connection_usage.go`, `unique.go`) holds the rule, never the error it returns. A new refusal goes there too, so a caller reading for
`errors.AsType` has one file to look at.

**Uniqueness is a property of the store, not of the record.** A `config.Connection` cannot
say whether it already exists — only the collection holding it can — so the rule lives beside
the other cross-record rule of its kind, `secret_usage.go`, in `repository/unique.go`. It is
the same species as `*SecretInUseError`: an invariant across records, a typed error, enforced
at the write point. The key is `config.Connection.Identity()`: each type builds its own, through
`config.identity(...)`, which joins the parts unambiguously so no type invents a separator. A
type puts in it every field that changes *where* you land — type, host, port, username, and
then its own: the database (Postgres adds the schema, Redis the db index), SSH its normalised
jump chain and local forward — and nothing about *how* you get there: TLS, auth, keepalive.
The repository never recomputes or inspects it; it only compares. `ConnectionRepository.check` supplies only the scope: `others(self)`, an
iterator over every row but the one being edited, which is what lets `Edit` pass its own
index and `Add` pass `-1`. Three deliberate choices in that key:

- **the password is not part of it** — two rows differing only by the secret they resolve are
  the same endpoint twice, which is the mistake being caught;
- **the schema is** — the same server under another `search_path` is a working context of its
  own, and refusing it would be a false positive on a hard error;
- **an empty name never collides**, since the name is optional on every form; two rows
  claiming the *same* label are ambiguous for real, because `SecretRefs()` yields `Name()` as
  the key the usage index is built on.

Rejection is the whole strategy: nothing is merged, renamed or overwritten behind the user's
back. The walk itself carries no domain and is `utils.FirstDuplicate`, called once per key:
`config.Connection.Identity` for the endpoint and the `config.Connection.Name` method expression for the label.

**A refused write never leaves the screen it came from.** `ui/form_loop.go` holds
`formLoop[T]`, and its `run` is the one shape every form in the tree takes: open, `save`, and
on failure open again with the message in the status line as `kindErr`. Nothing is retyped and
`conm` does not exit; cancelling is the way out. Every form shares it — `connectionForm`
behind `RunAddForm` (`conm <type> add`), `runAddConnection` and `runEditConnection` (the
connections table), `secretForm`, `transportForm` and `workspaceForm` (their screens) — and
they differ only in what they build the loop with: the title, `isEdit`, and the `save` func
(`repo.Add`, or the `Add`/`Edit` of the view the screen holds). **A screen
never calls `runForm` and then saves for itself.**

The retry re-opens the form with **the form's own values**, `formResult.values`, not with a
seed rebuilt from the record. That is not an optimisation: `SeedFunc` is a projection of what
is *stored*, and the secret form's password is deliberately never stored — `KeyringFormSpec`
seeds it blank, because conm does not write it to disk. Rebuilding the seed would therefore
blank the password on every retry, and would also make the loop need to know which field is
the password. Feeding back the values it already has keeps both problems from existing.
That is also why `DuplicateConnectionError` names the *existing* record rather than repeating
the endpoint when it can: the status line is `formInner` wide, the endpoint is already on
screen in the form above it, and a long host would push the one new fact — which connection it
collided with — past the truncation.

**A secret's material is probed outside the lock, and only when it has to be.**
`SecretRepository.Add` and `Edit` call the provider before taking the repository's lock: a
provider call is arbitrarily slow — it may prompt, unlock a store or cross the network —
and the render path reads this repository meanwhile. `Edit` does not probe at all when the
reference is unchanged, so a metadata-only edit never reaches the provider. `Add` writes
the password only when nothing is stored at that location yet; the two mismatches — material
already there with a password typed, or none there and none typed — come back as errors
*after* the record is saved, so they are information for the status line, not a refused
write. Re-pointing a secret that connections still reference is refused
(`*SecretInUseError`), because it would orphan every one of them.

**Known gaps.** A failed `Close` of a replaced or removed live connection is dropped, as is
a record that is not a `secret.Reference` when the usage index is fed; neither is logged,
because conm has no log. `config`'s wrappers do not check for collisions on `Add`/`Put` —
uniqueness is enforced here, at the write point, and nowhere below.

**A repository answers which client it launches.** `Connections.ClientInfo(ctx)` returns the
`cli.Info` of the launcher the repository was opened with — the name, path, installed flag
and version — so a screen asks the layer it already holds instead of building a launcher or
naming a binary. It is the only accessor on the interface that can start a subprocess (the
version probe), hence the context.

**A session hands out repositories, never a named one.** `Connections()` returns them
in `ConnTypes` order (what `view.NewConnections` is built from) and `ConnectionsOf(t)` looks
one up by `config.ConnType`, reporting `false` for a type it holds no repository for (what a
form saves through). There is no `Session.Postgres` field to
reach for, so **no layer above `repository` names a database type in its control flow** —
adding a type is one entry in `connectionRepositories`, and every screen keeps working.
