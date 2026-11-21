# `internal/importer` — connections read from a client's own config file

## Purpose

Turns the file a client keeps its connections in — `pg_service.conf`, a MySQL option file,
the ClickHouse client config, sqlcmd's `sqlconfig`, `~/.ssh/config` — into `config.Connection`s for `conm <type> import`. It decides what an entry of that file
*is* and how it maps onto a connection; it does not open files itself, save anything or
print anything.

May import (internal): `config`, `parser/params`, `secret`.

## Files

| File | Holds |
|---|---|
| `importer.go` | `Entry`, the `sources` table, `Supported`, `DefaultPath`, `Read`, `SecretLocation`, `withoutPassword` |
| `postgres.go` | `readPostgres` — one entry per service of a `pg_service.conf` |
| `mysql.go` | `readMySQL` — the entries of an option file |
| `clickhouse.go` | `readClickHouse` — one entry per `connections_credentials` connection |
| `mssql.go` | `readMSSQL` — one entry per sqlcmd context |
| `ssh.go` | `readSSH` — one entry per concrete `Host` alias |

## Main entities

- `Entry` is one connection found in a foreign file: its `Name`, the `Conn` built from it,
  the plain-text `Password` the file held, the parser's `Warnings`, and `Err` when the entry
  could not be read as a connection at all — in which case `Conn` is nil.
- `sources` maps a `config.ConnType` to its default path and its reader. A type that is not
  in it has no `import` command: `cmd` asks `Supported`. Redis and MongoDB are not in it:
  `.redisclirc` holds preferences and `mongosh.conf` shell settings, neither a connection.

## Rules

- **An entry is parsed as the command line it stands for.** A service is a keyword/value
  connection string and an option-file line `key=value` is `--key=value` to the client, so a
  reader assembles that input and hands it to `params.ParseArgs`. The mapping of keywords and
  flags to fields, the warnings and the TLS refusals stay in `internal/parser/params` and are
  the same ones `conm <type> add` applies.
- **One unreadable entry never fails the file.** A reader returns an error only when the file
  itself cannot be read; a refused entry carries its reason in `Entry.Err`.
- **The password never stays in `Entry.Conn`.** The parser returns it as a `literal:`
  reference; `withoutPassword` moves it to `Entry.Password` and leaves a
  `keyring:<type>/<name>` reference (`SecretLocation`) on the connection, or nothing when the
  file held no password. An SSH identity file is not a password: its `filepath:` reference
  stays on the connection as the parser built it.
- **A field the entry omits is filled in only when the client's own default is a fixed
  value** — ClickHouse's user and database `default`, the `master` database of a sqlcmd
  context — and the entry says so in a warning. A default that depends on the machine, such
  as the OS user psql and ssh fall back to, is not guessed: the entry is refused by
  validation and can be completed in the form. Storing the material is the caller's job, after the connection is
  saved.
- **File I/O stays in `config`**: `config.ReadINI`, `config.ReadPGPass` and the default
  paths. A reader for a new format adds its file reader there.

### Postgres

A service without a `password` gets the one `.pgpass` holds for its host, port, database and
user, the way libpq picks it: first matching row, `*` matching anything. `.pgpass` rows are
never connections of their own — they have no name and are mostly wildcards.

### MySQL

- `[client]` and `[mysql]` together are one connection, named `client`.
- Every `[client<suffix>]` group is another one, named after the suffix without its leading
  `_`, `-` or `.` — the connection `mysql --defaults-group-suffix=<suffix>` opens. Its options
  are those of `[client]` and `[mysql]` followed by `[client<suffix>]` and `[mysql<suffix>]`.
- Suffixes are discovered from `client…` groups only: `[mysqld]` and `[mysqldump]` start with
  `mysql` and are other programs, not suffixed groups. `[client-server]` and
  `[client-mariadb]` are shared MariaDB groups, not suffixes either.
- Only the last of a repeated option is passed on. The parser ignores an empty value, so
  without this an empty `password=` in a suffixed group could not take back the password it
  inherits.
- `!include` and `!includedir` lines are dropped, included files are not followed, a
  `# comment` after a value is not stripped, and the obfuscated `~/.mylogin.cnf` is not read.

### ClickHouse

- The file is the client's own config, XML or YAML by its extension; `config.ClickHouseClientPath`
  picks the first that exists in the client's search order.
- A connection without `hostname` uses its `name` as the host, as the client does.
- `secure` and `accept-invalid-certificate` become their flags only when true; the second is
  one the parser refuses, so such a connection is not importable.
- Every other key (`history_file`, `prompt`, …) is a warning.

### MSSQL

- A context is a connection: the endpoint it names gives the server, the user it names the
  login. A context whose endpoint is not in the file is an `Entry.Err`.
- **The stored password is base64 of the plain text when `password-encryption` is `none` or
  absent** — sqlcmd encodes it even unencrypted. With any other method (`dpapi` on Windows)
  the material cannot be read here: the entry is imported without a password and says so.
- `sqlconfig` has no database, hence the `master` default above.

### SSH

- Every alias on a `Host` line that holds none of `*`, `?`, `!` is a connection, once.
- **A pattern block is never a connection but its options apply.** The options of every
  block whose patterns match the alias are passed in file order as `-o Key=Value`, and the
  parser keeps the first value of each — which is how ssh itself reads the file. Options
  before the first `Host` are the block `*`. A negated pattern that matches excludes its
  block.
- The alias is passed as the destination, so it is the host unless a `HostName` came first.
- `Match` blocks are dropped by `config.ReadSSHConfig`: their conditions cannot be evaluated
  without running ssh. `Include` is an option like any other and ends as a warning.

## Adding a format

1. The file reader and the default path in `internal/config/foreign.go` (or `ini.go` if the
   file is sectioned key/value).
2. `internal/importer/<type>.go` with a `read<Type>(path) ([]Entry, error)`.
3. A row in `sources`.
4. The type's row in the table of `docs/user/add-connection.md`.
