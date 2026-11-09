# Add a connection

There are two ways to add a connection:
1. Fill in an empty form in the interactive mode
2. Parse a command line or URL you already use into a prefilled form

Either way you check the form, test it, and save it.

Every database works the same way under its own name: `conm mysql add`, `conm redis add`, and so on.

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
is missing (here the password and a name), then save.

You can also pass the command line after `--` instead of quoting it:
`conm postgres add -- psql -h localhost -U me mydb` or `conm postgres add -- -h localhost -U me mydb`.
`conm <database> add --help` prints the examples for each database.