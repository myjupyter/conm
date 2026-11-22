# `internal/exporter` — connections written as entries of a client's own config file

## Purpose

The inverse of `internal/importer`: turns `config.Connection`s into the entries of the file a
client keeps its connections in — a service of `pg_service.conf`, the groups of a MySQL
option file, a `connections_credentials` connection, a sqlcmd context, a `Host` block — for
`conm <type> export`. It decides what a connection *is* in that file and which of them can
go in, and merges them into what the file already holds; it resolves no secret and prints
nothing.

May import (internal): `config`, `secret`, `importer`.

## Files

| File | Holds |
|---|---|
| `exporter.go` | `Entry`, the `targets` table, `Supported`, `DefaultPath`, `WritesPasswords`, `Plan`, `Write`; `heldBy`, `appended` and `textBlocks`, the text merge the line-based formats share |
| `postgres.go`, `mysql.go`, `clickhouse.go`, `mssql.go`, `ssh.go` | per type: `<type>Sections` — one connection as `importer.Section`s, with the warnings for what the format cannot hold — and, beside it, what the file holds and its merger (`sshHeld`/`sshMerged`, `clickhouseMerged`, `sqlConfigHeld`/`sqlConfigMerged`) |
| `ini.go` | `iniMerged` — the merger `pg_service.conf` and the MySQL option file share |
| `yaml.go` | the `yaml.Node` helpers the ClickHouse YAML and sqlconfig mergers share |

## Main entities

- `Entry` is one connection on its way out: its `Name`, the `Sections` it becomes, the
  `Warnings` for what was left behind, and `Err` when it cannot be written.
- `targets` maps a `config.ConnType` to its default path, `held` — the names the file
  already has — `merged`, which turns the file's bytes and the new sections into its new
  bytes, the type's `sections` function, `lineBased` for the formats a line break would
  break, and `passwords` for the ones that can hold a password. A type not in it has no
  `export` command: `cmd` asks `Supported`. The types are the ones `importer.sources` has.
- **The default path is one the user owns.** It is the importer's, except ClickHouse:
  `importer.ClickHouseUserPath`, not the read path, which may be a `clickhouse-client.xml`
  in the working directory or the system config under `/etc`.

## Rules

- **`Plan` writes nothing and `Write` decides nothing.** `Plan` reads the names the file
  holds and marks every refusal on its `Entry`; `Write` merges the sections of the entries
  without an `Err`. The caller chooses between all-or-nothing and skipping. **Only `Write`
  touches the disk for writing, and only through `config.WriteForeign`** — atomic, and the
  file is left `0600`; a merger is bytes in, bytes out.
- **An entry is refused, never replaced**: when the file already has a section of one of its
  names, or an earlier entry of the same plan took it. A merger only adds, so `Write` must
  be given entries that went through `Plan` against the same file.
- **A name must be writable in every format**: letters, digits and `_ . @ -`. A conm name
  may hold spaces or brackets, which would be two `Host` aliases or a broken group header.
- **The password is the caller's to hand over**, already resolved, and empty when it is not
  to be exported. Nothing here resolves a reference; `ssh.go` only reads the scheme of one to
  tell a key file from stored key material. `WritesPasswords` tells the caller whether a
  type has anywhere to put one — SSH has not, so its secrets are never resolved.
- **What a format cannot hold is a warning, not a refusal**: a transport (the client would
  connect directly), the database and the encryption settings of an MSSQL connection, the
  password of an MSSQL connection without a user, an SSH password, an SSH key that is not a
  file.
- **A value with a line break is the entry's refusal**, marked by `Plan` for a `lineBased`
  target, so `--skip-errors` can pass over it. `textBlocks` refuses it again as the last
  guard against a line of its own.
- **The output is what `internal/importer` reads back** — `exporter_test.go` round-trips
  every type through `importer.Read`. A keyword or option added here must be one the
  importer's parser maps.

### Merging

- A missing file is an empty one.
- `appended` (INI) adds text after what is there and leaves the existing bytes alone.
  Values are written verbatim — quoting is the `sections` function's, which knows what a
  value means.
- `sshMerged` puts the `Host` blocks **before the first option of the file**, since ssh takes
  the first value it finds: below a `Host *`, or below a global option, the exported value
  would lose to it. Leading comments stay above them. When that first option is a global one
  rather than a `Host` or `Match` line, a `Host *` line follows the exported blocks, so the
  old global options — a top-level `Include` among them — still apply to every host instead
  of joining the last exported block. A file of comments only gets them at the end.
- `clickhouseMerged` follows the content as the reader does (XML starts with `<`); an empty
  file is YAML only by a `.yaml`/`.yml` name. XML is merged as text, so comments and layout
  survive; the place is found with `xml.Decoder` offsets, so a tag inside a comment is not
  one. The connections go before the last `</connections_credentials>` under the root, into
  a self-closing `<connections_credentials/>` opened up, or in a new one before the root's
  closing tag. The result is parsed and must hold exactly the old and the new connections.
- The YAML mergers go through `yaml.Node`: comments and unknown keys are kept,
  **indentation, blank lines and quoting are re-serialised**. A file with a second document
  is refused, since re-encoding keeps only the first. Every written scalar is tagged `!!str`,
  so a value YAML would read as a null, a number or a bool is quoted. A single ClickHouse
  `connection` map becomes a list.
- `sqlConfigMerged` writes an endpoint, a context and — when there is a username — a user,
  all three named after the section, through the `importer.SQLConfig*` types, and
  `version: v1` into a file that had no keys. `currentcontext` is set to the first exported
  context only when the file has none, so sqlcmd has a context to use. Hence `sqlConfigHeld` counts endpoints and users as taken
  names too, and `sshHeld` every alias of every `Host` line.

### Per format

- Postgres: a schema is `options=-c search_path=<schema>`, which the parser reads back.
- MySQL: `[client_<name>]` holds host, port, user, `loose-ssl-mode` and the password,
  `[mysql_<name>]` the database — `database` in a `client` group breaks every other MySQL
  program that reads it. The connection is the one `mysql --defaults-group-suffix=_<name>`
  opens. Every string value is double-quoted with `\`, `"` and the line breaks escaped, so
  `#` does not start a comment and a backslash is not an escape. `loose-` makes the MariaDB
  client, which has no `ssl-mode`, skip the option instead of failing. `preferred`, the
  client's own default, writes none.
- MSSQL: the password is base64 with `password-encryption: none`, as sqlcmd stores an
  unencrypted one.
- SSH: `LocalForward` is written as `port host:port`, the two arguments ssh_config takes; an
  `IdentityFile` with a space is double-quoted; `ProxyJump` is `config.NormalizeSSHJump`ed,
  since a hop list with spaces is more than one argument.

## Adding a format

1. `internal/exporter/<type>.go` with a `<type>Sections` and the type's merger (`iniMerged`
   if the file is sectioned key/value).
2. The reader of the names the file holds — `heldBy` over the importer's reader, or the
   type's own when an entry takes more than one name.
3. A row in `targets`, and a case in `TestPlan_roundTrip`.
4. The type's row in the export table of `docs/user/add-connection.md`.
