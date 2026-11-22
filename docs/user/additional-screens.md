# Additional screens

## Settings

Open it from the connections table. Tabs: **client**, **keybind** and **appearance**, which has
its own tabs — accent, text, glyphs, borders and table. Changes are saved at once to
`conm.toml` and shared by all [workspaces](workspaces.md).

### Switching clients

Pick which installed client each database opens. Until you do, it opens the first installed
one — for Postgres `psql`, then `pgcli`, then `usql`.

![Switching Postgres from psql to pgcli](../demo/settings-client.gif)

```toml
[[conm.connection]]
type = 'postgres'
cli = 'pgcli'
```

### Overriding keybindings

Add, change, remove or reset the keys of any action; a key can also be a sequence of two. A key
cannot do two things on one screen, and navigation, confirm and cancel keep their default keys.

![Adding o as a second key to open a session](../demo/settings-keybind.gif)

```toml
[conm.keybinding.connection]
connect = ['enter', 'o']
```

### Appearance

Every change applies to all screens at once. For a guided tour, see
[Make conm yours](customisation.md).

| Tab | Presets |
|---|---|
| accent | `green`, `lime`, `teal`, `blue`, `purple`, `pink`, `red`, `orange`, `amber`, your own `#rrggbb`, or a gradient |
| text | `default`, `plain`, `emphasis`, `soft` |
| borders | `single`, `rounded`, `double`, `heavy`, `dashed`, `dotted`, `mixed`, `block`, `solid`, `shade`, `ascii`, `hidden` |
| glyphs | `unicode`, `heavy`, `geometric`, `minimal`, `emoji`, `nerd` (needs a [Nerd Font](https://www.nerdfonts.com)), `ascii` (draws the frame too) |

The font itself is your terminal's setting, not conm's.

#### Colours

Text, borders and glyphs each have a **palette** and a colour per element: `auto`, a name —
`accent`, `fg`, `soft`, `dim`, `ok`, `warn`, `error`, `teal`, … — or `#rrggbb`. Status icons stay
green, amber and red unless you colour them yourself.

#### Gradients

Join colours with `>`, like `teal>pink`, for borders, text, backgrounds and the accent. They
need a terminal with 24-bit colour; otherwise the first colour is used.

#### Changing one element

Every text role, border piece and glyph can be changed on its own — its style or character and
its colour; the tab, heading and subheading also have a background. Reset an element to go back
to the preset.

```toml
[conm.appearance]
border = 'rounded'
accent = 'blue>pink'
glyphs = 'geometric'
border_color = 'ocean'

[conm.appearance.glyph_runes]
caret = '»'

[conm.appearance.text_colors]
title = 'pink'

[conm.appearance.text_backgrounds]
heading = 'teal>blue'
```

### Table

For each connection type: its **colour** — its own, `tab`, or any colour — and its **columns**:
tick to show, move to reorder. NAME is always shown; TAGS is hidden until you tick it.

```toml
[[conm.connection]]
type = 'postgres'
columns = ['name', 'host', 'username', 'tags', 'ping']
color = 'tab'
```

## Info screen

Everything about the connection under the cursor, without opening the edit form: **general**,
**endpoint** (the password shown as `•••`) and **links**. Copy any value, or go on to the
[edit form](edit-connection.md).

![The general, endpoint and links pages of a connection](../demo/info.gif)

---

[← Workspaces](workspaces.md) · [Contents](README.md) · [Make conm yours →](customisation.md)
