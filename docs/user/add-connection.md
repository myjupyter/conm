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

---

[← First run](first-run.md) · [Contents](README.md) · [Edit or remove a connection →](edit-connection.md)
