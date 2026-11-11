# Workspaces

A workspace is a separate set of connections, transports and secrets. Use one per context you
want to keep apart: work and personal, one per client, or production and staging. Only one
workspace is active at a time, and everything you see and do in conm belongs to that workspace.
Its name is shown in the top left corner of every screen (`conm:default`).

conm starts with a single workspace named `default`.

## What belongs to a workspace

| Separate in every workspace | Shared by all workspaces |
|---|---|
| Connections of every database type | Which databases are enabled and [which client each one opens](additional-screens.md#switching-clients) |
| [Transports](ssh-transport.md) (SSH tunnels) | [Keybindings](additional-screens.md#overriding-keybindings) |
| [Secrets](secrets.md): the records and the passwords in the system keyring | |

Each workspace has its own list of secrets, and only those can be picked for its connections.
They are stored in the keyring under the workspace's own ID, so no other workspace can read them,
even one with a secret of the same location. See [Secrets and workspaces](secrets.md#secrets-and-workspaces).

Connection names only have to be unique within their workspace: `orders` in `work` and `orders`
in `personal` are two different connections.

## From the workspaces screen

Open the workspaces screen from the connections table, or straight from the shell:

```sh
conm workspace
```

On the screen you can:
- **add** a workspace: it starts empty;
- **rename** one: its connections, transports and secrets move with it and keep working;
- **delete** one: its connections, transports, secret records and their passwords in the system
  keyring are removed. You are asked to confirm, and the workspace you are in cannot be deleted;
- **open** one: conm switches to it for this run. A workspace with no connections yet asks for
  the type of the first one.

![Adding, renaming and opening a workspace](../demo/workspace-screen.gif)

A workspace name has to be unique and a valid folder name on Linux, macOS and Windows: no
`<>:"/\|?*`, no trailing dot or space, and not a reserved Windows name such as `CON` or `NUL`.

## From the shell

```sh
conm workspace add staging          # create a workspace
conm workspace use staging          # open it from now on, every time conm starts
conm workspace use default          # go back to the default one
conm -w staging postgres list       # use another workspace for one command only
```

`-w` (or `--workspace`) works with every command, for example `conm -w staging postgres
orders` or `conm -w staging postgres add`.

![Creating and switching workspaces from the shell](../demo/workspace-shell.gif)

If the workspace does not exist, conm says so and exits:

```
Error: no workspace named "nope" · 'conm workspace add nope' creates it
```

---

[← Connect through an SSH tunnel](ssh-transport.md) · [Contents](README.md) · [Additional screens →](additional-screens.md)
