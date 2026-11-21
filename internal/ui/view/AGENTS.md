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

- `Connections`, `Secrets`, `Transports`, `Workspaces` — one view per entity, each embedding
  `tableState[K]` over the repositories of that entity. A screen holds a pointer to one and
  nothing else: no repository, no row index, no filtered slice of its own.
- `Ref[K]{Kind, Index}` — a record's address: which repository, which position in it.
  Everything a screen addresses is a *visible row*; `RefAt` translates.
- `Scroll` — which slice of the rows fits the terminal.

## How it works

`ui/view` is the layer between the Bubble Tea screens and the repositories. It holds no
rendering, no Bubble Tea and no file I/O; it holds the two things a screen would otherwise
grow by hand — the rows to draw, and the cursor over them.

One view per entity, all built from the same generic core:

| Entity      | View               | Files                                  |
|-------------|--------------------|----------------------------------------|
| connections | `view.Connections` | `connection.go`, `connection_state.go` |
| secrets     | `view.Secrets`     | `secret.go`, `secret_state.go`         |
| transports  | `view.Transports`  | `transport.go`                         |
| workspaces  | `view.Workspaces`  | `workspace.go`                         |
| *(shared)*  | `index[K]`, `tableState[K]`, `Ref[K]` | `index.go`, `state.go` |
| *(shared)*  | `searcher`, `document`                | `search.go`                            |
| *(shared)*  | `Scroll`                              | `scroll.go`                            |

**The state is the view's own and never leaves it.** `tableState[K]` and `index[K]` are
unexported and are embedded into the exported view, so a screen holds a `*view.Connections`
and moves the cursor by calling it. There is no state object for `ui` to hold, pass around
or mutate behind the view's back — if a screen needs to change what it is looking at, that
is a method on the view.

## The rows — `index.go`, `connection.go`, `secret.go`

A view owns the repositories of one entity keyed by their kind (`config.ConnType` for
connections, `secret.Scheme` for secrets), the order they were registered in, and which of
them is `Active()`. Its `index` carries one entry per record across *every* repository: a
`Ref[K]` (`{Kind, Index}` — the repository and the position inside it) plus the fields a
query matches against. `Reindex` rebuilds it after any mutation, and hands the fields to the
searcher as `document`s keyed by their position in `entries`.

The visible rows are `[]Ref[K]`: the active kind's records, in order, narrowed by the query
when one is standing. **The active kind is a filter of its own**, applied on both paths in
`refilter`, so a query never shows a row belonging to another kind and switching kind
re-matches the same query against the records of the kind switched to — the query survives
the switch, its result does not. Everything a screen addresses is indexed by *visible row*,
not by repository position; `RefAt` translates, and the `*For(ref)` accessors take a `Ref` for work
that outlives the current row order. `ConnRef` and `SecretRef` are the instantiated aliases
screens should spell.

`Connections.ClientInfo(ctx)` is the exception to "indexed by row": the client is a property of
the database, not of a record, so it is served by the active kind's repository and takes no
`Ref` — the visible rows are always the active kind's, so the cursor's database *is*
`Active()`. It returns the `cli.Info` the repository got from its launcher, which is the
only accessor here that can start a subprocess, and reports `false` when the active kind has
no repository.

## The search — `search.go`

**The matching algorithm lives behind `searcher` and nowhere else.** `index` owns the rows,
the query string and the cursor consequences; it knows nothing about *how* a query selects a
record. `search.go` holds that whole half:

```go
type document struct { id int; fields []string }

type searcher interface {
    Reindex(docs []document)
    Search(query string) (hits []int, filtered bool)
}
```

`Reindex` hands the searcher every record once, so an implementation may build whatever index
it wants and keep it — the default `substringSearcher` keeps its own case-folded copy of the
fields, so a keystroke costs one pass over that copy. `Search` returns the ids that match, in
index order, plus whether the query narrowed anything at all: `filtered == false` is what
tells `refilter` to fall back to "the active kind's records, in order", so what counts as an
empty query is the algorithm's call, not the index's. Swapping in fuzzy, prefix or scored
matching is one new type plus the line in `newSearcher`; `index.go` does not change, and
neither does any screen.

Field collection stays on the entity side (`connectionFields`, `secretFields` → `searchable`,
which only drops blanks); normalisation is the searcher's, since a different algorithm folds
differently.

**No password material is ever indexed.** The fields come from `config.Connection` and
`config.Secret`, and neither interface exposes the password or the reference standing in for
it (`config.Postgres.Password` is reachable only through `secret.Reference`), so the search
cannot leak what the tables already refuse to show. A new field added to `*Fields` must keep
that true.

## The cursor — `state.go`, `connection_state.go`, `secret_state.go`

`state.go` is the cursor, the delete confirmation and the help toggle, reached through
methods rather than fields (`MoveUp`/`MoveDown`/`Top`/`Bottom`/`SetCursor`,
`AskConfirm`/`ClearConfirm`/`Confirming`, `ToggleHelp`/`Help`, and the search prompt —
`StartSearch`/`CommitSearch`/`CancelSearch`/`Searching`, `AppendSearch`/`TrimSearch`,
`Query`/`Filtered`). Keeping it valid is the view's job: it clamps against its own length,
rides the same `Ref` through a `Search` that keeps it visible, drops to the top on a kind
switch, and re-clamps after every mutation.
Each view has a cursor-bound half — `Connection`/`Config`, `Secret`/`Usages`,
`Transport`/`Usages`, `Workspace`/`Current`, and the mutators *over the row under the cursor*
(`Add`/`Edit`/`Remove`; `Use`/`Rename` for a workspace) — so a screen never hands an index
back. The two larger views keep it in a `_state.go` file; `Transports` and `Workspaces` are
small enough to hold both halves in one. `Secrets` also owns the chooser mode
(`StartPicking`/`Pick`/`StopPicking`) that turns the secrets table into the form's picker.

What is left in `ui.Model` and `secretModel` is only what is genuinely theirs: the status
line, and the per-`Ref` ping state (spinner, result, error), which stays in `ui` because it
is made of `bubbles` and frame types the view layer must not import. A `Ref` survives a
re-filter but not a removal, so a removal that renumbers a repository drops that map
(`connChangedMsg.renumbered`).

`Searching()` is the input having focus, `Filtered()` is a query actually narrowing the rows
— a screen needs both: it draws the search input while either holds, paints it active only
for the first, and lets the second decide what cancelling means. Query editing is the view's too:
a screen appends the `tea.KeyPressMsg` text and never keeps a copy of the query.

`NextKind`/`PrevKind` are what the tables' kind switch calls; `SetActive` is how a caller
opens the view on a kind it already chose — the connections loop after the type picker.

`Transports` and `Workspaces` each sit over a single repository, so they have one kind and
nothing to switch. `Transports` is also a `transport.Provider` (`Resolve`), which is what
lets a form's preview ping route through a transport without holding the repository, and
`NamesOf(kind)` is the list a connection form offers.

## The window — `scroll.go`

`Scroll` is the other half of "where a screen is looking": which slice of the rows fits on
the terminal it was given. It counts rows and nothing else — it never names a connection, a
secret or a screen, and it holds no rows itself, so one lives next to whatever list needs a
window and takes the total on every call.

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
- `Follow` moves the window only when the cursor has left it, by exactly as much as it takes
  to bring it back to the edge it crossed; a cursor inside the window leaves `start` alone.
- Every call clamps against the `total` it is handed, so a list that shrinks under the
  window — rows removed, a query narrowing them — slides back into range instead of
  rendering past the end. Same for a `Resize` that grows the window past the tail.
- `Above`/`Below` are the hidden counts a screen renders as a "+n more" line; they are not
  scrollbar geometry, and the screen decides how to draw them.

**The window belongs to the view, not to the screen** — same rule as the cursor. A screen
sizes it from the `tea.WindowSizeMsg` and asks for the range; it does not keep its own
offset, slice the rows by hand or compute where the cursor went.

## The flow, end to end

```
NewX(repos…)      register(kind) per repo, first one becomes active   ──┐
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

Two invariants make the rest fall out: **a row number is a position in `visible`, never in a
repository**, and **every path that can change `visible` ends in `refilter` + a cursor
clamp**. That is why `Reindex` is called by the view itself after each mutation rather than
by the caller, and why a screen that wants a record asks for the one under the cursor.

## Adding a table for a new entity

`Connections` and `Secrets` are the two full examples and `Transports` the minimal one; a
new entity follows the same
shape and touches nothing generic. Given a repository surface for it (`Kind() K`, `Len`, a
positional accessor and the mutators — `K` being whatever names a store of that entity,
`comparable` and printable):

1. **`internal/ui/view/<entity>.go`** — the rows:
   - `type <Entity>Ref = Ref[K]`, the alias screens spell.
   - `type <Entities> struct { repos map[K]repository.<Entities>; tableState[K] }` — embed
     `tableState`, never `index`; that is what hides the cursor inside the view.
   - `New<Entities>(repos ...repository.<Entities>) *<Entities>` — skip a nil or duplicate
     repo with `v.register(repo.Kind())`, store it, finish with `v.Reindex()`.
   - `Reindex()` — walk `v.idx.order`, read each record, append
     `entry[K]{ref: <Entity>Ref{kind, i}, fields: <entity>Fields(rec)}`, then `v.reset(...)`.
   - `<entity>Fields(rec) []string` — everything a query should hit, through `searchable(…)`,
     which drops the empties; the searcher normalises them. Never pass a password or the
     reference standing in for one.
   - `CountOf(kind)`, the row accessors, and `rowRepo(action, i)` for the mutators. Name the
     accessors `At(i)`/`For(ref)` when the entity has one projection, as `Secrets` does; give
     each its own pair when it has several, the way `Connections` splits `ConnectionAt`/
     `ConnectionFor` (live) from `ConfigAt`/`ConfigFor` (stored).
2. **`internal/ui/view/<entity>_state.go`** — the cursor-bound half: the record under the
   cursor, `Add`/`Edit`/`Remove` (resolve through `rowRepo(action, v.Cursor())`, call the
   repository, `Reindex()`), and any screen mode that belongs to the entity rather than to
   the screen, the way `Secrets` owns picking.
3. **`ui.Run`** — build the view from the workspace's repositories and hand it to the screen.
4. **`internal/ui`** — the three/four-file screen split (`internal/ui/AGENTS.md`) over `*view.<Entities>`; the model
   keeps only its status line and its own async state, and gets a `keyhints()` method.
5. **`index.go` / `state.go` / `search.go` stay untouched.** If a new table genuinely needs
   a new cursor or index behaviour, add it there once, generically, and let both existing
   views have it —
   do not special-case an entity inside the shared core, and do not grow a private cursor in
   the screen.
