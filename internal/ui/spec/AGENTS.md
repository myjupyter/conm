# `internal/ui/spec` — declarative forms

## Purpose

What a form *is*, as data: one named spec per entity — its fields, sections, and the two
functions between a record and the flat values map a form edits. Also the only place the
composite value formats (secret reference, transport reference, link list, string array) are
assembled and taken apart. The form *drivers* live in `internal/ui`; the part of them that is
about specs is described here.

May import (internal): `config`, `cli`, `secret`, `transport`, `utils/names`.

## Files

| File | Holds |
|---|---|
| `spec.go` | the vocabulary — `FormField`, `FormSection`, `FormSpec[T]`, `FieldKind`, `FieldProperty`, `Sources` — and the registry `FormSpecs` |
| `postgres.go`, `mysql.go`, `mssql.go`, `clickhouse.go`, `redis.go`, `mongodb.go`, `ssh.go` | one connection spec each (`PostgresFormSpec`, …) |
| `keyring.go` | `KeyringFormSpec` — the secret record |
| `transport.go` | `TransportFormSpec`, the transport field pair a connection spec reuses, `JoinTransport`/`SplitTransport`, `TransportKindsOrder` |
| `workspace.go` | `WorkspaceFormSpec` |
| `secret.go` | the secret field pair, `SecretProvidersOrder`, `JoinSecret`/`SplitSecret` |
| `links.go` | the link list: `LinkNameKey`/`LinkURLKey`, `SeedLinks`/`LinksFrom`, `ValidateLink` |
| `utils.go` | `metadataSection`, `buildMeta`, `intValue`, `JoinStringArray`/`SplitStringArray` |

## Main entities

- `FormSpec[T]{AddTitle, EditTitle, Fields, Sections, BuildFunc, SeedFunc}` — one spec drives
  both add and edit. `SeedFunc(T)` projects a stored record into values; `BuildFunc(values)`
  is the inverse.
- `FormField` — a key, a label, a kind (`TextFieldKind`, `HiddenFieldKind`, `IntFieldKind`,
  `SelectFieldKind`, `SourceFieldKind`, `StringArrayFieldKind`), a property
  (required/optional), options and a `ValidateFunc` pointing
  at a `config.Validate*` function.
- `FormSection` — a tab: a title and the field keys it shows, in order.
- `FormSpecs` — `config.ConnType` → `FormSpec[config.Connection]`, the registry `cmd`, the
  forms and the info sheet read.

## How it works

Forms are data-driven: `formModel` and `specFormModel[T]` are generic drivers, and everything
entity-specific lives in `ui/spec`. **Every concrete entity gets its own named file and its
own named spec in `ui/spec` — `postgres.go` → `PostgresFormSpec`, `keyring.go` →
`KeyringFormSpec`, `transport.go` → `TransportFormSpec`.** Do not grow `spec.go` with entity detail, and do not special-case an
entity inside a form model.

`spec.go` holds only the vocabulary and the registry:

- `FormField{Key, Label, Kind, Property, Example, DefaultValue, DefaultFunc, Options,
  OptionsFunc, Source, TextFunc, ValidateFunc}`.
  - `Kind` is `TextFieldKind` (default), `HiddenFieldKind` (masked; whitespace is preserved,
    everything else is trimmed), `IntFieldKind`, `SelectFieldKind` (cycled through
    `Options`), `SourceFieldKind` (cycled through a list that exists only at run time — see
    below) or `StringArrayFieldKind` (a list of values edited as chips and stored joined by
    `StringArraySeparator`; `JoinStringArray`/`SplitStringArray` are the only spelling of
    that format).
  - `Property` is `RequiredFieldProperty` (default) or `OptionalFieldProperty`.
  - `DefaultValue` seeds an empty field; `DefaultFunc` does the same with a computed value
    and only on add — the connection name opens with `names.Random()`.
  - `OptionsFunc`, when set, replaces `Options` with a list computed from the form's current
    values — the ssh form narrows its secret providers to the ones its `auth` can use. After
    any select is cycled the connection form re-clamps every such field to its first option
    if the value it held is no longer offered, so a dependent select never shows a stale
    choice; `Options` stays the union, which `SplitSecret` reads.
  - `TextFunc`, when set, replaces the label and example with ones computed from the current
    values (`FieldText`) — the ssh secret value is a password, a key or a path depending on
    `auth`.
- **A `SourceFieldKind` field takes its options from the caller, not from the spec.** A spec
  cannot know which transports or SSH connections exist, so the field names a source —
  `FieldSource{Name}`, or `NameFunc(values)` when the list depends on another field — and the
  screen that opens the form supplies `Sources`, a `map[name][]string`, built from its views.
  The spec stays free of `repository` and `ui/view`.
- `FormSection{Title, Fields []FormFieldKey}` — a tab; fields are referenced by key,
  so ordering on screen is a section concern, not a field one. The forms show the tab strip
  alone, because a line of prose under it says nothing the fields and the footer do not.
- `FormSpec[T]{AddTitle, EditTitle, Fields, Sections, BuildFunc, SeedFunc}` — one spec drives
  both add and edit. `SeedFunc(T) map[key]value` type-asserts the concrete config type and
  returns nil if it does not match (the caller reports that as an error rather than silently
  showing an add form); `BuildFunc(map[key]value) (T, error)` is the inverse.
- The helpers a `BuildFunc` folds its shared halves through: `SplitStringArray`, `intValue` (an
  optional numeric field, falling back to the type's default and naming the field in its
  error) and `buildMeta` (the metadata keys plus the links → a `config.ConnMeta`). Every spec uses
  them; one that spells them out by hand instead is duplication `dupl` will fail on.
- Registries: `spec.FormSpecs` maps `config.ConnType` → `FormSpec[config.Connection]`;
  `ui/secret_form_commands.go`'s `secretFormSpecs` maps `secret.Scheme` →
  `FormSpec[config.Secret]`.

A per-entity spec file declares its own unexported field-key and example constants, the
ordered `Fields`, the `Sections`, and `SeedFunc`/`BuildFunc`. **Validators are not written
there** — a field points `ValidateFunc` at a `config.Validate*` function, so the form and
the persisted model can never disagree.

`spec/secret.go` is the cross-entity piece any connection spec reuses: the
`SecretProviderKey`/`SecretValueKey` field pair (with the shared `secretProviderField` and
`passwordField` values), `SecretProvidersOrder`, and `JoinSecret`/`SplitSecret`, which fold
a provider plus a value into the single reference stored on the connection and back. A spec
that omits the pair is literal-only. When the pair is present, the value field stops being
free text in a store mode and becomes a picker over the store's entries: `Cycle` steps through
them, `keyMap.Screen.Secrets` on the provider field opens the secrets screen as a chooser, and the
value is validated against the entries that exist rather than by its own validator. Under
`none` — the mode for a server that wants no password at all — the value field holds
nothing: it is not typed into, not revealed and not validated, and the form keeps the
literal and the store value stashed so cycling through `none` does not discard either.

## The transport pair — how a connection names its route

`spec/transport.go` is to a transport what `spec/secret.go` is to a secret: the field pair a
database spec reuses and the one place its stored format is spelled. `transportTypeField`
(`TransportTypeKey`) picks the kind — `TransportDirect` or one of `TransportKindsOrder` — and
`transportField` (`TransportKey`) picks a transport of that kind by name; both are
`SourceFieldKind`, the second with a `NameFunc` that follows the first, so the names offered
are always those of the kind chosen. `JoinTransport`/`SplitTransport` fold the pair into the
single `"<kind>:<name>"` reference the connection stores and back; a direct connection
stores nothing. A spec that omits the pair cannot be routed — SSH itself is one.

## Pasting into a field

`keyMap.Paste` (`ctrl+v` by default) pastes into the field under the cursor, in **every** form
and every text field —
`ui/clipboard_commands.go` holds the whole of it — the key, the message and the flattening; the
clipboard binaries themselves live in `internal/cli`. Two paths arrive at the same place:

- A terminal that pastes on its own (`cmd+v` on macOS, `ctrl+shift+v` on most Linux
  terminals) sends bracketed paste, which bubbletea delivers as `tea.PasteMsg`.
- `keyMap.Paste` is a plain key press, so it runs `readClipboardCmd` over `cli.ReadClipboard`,
  which is where the binaries are named (`internal/cli/AGENTS.md`), and comes back as a `clipboardMsg`. Reading the clipboard over OSC52 was not used: too many terminals refuse
  it, which would have made the key work only sometimes.

Both land in the model's `applyPaste(raw, err)`. `pasteText` flattens the payload first —
control characters become spaces and the ends are trimmed — because a url copied from a
browser arrives with a trailing newline and a config file should never receive a raw one.
`formModel.applyPaste` goes through `navEdit`, so a field that cannot be typed into refuses
a paste for the same reason and in the same words that it refuses `keyMap.Form.EditField`;
pasting into a field
that is merely hovered starts editing it.

## Links: a list the user grows

The `meta` section is drawn as **two titled blocks** — `COMMON` then `LINKS` — rather than
one undifferentiated pile. `frameHead` in `ui/frame.go` is the primitive: a blank line, the
title in caps and bold on a banded row, then another blank line. Its note is optional and
usually empty: the keys are already in the footer, and a header that repeats them is noise.
The one exception is an empty link list, where the block would otherwise be blank and
nothing would say which key fills it. A section with no blocks goes straight from the tab
strip to its fields, so only the `meta` tab changes shape.

`spec/links.go` is the cross-entity piece for the resource links a connection carries in its
metadata. The values map is flat `map[key]value`, so a list rides it as an indexed run of
key pairs — `link.0.name`, `link.0.url`, `link.1.name`, … — and **`spec/links.go` is the only
place that format is assembled or taken apart**, the way `JoinSecret`/`SplitSecret` is for a
secret reference:

- `LinkNameKey(i)`/`LinkURLKey(i)` spell a key and `LinkIndexOf` reads one back — `IsLinkKey`
  is that same parse, so the two can never disagree about what a link key is. `LinkCount`
  reports how many a values map holds: the keys are a gapless run, so the first index
  without a name is the end of it.
- `SeedLinks(links, values)` writes a connection's links into a seed map;
  `LinksFrom(values)` reads them back in index order. `buildMeta` calls `LinksFrom`, so
  **every connection spec gets links without declaring a single field** — the keys are
  type-independent.
- `LinkNameField(i)`/`LinkURLField(i)` build the two `FormField`s a link occupies. They are
  labelled plainly `name` and `url`: a link is drawn as one object, so the rail says which
  link the rows belong to and the labels do not have to repeat it.

Because the keys are indexed rather than declared, the link fields cannot live in the
package-level spec. `formModel` therefore holds **its own copy** of `Fields` and of the
sections' key lists (`m.fields`, `m.sections`), and `addLink`/`removeLink` edit that copy.
Everything else in the driver — insert mode, `fieldError`, the per-section error count, the
gutters, `values()` — reads those copies and never learns that a link is different from any
other field. The rows themselves are positional and identical, so `removeLink` shifts the
*values* down over the removed one and drops the last pair of fields; `linkBase`, the
section's field count before any link was appended, is where the run begins.

A link's two rows are bracketed by a rail — `┌` on the name, `└` on the url — and both
light up together when the cursor is on either of them: the row under the cursor takes the
full accent, its sibling a dim wash (`cPairBg`), and the message row under each carries the
wash too so the pair reads as one block. `formModel.linkPair(i)` answers whether a field is
half of a link, which half, and whether its link holds the cursor; the rail takes its two
columns out of the label width, so every input box still starts in the same column.

The keys, while the metadata section is showing:

| Binding | Does |
|---|---|
| `keyMap.Form.AddItem` | appends a link and starts typing its name; `Confirm` carries the cursor on to the url without leaving insert mode |
| `keyMap.Form.RemoveItem` | removes the link under the cursor |
| `keyMap.Confirm` | opens the link under the cursor in the system browser — it does **not** submit the form |

`Confirm` is the one key a link takes away from the form, so `navLink` runs before the rest of
`navKey` and reports whether it consumed the key. Opening is a `tea.Cmd`
(`ui/form_commands.go`): the url is checked first, the browser is only started — never
waited for, because `xdg-open` on some desktops does not return until the browser it
launched exits — and both the
refusal and whatever the browser reports land in the status line — the form never goes away
because a link would not open.

Validation is a property of the **pair**, not of either half, so it cannot be a `ValidateFunc`
(which sees one value): `spec.ValidateLink` takes both and `formModel.linkError` calls it. A
row with neither half filled is one the user added and left alone — `LinksFrom` drops it and
`ValidateLink` passes it. Fill in either half and the other is then required, and the url
must be an absolute `http`/`https` url with a host.
