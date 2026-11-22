# `internal/config` — the data model

## Purpose

The persisted, validatable data model: every connection type, the secret records, the
transports, the settings in `conm.toml`, the workspaces — and **all TOML file I/O and every
config path**. A leaf: it knows nothing about drivers, clients, repositories or screens.

May import (internal): `utils/keys`, `utils/pathname`.

## Files

| File | Holds |
|---|---|
| `config.go` | `Connection`, the persistence generics `ConfigWrapper[C]` / `File[C]` / `Config[W, C]`, `OpenConfig`, the `validationErrors` collector, `WriteForeign` |
| `database.go` | `DBConnection` — a `Connection` with a database and a schema; `Databases`/`ConnTypes`, `ConnTypeNames()` |
| `distinct.go` | `Distinct` and `identity(...)`, the uniqueness key a record builds for itself |
| `meta.go` | `ConnMeta`, `WebLink` and their validators, `ValidateConnectionName`, `ValidateTags` |
| `conm.go` | `Conm`, `ConnectionSettings`, `ConnType`, `ConnKind`, `ReadConm`/`CreateConm`, `withKnownTypes` |
| `keybinding.go` | `KeybindingSettings` — the keybinding overrides stored in `conm.toml` |
| `appearance.go` | `AppearanceSettings` — border, accent, glyph set, text style and the border and glyph colour palettes, plus per-element overrides; `BorderStyles`, `AccentColors`, `GlyphSets`, `TextStyles` list the preset names, `BorderElements`, `GlyphElements`, `TextElements` the overridable elements |
| `postgres.go`, `mysql.go`, `mssql.go`, `clickhouse.go`, `redis.go`, `mongodb.go`, `ssh.go` | one connection type each: the struct, its `*ConfigWrapper`, its mode constants and validators; `postgres.go` also holds the limits and regexps the types share |
| `secret.go`, `keyring.go` | `Secret`, `SecretParam`; `Keyring`, the one stored secret record, and its wrapper |
| `transport.go` | `Transport`, `TransportKind`, its wrapper and validators |
| `workspace.go` | the active workspace, the startup workspace, create/rename/remove of a workspace's directories |
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

- `Connection` — the endpoint every saved connection has (`Username`, `Host`, `Port`,
  `ConnectionString`, `ConnType`). `Meta()` is the user's annotation, `ConnectionString(secret)`
  the URL with the resolved password injected, `Identity()` the uniqueness key the repository
  compares.
- `DBConnection` — what every database type implements; `SSH` stops at `Connection`.
  `SetHost`/`SetPort` return the copy a transport re-points at its local end.
- `Transport` — a named route (`Kind`, `Via` — the connection it goes through) a database
  connection references.
- `Keyring` — the only `Secret`.
- `Conm` — `conm.toml`: one `ConnectionSettings` row per connection type plus
  `KeybindingSettings` and `AppearanceSettings`.
- `Config[W, C]` — the one `File[C]`: a typed TOML file opened by `OpenConfig[W](path)`.

## Per-type deviations from the shared model

- The network drivers take a `config.DBConnection`. A consumer holding only a
  `config.Connection` (the uniqueness key, the search index, the table row) asserts
  `cfg.(config.DBConnection)` before reading a database name.
- `SSH` has a host, a port, a user and one secret reference but no catalog: the table's
  database column and the uniqueness key's database part are empty for it.
- `Databases` lists the database types, `ConnTypes` every type conm knows (`Databases` plus
  `SSHConnType`).
- Anything a type needs *beyond* the shared accessors is a plain extra field on its own
  struct (`Postgres.SSLMode`, `Redis.TLSMode`).
- Where a type has no honest answer *for* an accessor, its own file takes a documented
  default:

| Type | Accessor | Value | Why |
|---|---|---|---|
| MySQL | `Schema()` | `""` | a MySQL schema *is* a database, which `Database()` already names |
| MSSQL | `Schema()` | `""` | no connection string selects a schema — the default one is a property of the user |
| ClickHouse | `Schema()` | `""` | a database *is* its namespace; single-host form only, a cluster seed list would not fit `Host()` |
| Redis | `Database()` | the decimal DB index — `"0"`, `"3"` | a numbered database, not a named one; stored as `DBIndex int` |
| Redis | `Schema()` | `""` | Redis has no schemas |
| Redis | `Username()` | `""` when unset | ACL users are optional; password-only AUTH is the common case |
| MongoDB | `Schema()` | `""` | a database *is* its namespace; single-host form only, a replica-set seed list would not fit `Host()` |
| MongoDB | `Username()` | `""` when unset | a mongod with authorization off refuses a URI that carries credentials |
| SSH | `ConnectionString(secret)` | `ssh://user@host:port`, the secret ignored | OpenSSH takes that URI as a destination but no password inside it; `auth` decides how `cli` hands the secret over, and `agent`/`key` need none |
| SSH | `Password` | what `auth` needs — `key`: `filepath:<identity>`, or the key material itself as a literal or store ref; `password`: a literal, a store ref or `none:` (ssh prompts); `agent`: nothing | `auth` says what the credential *is*, the scheme *where* it sits; one reference, one credential. `ValidateSSHCredential` refuses a mismatched pair at the write point: it reads the scheme prefix and, for `key` auth with a `filepath:` ref, hands the location to `ValidateSSHIdentity`, which requires a path starting with `~` or `/`. A key's passphrase is a second credential the single `SecretRef()` cannot hold — ssh prompts for it, and `agent` auth avoids the prompt |

## The metadata a connection carries

- Every connection type embeds `ConnMeta` as a field named `Metadata`, tagged `toml:"meta"`,
  and exposes it through `Meta()` on `config.Connection` — Go forbids a field and a method
  under one name. Read it through `Meta()` from outside the package.
- `ConnMeta` — `Name`, `Description`, `Tags` and `Links` — is the user's own annotation and
  is **never sent to the server**: none of it reaches `ConnectionString` or a client's argv.
  Name, description and tags are searchable; links stay out of the search index
  (`internal/ui/view/AGENTS.md`).
- `Links` is an ordered `[]WebLink`, each a `Name` (`grafana`, `logs`, `runbook`) and a `URL`
  — the dashboards, log searches and runbooks of the server the connection reaches. It is
  `toml:"links,omitempty"`: no links writes no `links` key and reads back nil. It sits on
  `ConnMeta` because a link is metadata, not per-database configuration.
- `ValidateWebLinkName`: non-blank, at most 64 characters. `ValidateWebLinkURL`: an absolute
  `http` or `https` url with a host. The pair rule — a row with neither half filled is not a
  link — is `spec.ValidateLink`; see `internal/ui/spec/AGENTS.md`.

## Persistence generics

- `config.Config[W ConfigWrapper[C], C any]` is the concrete `config.File[C]`. A
  `ConfigWrapper` (`PostgresConfigWrapper{Conns []Postgres}`,
  `KeyringConfigWrapper{WorkspaceID string; Keyring []Keyring}`, `ConmConfigWrapper`) owns
  TOML marshal/unmarshal for one file shape.
- Wrappers backing a repository are typed on the *interface* (`ConfigWrapper[Connection]`,
  `ConfigWrapper[Secret]`) and narrow to their concrete struct internally, which keeps the
  repositories non-generic.
- `OpenConfig[W](path)` opens, parses and validates. It creates the file's directory first:
  the config and data directories are under different XDG roots and either may be missing.
- `ConfigWrapper.Add` and `Put` return no error: a record of the wrong concrete type is
  silently ignored, and collisions are not checked — uniqueness is the repository's rule.
- `ValidateSecretLocation` only refuses an empty location.
- `WriteForeign` is the same atomic write for a file conm does not own — a client's config
  `internal/exporter` merged entries into: the directory is created, the file replaced in
  one rename and **left `0600`** whatever it was, since it may now hold a password.
- `SplitSSHOption` (`ssh.go`) cuts an ssh option line into keyword and value: the keyword
  ends at a space, a tab or `=`, the value loses one pair of surrounding double quotes. It
  sits here because both `internal/parser/params` (an `-o` value) and `internal/importer`
  (a line of `~/.ssh/config`) read with it.

## The secret reference stored on a connection

A connection's password field holds what `SecretRef()` returns: a bare literal or
`"<scheme>:<location>"`. `spec.JoinSecret`/`spec.SplitSecret` assemble and split it on the
form side; `internal/parser/params` and `internal/repository` use `secret.Ref`/
`secret.ParseRef`, and `ValidateSSHCredential` reads the scheme. The cross-module rule is in
`internal/AGENTS.md`, the providers in `internal/secret/AGENTS.md`. A bare password
containing a colon stays literal, matching how `secret.Resolver` reads it.

**Not every scheme is a store.** `literal`, `none` and `filepath` have no entries, no
repository and no tab on the secrets screen:

- `literal` carries the password inline.
- `none` carries no password at all: `NoneProvider` resolves to the empty string, and
  `ConnectionString` then omits the password entirely. It is stored as `"none:"` — a bare
  `""` reads back as an empty literal, which is what an unset password already is.
- `filepath` carries a path conm hands to the client but never opens: `FilepathProvider`
  checks the file exists (`~` expanded only for that check) and resolves to the path verbatim.
- `secret.IsStore` is the one test for "this reference points into a store" — the usage
  index's `trackable` and the form's `isRef` ask it, never `!= secret.Literal`.
- `secret.IsMaterial` — `literal` only — is what the form masks and keeps untrimmed; a
  `filepath` value is typed plain like a location.

## The settings row — `conm.toml`

- `conm.toml` holds one `[[conm.connection]]` row per connection type —
  `ConnectionSettings{Type, CLI}`, the type written by name — plus the keybinding overrides.
- `Conm.complete` runs everything read through `withKnownTypes`: one row per type in
  `ConnTypes`, in that order, whatever the file holds. Every layer above may take "a row for
  every known type" for granted, and the tab order is this package's to decide. A type the
  file does not mention gets a row with no client.
- `[conm.appearance]` holds names — `border`, `accent`, `glyphs`, `text` — and `accent` may
  instead be a `#rrggbb` colour (`IsHexColor`). `[conm.appearance.border_runes]`,
  `[conm.appearance.glyph_runes]` and `[conm.appearance.text_style]` map an element name to
  the value that replaces the preset's; each is omitted when empty.
- `border_color` / `glyph_color` / `text_color` name a palette (`BorderPalettes`,
  `GlyphPalettes`, `TextPalettes`) and `[conm.appearance.border_colors]` /
  `[conm.appearance.glyph_colors]` / `[conm.appearance.text_colors]` override one element
  (text colours cover `TextColorElements`: the styled roles plus `body` and `input`, which
  have a colour and no style). `[conm.appearance.text_backgrounds]` sets the background of
  the `TextBackgroundElements` — `tab` and `heading` — with no palette:
  a colour is `auto`, a name from `ColorNames` or `#rrggbb` (`IsColor`). Glyph colours exist
  only for `GlyphColorElements` — the glyphs drawn on their own, not inside text.
- A border or text colour may be a gradient — two or more
  colours joined by `GradientSeparator` (`>`), none of them `auto` (`IsGradient`,
  `IsColorOrGradient`); glyph colours stay flat. The accent may be a gradient of accent
  presets and `#rrggbb` only (`IsAccent`) — no colour names, since `accent` would name itself.
  Whether a gradient is drawn as one is `ui`'s decision, made from the terminal.
- What an override may hold: a border rune is one grapheme one column wide (`IsBorderRune`);
  a glyph is one grapheme one or two columns wide (`IsGlyph`), so an emoji with a modifier
  counts as one — grapheme counting is `rivo/uniseg`; a text style is `plain` or distinct
  `TextAttributes` joined by `+` (`IsTextStyle`).
- `Conm.complete` replaces a missing or unknown name with the first entry of its list and
  drops an override for an unknown element or one that does not fit its kind, so every layer
  above gets values it can draw;
  `AppearanceSettings.Validate` refuses the same at the write point. What a name looks like
  is `ui/theme.go`'s to decide.
- `ConnType.Kind()` says which `ConnKind` (`database`, `ssh`) a type belongs to; it is
  derived, never stored.
