# Connect and run queries

conm does not talk to the database itself. It opens the client picked for that database in the
settings (for Postgres: `psql`, `pgcli` or `usql`) with the host, user, database and password
of the connection already passed in. There are three ways to use it:
1. Open a session from the connections table
2. Open a session from the shell, by the connection's name
3. Run a query or a script from the shell, without an interactive session

Every database works the same way under its own name: `conm mysql orders-local`,
`conm redis exec cache -- "ping"`, and so on.

## From the connections table

Open the table, select the connection and open a session. The client takes over the terminal;
when you quit it, you are back in the table.

```sh
conm postgres
```

![Opening a psql session from the connections table](../demo/connect-table.gif)

Inside the session you work with the client as usual, including running script files with its
own commands (`\i script.sql` in `psql`).

## From the shell

Give the name of the connection. The client opens straight away, and quitting it returns you
to the shell.

```sh
conm postgres orders-local
```

![Opening a psql session by the connection's name](../demo/connect-shell.gif)

## Run a query or a script from the shell

`exec` runs one query and exits. The result goes to stdout, errors go to stderr, and conm exits
with the client's exit code, so it works in scripts and pipes. Quote the query, or the shell
expands `*` and ends the command at `;`.

```sh
conm postgres exec orders-local -- "select current_user, current_database();"
```

To run a whole script, feed it to the session on stdin:

```sh
conm postgres orders-local < script.sql
```

![Running a query and a script from the shell](../demo/exec-shell.gif)

Everything after `--` is the query. Client flags cannot be passed there.

If no connection has that name, conm says so and exits:

```
Error: no postgres connection named "nope"
```
