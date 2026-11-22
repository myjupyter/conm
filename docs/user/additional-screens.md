# Additional screens

## Settings

The settings screen is opened from the connections table and has three pages: which client each
database opens, which keys do what, and how conm looks. A change is saved as soon as you make it. Settings are
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

### Appearance

The appearance page changes the look of every screen at once, as soon as you pick a value.
It has four sections:
- **colour**: the **accent** — the colour of the highlighted row and the active tab, `green`,
  `lime`, `teal`, `blue`, `purple`, `pink`, `red`, `orange` or `amber`, your own `#rrggbb`, or a
  gradient of them like `teal>pink` (below). Status colours stay as they are: a success is always
  green, a warning amber, a failure red;
- **text**: how text is drawn in your terminal's font — `default` (titles, tabs, headings,
  the selection, keys and values in bold), `plain` (nothing bold), `emphasis` (underlined
  titles and headings, italic notes) or `soft` (bold titles and headings only, faint italic
  notes). conm cannot change the font itself, its size or its family: that is your terminal's
  setting;
- **border**: the frame around each screen — `single` `┌──┐`, `rounded` `╭──╮`, `double`
  `╔══╗`, `heavy` `┏━━┓`, `dashed` `┌╌╌┐`, `dotted` `┌┄┄┐`, `mixed` `╒══╕`, `block` `▛▀▀▜`,
  `solid` `████`, `shade` `░░░░`, `ascii` `+--+` or `hidden`, which draws no frame at all;
- **glyphs**: the symbols for status, cursor and markers — `unicode`, `heavy` (`➜ ✔ ✘`),
  `geometric` (`▶ ◆ ■`), `minimal` (`▸ • ×`), `emoji` (`👉 ✅ ❌`), `nerd` (icons from a
  [Nerd Font](https://www.nerdfonts.com), which your terminal must use) or `ascii`, for a
  terminal or font that draws symbols like `❯ ✓ ◐` as boxes. `ascii` draws the frame with
  `+ - |` too, so the border section is locked while it is on.

Emoji take two columns, and conm makes room for them. Some emoji, mostly those written
with an invisible variation selector like `✔️` or `⚠️`, are one column in one terminal and two in
another; conm cannot tell, and such a glyph can push the right edge of a line out of place.
The `emoji` set only uses emoji every terminal draws two columns wide.

#### Colours

The text, border and glyph sections each have a **palette** row under their style. Every text
role, border piece and glyph then has a heading of its own with two rows under it — its style,
character or glyph, and its **colour**, with a swatch showing the colour. A heading says
"changed" while that element differs from the preset.

The palettes:
- **text**: `default` (every role keeps its usual colour), `accent` (titles, headings and keys
  in the accent) or `muted` (softer headings, keys and values, dim notes);
- **border**: `default` (the usual grey), `accent`, `soft`, `bright` or `duotone` (the accent
  across the top, the bottom and the rules, grey at the sides);
- **glyphs**: `default` (every glyph keeps the colour the screen gives it), `accent`, `bright`
  or `muted` — for the caret, the dots, the arrows and the cursor.

A colour row sets that one element: `auto` (what the palette and the screen decide), a named
colour — `accent`, `fg`, `soft`, `dim`, `border`, `ok`, `warn`, `error` or any accent preset
like `teal` — or your own `#rrggbb`. A glyph palette never recolours the status icons — a
success stays green, a warning amber, a failure red — but a colour you set on one of them
yourself does. The highlighted row keeps its own colours so it stays readable; only the caret,
which only ever sits on that row, takes its colour there. The `…`, `—`, the password mask and
the scroll arrows sit inside text and keep the text's colour, so they have no colour row.

#### Gradients

A border or text colour can be a gradient: colours joined by `>`, like `teal>pink` or
`#ff8800>accent>#ff0088`. The top, bottom and the rules between sections run from left to right
across the screen, the left and right sides from top to bottom, and a text role across its
characters. The `sunset` and `ocean` palettes, for the border and for text, are made of
gradients. Glyph colours are always one colour.

A tab or heading background can be a gradient as well: a heading's band runs across the line,
a tab across the tab.

The accent can be a gradient too — step past the single colours to a few ready ones like
`teal>pink`, or type your own: the highlighted row then runs from the first colour to the last
across its width, and the active tab across the tab. Where the accent colours a line or a
letter — the caret, a key, the workspace name — it uses the first colour. The text on the
highlighted row stays dark, so pick light colours for an accent gradient.

Gradients need a terminal with 24-bit colour, which most current terminals have. conm checks
what your terminal supports when it starts; without 24-bit colour every gradient is drawn in its
first colour, and the settings screen says so when you pick one.

#### Changing one element

Under the text style, the border style and the glyph set, every element has its own row (and,
as above, a colour row under it):
- **text**: title, tab, heading, subheading, selection, key, value, note — each `plain` or a mix of
  `bold`, `italic`, `underline` and `faint` joined by `+`, like `bold+italic` — and, with a
  colour only, **body** (ordinary text: table cells, values, messages) and **input** (what you
  type into a field or the search). The **tab**, the **heading** and the **subheading** also
  have a **background**: the colour behind the active tab (by default the accent, or the type's
  colour on the connections table), behind a section heading's band, and behind a subheading —
  the one-line headings of this page and the group titles of the key help, which have no band
  by default;
- **border**: each corner, the joins at either end of a rule, and the top, bottom, left and
  right lines and the rule between sections — a border can be different on every side;
- **glyphs**: each status icon, the cursor, the markers, the `<` `>` arrows of a selector.

Step through the choices for an element, or edit the row and type your own: a style, or any
single character — one column wide for a border, one or two for a glyph. The preset stays as it is:
your value replaces just that element, the row shows the preset's value beside it, and you can
reset the row to the preset again. The section heading counts how many elements you changed,
and they are kept when you pick another preset.

```toml
[conm.appearance]
border = 'rounded'
accent = '#ff8800'
glyphs = 'emoji'
text = 'default'
border_color = 'duotone'
glyph_color = 'accent'
text_color = 'default'

[conm.appearance.border_runes]
top = '═'
bottom = '═'

[conm.appearance.glyph_runes]
caret = '»'

[conm.appearance.text_style]
note = 'italic'

[conm.appearance.border_colors]
left = '#ff0000'
top = 'teal>pink'

[conm.appearance.glyph_colors]
caret = 'pink'

[conm.appearance.text_colors]
note = 'dim'
input = '#e8e2d4'

[conm.appearance.text_backgrounds]
tab = 'pink'
heading = '#202020>#3a2a40'
```

A name conm does not know falls back to the first one in its list; a value that does not fit
its element is ignored.

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

[← Workspaces](workspaces.md) · [Contents](README.md) · [Commands →](commands.md)
