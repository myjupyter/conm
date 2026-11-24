# Settings

Open it from the connections table. Tabs: **client**, **keybind** and **appearance**, which has
its own tabs — accent, text, glyphs, borders and table. Changes are saved at once and shared by
all [workspaces](workspaces.md).

## Switching clients

Pick which installed client each database opens. Until you do, it opens the first installed
one — for Postgres `psql`, then `pgcli`, then `usql`.

![Switching Postgres from psql to pgcli](../demo/settings-client.gif)

## Overriding keybindings

Add, change, remove or reset the keys of any action; a key can also be a sequence of two. A key
cannot do two things on one screen, and navigation, confirm and cancel keep their default keys.

![Adding o as a second key to open a session](../demo/settings-keybind.gif)

## Appearance

Every change applies to all screens at once.

![Changing the accent, the glyphs and the border](../demo/settings-appearance.gif)

| Tab | Presets |
|---|---|
| accent | `green`, `lime`, `teal`, `blue`, `purple`, `pink`, `red`, `orange`, `amber`, your own `#rrggbb`, or a gradient |
| text | `default`, `plain`, `emphasis`, `soft` |
| borders | `single`, `rounded`, `double`, `heavy`, `dashed`, `dotted`, `mixed`, `block`, `solid`, `shade`, `ascii`, `hidden` |
| glyphs | `unicode`, `heavy`, `geometric`, `minimal`, `emoji`, `nerd` (needs a [Nerd Font](https://www.nerdfonts.com)), `ascii` (draws the frame too) |

The font itself is your terminal's setting, not conm's.

### Colours

Text, borders and glyphs each have a **palette** and a colour per element: `auto`, a name —
`accent`, `fg`, `soft`, `dim`, `ok`, `warn`, `error`, `teal`, … — or `#rrggbb`. Status icons stay
green, amber and red unless you colour them yourself.

### Gradients

Join colours with `>`, like `teal>pink`, for borders, text, backgrounds and the accent. They
need a terminal with 24-bit colour; otherwise the first colour is used.

### Changing one element

Every text role, border piece and glyph can be changed on its own — its style or character and
its colour; the tab, heading and subheading also have a background. Reset an element to go back
to the preset.

## Table

For each connection type: its **tab colour** — its own, another type's, `auto` (the tab background,
or the accent) or any colour — and its **columns**: tick to show, move to reorder. NAME is always
shown; TAGS is hidden until you tick it.

The tab colour paints the type's tab, active or not, and nothing else.

---

[← Workspaces](workspaces.md) · [Contents](README.md) · [Additional screens →](additional-screens.md)
