# `internal/ui` — the Bubble Tea TUI

## Purpose

Every screen conm draws, and the blocking entry points `cmd` calls. Screens are composed
from shared primitives, reach entities only through `ui/view`, and take their forms from
`ui/spec`.

May import (internal): `ui/view`, `ui/spec`, `repository`, `importer`, `config`, `network`, `cli`,
`secret`, `transport`, `utils/keys`, `exporter`.

## Files

A screen is split `<name>.go` (model, `Update`) / `<name>_view.go` (render) /
`<name>_commands.go` (`tea.Cmd`s, I/O, the blocking `run<Screen>` entry points).

| Files | Holds |
|---|---|
| `ui.go` | the entry points `Run`, `RunOn`, `Connect`, `Ping`, `Connections`, `RemoveConnection`; the loop that rebuilds a `repository.Session` per pass |
| `screen_stack.go` | `WithScreens`, `runScreen` — one `tea.Program` hosting a stack of screens |
| `screen_size.go` | `screenSize` — widths and visible rows from the terminal size |
| `frame.go`, `theme.go` | the line builders; the colours and glyphs |
| `keybinding.go`, `keyhint_table.go` | `keyMap`; the footer |
| `conn_error.go` | `connError`, `hintFor` — a `*network.OpError` as display strings |
| `clipboard_commands.go` | paste and yank over `cli.ReadClipboard`/`cli.WriteClipboard` |
| `table*.go` | the connections table — `Model`, the main screen |
| `info*.go` | the read-only info sheet |
| `form*.go` | `formModel`, the connection add/edit form; `form_array.go` its chip fields; `form_loop.go` the open → save → retry loop of every form |
| `spec_form*.go` | `specFormModel[T]`, the generic form over any `spec.FormSpec[T]` |
| `secret_table*.go`, `secret_form_commands.go` | the secrets screen and its form |
| `transport_table*.go`, `transport_form*.go` | the transports screen and its form |
| `workspace_table*.go`, `workspace_form_commands.go` | the workspaces screen and its form |
| `settings*.go`, `settings_keybind*.go`, `settings_appearance.go` | settings: the client per type, keybinding overrides, appearance |
| `init_table*.go` | the type picker: before any connection exists, and when adding one of another type |
| `import.go` | `ImportConnections`, `RenderImport` — saves `importer.Entry`s for `conm <type> import` and prints what became of each |
| `export.go` | `ExportConnections` — hands saved connections to `exporter` for `conm <type> export` and reports what became of each |
| `doctor.go` | `Doctor` — renders `repository.CheckClients`/`CheckFiles` for `conm doctor` |

## Main entities

- `Model` — the connections table; holds a `*view.Connections`, the status line and the
  per-`Ref` ping state.
- `formModel` / `specFormModel[T]` — the form drivers; everything entity-specific is a
  `spec.FormSpec` (`internal/ui/spec/AGENTS.md`).
- `formLoop[T]` — runs any form to completion; a screen supplies `open` and `save`.
- `keyMap` — every key the TUI reacts to, named by meaning.
- `screenStack` — the host a screen pushes another screen onto.

## Import — `import.go`

`ImportConnections` is a shell entry point like `RemoveConnection`: it works on the
repository, not through `ui/view`, and opens a screen only in `ImportInteractive`.

- `planImport` sorts the entries with `repository.Connections.Check` before any write:
  saveable together, refused, already present. **A `*repository.DuplicateConnectionError` is
  not a refusal** — the same target under any name is `ImportPresent` in every mode, so
  importing a file twice is not an error. A taken name with another target is a refusal.
- **An entry repeating one of the same file is a twin, not `ImportPresent`**: it stays
  `ImportNew` with a note naming the entry it repeats, and becomes `ImportPresent` only once
  that entry is saved.
- `ImportStrict` writes nothing when one entry is refused and leaves the rest `ImportNew`;
  the other modes save what was admitted with one `AddAll`. `ImportShow` stops after the plan.
- `ImportInteractive` then opens each refused entry that has a connection in the add form,
  seeded from it with the refusal as status, through the same `formLoop` as
  `conm <type> add`, all under one `WithScreens`. An entry the importer could not read has
  nothing to seed a form and stays `ImportFailed`.
- **A password arrives as a `literal:` reference and `keyringBound` moves it to the
  keyring**: the connection is rebuilt through the spec's `SeedFunc`/`BuildFunc` with
  `keyring:<type>/<name>` replacing the literal, beside the `repository.SecretWrite` storing
  the password under a secret named after the connection. A reference that is not a literal
  password — an SSH key file — is left alone.
- **Passwords are stored before connections are saved**, with one `Secrets.PutAll`. An entry
  whose password was refused — its location is used by another connection — is
  `ImportFailed` and not saved; in `ImportStrict` that stops the import before `AddAll`, and
  the secrets already stored stay as unused records the next run rewrites.
- **The form is seeded with the literal password**, never a keyring reference: the location
  does not exist yet and the form refuses a reference to nothing. `keyringBound` runs on
  what the form saved, so the location follows the final name; a connection the form pointed
  at another provider keeps that choice and nothing is stored.
- `ImportResult.Warnings` — what the importer left out of an entry, an unreadable password
  included — is passed on in every mode.
- `RenderImport` prints `[glyph] name dsn` with the reason and warnings underneath, as
  `doctor` prints a check: `kindOK` imported, `kindErr` refused, `kindWarn` already there,
  `gImportNew` in `cPostgres` importable but not imported. The DSN is
  `ConnectionString("")`, so it never carries the password.

## Export — `export.go`

`ExportConnections` is a shell entry point like `ImportConnections` and opens no screen. It
picks the connections (all, or the named ones — an unknown name fails the run), resolves
their passwords through `repository.Session.Password` only under
`ExportOptions.WithPasswords` and for a type `exporter.WritesPasswords` — never SSH, so no
keychain is asked for a secret that would not be written — and lets `exporter.Plan` and
`exporter.Write` do the rest.

- **It reports through `ImportResult` and `RenderImport`**, with the same marks:
  `ImportNew` exportable but not written, `ImportDone` written, `ImportFailed` refused.
  `ImportPresent` does not occur: an entry the file already has is a refusal.
- A refusal writes nothing unless `SkipErrors`; `Show` stops after the plan. A password that
  cannot be resolved is that connection's refusal, never an export without it.

## Shared primitives — reuse, never re-implement

**Compose every UI element from these files; if a piece is missing, add it there. Never
inline a private copy, a raw rune, a raw colour or a raw key string in a screen.**

- **`ui/frame.go` — the frame blanks.** A screen is one box of fixed inner width, built line
  by line from `span` (a styled run of text), `truncPad`, `border` and the builders
  `frameTop`, `frameLine`, `frameRule`, `frameBottom`, `frameBadge`, `frameStatus`,
  `frameKeybar`, `frameSearch`, `framePong`, `frameErrPanel`. Each builder takes the inner
  width `w`, so it serves both the wide table and the narrower forms. Also here: `keybind`,
  `statusKind`/`statusGlyph`, `cursorPos`.
- **`ui/conn_error.go` — the failure panel's data.** `connError` is a `*network.OpError` as
  display strings; `newConnError` is the only caller of `errors.AsType`. `hintFor` is the
  advice keyed by `network.ErrorCode`, owned by the UI because it spells the key to press
  from `keyMap`. Classifying an error belongs to `internal/network/AGENTS.md`.
- **`ui/theme.go` — the look.** The palette (`c*`), the glyphs (`g*`: box-drawing runes,
  status icons, row/field markers, input chrome, ellipsis) and `typeColor`. `frame.go`
  decides *how* a line is built, `theme.go` *what it looks like*: a view never uses a
  literal rune or `lipgloss.Color`, so a restyle never touches drawing code.
  - The glyphs and `cAccent` are variables set by `applyAppearance` from
    `config.AppearanceSettings`. `borderPresets` and `glyphPresets` map a preset name to its
    runes keyed by `config.BorderElements` / `config.GlyphElements` (built positionally by
    `box` and `glyphs`); `borderVars` and `glyphVars` map each element to the `g*` variable
    it sets. `accentColors` holds the presets and `accentColor` also takes a `#rrggbb` value.
  - **`cAccent` is for highlight only** — the selected row, the active tab, the caret, the
    key in a hint. A colour that means a state stays fixed whatever the accent: `cOK` for
    success (the ✓ of a status line and of `doctor`/`import`, PONG, a fast ping, an installed
    client, a filled form field), beside `cAmber` and `cRed`.
  - `borderRunes` and `glyphRunes` are the one place a preset and the user's per-element
    overrides are merged. `borderRunes` ignores both border preset and overrides while the
    glyphs are `ascii` (`borderLocked`).
  - Every rune is one cell wide, as the width arithmetic in `frame.go` assumes —
    `theme_test.go` checks every preset has every element at that width, and that every
    `config` name has an entry.
  - `applyConm` sets both `keyMap` and the appearance; every entry point that draws calls it
    (or `applyAppearance`, for `doctor` and `export`) before its first line.
- **`ui/form_loop.go` — running a form to completion.** `formStatus` (message plus
  `statusKind`), `formResult[T]` (what one run produced, plus its raw values) and
  `formLoop[T]` (open → save → on failure open again). A screen supplies only `open` and
  `save`; it never drives a form itself.
- **`ui/keybinding.go` — every key the TUI reacts to.** A `binding` is a set of keys plus
  their spelling in a hint.
  - `newKeymap` rebuilds `keyMap` from `config.KeybindingSettings` on every pass of the
    connections loop, so a changed key is live on the next screen; invalid settings fall
    back to `config.DefaultKeybindings()`.
  - Shared keys are top-level (`Up`, `Confirm`, `Cancel`, `Paste`, `Copy`, `Search`, `Help`,
    …); the rest are grouped by target: `keyMap.Screen` (opening another screen),
    `keyMap.Connection`, `keyMap.Form`, `keyMap.Secret` / `Transport` / `Workspace` (add,
    edit, delete), `keyMap.Settings`, `keyMap.Prompt`.
  - `Update` matches with `switch { case keyMap.X.matches(key): … }`, never against a raw
    string; a hint takes its key from `keyMap.X.hint` and supplies only the label.
  - Keys treated as one group are merged bindings, spelled and matched once: `MoveVertical`,
    `Cycle`, `Quit`, `Back`, `SwitchKind` (next/previous section plus left/right). Only the
    connections table switches on `SwitchKind` (`Model.switchKind` → `NextKind`/`PrevKind`);
    the secrets and transports tables answer it with an "only one" status, the workspaces
    table does not match it.
  - `pressedTwice` only doubles the hint; `Copy` is built with it but matches on one press.
    Waiting for the second press is the consumer's job — the info sheet's `copyRow`.
  - `Search` is also the prompt marker before the query, so `frame.go` spells it from the
    binding, not from `theme.go`.
  - **Every key is rebindable, so these documents name the binding, not the key.** A key
    that is given is the default from `config.DefaultKeybindings()` and marked as such.
- **`ui/keyhint_table.go` — the whole footer, in order.** `keyhintLines` assembles the
  expanded key table, the search field and the prompt line, each closed by its own rule. A
  screen contributes only `keyhints()` and one
  `keyhintFooter{groups, open, search, typing, query, pos}`:

  | Field | Effect |
  |---|---|
  | `open` | expands the table; the prompt becomes "exit from help" |
  | `search` | shows the field; without `typing` it is a standing filter, where `Cancel` means "clear the filter", not "quit", and the prompt says so |
  | `typing` | gives the field the caret; the prompt shrinks to keep (`Confirm`) and clear (`Cancel`) |
  | `pos` | the cursor counter, pinned bottom-right |

  A searchable screen contributes a `keyhintSearch` group
  (`searchKeyhints(label, filtered)`), which also puts the search hint on the prompt line.
  `keyhintBar` and `keyhintStatus` are the unframed variants forms use.

A table's footer is `footerLines` inside `render`, never a screen of its own: optional
blocks, each followed by its own rule.

1. The status line, **only when it has something to say** — `statusKind != kindIdle`, or a
   delete confirmation is pending. The resting "› ready" is not drawn, so a message that
   must survive gets a kind (`kindOK` for a result, `kindWarn` for a refusal).
2. The expanded key table, when `Help` is toggled.
3. The search field (`frameSearch`: the `Search` marker, the query and, while focused, the
   caret) — **whenever the field has focus or a filter is standing**. It is not a modal
   overlay: the key table and the prompt stay, and only `Cancel` removes it.
4. The prompt line, always last, carrying the counter.

A searchable screen reads keys in three passes, most modal first: the focused field
(`Searching()`), the delete confirmation, the screen's own keys. The focused field matches
`keyMap.Interrupt` first (quit), then `Cancel`/`Confirm`/`Backspace`, plus `Search` on an
empty query as a second way out; anything else is text from `tea.KeyPressMsg.Text`.
`Confirm` keeps the filter and drops the focus; `Cancel` then clears the filter before it
means quit.

Prose keeps its own punctuation: the `·` in a status message is copy, written where it is
used. The arrows in a key hint (`↑/k`, `←/→`) spell a key and come from `keybinding.go`.

## The connections table's columns

Columns follow the *kind* of the active `ConnType` (`ConnType.Kind()`), never a named type:
a database tab shows `NAME · USERNAME · HOST · PORT · DBNAME · PING`, any other kind (ssh)
drops `DBNAME` and widens `HOST`. Only `hasDatabaseColumn`/`hostWidth` in `table_view.go`
decide it; `tableInner` never changes.

## The three/four-file screen split

A complex screen is a model split across files sharing a base name — `table.go` /
`table_view.go` / `table_commands.go`, likewise `secret_table*`, `transport_table*`,
`workspace_table*`, `form*`, `transport_form*`:

| File                 | Contents |
|----------------------|----------|
| `<name>.go`          | the model struct, its message types, `Init`, `Update`, the key/state handlers. The reducer: value receivers, return a new model, no I/O, no rendering, no `tea.Cmd` bodies — only the decision to run one. |
| `<name>_view.go`     | `View`, `render`, the per-line helpers. Read-only over the model; owns the screen's width constants (`tableInner`, `formInner`, `secretInner`, column widths) and composes `frame*` builders only. |
| `<name>_commands.go` | every `tea.Cmd` producer and side effect: calls into the view, pings, and the blocking `run<Screen>` entry points that push a model with `runScreen`, wait, assert the returned model and hand back a typed result. |

- `Update` returns a `tea.Cmd` from `_commands.go`; the command does the slow or blocking
  work and folds the outcome back as a message struct (`pingResultMsg`, `runResultMsg`,
  `connChangedMsg`, `secretChangedMsg`, `secretPickedMsg`, `formPingMsg`, …) that `Update`
  applies. Concurrent work such as spinner ticks goes through `tea.Batch`.
- One screen serves two roles by a flag, not a second model: the secrets table is also the
  form's location chooser. The flag is on the view — `view.Secrets.Picking()`, set by
  `StartPicking` (`internal/ui/view/AGENTS.md`) — and read in `secret_table.go`,
  `secret_table_view.go` and `keyhint_table.go`.
- A new screen follows this split and has a `keyhints()` method, which is all the help
  footer needs.

## One program, a stack of screens — `screen_stack.go`

**There is one `tea.Program` for the whole run.** `WithScreens(run)` starts it around a
`screenStack` and keeps it alive until `run` returns; every `cmd` entry point that opens a
screen is wrapped in it, and a nested call joins the running program. A screen never builds
a `tea.Program` of its own.

- **`runScreen(model)` opens a screen and blocks until it is done.** The stack draws and
  feeds only the top screen (a `tea.WindowSizeMsg` goes to all, and a pushed screen gets the
  last known size at once). On `tea.Quit` the stack pops the screen instead of quitting and
  hands its final model to the caller, which asserts the type and reads the result. Every
  `run<Screen>` entry point is that.
- **`screenCmd(run, done)` is how a screen opens another one.** `run` is the blocking
  `run<Screen>` call, executed as an ordinary command while the parent stays on the stack;
  `done` turns its error into the parent's message. `changeExec` and its per-screen siblings
  fill in the "something may have changed" message.
- **`execCmd(run, done)` is for work that needs the real terminal** — the database client.
  Only this path uses `tea.Exec`, which releases the terminal; opening a screen never does.
- Messages are routed by screen: `wrapScreenCmd` tags what a screen's commands produce with
  its id, so a result arriving under a newer screen still reaches the one that asked, and
  one for a popped screen is dropped.

`screen_size.go`: `screenSize` observes the window size, answers the frame width, the modal
width and the rows that fit, and drives the screen's `view.Scroll` so the cursor stays
visible. A screen never does that arithmetic itself.

## The info sheet — `info*.go`

A read-only sheet over the connection under the table's cursor, so an entry can be read
without opening the edit form. `keyMap.Connection.Info` pushes it as its own screen
(`runInfo`); it reports only whether the user left through `keyMap.Connection.Edit`, which
the table turns into its own `editCmd`, so the screen that owns the write opens the form. No
other key on the sheet writes, and the table's cursor never moves while it is open.

Up to three pages, switched with next/previous section or left/right:

| Page | Rows |
|---|---|
| `general` | `META` — the metadata fields; `CLI` — the resolved client: `cli`, `version`, `path` |
| `endpoint` | `ENDPOINT` — the connection fields, then `dsn`: the target `ConnectionString` builds, with the password as `•••` |
| `links` | `LINKS` — one row per `ConnMeta.Links` entry; `Confirm` opens the highlighted one |

Each of the four blocks opens with a `frameHead` band, so no page starts on a bare field row.

- **A connection with no links has no links page.** `infoModel.tabs()` is the one list of
  what exists — the tab strip, the page cycle and the footer's page list read it. It is a
  property of the record, not a mode: a link added in the edit form brings the page back on
  the next open.
- **The pages are not per-database.** `specRows` reads them off `spec.FormSpecs[kind]`: the
  metadata section is `general`, every other section `endpoint`, and `SecretProviderKey` /
  `SecretValueKey` are dropped at that one point. A new database type needs no change here,
  and nothing else on the sheet could name a password — `infoDSN` masks the one value that
  could carry it.
- **The cursor does not exist until it is reached for.** `infoModel.cursor` is `-1` until
  the first `MoveVertical` key creates it on the first or last field; it then steps field to
  field, wraps at both ends and skips section heads, which are not picks. `Cancel` clears
  it, a second `Cancel` leaves; switching page drops it and resets the window. The focused
  row carries the accent background, `gCaret`, and what it accepts — yank, or open and yank
  on a link.
- **`copyRow` copies on the second `Copy` press**, so a stray keystroke never reaches the
  clipboard; the first press arms it and says so in the status line. `writeClipboardCmd`
  sits beside `readClipboardCmd` in `ui/clipboard_commands.go`, over
  `cli.WriteClipboard`/`cli.ReadClipboard`, which name the binaries.
- **Every page follows the terminal height** through `infoModel.body`, a `view.Scroll`,
  with the scroll note in the top border. The links page also caps its list at
  `infoLinkWindow` rows through a second `view.Scroll` the cursor drags (`Follow`); its
  section head counts what is out of sight as `↑ N above · ↓ N below`. Rows outside that
  window are still built and counted, only marked `hidden`, so the cursor keeps its place.

## The type picker — `init_table*.go`

`initModel` lists the connection types (`config.ConnTypes`) with the clients `cli.Detect`
found for each. `Confirm` picks the type under the cursor and the screen returns it
(`runTypePicker`); a type with no installed client cannot be picked — the status line names
the binaries to install. The cursor starts on the first type that has one.

It is reached two ways, told apart by its one mode flag, `initModel.adding`:

- **Before any connection exists.** `runConnections` finds every repository empty and calls
  `runFirstConnection`: the picker, then the add form of the picked type, back to the picker
  when the form is cancelled. A first `conm` shows this instead of an empty table.
- **From the connections table**, to add a connection of another type
  (`Model.addOfTypeCmd`, built with `newAddConnectionModel`): the picker, then the add form,
  and the table switches to the picked type.

**A session is rebuilt, never grown.** A `repository.Session` is built from one
`config.Conm` and one workspace, so a change to either — a client switched in settings,
another workspace chosen — ends the pass with `Model.reload` set; `runConnectionsLoop`
closes the session, re-reads `conm.toml` and builds the next one, opened on `Model.active`.

## The entity screens — secrets, transports, workspaces

Three tables shaped like the connections table, each over its own view, with
`keyMap.<Entity>.Add` / `Edit` / `Delete`, a search, a delete confirmation and a form run
through `formLoop`:

| Screen | Model | View | Form |
|---|---|---|---|
| secrets | `secretModel` | `*view.Secrets` | `specFormModel[config.Secret]` over `secretFormSpecs[scheme]` |
| transports | `transportModel` | `*view.Transports` | `transportFormModel` over `spec.TransportFormSpec` |
| workspaces | `workspaceModel` | `*view.Workspaces` | `specFormModel[string]` over `spec.WorkspaceFormSpec` |

- **Each row says who uses it.** Secrets and transports show their usages (`UsagesAt`); the
  repository refuses to remove one still referenced (`*SecretInUseError`,
  `*TransportInUseError`) and the screen only renders the refusal.
- **The transport form is its own model** because its fields depend on live data: `via` is
  picked from the existing SSH connections (a `spec.Sources` entry built in
  `transportModel.formContext`), and it can ping that connection before saving
  (`pingViaCmd`).
- **The workspaces screen switches, it does not edit in place.** `Confirm` calls
  `view.Workspaces.Use` and leaves with `entered` set. Opened from the connections table,
  that returns as `workspaceSwitchedMsg`, ending the pass with `Model.reload` and
  `Model.workspace` so the loop rebuilds the session on the chosen workspace. Opened from the
  shell (`RunWorkspaces`, behind `conm workspace`), it selects the workspace and continues
  into `Run`. Removing a workspace also removes its secret material; the active one cannot
  be removed (`*ActiveWorkspaceError`).

## Settings — `settings*.go`, `settings_keybind*.go`, `settings_appearance.go`

`settingsModel` edits `conm.toml` and nothing else, in three tabs.

- **Clients.** One row per connection type with the clients `cli.Detect` found and their
  versions, probed in the background on `Init`. `Cycle` moves to the next *installed* client
  of that type; a type with no alternative says which binaries to install. `saveClientCmd`
  checks the name through `cli.Validate` before it writes.
- **Keybindings.** One row per `config.Keybinding` that is not `Locked`
  (`settingsModel.bindings()`), its keys drawn as chips.
  - `keyMap.Settings.AddKey` / `ChangeKey` start a capture: the next key press is the
    binding, a second press within `chordTimeout` makes it a two-stroke sequence, `Confirm`
    keeps the first stroke alone, `Cancel` abandons it.
  - `RemoveKey` drops a chip; `Reset` restores that binding's defaults after a confirmation.
  - A binding marked `OnlyAppend` keeps its default keys: they can be added to, not replaced
    or removed (`isPinned`).
  - Every change is checked through `KeybindingSettings.Validate` before it is written.
- **Appearance.** Three sections from `appearanceRows`, scrolled like the keybindings: the
  accent; the border style and one row per border element; the glyph set and one row per
  glyph. Every row is a select box with a preview or note beside it, and every change is
  saved at once; the saved message runs `applyConm`, so the screen redraws in the new look.
  - A preset row (`options`, `field`) cycles its `config` name list. An element row
    (`group`, `element`) cycles the distinct runes the presets use for that element
    (`runeGroup.variants`) and stores the choice as an override — or drops the override when
    the choice is the current preset's rune, so the file keeps only real differences. Its
    note names the preset's rune while it is overridden; `keyMap.Settings.Reset` drops it.
  - `keyMap.Settings.EditValue` opens typed input: the accent keeps `#` and hex digits up
    to seven and needs a full `#rrggbb`; an element keeps the last character typed and needs
    one cell (`config.IsOneCell`). `Confirm` saves, `Cancel` drops the draft; an unfit draft
    keeps the input open. A value outside the choices shows `custom` in the box.
  - While the glyphs are `ascii`, every border row is `locked` and refuses to change.
  - A change copies the settings (`AppearanceSettings.Clone`) before it touches the
    override maps, so the screen's own `conm` is never mutated before the write succeeds.

Every write goes through `config.CreateConm` at once; there is no "save" step. The screen
reports only whether anything changed, and the connections table answers a change with a
reload, which is how a new client or key takes effect.
