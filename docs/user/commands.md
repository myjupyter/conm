# Commands

Everything conm does from its screens can also be done from the shell. Run `conm --help`, or
`--help` after any command, for the full list of flags.

## Connections

Each connection type has its own command, and most have short aliases (`conm pg list` is
`conm postgres list`):

| Command | Aliases |
|---|---|
| `postgres` | `pg` |
| `mysql` | `my` |
| `mssql` | `ms`, `sqlserver` |
| `clickhouse` | `ch` |
| `redis` | `rs` |
| `mongodb` | `mongo`, `mg` |
| `ssh` | - |

The examples use Postgres.

| Command | What it does |
|---|---|
| `conm` | Open the connections table |
| `conm postgres` | Open the table on the Postgres connections |
| `conm postgres <name>` | Open the client on a connection ([more](connect-and-execute.md#from-the-shell)) |
| `conm postgres list` | Print the Postgres connections |
| `conm postgres add [command line]` | Add a connection, optionally prefilled from a command line or URL ([more](add-connection.md#prefill-from-a-command-line)) |
| `conm postgres import [name...]` | Import connections from the client's own config file ([more](add-connection.md#import-from-a-clients-config-file)) |
| `conm postgres export [name...]` | Export connections to the client's own config file ([more](add-connection.md#export-to-a-clients-config-file)) |
| `conm postgres edit <name>` | Edit a connection ([more](edit-connection.md#from-the-shell)) |
| `conm postgres remove <name>` | Remove a connection ([more](edit-connection.md#remove-a-connection)) |
| `conm postgres ping <name>` | Check that a connection answers |
| `conm postgres exec <name> -- <query>` | Run one query and exit ([more](connect-and-execute.md#run-a-query-or-a-script-from-the-shell)) |

## Workspaces

| Command | What it does |
|---|---|
| `conm workspace` | Open the workspaces screen |
| `conm workspace add [name]` | Create a workspace |
| `conm workspace use <name>` | Open this workspace from now on |
| `conm -w <name> …` | Run any command in another workspace, just this once |

See [Workspaces](workspaces.md#from-the-shell).

## Doctor

| Command | What it does |
|---|---|
| `conm doctor` | Check that conm is ready to use: `conm.toml`, the clients, the saved connections |

```
Doctor summary (conm v0.1.0)
[✓] Conm internal configs

Clients:
[✓] Postgres (psql 17.11)
[!] Mongodb (no client installed)
  - no client installed
...

Connections:
[✓] Postgres
[✗] Mysql
  - "orders": port must be between 1 and 65535
...

! Doctor found issues in 2 categories.
```

A missing client is a warning (`!`); a missing [chosen client](settings.md#switching-clients) or
an invalid connection is an error (`✗`). It never changes anything and exits with status 1 on an
error, so it works in scripts.

---

[← Additional screens](additional-screens.md) · [Contents](README.md)
