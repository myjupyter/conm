# Edit or remove a connection

Edit a connection from the connections table, or from the shell by its name. Either way you get
the same form as when [adding it](add-connection.md), filled in with the saved values.

## From the connections table

Open the table, select the connection and open its edit form. After saving you are back in the
table.

```sh
conm postgres
```

![Editing a Postgres connection from the table](../demo/edit-table.gif)

If a field is invalid, or the change makes it a duplicate of another connection, the form stays
open with your values and shows the reason in the status line. Cancelling leaves the
connection as it was.

## From the shell

Give the name of the connection. The edit form opens straight away. After saving, conm opens
the Postgres connections table.

```sh
conm postgres edit orders-local
```

![Editing a Postgres connection by its name](../demo/edit-shell.gif)

If no connection has that name, conm says so and exits:

```
Error: no postgres connection named "nope"
```

## Remove a connection

In the connections table, select the connection and delete it; conm asks you to confirm. From
the shell, `remove` deletes it straight away, without asking:

```sh
conm postgres remove orders-local
```

```
removed postgres connection "orders-local"
```

![Removing a connection from the table and from the shell](../demo/remove.gif)

The [secret](secrets.md) the connection used stays on the secrets screen. An SSH connection that
a [transport](ssh-transport.md) goes through cannot be removed until no transport uses it:

```
Error: remove: connection "ssh:bastion" is still in use by 1 transport(s): transport/prod
```

---

[← Add a connection](add-connection.md) · [Contents](README.md) · [Connect and run queries →](connect-and-execute.md)
