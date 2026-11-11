# Edit a connection

There are two ways to edit a connection:
1. From the connections table
2. From the shell, by its name

Either way you get the same form as when adding it, filled in with the saved values. Change
what you need, test it, and save it.

Every database works the same way under its own name: `conm mysql edit`, `conm redis edit`,
and so on.

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
