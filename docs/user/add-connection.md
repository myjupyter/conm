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

---

[← First run](first-run.md) · [Contents](README.md) · [Edit or remove a connection →](edit-connection.md)
