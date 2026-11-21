# `internal/ui` — the Bubble Tea TUI

## Purpose

Every screen conm draws, and the blocking entry points `cmd` calls. Screens are composed
from shared primitives, reach entities only through `ui/view`, and take their forms from
`ui/spec`.

May import (internal): `ui/view`, `ui/spec`, `repository`, `importer`, `config`, `network`, `cli`,
`secret`, `transport`, `utils/keys`.

## Files

A screen is split `<name>.go` (model, `Update`) / `<name>_view.go` (render) /
`<name>_commands.go` (`tea.Cmd`s, I/O, the blocking `run<Screen>` entry points).

| Files | Holds |
|---|---|
| `ui.go` | the entry points: `Run`, `RunOn`, `Connect`, `Ping`, `Connections`, `RemoveConnection`, and the loop that rebuilds a `repository.Session` per pass |
| `screen_stack.go` | `WithScreens`, `runScreen` — one `tea.Program` hosting a stack of screens |
| `screen_size.go` | `screenSize` — widths and visible rows derived from the terminal size |
| `frame.go`, `theme.go` | the line builders; the colours and glyphs |
| `keybinding.go`, `keyhint_table.go` | `keyMap`; the footer |
| `conn_error.go` | `connError`, `hintFor` — a `*network.OpError` reduced to display strings |
| `clipboard_commands.go` | paste and yank over `cli.ReadClipboard`/`cli.WriteClipboard` |
| `table*.go` | the connections table — `Model`, the main screen |
| `info*.go` | the read-only info sheet |
| `form*.go` | `formModel`, the connection add/edit form; `form_array.go` its chip fields; `form_loop.go` the open → save → retry loop every form runs in |
| `spec_form*.go` | `specFormModel[T]`, the generic form over any `spec.FormSpec[T]` |
| `secret_table*.go`, `secret_form_commands.go` | the secrets screen and its form |
| `transport_table*.go`, `transport_form*.go` | the transports screen and its form |
| `workspace_table*.go`, `workspace_form_commands.go` | the workspaces screen and its form |
| `settings*.go`, `settings_keybind*.go` | settings: the client per type, keybinding overrides |
| `init_table*.go` | the type picker: shown before any connection exists, and when adding a connection of another type |
| `import.go` | `ImportConnections`, `RenderImport` — saves `importer.Entry`s for `conm <type> import` and prints what became of each |
| `doctor.go` | `Doctor` — renders `repository.CheckClients`/`CheckFiles` for `conm doctor` |

## Main entities

- `Model` — the connections table; holds a `*view.Connections`, the status line and the
  per-`Ref` ping state.
- `formModel` / `specFormModel[T]` — the form drivers; everything entity-specific is a
  `spec.FormSpec` (see `internal/ui/spec/AGENTS.md`).
- `formLoop[T]` — how any form is run to completion; a screen supplies `open` and `save`.
- `keyMap` — every key the TUI reacts to, named by meaning.
- `screenStack` — the host a screen pushes another screen onto.

## Import — `import.go`

`ImportConnections` is a shell entry point like `RemoveConnection`: it works on the
repository, not through `ui/view`, and opens a screen only in `ImportInteractive`.

- `planImport` sorts the entries with `repository.Connections.Check` before anything is
  written: the ones that can be saved together, the ones refused, and the ones conm already
  has a connection to. **A `*repository.DuplicateConnectionError` is not a refusal** — the
  same target under any name is `ImportPresent` in every mode, so importing a file twice is
  not an error. A taken name with another target is a refusal.
- `ImportStrict` writes nothing when one entry is refused and leaves the rest `ImportNew`;
  the others save what was admitted with one `AddAll`. `ImportShow` stops after the plan.
- `ImportInteractive` then opens each refused entry that has a connection in the add form,
  seeded from it and with the refusal as its status, through the same `formLoop` as
  `conm <type> add`. An entry the importer could not read has nothing to seed a form with
  and stays `ImportFailed`.
- **The password is stored after the connection is saved**, in the keyring at the location
  the saved connection refers to, as a secret named after the connection. A connection the
  form pointed at another provider keeps what the user chose and nothing is stored. A
  keyring failure leaves the connection saved and is reported in `ImportResult.Note`; it is
  not rolled back.
- `ImportResult.Warnings` — what the importer left out of an entry — is filled in
  `ImportShow` only, so an import of a large file prints one line per connection.
- `RenderImport` prints `[glyph] name dsn` with the reason and the warnings underneath, the way `doctor`
  prints a check: `kindOK` imported, `kindErr` refused, `kindWarn` already there, and
  `gImportNew` in `cPostgres` for an entry that could be imported but was not. The DSN is
  `ConnectionString("")`, so it never carries the password.

## Shared primitives — reuse, never re-implement

Three files own everything a screen is drawn and driven with. **When you add a UI element,
compose it from these; if a piece is missing, add it there and use it from both places.
Never inline a private copy, a raw rune, a raw colour or a raw key string in a screen.**

- **`ui/frame.go` — the frame blanks.** Every screen is one box of a fixed inner width,
  assembled line by line: `span` (a styled run of text), `truncPad`, `border`, and the
  builders `frameTop`, `frameLine`, `frameRule`, `frameBottom`, `frameBadge`, `frameStatus`,
  `frameKeybar`, `frameSearch`, `framePong`, `frameErrPanel`. Every builder takes the inner
  width `w` explicitly, so the same primitive serves the wide connections table and the
  narrower forms. Also here: `keybind`, `statusKind`/`statusGlyph`, `cursorPos`.
- **`ui/conn_error.go` — the failure panel's data.** `connError` is a `*network.OpError`
  reduced to the strings a line is drawn from; `newConnError` is the only place that
  `errors.AsType` is called, and `hintFor` is the advice copy keyed by `network.ErrorCode` —
  the only part of a failure the UI owns, since it spells the key to press from `keyMap`.
  Classifying an error is not the UI's job — see `internal/network/AGENTS.md`.
- **`ui/theme.go` — the colours and the frame elements themselves.** The palette (`c*`),
  the glyph set (`g*`: box-drawing runes, status icons, row/field markers, input chrome,
  ellipsis) and `typeColor`. `frame.go` decides *how* a line is built, `theme.go` decides
  *what it looks like* — a restyle must never touch drawing code, so a view may not reach
  for a literal rune or `lipgloss.Color`.
- **`ui/form_loop.go` — how any form is run to completion.** `formStatus` (the message plus
  its `statusKind`), `formResult[T]` (what one run of a form produced, plus the raw values it
  produced them from) and `formLoop[T]` (open → save → on failure open again). A screen
  supplies only `open` and `save`; it never drives a form itself.
- **`ui/keybinding.go` — every key the TUI reacts to.** A `binding` is a set of keys plus
  the way that key is spelled in a hint. `keyMap` is rebuilt by `newKeymap` from
  `config.KeybindingSettings` on every pass of the connections loop, so a key changed in
  settings is live on the next screen; settings that do not validate fall back to
  `config.DefaultKeybindings()`. The shared keys sit at the top level (`Up`, `Confirm`,
  `Cancel`, `Paste`, `Copy`, `Search`, `Help`, …) and the rest are grouped by what they act
  on: `keyMap.Screen` (opening another screen), `keyMap.Connection`, `keyMap.Form`,
  `keyMap.Secret` / `Transport` / `Workspace` (add, edit, delete of that entity),
  `keyMap.Settings`, `keyMap.Prompt`. `Update` matches with
  `switch { case keyMap.X.matches(key): … }`, never against a raw string; a hint takes its
  key from `keyMap.X.hint` and supplies only the label. Keys a screen treats as one group are
  merged bindings, spelled and matched once: `MoveVertical`, `Cycle`, `Quit`, `Back`, and
  `SwitchKind` (next/previous section plus left/right — the connections table switches
  connection type, the secrets table switches store, and a table with one kind says so
  rather than doing nothing silently). `Copy` is `pressedTwice`: it fires on the second
  press, so a stray keystroke never reaches the clipboard. `Search` is also the prompt
  marker the query is echoed behind, so `frame.go` spells it from the binding rather than
  from `theme.go`.
  **Every key is rebindable, so these documents name the binding, not the key.** Where a key
  is given, it is the default from `config.DefaultKeybindings()` and is marked as such.
- **`ui/keyhint_table.go` — the whole footer, in order.** `keyhintLines` assembles the
  expanded key table, the search field and the prompt line, each block closed by its own
  rule, so no screen decides that order for itself. Each screen contributes only its
  `keyhints()` method and one `keyhintFooter{groups, open, search, typing, query, pos}`:
  `open` expands the table and flips the prompt to "exit from help", `search` puts the
  field on screen, `typing` gives it the caret and reduces the prompt to keep (`Confirm`) and
  clear (`Cancel`) — the caret already says the field has focus, so the prompt drops everything the
  focused field does not answer — and `pos` is the cursor counter pinned to the bottom-right
  corner — the one thing on screen that is always in the same place. `search` without
  `typing` is a filter standing over an unfocused field, the one state where `Cancel` means
  "clear the filter" rather than "quit", so the prompt spells that out as well. A screen declares it can be searched by
  contributing a `keyhintSearch` group (`searchKeyhints(label, filtered)`) — that group is also what puts the search hint on the prompt line. `keyhintBar`
  and `keyhintStatus` are the unframed variants forms use.

A table's footer is a stack of optional blocks, each followed by its own rule and ending in
the prompt line — `footerLines` inside `render`, never a screen of its own:

1. the status line, **only when it has something to say** — `statusKind != kindIdle`, or a
   delete confirmation is pending. The resting "› ready" is not drawn at all, so a message
   that must survive gets a kind (`kindOK` for a result, `kindWarn` for a refusal); an idle
   status is a deliberate "say nothing". The cursor counter moved off this line, since the
   line is not always there.
2. the expanded key table, when `Help` is toggled.
3. the search field (`frameSearch`) — the `Search` marker, the query and the caret, which is only there
   while the field has focus — **whenever the field has focus or a filter is standing**. It is a field above the footer, not a modal overlay: the key table
   above and the prompt below both stay, and only `Cancel` takes it away.
4. the prompt line, which always ends the screen and carries the counter.

A screen with a search reads its keys in three passes, most modal first: the field while it
has focus (`Searching()`), then the delete confirmation, then the screen's own keys. While
the field has focus the screen matches only `Cancel`/`Confirm`/`Backspace`, plus `Search` on
an empty query as a second way out, and otherwise appends `tea.KeyPressMsg.Text`, so every
other key is text; `Confirm` keeps the filter and drops the focus, and once the focus is gone
`Cancel` clears the filter before it means quit.

Prose keeps its own punctuation: the `·` in a status message is copy, not chrome, and stays
written where it is used. The arrows in a key hint (`↑/k`, `←/→`) spell a key and come from
`keybinding.go`.

## The connections table's columns

The table is drawn per active tab, so its columns follow the *kind* of the active
`ConnType` (`ConnType.Kind()`), never a named type: a database tab shows
`NAME · USERNAME · HOST · PORT · DBNAME · PING`, any other kind (ssh today) drops `DBNAME`
and gives its width to `HOST`. `hasDatabaseColumn`/`hostWidth` in `table_view.go` are the
only two places that decide it, and `tableInner` never changes.

## The three/four-file screen split

A complex screen is a Bubble Tea model split by responsibility across files sharing a base
name — `table.go` / `table_view.go` / `table_commands.go`, and likewise `secret_table*`,
`transport_table*`, `workspace_table*`, `form*`, `transport_form*`. Follow this split and keep
each piece in its file:

| File                 | Contents |
|----------------------|----------|
| `<name>.go`          | the model struct, its message types, `Init`, `Update`, and the key/state handlers. The reducer: value receivers, return a new model, no I/O, no rendering, no `tea.Cmd` bodies — only the decision to run one. |
| `<name>_view.go`     | `View`, `render` and the per-line helpers. Read-only over the model; owns this screen's width constants (`tableInner`, `formInner`, `secretInner`, column widths) and composes `frame*` builders only. |
| `<name>_commands.go` | every `tea.Cmd` producer and side effect: calls into the view, pings, and the blocking `run<Screen>` entry points that push a model with `runScreen`, wait for it, assert the returned model and hand back a typed result. |

How they talk: `Update` returns a `tea.Cmd` from `_commands.go`; the command does the slow
or blocking work and folds the outcome back as one of the model's message structs
(`pingResultMsg`, `runResultMsg`, `connChangedMsg`, `secretChangedMsg`, `secretPickedMsg`,
`formPingMsg`, …), which `Update` applies. Concurrent work such as spinner ticks goes
through `tea.Batch`. One screen can serve two roles by a flag rather than a second model:
`secretModel.picking` turns the secrets table into a chooser that returns a location to the
form.

Add a new screen by following the same split, and give it a `keyhints()` method so the help
footer works for free.

## One program, a stack of screens — `screen_stack.go`

**There is one `tea.Program` for the whole run.** `WithScreens(run)` starts it around a
`screenStack` and keeps it alive until `run` returns; every `cmd` entry point that opens a
screen is wrapped in it, and a nested call simply joins the program already running. A
screen never builds a `tea.Program` of its own.

- **`runScreen(model)` opens a screen and blocks until it is done.** It pushes the model
  onto the stack and waits; the stack draws and feeds only the top screen (a
  `tea.WindowSizeMsg` goes to all of them, and a pushed screen is handed the last known size
  at once). When the screen returns `tea.Quit`, the stack pops it instead of quitting the
  program and hands its final model back to the caller, which asserts the type and reads
  the result off it. That is what every `run<Screen>` entry point is.
- **`screenCmd(run, done)` is how a screen opens another one.** `run` is the blocking
  `run<Screen>` call, executed as an ordinary command while the parent stays on the stack
  underneath; `done` turns its error into the parent's message. `changeExec` and its
  per-screen siblings are that with the "something may have changed" message filled in.
- **`execCmd(run, done)` is for work that needs the real terminal** — handing it to the
  database client. Only this path goes through `tea.Exec`, which releases the terminal for
  the duration; opening a screen never does.
- Messages are routed by screen: `wrapScreenCmd` tags what a screen's commands produce with
  that screen's id, so a result that arrives after another screen was pushed on top still
  reaches the screen that asked for it, and one for a screen already popped is dropped.

`screen_size.go` is the other thing every screen shares: `screenSize` observes the window
size and answers the frame width, the modal width and how many rows fit, and drives the
screen's `view.Scroll` so the cursor stays visible. A screen never does that arithmetic
itself.

## The info sheet — `info*.go`

`ui/info*.go` is a read-only sheet over the connection the table is sitting on: the place to
look at an entry without opening the edit form, which is one keystroke away from mutating
the record. It is opened from the connections table by `keyMap.Connection.Info`, pushed as its
own screen (`runInfo`) like the secrets table, and reports only whether the user left through
`keyMap.Connection.Edit`; the table turns that into its own `editCmd`, so the edit form is still opened by the screen that owns the write. No other key
on the sheet writes anything, and the table's cursor never moves while it is open.

Up to three pages, switched with next/previous section or left/right:

| Page | Rows |
|---|---|
| `general` | `META` — the metadata fields; `CLI` — the resolved client: `cli`, `version`, `path` |
| `endpoint` | `ENDPOINT` — the connection fields, then `dsn`: the target `ConnectionString` builds, with the password as `•••` |
| `links` | `LINKS` — one row per `ConnMeta.Links` entry; `Confirm` opens the highlighted one |

Every block on every page opens with a `frameHead` band, so a page never starts on a bare
field row and the four blocks read the same.

**A connection with no links has no links page.** `infoModel.tabs()` is the one list of what
exists — the tab strip, the page cycle and the footer's page list all read it — so an
unlinked connection cycles between two pages and never lands on an empty sheet. It is a
property of the record, not a mode: adding a link in the edit form brings the page back on
the next open.

**The pages are not per-database.** `specRows` reads them off `spec.FormSpecs[kind]`: the
metadata section is `general`, every other section is `endpoint`, and `SecretProviderKey` /
`SecretValueKey` are dropped at that one point. A new database type therefore shows up here
with no change, and there is no second place where the sheet could start naming a password —
`infoDSN` masks the one value that could carry it.

**The cursor does not exist until it is reached for.** `infoModel.cursor` is `-1` until the
first `MoveVertical` key, which creates it on the first or last field; it then steps field to field,
wraps at both ends, and skips section heads because heads are not counted among the picks.
`Cancel` clears it, a second `Cancel` leaves; switching page drops it and resets the window.
The focused row carries the accent background, `gCaret`, and what it accepts — yank, or open
and yank on a link. `Copy` fires on its second press, so a stray keystroke never reaches the
clipboard; `writeClipboardCmd` sits beside `readClipboardCmd` in
`ui/clipboard_commands.go`, over the `cli.WriteClipboard`/`cli.ReadClipboard` that name the binaries.

**Only the links page scrolls.** It renders at most `infoLinkWindow` rows through a
`view.Scroll` the cursor drags (`Follow`), and its section head counts what is out of sight
as `↑ N above · ↓ N below`. The rows outside the window are still built and still counted,
only marked `hidden`, so the cursor keeps its place in the whole list.

## The type picker — `init_table*.go`

`initModel` is the list of connection types conm knows (`config.ConnTypes`) with the clients
`cli.Detect` found for each on this host. `Confirm` picks the type under the cursor and the
screen returns it (`runTypePicker`); a type with no installed client cannot be picked — the
status line names the binaries to install instead. The cursor starts on the first type that
has one.

It is reached two ways, and that is its one mode flag, `initModel.adding`:

- **Before any connection exists.** `runConnections` finds every repository empty and calls
  `runFirstConnection`: the picker, then the add form of the picked type, looping back to
  the picker when the form is cancelled. This is what a first `conm` shows instead of an
  empty table.
- **From the connections table**, to add a connection of another type
  (`Model.addOfTypeCmd`, built with `newAddConnectionModel`): the picker, then the add form,
  and the table switches to the picked type.

**A session is rebuilt, never grown.** A `repository.Session` is built from one
`config.Conm` and one workspace, so anything that changes either — a client switched in
settings, another workspace chosen — ends the pass with `Model.reload` set, and
`runConnectionsLoop` closes the session, re-reads `conm.toml` and builds the next one, opened
on `Model.active`.

## The entity screens — secrets, transports, workspaces

Three tables of the same shape as the connections table, each over its own view and each
with `keyMap.<Entity>.Add` / `Edit` / `Delete`, a search, a delete confirmation and a form
run through `formLoop`:

| Screen | Model | View | Form |
|---|---|---|---|
| secrets | `secretModel` | `*view.Secrets` | `specFormModel[config.Secret]` over `secretFormSpecs[scheme]` |
| transports | `transportModel` | `*view.Transports` | `transportFormModel` over `spec.TransportFormSpec` |
| workspaces | `workspaceModel` | `*view.Workspaces` | `specFormModel[string]` over `spec.WorkspaceFormSpec` |

- **Each row says who uses it.** Secrets and transports show their usages
  (`UsagesAt`), and removing one that is still referenced is refused by the repository
  (`*SecretInUseError`, `*TransportInUseError`); the screen only renders the refusal.
- **The transport form is its own model** because its fields depend on live data: `via` is
  picked from the SSH connections that exist (a `spec.Sources` entry built in
  `transportModel.formContext`), and it can ping the connection it goes through before
  saving (`pingViaCmd`).
- **The workspaces screen switches, it does not edit in place.** `Confirm` calls
  `view.Workspaces.Use` and leaves with `entered` set. Opened from the connections table,
  that comes back as `workspaceSwitchedMsg`, which ends the pass with `Model.reload` and
  `Model.workspace` so the loop rebuilds the session on the chosen workspace. Opened from the
  shell (`RunWorkspaces`, behind `conm workspace`), it selects the workspace and continues
  into `Run`. Removing a workspace also removes its secret material, and the active one
  cannot be removed (`*ActiveWorkspaceError`).

## Settings — `settings*.go`, `settings_keybind*.go`

`settingsModel` edits `conm.toml` and nothing else, in two tabs.

- **Clients.** One row per connection type with the clients `cli.Detect` found and, probed
  in the background on `Init`, their versions. `Cycle` moves to the next *installed* client
  of that type; a type with no alternative says which binaries to install.
  `saveClientCmd` checks the name through `cli.Validate` before it writes.
- **Keybindings.** One row per `config.Keybinding`, its keys drawn as chips.
  `keyMap.Settings.AddKey` / `ChangeKey` start a capture: the next key press is the binding,
  a second press within `chordTimeout` makes it a two-stroke sequence, `Confirm` keeps the
  first stroke alone, `Cancel` abandons it. `RemoveKey` drops a chip and `Reset` restores that
  binding's defaults after a confirmation. A binding marked `OnlyAppend` keeps its default keys:
  they cannot be replaced or removed, only added to (`isPinned`). Every change is checked through
  `KeybindingSettings.Validate` before it is written.

Every write goes through `config.CreateConm` at once; there is no "save" step. The screen
reports only whether anything changed, and the connections table answers a change with a
reload, which is how a new client or key takes effect.
