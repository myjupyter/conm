# `internal/importer` — connections read from a client's own config file

## Purpose

Turns the file a client keeps its connections in — `pg_service.conf`, a MySQL option file,
the ClickHouse client config, sqlcmd's `sqlconfig`, `~/.ssh/config` — into
`config.Connection`s for `conm <type> import`. It knows where each file is, reads it into
`Section`s and decides what an entry *is* and how it maps onto a connection; it saves
nothing and prints nothing. The files are only read here — `internal/exporter` is what adds
entries to them.

May import (internal): `config`, `parser/params`.

## Files

| File | Holds |
|---|---|
| `importer.go` | `Entry`, `Section`/`Param`, the `sources` table, `Supported`, `DefaultPath`, `Read` |
| `ini.go` | `ReadINI` — the sectioned `key=value` format `pg_service.conf` and MySQL option files share |
| `postgres.go` | `readPostgres` — one entry per service of a `pg_service.conf`; `PGServicePath`, `PGPassPath`, `PGPass` with libpq's lookup, `ReadPGPass` |
| `mysql.go` | `readMySQL` — the entries of an option file; `MySQLOptionPath` |
| `clickhouse.go` | `readClickHouse` — one entry per `connections_credentials` connection; `ClickHouseClientPath`, `ClickHouseUserPath`, `ReadClickHouseConnections` |
| `mssql.go` | `readMSSQL` — one entry per sqlcmd context; `SQLConfigPath`, `ReadSQLConfig`, the `SQLConfig*` keys and types |
| `ssh.go` | `readSSH` — one entry per concrete `Host` alias; `SSHConfigPath`, `ReadSSHConfig` |

## Main entities

- `Entry` is one connection found in a foreign file: its `Name`, the `Conn` built from it,
  the parser's `Warnings`, and `Err` when the entry could not be read as a connection — in
  which case `Conn` is nil.
- `Section` is a file's own shape — a name and its `Param`s in file order, whatever the
  syntax — and knows nothing of connections. `internal/exporter` writes the same shape back.
- `sources` maps a `config.ConnType` to its default path and its reader. A type not in it
  has no `import` command: `cmd` asks `Supported`. Redis and MongoDB are not in it:
  `.redisclirc` holds preferences and `mongosh.conf` shell settings, neither a connection.

## Rules

- **An entry is parsed as the command line it stands for.** A service is a keyword/value
  connection string and an option-file line `key=value` is `--key=value` to the client; a
  reader assembles that input and hands it to `params.ParseOptions`, which takes arguments
  only, so no value of a file is read as a client binary or an environment assignment. The
  mapping of keywords and flags to fields, the warnings and the TLS refusals stay in
  `internal/parser/params` and are the ones `conm <type> add` applies.
- **One unreadable entry never fails the file.** A reader returns an error only when the file
  itself cannot be read; a refused entry carries its reason in `Entry.Err`.
- **The password reference is the parser's and is not touched here.** A password the file
  held is on `Entry.Conn` as the `literal:` reference `params` wrote, an SSH identity file as
  its `filepath:` one. Moving a literal to the keyring is the caller's job
  (`ui.keyringBound`): it goes through the form's secret format and depends on the name the
  connection is finally saved under.
- **A field the entry omits is filled in only when the client's own default is a fixed
  value** — ClickHouse's user and database `default`, the `master` database of a sqlcmd
  context — and the entry says so in a warning. A machine-dependent default, such as the OS
  user psql and ssh fall back to, is not guessed: validation refuses the entry and it can be
  completed in the form.
- **Each type's file reader and default path sit in that type's file**, exported where
  `internal/exporter` needs them to see what a file already holds. So are the format's names
  both sides spell — `ClickHouseNameKey`, `YAMLNullTag`, `MSSQLNoEncryption`, the
  `SQLConfig*` keys and types — and `Values`, so the reader and the writer of a format keep
  one copy. This is the one place outside `config` that opens a file, and it never writes
  one.

### Postgres

- A service without a `password` gets the one the password file holds for its host, port,
  database and user, as libpq picks it: first matching row, `*` matching anything, a service
  without a host looked up as `localhost`. The file is the service's own `passfile` when it
  names one, else `config.PGPassPath()`. The password found is put in front of the service's
  keywords and the service parsed again, so the reference is still the parser's.
- A password file that cannot be read is one without rows, as for libpq: it is optional and
  never fails the import.
- **The first of a repeated keyword is kept** in what reaches the parser, as libpq never
  overrides a keyword a service file already set: `conninfo` drops the later ones. The
  `passfile` lookup alone goes through `values()` and reads the last.
- `.pgpass` rows are never connections of their own — no name, mostly wildcards.

### MySQL

- `[client]` and `[mysql]` together are one connection, named `client` — when they name a
  `host`, or when the file has no suffixed group. Otherwise they are the defaults the
  suffixed groups share: an entry made of them would only be a refusal that fails a strict
  import of the connections the file does hold.
- Every `[client<suffix>]` group is another one, named after the suffix without its leading
  `_`, `-` or `.` — the connection `mysql --defaults-group-suffix=<suffix>` opens. A suffix
  whose short name is empty or already taken (`[client_a]` beside `[client-a]`) keeps its
  whole spelling. The name `client` is taken only when `[client]` or `[mysql]` hold
  options: `[client_client]`, which export writes for a connection named `client`, is
  `client` again in a file without them. Its options are those of `[client]` and `[mysql]` followed by
  `[client<suffix>]` and `[mysql<suffix>]`.
- Suffixes are discovered from `client…` groups only: `[mysqld]` and `[mysqldump]` start with
  `mysql` but are other programs, `[client-server]` and `[client-mariadb]` shared MariaDB
  groups.
- Only the last of a repeated option is passed on. The parser ignores an empty value, so
  otherwise an empty `password=` in a suffixed group could not take back the password it
  inherits.
- **Every option is passed as `--key=value`, the `=` kept when the value is empty.** A bare
  `--key` would take the argument after it as its value.
- **A value is unquoted, then unescaped as the client does** — `\\`, `\"`, `\'`, `\n`, `\r`,
  `\t`, `\b`, `\s` — quoted or not. A `loose-` prefix on an option is dropped: it only tells
  a client that does not know the option to skip it.
- `!include` and `!includedir` lines are dropped, included files are not followed, a
  `# comment` after a value is not stripped, and the obfuscated `~/.mylogin.cnf` is not read.

### ClickHouse

- The file is the client's own config, XML or YAML by its content;
  `ClickHouseClientPath` picks the first that exists in the client's search order.
- A connection without `hostname` uses its `name` as the host, as the client does.
- `secure` and `accept-invalid-certificate` become their flags only when true; the parser
  refuses the second, so such a connection is not importable.
- Every other key (`history_file`, `prompt`, …) is a warning.

### MSSQL

- A context is a connection: its endpoint gives the server, its user the login. A context
  whose endpoint is not in the file is an `Entry.Err`.
- **The stored password is base64 of the plain text when `password-encryption` is `none` or
  absent** — sqlcmd encodes it even unencrypted. With any other method (`dpapi` on Windows)
  the material cannot be read: the entry is imported without a password and says so.
- `sqlconfig` has no database, hence the `master` default above.

### SSH

- Every alias on a `Host` line that holds none of `*`, `?`, `!` is a connection, once.
- **A pattern block is never a connection but its options apply.** The options of every
  block whose patterns match the alias are passed in file order as `-o Key=Value`, and the
  parser keeps the first value of each, as ssh does. Options before the first `Host` are the
  block `*`. A negated pattern that matches excludes its block. `LocalForward` and
  `IdentityFile` are the options ssh adds up instead: the parser keeps the first and warns
  about the rest.
- The alias is passed as a last `-o HostName=<alias>`, so it is the host unless a `HostName`
  came first. It is not passed as the destination, which the parser would take apart as
  `user@host`.
- `%h` in a `HostName` is the alias and `%%` a percent sign, as ssh expands them. No other
  token is expanded.
- `Match` blocks are dropped by `ReadSSHConfig`: their conditions cannot be evaluated
  without running ssh. `Include` is an option like any other and ends as a warning.

## Reading the files

- Every reader returns `Section`s — a name and its `Param`s in file order —
  whatever the file's syntax.
- `ReadClickHouseConnections` reads `connections_credentials` from XML (under any root
  element) or YAML, told apart by content — XML starts with `<` — since `--path` may name a
  copy under any name. In YAML one connection is a map and several are a list; a null value
  is a key that is not there. The `name` key becomes the section's name.
- `ClickHouseClientPath` follows the client's search order — `./clickhouse-client.*`,
  `$XDG_CONFIG_HOME/clickhouse/config.*` (`~/.config` when unset, not the platform directory
  `xdg` would give), `~/.clickhouse-client/config.*`, `/etc/clickhouse-client/config.*` — over
  `.xml`, `.yaml`, `.yml`, and names the home XML file when none exists.
  `ClickHouseUserPath` searches the two user locations only — the file export writes to.
- `ReadSQLConfig` joins each context of a sqlcmd `sqlconfig` with the endpoint and the user it
  names, under the `SQLConfig*` keys. A name that resolves to nothing leaves those keys out,
  as does an endpoint without a port. The password is passed on as stored.
- `ReadSSHConfig` returns one section per `Host` line, named by the whole pattern list, after
  a first section `*` holding the options given before any `Host`. The options of a `Match`
  block are dropped up to the next `Host`. A line is cut by `config.SplitSSHOption`, which
  `internal/parser/params` also reads an `-o` value with.
- `ReadINI` drops comment lines (`#`, `;`), blank lines and anything before the first
  section. A line without `=` is a key with an empty value — an option file's flag.
- `PGServicePath` and `PGPassPath` honour `PGSERVICEFILE` and `PGPASSFILE` before falling
  back to the home directory, as libpq does.
- `ReadPGPass` treats a missing file as an empty one. A row has five fields, everything
  after the fourth colon is the password, `\:` and `\\` are the escapes, and
  `PGPass.Password` returns the first row whose host, port, database and user each match or
  are `*`.

## Adding a format

1. `internal/importer/<type>.go`: the default path, the reader of the file into `Section`s
   (`ReadINI` if it is sectioned key/value).
2. In the same file, a `read<Type>(path) ([]Entry, error)`.
3. A row in `sources`.
4. The type's row in the table of `docs/user/add-connection.md`.
5. The way back, if the format can be written: `internal/exporter/AGENTS.md`.
