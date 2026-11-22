# Add a connection

Fill in an empty form, or prefill it from a command line or URL you already use. Either way you
check the form, test it, and save it.

## Interactive mode

```sh
conm postgres add
```

![Adding a Postgres connection through the form](../demo/add-interactive.gif)

If a field is invalid, or the connection already exists, the form stays open with your values
and shows the reason in the status line.

## Prefill from a command line

Paste a command line or URL you already use, in quotes. Put `--` before one that starts with
a flag.

```sh
conm postgres add "psql -h localhost -p 5432 -U me mydb"
conm postgres add -- "-h localhost -p 5432 -U me mydb"
conm postgres add "pgcli --host localhost --dbname mydb --user me"
conm postgres add -- "--host localhost --dbname mydb --user me"
conm postgres add "postgres://me@localhost:5432/mydb?sslmode=require"
```

![Adding a Postgres connection prefilled from a psql command line](../demo/add-parse.gif)

conm reads the host, port, user, database and the other options it recognises, and opens the
same form filled in. Anything it cannot use is listed as a warning in the form. Fill in what
is missing (here the password), then save. The name is prefilled with a random one such as
`calm-otter`; change it to something you will recognise.

`conm <database> add --help` prints the examples for each database.

## Import from a client's config file

If a client already keeps your connections in its own config file, import them instead of
typing them again. The file is only read, never changed.

```sh
conm postgres import                  # every connection in the file
conm postgres import orders billing   # only these
conm postgres import --show           # print what would happen, change nothing
```

| Type | File | Passwords |
|---|---|---|
| `postgres` | `~/.pg_service.conf`: one connection per service | from the service, else from its `passfile` or `~/.pgpass` |
| `mysql` | `~/.my.cnf`: each `[client_<name>]` as `<name>`, and `[client]` as `client` when it names a host or is the only group | from the group |
| `clickhouse` | the client's `config.xml` or `config.yaml`: one connection per `connections_credentials` entry | from the entry |
| `mssql` | `~/.sqlcmd/sqlconfig`: one connection per context | from the context's user |
| `ssh` | `~/.ssh/config`: one connection per `Host` that is not a pattern | none: a key file stays a path |

Redis and MongoDB clients keep no connections in a file, so they have no `import`.

Each connection is named after its entry and printed with what happened to it:

| Mark | Meaning |
|---|---|
| `✓` | imported |
| `○` | can be imported, but was not |
| `!` | already in conm, so left alone |
| `✗` | cannot be imported, for the reason printed below it |

Under a connection conm lists what it left out of the entry, such as an option it has no
field for or a password it could not read. An entry that repeats a connection of the same
file is not imported a second time. If one connection cannot be imported, nothing is. Add `--skip-errors` to import the rest.

![Importing Postgres services from the shell](../demo/import-shell.gif)

With `-i`, each connection that was read but cannot be saved opens in the add form with the
reason shown: fix it and save, or cancel to skip it. An entry conm could not read stays `✗`.

![Fixing a service that cannot be imported in the add form](../demo/import-interactive.gif)

A password found in the file goes to the system [keyring](secrets.md), not into conm's files,
and a connection whose password cannot be stored there is not imported.
`conm <database> import --help` lists the other options, such as another file to read.

## Export to a client's config file

The way back: write conm's connections into the client's own file, for the client itself or
for a teammate without conm.

```sh
conm postgres export                  # every Postgres connection
conm postgres export orders billing   # only these
conm postgres export --show           # print what would happen, change nothing
conm postgres export --path ./pg_service.conf --with-passwords
```

The entries are added to the file, which is created if it is not there. Nothing already in
it is replaced: a connection whose name the file already has is `✗`, and then nothing is
exported unless you add `--skip-errors`. The marks are the ones import prints.

| Type | Written as | Left out |
|---|---|---|
| `postgres` | a service in `~/.pg_service.conf`; use it with `psql service=<name>` | the transport |
| `mysql` | `[client_<name>]` and `[mysql_<name>]` in `~/.my.cnf`; use it with `mysql --defaults-group-suffix=_<name>` | the transport |
| `clickhouse` | a `connections_credentials` entry in your own `config.xml` or `config.yaml` — `~/.config/clickhouse/` or `~/.clickhouse-client/`, never one in the current directory or `/etc`; use it with `clickhouse client --connection <name>` | the transport |
| `mssql` | a context in `~/.sqlcmd/sqlconfig`, the current one if the file had none; select another with `sqlcmd config use-context <name>` | the database, the encryption settings, a password without a user, the transport |
| `ssh` | a `Host` block at the top of `~/.ssh/config`, so neither a `Host *` nor an option above it can override it (what was global stays global, under a `Host *`); use it with `ssh <name>` | a password, and a key that is not a file |

Passwords stay in conm. `--with-passwords` writes them into the file **in plain text**
(base64 in `sqlconfig`, which is not encryption); an SSH config cannot hold one. The file is
left readable by you alone.

A name can be exported when it is made of letters, digits and `_ . @ -`. A password with a
line break cannot go into `~/.pg_service.conf`; that connection is `✗`. Comments in the
file are kept; a YAML file (`sqlconfig`, a ClickHouse `config.yaml`) is re-indented, and one
holding more than one `---` document is not changed.

---

[← First run](first-run.md) · [Contents](README.md) · [Edit or remove a connection →](edit-connection.md)
