# Additional screens

## Settings

The settings screen is opened from the connections table and has two pages: which client each
database opens, and which keys do what. A change is saved as soon as you make it. Settings are
shared by all [workspaces](workspaces.md), and live in `conm.toml`.

### Switching clients

Every database lists the clients conm supports for it, with the installed ones highlighted, and
the client it opens now, with its path and version. Switch to another installed client and
every [session, `exec`](connect-and-execute.md) and info screen of that database uses it from then on. Clients that are
not installed are skipped; if no other one is installed, conm tells you which to install.

Until you pick one, a database opens the first installed client in its list, for Postgres
`psql`, then `pgcli`, then `usql`.

![Switching Postgres from psql to pgcli](../demo/settings-client.gif)

The choice is stored per database:

```toml
[[conm.connection]]
type = 'postgres'
cli = 'pgcli'
```

### Overriding keybindings

The keybinding page lists every action, grouped by where it applies. An action
can have several keys. For each action you can add a key, change or remove one, or reset the
action to its default keys (you are asked to confirm).

A key can also be a sequence of two keys, like the default `s s` for settings: press the second
key right after the first, or confirm to keep the first key alone.

![Adding o as a second key to open a session](../demo/settings-keybind.gif)

A few rules keep every screen usable:
- a key cannot do two things on the same screen;
- a sequence cannot start with a key another action on that screen uses alone;
- navigation, confirm and cancel keep their default keys; you can add more but not remove them;
- `ctrl+c` and `backspace` are fixed and not listed.

The keys are stored in `conm.toml` and can also be edited there:

```toml
[conm.keybinding.connection]
connect = ['enter', 'o']
```

## Info screen

The info screen shows everything about the connection under the cursor without opening the edit
form. Open it from the connections table.

It has up to three pages:
- **general**: name, description and tags, and the client the connection opens, with its
  version and path;
- **endpoint**: host, port, user, database and the other connection fields, the transport, and
  the connection string, with the password shown as `•••`;
- **links**: the connection's links, each one opens in the browser. A connection without links
  has no links page.

Any value can be copied to the clipboard, and the [edit form](edit-connection.md) is one step away when you do want
to change something.

![The general, endpoint and links pages of a connection](../demo/info.gif)

---

[← Workspaces](workspaces.md) · [Contents](README.md)
