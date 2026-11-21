# `internal/config` — the data model

## Purpose

The persisted, validatable data model: every connection type, the secret records, the
transports, the settings in `conm.toml`, the workspaces — and **all TOML file I/O and every
config path**, plus the read-only readers of the files other clients keep their connections
in. A leaf: it knows nothing about drivers, clients, repositories or screens.

May import (internal): `utils/keys`, `utils/pathname`.

## Files

| File | Holds |
|---|---|
| `config.go` | `Connection`, the persistence generics `ConfigWrapper[C]` / `File[C]` / `Config[W, C]`, `OpenConfig`, the shared validators |
| `database.go` | `DBConnection` — a `Connection` with a database and a schema |
| `distinct.go` | `Distinct` and `identity(...)`, the uniqueness key a record builds for itself |
| `meta.go` | `ConnMeta`, `WebLink` and their validators |
| `conm.go` | `Conm`, `ConnectionSettings`, `ConnType`, `ConnKind`, `Databases`/`ConnTypes`, `ReadConm`/`CreateConm`, `withKnownTypes` |
| `keybinding.go` | `KeybindingSettings` — the keybinding overrides stored in `conm.toml` |
| `postgres.go`, `mysql.go`, `mssql.go`, `clickhouse.go`, `redis.go`, `mongodb.go`, `ssh.go` | one connection type each: the struct, its `*ConfigWrapper`, its mode constants and validators |
| `secret.go`, `keyring.go` | `Secret`, `SecretParam`; `Keyring`, the one stored secret record, and its wrapper |
| `transport.go` | `Transport`, `TransportKind`, its wrapper and validators |
| `workspace.go` | the active workspace, the startup workspace, create/rename/remove of a workspace's directories |
| `ini.go` | `ReadINI` — the sectioned `key=value` format of `pg_service.conf` and MySQL option files, as `INISection`s in file order |
| `foreign.go` | where the clients' own files are (`PGServicePath`, `PGPassPath`, `MySQLOptionPath`) and `PGPass`, the rows of a `.pgpass` with libpq's lookup |
| `xdg.go` | every path: `ConmPath()`, `PostgresPath()`, `TransportPath()`, `SecretConfigPath()`, … — workspace-scoped except `conm.toml` |

## Main entities

```go
type Connection interface {
    Meta() ConnMeta
    Username() string; Host() string; Port() int
    ConnectionString(secret string) string
    ConnType() ConnType
    IsValid() bool; Validate() []error
    Distinct
}

type DBConnection interface {
    Connection
    Database() string; Schema() string
    SetHost(host string) DBConnection; SetPort(port int) DBConnection
}

type Distinct interface{ Identity() string }

type Secret interface {
    ID() string; Provider() string; Location() string; Description() string
    Params() []SecretParam
    IsValid() bool; Validate() []error
}
```

- `Connection` is the endpoint every saved connection has; `Meta()` is the user's annotation
  and never reaches the server, `ConnectionString(secret)` is the URL with the resolved
  password injected, `Identity()` the uniqueness key the repository compares.
- `DBConnection` is what every database type implements; `SSH` stops at `Connection`.
  `SetHost`/`SetPort` return the copy a transport re-points at its local end.
- `Transport` is a named route (`Kind`, `Via` — the connection it goes through) a database
  connection references; `Keyring` is the only `Secret`.
- `Conm` is `conm.toml`: one `ConnectionSettings` row per connection type plus
  `KeybindingSettings`.
- `Config[W, C]` is the one `File[C]`: a typed TOML file opened by `OpenConfig[W](path)`.

## Per-type deviations from the shared model

`config.Connection` is the endpoint every saved connection has — `Username`, `Host`, `Port`,
`ConnectionString`, `ConnType` — and `config.DBConnection` embeds it and adds `Database()` and
`Schema()`. Every database type implements `DBConnection`; the network drivers take it, and
a consumer that only holds a `Connection` (the uniqueness key, the search index, the table
row) asserts `cfg.(config.DBConnection)` before reading a database name. `SSH` is the one
type that stops at `Connection`: it has a host, a port, a user and one secret reference,
but no catalog, so the table's database column and the uniqueness key's database part are
simply empty for it. `Databases` lists the database types, `ConnTypes` every type conm
knows (`Databases` plus `SSHConnType`), and `ConnType.Kind()` says which `ConnKind` a type
belongs to. Anything a type
needs *beyond* them is a plain
extra field on its own struct (`Postgres.SSLMode`, `Redis.TLSMode`) and costs nothing
shared. Where a database has no honest answer *for* one of the database accessors, its own file takes a
documented default — these are the ones in the tree:

| Type  | Accessor     | Value                              | Why |
|-------|--------------|------------------------------------|-----|
| MySQL | `Schema()`   | `""`                                | in MySQL a schema *is* a database, which `Database()` already names |
| MSSQL | `Schema()`   | `""`                                | SQL Server has schemas, but no connection string selects one — the default schema is a property of the user |
| ClickHouse | `Schema()` | `""`                              | a ClickHouse database *is* its namespace; only the single-host form is supported, a cluster seed list would not fit `Host()` |
| Redis | `Database()` | the decimal DB index — `"0"`, `"3"` | Redis selects a numbered database, not a named one; it is stored as `DBIndex int` |
| Redis | `Schema()`   | `""`                                | Redis has no schemas |
| Redis | `Username()` | `""` when unset                     | ACL users are optional; password-only AUTH is the common case |
| MongoDB | `Schema()` | `""`                              | a MongoDB database *is* its namespace; only the single-host form is supported, a replica-set seed list would not fit `Host()` |
| MongoDB | `Username()` | `""` when unset                   | a mongod with authorization off refuses a URI that carries credentials, so an authless server is reachable only without them |
| SSH | `ConnectionString(secret)` | `ssh://user@host:port`, the secret ignored | OpenSSH takes that URI as a destination but no password inside it; `auth` decides how `cli` hands the secret over, and `agent`/`key` need none |
| SSH | `Password` | what `auth` needs — `key`: `filepath:<identity>`, or the key material itself as a literal or store ref; `password`: a literal, a store ref or `none:` (ssh prompts); `agent`: nothing | `auth` says what the credential *is*, the scheme says *where* it sits; one reference, one credential. `ValidateSSHCredential` refuses a mismatched pair at the write point, reading only the scheme prefix. A key's passphrase is a second credential the single `SecretRef()` cannot hold — ssh prompts for it, and `agent` auth avoids the prompt |

## The metadata a connection carries

Every connection type embeds `ConnMeta` as a field named `Metadata`, tagged `toml:"meta"`,
and exposes it through the `Meta()` accessor on `config.Connection`. The field and the
accessor are spelled differently because Go forbids a type having both under one name — read
the metadata through `Meta()` from outside the package.

`ConnMeta` is the user's own annotation and is **never sent to the server**: `Name`,
`Description`, `Tags` and `Links`. None of them reaches `ConnectionString`, a client's argv
or the search index's reachable text.

`Links` is an ordered `[]WebLink`, each a `Name` (`grafana`, `logs`, `runbook`) and a `URL`
— the dashboards, log searches and runbooks that belong to the server this connection
reaches. It is `toml:"links,omitempty"`, so a connection with no links writes no `links`
key at all and reads back nil. It is on `ConnMeta` rather than on any concrete type because
a link is metadata, not per-database configuration: every connection type has them, and none
of them had to declare anything.

`ValidateWebLinkName` and `ValidateWebLinkURL` are the validators: a name must be non-blank
and at most 64 characters, and a url must parse as an
absolute `http` or `https` url with a host. The pair rule — that a row with neither half
filled is not a link at all — lives one layer up in `spec.ValidateLink`, because it is a
property of a form row rather than of a stored record; see `internal/ui/spec/AGENTS.md`.

## Persistence generics

`config.Config[W ConfigWrapper[C], C any]` is the concrete
`config.File[C]`. A `ConfigWrapper` (`PostgresConfigWrapper{Conns []Postgres}`,
`KeyringConfigWrapper{Secrets []Keyring}`, `ConmConfigWrapper`) owns TOML marshal/unmarshal
for one file shape. Wrappers backing a repository are typed on the *interface*
(`ConfigWrapper[Connection]`, `ConfigWrapper[Secret]`) and narrow to their concrete struct
internally, which is what lets the repositories stay non-generic. `OpenConfig[W](path)`
opens, parses and validates; it creates the file's directory first, since the config and
data directories live under different XDG roots and either may be missing.

`ConfigWrapper.Add` and `Put` return no error: a record of the wrong concrete type is
silently ignored, and collisions are not checked at this level — uniqueness is the
repository's rule. `ValidateSecretLocation` only refuses an empty location.

## The secret reference stored on a connection

A connection's password field holds what `SecretRef()` returns: either a bare literal or
`"<scheme>:<location>"`. `spec.JoinSecret`/`spec.SplitSecret` are the only place that format
is assembled and taken apart. A bare password containing a colon stays literal, matching how
`secret.Resolver` reads it, so existing configs need no migration.

**Not every scheme is a store.** `secret.IsStore` is the one test for that: `literal` carries
the password inline, `none` carries no password at all (`NoneProvider` resolves to the
empty string, and `ConnectionString` then omits the password entirely) and `filepath`
carries a path conm hands to the client but never opens (`FilepathProvider` resolves to the
path verbatim), so none of them has entries, a repository or a tab on the secrets screen.
`secret.IsMaterial` is the other test — `literal` only — and is what the form masks and
keeps untrimmed; a `filepath` value is typed plain like a location. Everything that means "this reference
points into a store" — the usage index's `trackable`, the form's `isRef` — asks
`secret.IsStore`, never `!= secret.Literal`. The `none` reference is stored as `"none:"`: a
bare `""` would read back as an empty literal, which is what a connection whose password is
merely unset already is.

## The clients' own files

`ini.go` and `foreign.go` read files conm does not own and never writes. They return the
file's own shape — sections and pairs, password rows — and know nothing of connections: what
a section means is `internal/importer`'s to decide.

- `ReadINI` drops comment lines (`#`, `;`), blank lines and anything before the first
  section. A line without `=` is a key with an empty value, which is how an option file
  spells a flag.
- `PGServicePath` and `PGPassPath` honour `PGSERVICEFILE` and `PGPASSFILE` before falling
  back to the home directory, as libpq does.
- `ReadPGPass` treats a missing file as an empty one, since the file is optional. A row has
  exactly five fields, `\:` and `\\` are the escapes, and `PGPass.Password` returns the first
  row whose host, port, database and user each match or are `*`.

## The settings row — `conm.toml`

`conm.toml` holds one `[[conm.connection]]` row per connection type —
`ConnectionSettings{Type, CLI}`, the type written by name — plus the keybinding overrides.
`Conm.complete` runs everything read through `withKnownTypes`: one row per type in
`ConnTypes`, in that order, whatever the file holds, so every layer above may take "a row
for every known type" for granted and the tab order is this package's to decide. A type the
file does not mention gets a row with no client. `ConnType.Kind()` says which `ConnKind`
(`database`, `ssh`) a type belongs to; it is derived, never stored.
