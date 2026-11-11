# Connect and run queries

conm does not talk to the database itself. It opens the client picked for that database in the
[settings](additional-screens.md#switching-clients) (for Postgres: `psql`, `pgcli` or `usql`) with the host, user, database and password
of the connection already passed in. A connection behind a bastion works the same way once it
has a [transport](ssh-transport.md).

## From the connections table

Open the table, select the connection and open a session. The client takes over the terminal;
when you quit it, you are back in the table.

```sh
conm postgres
```

![Opening a psql session from the connections table](../demo/connect-table.gif)

## From the shell

Give the name of the connection. The client opens straight away, and quitting it returns you
to the shell.

```sh
conm postgres orders-local
```

![Opening a psql session by the connection's name](../demo/connect-shell.gif)

## Run a query or a script from the shell

`exec` runs one query and exits. The result goes to stdout, errors go to stderr, and conm exits
with the client's exit code, so it works in scripts and pipes. Quote an SQL query, or the shell
expands `*` and ends the command at `;`. Redis is the exception: it takes the command as
separate words, `conm redis exec cache -- get foo`, not `"get foo"`.

```sh
conm postgres exec orders-local -- "select current_user, current_database();"
```

To run a whole script, feed it to `psql` on stdin (`pgcli` runs it but then stays open at its
prompt):

```sh
conm postgres orders-local < script.sql
```

![Running a query and a script from the shell](../demo/exec-shell.gif)

Everything after `--` is the query, not client flags; only the Redis clients read flags there too.

---

[← Edit or remove a connection](edit-connection.md) · [Contents](README.md) · [Passwords and secrets →](secrets.md)
