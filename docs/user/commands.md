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

`conm doctor` checks that conm is ready to use and prints one line per check, followed by the
problems it found:

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

It checks three things:

- **Conm internal configs**: `conm.toml` can be read and its keybindings are valid.
- **Clients**: for every database type, the client conm would open is installed, with its
  version. A type with no client at all is a warning (`!`), since you may not use that database.
  A [chosen client](additional-screens.md#switching-clients) that is missing or not valid is an
  error (`✗`).
- **Connections**: every saved connection and transport in the current workspace can be read and
  is valid.

It only reports and never changes anything. It exits with status 1 when any check fails with
`✗`, so it can be used in scripts.

---

[← Additional screens](additional-screens.md) · [Contents](README.md)
