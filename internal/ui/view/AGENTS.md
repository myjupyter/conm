# `internal/ui/view` — what a screen shows and where it is looking

## Purpose

The layer between the Bubble Tea screens and the repositories: the rows to draw, the search
over them and the cursor. No rendering, no Bubble Tea, no file I/O. **The runtime TUI touches
entities only through this package.**

May import (internal): `repository`, `config`, `network`, `cli`, `secret`, `transport`.

## Files

| File | Holds |
|---|---|
| `connection.go`, `connection_state.go` | `Connections`, `ConnRef` — rows, then the cursor-bound half |
| `secret.go`, `secret_state.go` | `Secrets`, `SecretRef`, the picker mode |
| `transport.go` | `Transports`, `TransportRef` |
| `workspace.go` | `Workspaces`, `WorkspaceRef` |
| `index.go`, `state.go` | the shared core: `Ref[K]`, `index[K]`, `tableState[K]` |
| `search.go` | `searcher`, `document`, `substringSearcher` |
| `scroll.go` | `Scroll` — the window over the rows |

## Main entities

- `Connections`, `Secrets`, `Transports`, `Workspaces` (`view.Connections`, `view.Secrets`,
  `view.Transports`, `view.Workspaces`) — one view per entity, each embedding `tableState[K]`
  over the repositories of that entity. A screen holds a pointer to one and nothing else: no
  repository, no row index, no filtered slice of its own.
- `Ref[K]{Kind, Index}` — a record's address: which repository, which position in it.
  Everything a screen addresses is a *visible row*; `RefAt` translates.
- `Scroll` — which slice of the rows fits the terminal.

## How it works

**The state is the view's own and never leaves it.** `tableState[K]` and `index[K]` are
unexported and embedded into the exported view of `ui/view`, so a screen holds a
`*view.Connections` and changes what it is looking at only through the view's methods.

## The rows — `index.go`, `connection.go`, `secret.go`

A view owns the repositories of one entity keyed by their kind (`config.ConnType` for
connections, `secret.Scheme` for secrets), the order they were registered in, and which is
`Active()`. Its `index` carries one entry per record across *every* repository: a `Ref[K]`
plus the fields a query matches against. `Reindex` rebuilds it after any mutation and hands
the fields to the searcher as `document`s keyed by their position in `entries`.

The visible rows are `[]Ref[K]`: the active kind's records, in order, narrowed by the query
when one is standing. **The active kind is a filter of its own**, applied on both paths in
`refilter`: a query never shows a row of another kind, and switching kind re-matches the
same query against the kind switched to. The `*For(ref)` accessors take a `Ref` for work
that outlives the current row order. `ConnRef` and `SecretRef` are the instantiated aliases
screens should spell.

`Connections.ClientInfo(ctx)` is the exception to "indexed by row": the client is a property
of the database, not of a record, so it is served by the active kind's repository and takes
no `Ref`. It returns the `cli.Info` the repository got from its launcher — the only accessor
here that can start a subprocess — and reports `false` when the active kind has no
repository.

## The search — `search.go`

**The matching algorithm lives behind `searcher` and nowhere else.** `index` owns the rows,
the query string and the cursor consequences; it knows nothing about *how* a query selects a
record:

```go
type document struct { id int; fields []string }

type searcher interface {
    Reindex(docs []document)
    Search(query string) (hits []int, filtered bool)
}
```

- `Reindex` hands the searcher every record once, so an implementation may build and keep
  any index — the default `substringSearcher` keeps a case-folded copy of the fields.
- `Search` returns the matching ids in index order, plus whether the query narrowed
  anything: `filtered == false` tells `refilter` to fall back to "the active kind's records,
  in order", so what counts as an empty query is the algorithm's call.
- Swapping in fuzzy, prefix or scored matching is one new type plus the line in
  `newSearcher`; `index.go` and the screens do not change.
- Field collection stays on the entity side (`connectionFields`, `secretFields` →
  `searchable`, which only drops blanks); normalisation is the searcher's.

**No password material is ever indexed.** The fields come from `config.Connection` and
`config.Secret`, and neither interface exposes the password or the reference standing in for
it (`config.Postgres.Password` is reachable only through `secret.Reference`). A new field
added to `*Fields` must keep that true.

## The cursor — `state.go`, `connection_state.go`, `secret_state.go`

`state.go` is the cursor, the delete confirmation, the help toggle and the search prompt,
reached through methods, never fields:

| State | Methods |
|---|---|
| cursor | `MoveUp`/`MoveDown`/`Top`/`Bottom`/`SetCursor` |
| delete confirmation | `AskConfirm`/`ClearConfirm`/`Confirming` |
| help | `ToggleHelp`/`Help` |
| search prompt | `StartSearch`/`CommitSearch`/`CancelSearch`/`Searching`, `AppendSearch`/`TrimSearch`, `Query`/`Filtered` |

Keeping the cursor valid is the view's job: it clamps against its own length, rides the same
`Ref` through a `Search` that keeps it visible, drops to the top on a kind switch, and
re-clamps after every mutation.

Each view has a cursor-bound half — `Connection`/`Config`, `Secret`/`Usages`,
`Transport`/`Usages`, `Workspace`/`Current`, and the mutators *over the row under the cursor*
(`Add`/`Edit`/`Remove`; `Use`/`Rename` for a workspace) — so a screen never hands an index
back. The two larger views keep it in a `_state.go` file; `Transports` and `Workspaces` hold
both halves in one. `Secrets` also owns the chooser mode (`StartPicking`/`Pick`/
`StopPicking`) that turns the secrets table into the form's picker.

What stays in the screen models: `ui.Model` and `secretModel` each hold the status line, a
`screenSize` and a `view.Scroll`. `Model` also holds the reload/active/workspace loop state,
the chord state and the per-`Ref` ping state (spinner, result, error), which is made of
`bubbles` and frame types the view layer must not import; `secretModel` has no ping state. A
`Ref` survives a re-filter but not a removal, so a removal that renumbers a repository drops
the ping map (`connChangedMsg.renumbered`).

`Searching()` is the input having focus, `Filtered()` is a query narrowing the rows. A
screen draws the search input while either holds, paints it active only for the first, and
lets the second decide what cancelling means. Query editing is the view's too: a screen
appends the `tea.KeyPressMsg` text and never keeps a copy of the query.

Only the connections table switches kind through `NextKind`/`PrevKind` (`ui/table.go`);
`Connections` overrides both (`view/connection.go`) to step over kinds with no records.
`SetActive` is how a caller opens the view on a kind it already chose — the connections loop
after the type picker.

`Transports` and `Workspaces` each sit over a single repository. `Transports` is also a
`transport.Provider` (`Resolve`), which lets a form's preview ping route through a transport
without holding the repository, and `NamesOf(kind)` is the list a connection form offers.

## The window — `scroll.go`

`Scroll` counts rows and nothing else — it names no connection, secret or screen and holds
no rows, so one lives next to whatever list needs a window and takes the total on every call.

```go
type Scroll struct{ size, start int }

func (s *Scroll) Resize(size int)          // how many rows fit
func (s *Scroll) Follow(cursor, total int) // keep the cursor visible
func (s *Scroll) Range(total int) (start, end int)
func (s *Scroll) Above() int               // rows hidden before the window
func (s *Scroll) Below(total int) int      // rows hidden after it
func (s *Scroll) Top()                     // back to the first row
```

The contract:

- `Range` is half-open and always in bounds: `0 <= start <= end <= total`. A `total` that
  fits inside the window — and a `size` of zero, which means "no window given yet" — returns
  `0, total` and hides nothing, so a screen that never calls `Resize` renders every row.
- `Follow` moves the window only when the cursor has left it, by exactly as much as brings
  it back to the edge it crossed; a cursor inside the window leaves `start` alone.
- Every call clamps against the `total` it is handed, so a list that shrinks under the
  window — rows removed, a query narrowing them — slides back into range. Same for a
  `Resize` that grows the window past the tail.
- `Above`/`Below` are the hidden counts, not scrollbar geometry. `scrollNote`
  (`ui/info_view.go`) draws them as `↑ N above · ↓ N below`; tables and forms derive them
  from `Range`.

The `Scroll` type lives here, but each screen model holds the instance and drives it through
`screenSize` (`ui/screen_size.go`: `follow`/`fit`/`visibleRows`), which sizes it from the
`tea.WindowSizeMsg`. **A screen does not keep its own offset, slice the rows by hand or
compute where the cursor went.**

## The flow, end to end

```
NewX(…)           register(kind)s, below — the first registered is active   ──┐
                                                                        ▼
Reindex()         walk idx.order → repo.Len() → one entry{ref, fields} per record
   │                                                                    │
   └─► reset(entries) ─► searcher.Reindex(docs)
                     └─► refilter() ─► searcher.Search(q) ─► visible []Ref ─► Sync() clamps

read     Len / RefAt(i) / At(i) / For(ref)          — always by visible row
move     MoveUp · MoveDown · SetCursor · Top/Bottom — cursor only, no repository touched
narrow   Search(q) · SetActive(k) · NextKind()      — refilter, then re-home the cursor
type     StartSearch · Append/TrimSearch · Commit/CancelSearch — the prompt over Search(q)
write    Add(rec) · Edit(rec) · Remove()            — resolve Ref → repo call → Reindex
```

What each constructor registers:

| Constructor | Kinds |
|---|---|
| `NewConnections(repos…)` | one per repository, then re-homes to the first kind that has records |
| `NewSecrets(repos…)` | one per repository |
| `NewTransports(repo)` | one per `config.TransportKinds` |
| `NewWorkspaces(repo)` | one constant kind |

Two invariants: **a row number is a position in `visible`, never in a repository**, and
**every path that can change `visible` ends in `refilter` + a cursor clamp**. So the view
calls `Reindex` itself after each mutation, and a screen that wants a record asks for the
one under the cursor.

## Adding a table for a new entity

`Connections` and `Secrets` are the two full examples and `Transports` the minimal one; a
new entity follows the same shape and touches nothing generic. Given a repository surface
for it (`Kind() K`, `Len`, a positional accessor and the mutators — `K` being whatever names
a store of that entity, `comparable` and printable):

1. **`internal/ui/view/<entity>.go`** — the rows:
   - `type <Entity>Ref = Ref[K]`, the alias screens spell.
   - `type <Entities> struct { repos map[K]repository.<Entities>; tableState[K] }` — embed
     `tableState`, never `index`; that is what hides the cursor inside the view.
   - `New<Entities>(repos ...repository.<Entities>) *<Entities>` — skip a nil or duplicate
     repo with `v.register(repo.Kind())`, store it, finish with `v.Reindex()`.
   - `Reindex()` — walk `v.idx.order`, read each record, append
     `entry[K]{ref: <Entity>Ref{kind, i}, fields: <entity>Fields(rec)}`, then `v.reset(...)`.
   - `<entity>Fields(rec) []string` — everything a query should hit, through `searchable(…)`.
     Never pass a password or the reference standing in for one.
   - `CountOf(kind)`, the row accessors, and `rowRepo(action, i)` for the mutators. Name the
     accessors `At(i)`/`For(ref)` when the entity has one projection, as `Secrets` does; give
     each its own pair when it has several, the way `Connections` splits `ConnectionAt`/
     `ConnectionFor` (live) from `ConfigAt`/`ConfigFor` (stored).
2. **`internal/ui/view/<entity>_state.go`** — the cursor-bound half: the record under the
   cursor, `Add`/`Edit`/`Remove` (resolve through `rowRepo(action, v.Cursor())`, call the
   repository, `Reindex()`), and any screen mode that belongs to the entity rather than to
   the screen, the way `Secrets` owns picking.
3. **Build the view where its repositories are opened** and hand it to the screen. `ui.Run`
   only delegates to `RunOn`; the views are built in `runConnections` (`ui/ui.go`), in
   `openConnectionForm` (`ui/form_commands.go`) and, for workspaces, in `runWorkspaceTable`.
4. **`internal/ui`** — the three/four-file screen split (`internal/ui/AGENTS.md`) over
   `*view.<Entities>`; the model keeps only its status line and its own async state, and
   gets a `keyhints()` method.
5. **`index.go` / `state.go` / `search.go` stay untouched.** If a new table needs a new
   cursor or index behaviour, add it there once, generically — do not special-case an entity
   inside the shared core, and do not grow a private cursor in the screen.
