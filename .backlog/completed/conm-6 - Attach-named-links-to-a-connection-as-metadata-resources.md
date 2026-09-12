---
id: CONM-6
title: Attach named links to a connection as metadata resources
status: Done
assignee:
  - myjupyter
created_date: '2026-09-06 15:06'
updated_date: '2026-09-08 19:56'
labels:
  - ready-to-work
milestone: m-0
dependencies: []
documentation:
  - docs/forms.md
  - docs/config.md
  - docs/ui.md
  - dist/DB Connections TUI.dc.html
modified_files:
  - internal/config/conm.go
  - internal/config/config.go
  - internal/config/conm_test.go
  - internal/config/postgres.go
  - internal/config/mysql.go
  - internal/config/mssql.go
  - internal/config/clickhouse.go
  - internal/config/redis.go
  - internal/config/mongodb.go
  - internal/ui/spec/links.go
  - internal/ui/spec/links_test.go
  - internal/ui/spec/spec.go
  - internal/ui/spec/utils.go
  - internal/ui/spec/postgres.go
  - internal/ui/spec/mysql.go
  - internal/ui/spec/mssql.go
  - internal/ui/spec/clickhouse.go
  - internal/ui/spec/redis.go
  - internal/ui/spec/mongodb.go
  - internal/ui/form.go
  - internal/ui/form_view.go
  - internal/ui/form_commands.go
  - internal/ui/form_links_test.go
  - internal/ui/paste.go
  - internal/ui/paste_test.go
  - internal/ui/secret_form.go
  - internal/ui/keybinding.go
  - internal/ui/keyhint_table.go
  - internal/repository/unique_test.go
  - docs/config.md
  - docs/forms.md
  - .claude/skills/add-new-database-type/references/config.md
  - .claude/skills/add-new-database-type/references/spec.md
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
- [x] #1 A connection carries an ordered list of links as metadata; each link has a name (e.g. `grafana`, `logs`) and a URL
- [x] #2 Links round-trip through the connection's TOML file: saved on add/edit, read back on load, and absent from the file when a connection has no links
- [x] #3 The add/edit connection form shows the links under a `Links` header inside the metadata group, listing each link's name and URL
- [x] #4 With the Links field focused, `a` adds a new link and puts the user in editing it (name, then URL)
- [x] #5 With a link focused, `d` removes that link from the list
- [x] #6 With a link focused, `enter` opens its URL in the system browser, and a failure to open is reported in the status line rather than crashing or leaving the form
- [x] #7 A link is rejected at save time if its name is empty or its URL is not a valid absolute URL, and the refusal keeps the form open with its values and the reason in the status line
- [x] #8 Adding, removing and reordering links never mutates any other field of the connection, and links never appear in a connection string or in any displayed target
- [x] #9 The links section follows the shared form/frame/theme/keybinding primitives — no raw rune, colour or key string in the screen — and its keys appear in the keyhint footer while it is focused
- [x] #10 Every connection type supports links (they are metadata, not per-database configuration)
- [x] #11 Table-driven tests with testify cover the round-trip through TOML, the validation refusals, and add/remove behaviour of the list
- [x] #12 docs/forms.md and docs/config.md document the links field and the keys that operate on it
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

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented as planned. `go vet ./...`, `go test ./...` and `make lint` are all clean; `make build` produces bin/conm.

Two things the plan did not foresee:

- **Link validation could not be a `ValidateFunc`.** A link is only whole as a *pair*, and
  `FormField.ValidateFunc` is `func(string) error` — it cannot see the sibling half. The rule
  moved to `spec.ValidateLink(name, url)` (which owns the key format anyway) called from
  `formModel.linkError`, and the link fields carry no `ValidateFunc` at all. A row with
  neither half filled is dropped by `LinksFrom` and passed by `ValidateLink`, matching the
  design; fill either half and the other becomes required.
- **`navKey` broke the gocyclo ceiling** (27 > 20) once the three link keys were added as
  cases. They were extracted into `navLink(key, cur) (model, cmd, handled)`, the same shape
  as the existing `navSecret`, and run before the main switch because `enter` is the one key
  a link takes away from the form.

Smaller deviations:

- `insertKey` keeps insert mode when `enter` steps from a link's name to its url, so a new
  link is typed as one thing (AC#4's "name, then URL").
- `openURLCmd` starts the opener and reaps it in a goroutine rather than waiting: `xdg-open`
  on some desktops does not return until the browser it launched exits, but an unreaped
  child would be a zombie per link opened.
- `spec.metadataSectionTitle` had to be exported as `MetadataSectionTitle` for the model to
  find the section its links live in.
- The `add-new-database-type` skill's reference templates (`references/config.md`,
  `references/spec.md`) declared `Meta ConnMeta` and `func (m MySQL) ConnMeta()`, so a type
  copied from them would no longer compile after the rename. Both updated, per CLAUDE.md's
  rule that a seam change is not finished until the skill says so too.

Follow-up: **paste did not work in any field**, not just the link ones. Bracketed paste is on by default in bubbletea v2, so a terminal paste arrived as `tea.PasteMsg` and every form model silently dropped it — no form had ever handled the message. Fixed at the root rather than on the link fields.

`ui/paste.go` is the whole of it, and both input paths land in the same `applyPaste(raw, err)`:

- `tea.PasteMsg` — the terminal's own paste (`cmd+v` on macOS, `ctrl+shift+v` on Linux).
- `ctrl+v` (`keyMap.Paste`) — a plain key press, so it runs `readClipboardCmd`: `pbpaste` on
  darwin, else the first of `wl-paste` / `xclip` / `xsel` installed. OSC52 clipboard *read*
  (`tea.ReadClipboard`) was rejected — too many terminals refuse it, which would have made
  the key work only sometimes.

`pasteText` flattens the payload (control characters → spaces, ends trimmed) because a url
copied from a browser arrives with a trailing newline and a config file should never take a
raw one. `formModel.applyPaste` routes through `navEdit`, so a field that cannot be typed
into refuses the paste in the same words it refuses `e`, and pasting into a merely-hovered
field starts editing it.

Handled in `Update` before the insert/nav split in both `formModel` and `secretFormModel`,
which is also what kept `navKey` under the gocyclo ceiling. `initFormModel` has no text
field, so it needed nothing. `ctrl+v` is in the editing keyhints of both forms and
`docs/forms.md` has a section on it. Verified against the real clipboard on darwin.

Caveat for the user: pressing ctrl+v in a terminal that maps it to something else (literal-next) still reaches conm as a key press, so this works — but a terminal that swallows ctrl+v entirely would not. cmd+v keeps working through the `tea.PasteMsg` path either way.

Follow-up: the metadata section's headers did not match the design. The group note and my
links line were both drawn as bare faint text, where `dist/DB Connections TUI.dc.html` draws
the metadata tab as **two titled blocks** via its `head(title, note)` — a blank line, the
title uppercased and bold on a `#1a1912` band with the note beside it in ghost, then another
blank line.

Added `frameHead(w, title, note) []string` to `ui/frame.go` (and `cHeadBg` to `theme.go`), a
shared primitive rather than anything link-specific. The metadata tab now renders `COMMON`
with the section note and `LINKS` with its own; `linkHeadLines` is a one-line call into it.
Sections without blocks keep the plain `noteLine`, so only the metadata tab changes shape.

One deliberate deviation from the design: it emits the group note **twice** on that tab —
once under the tab divider (outside its entries loop) and again as the `common` head's note,
the same string two lines apart. Here the standalone note line is suppressed when the
section is drawn as blocks, so it reads once, in the header. Say the word if the literal
duplication is wanted.

Follow-up: the link rows are now drawn as the design's **pair**, which the plan had listed
as a deliberate skip.

- Labels are plainly `name` and `url` (`linkNameLabel`/`linkURLLabel` in `spec/links.go`),
  not `link 1 name` / `link 1 url`.
- A rail brackets the two rows — `gCornerTL` on the name, `gCornerBL` on the url — and takes
  its two columns out of the label width (`railW`), so every input box still starts in the
  same column as on the other tab.
- Both rows light up together: the one under the cursor takes the full accent, its sibling a
  dim wash (`cPairBg`), and the message row under each carries the wash too, so the pair
  reads as one block. The rail itself goes accent when the pair holds the cursor, `cRail`
  otherwise.
- `formModel.linkPair(i)` is the single query behind all of it — is this field half of a
  link, which half, does its link hold the cursor — and `rowBg`/`railGlyph`/`railColor` are
  the small helpers `fieldRow` and `messageRow` share.

Still not done from the design, and worth a look: the url field **clips with a trailing
ellipsis while you type into it**, so a long url is edited blind past column 44. The design
shows the tail with a leading `…` whenever the field is in insert mode. Small change to
`inputBox`, not made because it was not asked for.

Follow-up: trimmed the header copy and renamed the tab.

- `frameHead`'s note is now optional. `COMMON` and `LINKS` carry no note — the keys are
  already in the footer and "yours — never sent to the server" restated what the tab name
  says.
- Kept: the `LINKS` header still reads `nothing linked yet — a adds one` **when the list is
  empty**, because with no rows under it the block is otherwise blank and nothing names the
  key that fills it. It disappears as soon as a link exists.
- Kept: the `connection` tab's `how conm reaches the server` note line, which was not part
  of the ask, and the section note in the status line on tab switch — a different surface
  from the header.
- `spec.MetadataSectionTitle` is now `"meta"`, so the tab bar reads `connection  meta`. The
  constant keeps its name; only the displayed value changed, and `linkSection()` matches on
  the same constant so nothing else moved.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Connections now carry an ordered list of named web links in their metadata, authored in the add/edit form.

## The model

`config.WebLink{Name, URL}` and `ConnMeta.Links []WebLink` (`toml:"links,omitempty"`). Because every connection type already embeds `ConnMeta`, **all six types got links without one line of per-type code** — and a connection with no links writes no `links` key at all:

```toml
[postgres.meta]
name = 'prod-primary'

[[postgres.meta.links]]
name = 'grafana'
url = 'https://grafana.internal/d/pg'
```

`ValidateWebLinkName`/`ValidateWebLinkURL` reject a blank name and anything that is not an absolute `http`/`https` url with a host.

`config.Connection` gained `Meta() ConnMeta`, renamed from the existing `ConnMeta()` on all six types. That forced the field `Meta ConnMeta` to become `Metadata` (Go forbids a type having both under one name); the `toml:"meta"` tag is unchanged, so nothing changed on disk.

## The form

The values map is flat `map[string]string`, so the list rides it as an indexed run — `link.0.name`, `link.0.url`, … — and `internal/ui/spec/links.go` is the only place that format is assembled or taken apart, the way `JoinSecret`/`SplitSecret` is for a secret reference. `buildMeta` calls `LinksFrom`, so the keys being type-independent is what makes the per-type cost zero.

Indexed keys cannot live in a package-level spec, so `formModel` holds its own copy of `Fields` and the section key lists, and `addLink`/`removeLink` edit that copy. Insert mode, `fieldError`, the per-section error badge, the gutters and `values()` all read those copies and never learn a link is different from any other field.

Keys, while the metadata section shows: `a` appends a link and starts typing its name (`enter` carries on to the url without leaving insert mode), `d` removes the one under the cursor, `enter` opens it in the system browser instead of submitting. Opening validates first, only starts the opener, and reports both the refusal and whatever the browser says in the status line — the form never goes away because a link would not open.

## Verification

`go vet ./...`, `go test ./...`, `make lint` (0 issues), `make build` all clean. New table-driven tests cover the validators, the TOML round-trip including the absent-when-empty case and that no link reaches `ConnectionString`, the `SeedLinks`/`LinksFrom` round-trip, and add/remove including renumbering after a middle removal, cursor clamping and that no other field is touched. The rendered form was inspected in both the populated and empty states.

Skipped, as planned: the design's 3-row windowing of the link list, reorder keys, and paired row bracketing.
<!-- SECTION:FINAL_SUMMARY:END -->
