---
id: CONM-6
title: Attach named links to a connection as metadata resources
status: In Progress
assignee:
  - myjupyter
created_date: '2026-09-06 15:06'
updated_date: '2026-09-06 15:31'
labels:
  - ready-to-work
milestone: m-0
dependencies: []
documentation:
  - docs/forms.md
  - docs/config.md
  - docs/ui.md
  - dist/DB Connections TUI.dc.html
priority: medium
type: feature
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A connection is rarely the only thing a user needs to reach for a given database: the Grafana board, the log search, the runbook and the ticket queue all belong to the same server, and today they live in a browser bookmark folder that has no relation to the entry in conm. The connection record already carries user-owned metadata (name, description, tags) that is never sent to the server, and that is the natural home for those pointers — so that opening the right dashboard is one keystroke from the connection it describes, and so the links travel with the connection when it is edited, exported or read back.

The add/edit connection form is the only place a connection's metadata is authored, so the links have to be authored there too, as a list that can grow and shrink rather than a fixed set of fields.

Design reference for the form modal's framing and grouping: the terminal-spec section of `dist/DB Connections TUI.dc.html` (the metadata group is where the new Links header sits).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A connection carries an ordered list of links as metadata; each link has a name (e.g. `grafana`, `logs`) and a URL
- [ ] #2 Links round-trip through the connection's TOML file: saved on add/edit, read back on load, and absent from the file when a connection has no links
- [ ] #3 The add/edit connection form shows the links under a `Links` header inside the metadata group, listing each link's name and URL
- [ ] #4 With the Links field focused, `a` adds a new link and puts the user in editing it (name, then URL)
- [ ] #5 With a link focused, `d` removes that link from the list
- [ ] #6 With a link focused, `enter` opens its URL in the system browser, and a failure to open is reported in the status line rather than crashing or leaving the form
- [ ] #7 A link is rejected at save time if its name is empty or its URL is not a valid absolute URL, and the refusal keeps the form open with its values and the reason in the status line
- [ ] #8 Adding, removing and reordering links never mutates any other field of the connection, and links never appear in a connection string or in any displayed target
- [ ] #9 The links section follows the shared form/frame/theme/keybinding primitives — no raw rune, colour or key string in the screen — and its keys appear in the keyhint footer while it is focused
- [ ] #10 Every connection type supports links (they are metadata, not per-database configuration)
- [ ] #11 Table-driven tests with testify cover the round-trip through TOML, the validation refusals, and add/remove behaviour of the list
- [ ] #12 docs/forms.md and docs/config.md document the links field and the keys that operate on it
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## Approach

Links are metadata, so they go into the one struct every connection type already
embeds — `config.ConnMeta`. That covers AC#10 and most of AC#2 with no per-type work.

The form is the hard part: `formModel` walks the *static* `spec.Fields` / `Sections`.
A growable list does not fit a static spec, so the model gets its **own copy** of the
field list and the section key list, and `add`/`remove` mutate that copy. Every
existing mechanism — insert mode, `fieldError`, error counts, gutters, `values()` —
then works on link fields for free, including AC#7 (a failed `submit()` already keeps
the form open with its values and a status line).

Link values ride the existing `map[string]string` as `link.0.name` / `link.0.url`,
the same trick the design uses (`lname0`/`lurl0`). One place assembles and parses
that shape, mirroring `JoinSecret`/`SplitSecret`.

Skipped deliberately (say the word and they go in):
- **Windowing** the link list to 3 visible rows (design's `LINK_WINDOW`, `↑ n above`).
  Add when a real connection has enough links to overflow the frame.
- **Reordering** keys. AC#8 only asks that reordering not corrupt other fields;
  insertion order is the order, and no key in the design reorders.
- **Paired rendering** (`pairStart`/`pairEnd` bracketing). Two ordinary field rows
  labelled `link name` / `link url` read fine.

## Steps

### 1. `internal/config/conm.go` — the model (AC#1, #2, #10)

```go
type WebLink struct {
    Name string `toml:"name"`
    URL  string `toml:"url"`
}

type ConnMeta struct {
    Name        string    `toml:"name,omitempty"`
    Description string    `toml:"description,omitempty"`
    Tags        []string  `toml:"tags,omitempty"`
    Links       []WebLink `toml:"links,omitempty"`
}
```

- `omitempty` on `Links` is the whole of "absent from the file when a connection has
  no links".
- `ValidateWebLink(WebLink) error`: empty name rejected; URL parsed with `net/url`,
  requiring an absolute `http`/`https` URL with a host (AC#7). The type name earns
  itself — the validator says what it will accept.
- Add `Meta() ConnMeta` to the `config.Connection` **interface**, so seeding stays
  generic instead of six edited `SeedFunc`s.

  All six types already have this accessor, spelled `ConnMeta()` (`postgres.go:268`,
  `mysql.go:171`, `mssql.go:172`, `clickhouse.go:150`, `redis.go:167`,
  `mongodb.go:166`). Renaming it to `Meta()` collides with the field `Meta ConnMeta`
  each type declares (`postgres.go:61`) — Go forbids a field and a method of the same
  name — so the field is renamed too:

  - field `Meta` → **`Metadata`** on all six types, keeping the tag `toml:"meta"`.
    Nothing changes on disk: no migration, no round-trip churn beyond the new
    `links` key.
  - method `ConnMeta()` → `Meta()`, returning `p.Metadata`.
  - Mechanical sweep of ~54 `.Meta` field reads (`Name()`, `Description()`, `Tags()`,
    the six `SeedFunc`s, `repository/unique_test.go:34`) and 6 method declarations.
    `go build ./...` catches every miss.

  Rejected: embedding `ConnMeta` to dodge the collision. It would promote `Name`,
  `Description` and `Tags` into every type, where the existing same-named methods
  shadow them — legal, but exactly the kind of thing someone decodes at 3am.

### 2. `internal/ui/spec` — the key format (AC#1)

New `links.go`:
- `LinkNameKey(i)`, `LinkURLKey(i)`, `IsLinkKey(key)`, `LinkIndexOf(key)`.
- `SeedLinks([]config.WebLink, values)` — writes the pairs into a seed map.
- `LinksFrom(values) []config.WebLink` — reads them back, in index order, dropping
  pairs that are entirely empty.
- `LinkNameField(i)` / `LinkURLField(i)` build the two `FormField`s, both
  `OptionalFieldProperty`, with `ValidateFunc`s that call `config.ValidateWebLink`
  so refusals flow through the existing `fieldError` path.

`utils.go`: `buildMeta` sets `Links: LinksFrom(values)`. Because the keys are
type-independent, **no per-database spec file changes at all**.

### 3. `internal/ui/form_commands.go` — seeding and opening (AC#6)

- `seedConnection` calls `spec.SeedLinks(existing.Meta().Links, initial)` after
  `SeedFunc` — one place, all six types.
- `openURLCmd(url)`: `exec.Command` on `open` (darwin) / `xdg-open` (else),
  `Start()` only, result folded back as a `linkOpenedMsg` so a failure lands in the
  status line and the form stays up (AC#6). It lives here, not in `cli`, because
  `cli` owns *database client* binaries; this is screen-local I/O, which is what
  `*_commands.go` is for.

### 4. `internal/ui/form.go` — the growable list (AC#4, #5, #8)

- `formModel` gains `fields []spec.FormField` (copied from `spc.Fields` in
  `newFormModel`) and `linkSection int` (index of the section titled `metadata`).
  Replace every `m.spec.Fields` read with `m.fields` (~10 sites in `form.go`,
  `form_view.go`, `keyhint_table.go:249`).
- `newFormModel` rebuilds the link fields from whatever `link.N.*` keys the seed
  map carries, appending them to `m.fields` and to the metadata section's key list.
- `addLink()`: append a name/url pair, move the cursor to the new name field and
  enter insert mode → "type a name, enter for the url" (AC#4; `enter` in insert
  mode already advances to the url field).
- `removeLink(i)`: rebuild the pairs without `i`, renumber the keys, drop the stale
  entries from `m.vals`, clamp the cursor (AC#5). Both operate only on link keys —
  no other field is touched (AC#8).
- `navKey`: `keyMap.Add` when the metadata section is active → `addLink`;
  `keyMap.Delete` on a link field → `removeLink`; `keyMap.Confirm` on a link field
  → `openURLCmd` instead of `submit()`.

### 5. `internal/ui/form_view.go` + `keyhint_table.go` (AC#3, #9)

- Render loop emits a `Links` header line (existing `frameLine`/`cSoft`/`cFaint`
  spans, no new primitive) before the first link field, or in place of it with
  "nothing linked yet — a adds one" when the list is empty.
- `keyhints()`: on a link field add `a add link`, `d remove link`, `enter open in
  browser`, and in the metadata section without links just `a add link`. No raw
  rune, colour or key string anywhere — everything through `keyMap` and `theme.go`.

### 6. Tests — testify, table-driven (AC#11)

- `internal/config/conm_test.go` — `ValidateWebLink` table (empty name, relative URL,
  no scheme, no host, valid http/https); TOML round-trip of a `Postgres` with links,
  and the absence of a `links` key when the slice is empty.
- `internal/ui/spec/links_test.go` — `SeedLinks` → `LinksFrom` round-trip, ordering,
  empty-pair dropping, and `buildMeta` carrying links through.
- `internal/ui/form_links_test.go` — `addLink`/`removeLink`: field and section
  growth, key renumbering after a middle removal, cursor clamping, and that no
  non-link value in `m.vals` changes.

### 7. Docs (AC#12)

- `docs/forms.md` — new section: the `link.N.*` key format, that it is the only
  place the shape is assembled, and the `a`/`d`/`enter` keys.
- `docs/config.md` — `WebLink` under the shared-model section: what validates it,
  why `omitempty`, and that it never reaches `ConnectionString`. Note the
  `Metadata` field / `Meta()` accessor split while you are there.

## Checks

`go vet ./... && go test ./... && make lint`. AC#8's "never in a connection string"
needs no code — `ConnectionString` reads the transport fields, not the metadata — but
the config round-trip test asserts it.
<!-- SECTION:PLAN:END -->
