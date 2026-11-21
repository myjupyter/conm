# `internal/ui/spec` — declarative forms

## Purpose

What a form *is*, as data: one named spec per entity — its fields, sections, and the two
functions between a record and the flat values map a form edits. Also the form-side place the
composite value formats (secret reference, transport reference, link list, string array) are
assembled and taken apart. The form *drivers* live in `internal/ui`; the part of them that is
about specs is described here.

May import (internal): `config`, `secret`, `transport`, `utils/names`.

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
  both add and edit. `SeedFunc(T) map[key]value` projects a stored record into values: it
  type-asserts the concrete config type and returns nil on a mismatch, which the caller
  reports as an error rather than showing an add form. `BuildFunc(map[key]value) (T, error)`
  is the inverse.
- `FormField` — a key, a label, a kind, a property, options and a `ValidateFunc`; see below.
- `FormSection{Title, Fields []FormFieldKey}` — a tab: a title and the field keys it shows,
  so ordering on screen is a section concern, not a field one. The forms show the tab strip
  alone, with no prose under it.
- Registries: `spec.FormSpecs` maps `config.ConnType` → `FormSpec[config.Connection]` and is
  read by `cmd`, the forms and the info sheet; `ui/secret_form_commands.go`'s
  `secretFormSpecs` maps `secret.Scheme` → `FormSpec[config.Secret]`.

## How it works

Forms are data-driven: `formModel` and `specFormModel[T]` are generic drivers, and everything
entity-specific lives in `ui/spec`. **Every concrete entity gets its own named file and its
own named spec in `ui/spec` — `postgres.go` → `PostgresFormSpec`, `keyring.go` →
`KeyringFormSpec`, `transport.go` → `TransportFormSpec`.** Do not grow `spec.go`, which holds
only the vocabulary and the registry, with entity detail, and do not special-case an entity
inside a form model.

A per-entity spec file declares its own unexported field-key and example constants, the
ordered `Fields`, the `Sections`, and `SeedFunc`/`BuildFunc`. **Validators are not written
there** — a field points `ValidateFunc` at a `config.Validate*` function, so the form and
the persisted model can never disagree.

`FormField{Key, Label, Kind, Property, Example, DefaultValue, DefaultFunc, Options,
OptionsFunc, Source, TextFunc, ValidateFunc}`:

| `Kind` | Behaviour |
|---|---|
| `TextFieldKind` | the default |
| `HiddenFieldKind` | masked; whitespace is preserved, every other kind is trimmed |
| `IntFieldKind` | numeric |
| `SelectFieldKind` | cycled through `Options` |
| `SourceFieldKind` | cycled through a list that exists only at run time |
| `StringArrayFieldKind` | values edited as chips, stored joined by `StringArraySeparator`; `JoinStringArray`/`SplitStringArray` are the only spelling of that format |

- `Property` is `RequiredFieldProperty` (default) or `OptionalFieldProperty`.
- `DefaultValue` seeds an empty field; `DefaultFunc` does so with a computed value and only
  on add — the connection name opens with `names.Random()`.
- `OptionsFunc` replaces `Options` with a list computed from the form's current values — the
  ssh form narrows its secret providers to the ones its `auth` can use. After any select is
  cycled the connection form re-clamps every such field to its first option if its value is
  no longer offered; `Options` stays the union, which `SplitSecret` reads.
- `TextFunc` replaces the label and example with ones computed from the current values
  (`FieldText`) — the ssh secret value is a password, a key or a path depending on `auth`.
- **A `SourceFieldKind` field takes its options from the caller, not from the spec**, which
  cannot know which transports or SSH connections exist. The field names a source —
  `FieldSource{Name}`, or `NameFunc(values)` when the list depends on another field — and the
  screen that opens the form supplies `Sources`, a `map[name][]string`, built from its views.
  The spec stays free of `repository` and `ui/view`.

A `BuildFunc` folds its shared halves through `SplitStringArray`, `intValue` (an optional
numeric field, falling back to the type's default and naming the field in its error) and
`buildMeta` (the metadata keys plus the links → a `config.ConnMeta`). Spelling them out by
hand is duplication `dupl` fails on.

## The secret pair

`spec/secret.go` is the cross-entity piece any connection spec reuses: the
`SecretProviderKey`/`SecretValueKey` field pair (with the shared `secretProviderField` and
`passwordField` values), `SecretProvidersOrder`, and `JoinSecret`/`SplitSecret`, which fold
a provider plus a value into the single reference stored on the connection and back. They
are the form-side place for that format; the reference itself is spelled by
`secret.Ref`/`secret.ParseRef` (`internal/AGENTS.md`). A spec that omits the pair is
literal-only. With the pair present:

- In a store mode the value field is a picker over the store's entries, not free text:
  `Cycle` steps through them, `keyMap.Screen.Secrets` on the provider field opens the secrets
  screen as a chooser, and the value is validated against the entries that exist rather than
  by its own validator.
- Under `none` — a server that wants no password — the value field holds nothing: it is not
  typed into, not revealed and not validated, and the form keeps the literal and the store
  value stashed so cycling through `none` discards neither.

## The transport pair — how a connection names its route

`spec/transport.go` is to a transport what `spec/secret.go` is to a secret: the field pair a
database spec reuses and the one place its stored format is spelled. Both fields are
`SourceFieldKind`:

- `transportTypeField` (`TransportTypeKey`) picks the kind from `TransportTypesSource`. The
  caller builds that list (`ui/form_commands.go`): `TransportDirect` plus each of
  `config.TransportKinds` that has a saved transport.
- `transportField` (`TransportKey`) picks a transport of that kind by name; its `NameFunc`
  follows the first field, so the names offered are those of the kind chosen.

`TransportKindsOrder` is only the `Options` of the transport form's own kind select.
`JoinTransport`/`SplitTransport` fold the pair into the single `"<kind>:<name>"` reference
the connection stores and back; a direct connection stores nothing. A spec that omits the
pair cannot be routed — SSH itself is one.

## Pasting into a field

`keyMap.Paste` (`ui/keybinding.go`; `ctrl+v` by default, from `config/keybinding.go`) pastes
into the field under the cursor, in **every** form and every text field.
`ui/clipboard_commands.go` holds the message, the commands and the flattening; the clipboard
binaries are named in `internal/cli` (`internal/cli/AGENTS.md`). Two paths:

- A terminal that pastes on its own (`cmd+v` on macOS, `ctrl+shift+v` on most Linux
  terminals) sends bracketed paste, which bubbletea delivers as `tea.PasteMsg`.
- `keyMap.Paste` is a plain key press, so it runs `readClipboardCmd` over
  `cli.ReadClipboard` and comes back as a `clipboardMsg`. OSC52 is not used: too many
  terminals refuse it.

Both land in the model's `applyPaste(raw, err)`. `pasteText` flattens the payload first —
control characters become spaces and the ends are trimmed — so a trailing newline copied
from a browser never reaches a config file. `formModel.applyPaste` goes through `navEdit`,
so a field that cannot be typed into refuses a paste in the same words as it refuses
`keyMap.Form.EditField`; pasting into a field that is merely hovered starts editing it.

## Links: a list the user grows

The `meta` section is drawn as **two titled blocks** — `COMMON` then `LINKS`. `frameHead` in
`ui/frame.go` is the primitive: a blank line, the title in caps and bold on a banded row,
another blank line. Its note is optional and usually empty, since the keys are in the
footer; the one exception is an empty link list, where nothing else would say which key
fills it. A section with no blocks goes straight from the tab strip to its fields, so only
the `meta` tab changes shape.

`spec/links.go` is the cross-entity piece for the resource links a connection carries in its
metadata. The values map is flat `map[key]value`, so a list rides it as an indexed run of
key pairs — `link.0.name`, `link.0.url`, `link.1.name`, … — and **`spec/links.go` is the only
place that format is assembled or taken apart**:

- `LinkNameKey(i)`/`LinkURLKey(i)` spell a key and `LinkIndexOf` reads one back; `IsLinkKey`
  is that same parse. `LinkCount` reports how many a values map holds: the keys are a
  gapless run, so the first index without a name ends it.
- `SeedLinks(links, values)` writes a connection's links into a seed map;
  `LinksFrom(values)` reads them back in index order. `buildMeta` calls `LinksFrom`, so
  **every connection spec gets links without declaring a single field** — the keys are
  type-independent.
- `LinkNameField(i)`/`LinkURLField(i)` build the two `FormField`s a link occupies, labelled
  plainly `name` and `url`: the rail says which link the rows belong to.

The keys are indexed rather than declared, so the link fields cannot live in the
package-level spec. `formModel` holds **its own copy** of `Fields` and of the sections' key
lists (`m.fields`, `m.sections`), and `addLink`/`removeLink` edit that copy. The rest of the
driver — insert mode, `fieldError`, the per-section error count, the gutters, `values()` —
reads those copies and treats a link like any other field. The rows are positional and
identical, so `removeLink` shifts the *values* down over the removed one and drops the last
pair of fields; `linkBase`, the section's field count before any link was appended, is where
the run begins.

A link's two rows are bracketed by a rail — `┌` on the name, `└` on the url — and light up
together when the cursor is on either: the row under the cursor takes the full accent, its
sibling a dim wash (`cPairBg`), and the message row under each carries the wash too.
`formModel.linkPair(i)` answers whether a field is half of a link, which half, and whether
its link holds the cursor; the rail takes its two columns out of the label width, so every
input box starts in the same column.

The keys, while the metadata section is showing:

| Binding | Does |
|---|---|
| `keyMap.Form.AddItem` | appends a link and starts typing its name; `Confirm` carries the cursor on to the url without leaving insert mode |
| `keyMap.Form.RemoveItem` | removes the link under the cursor |
| `keyMap.Confirm` | opens the link under the cursor in the system browser — it does **not** submit the form |

`Confirm` is the one key a link takes away from the form, so `navLink` runs before the rest
of `navKey` and reports whether it consumed the key. Opening is a `tea.Cmd`
(`ui/form_commands.go`): the url is checked first and the browser is only started, never
waited for — `xdg-open` on some desktops does not return until the browser exits. The
refusal and whatever the browser reports land in the status line; the form never goes away
because a link would not open.

Validation is a property of the **pair**, so it cannot be a `ValidateFunc`, which sees one
value: `spec.ValidateLink` takes both and `formModel.linkError` calls it. A row with neither
half filled is dropped by `LinksFrom` and passed by `ValidateLink`. Fill in either half and
the other is required, and the url must be an absolute `http`/`https` url with a host.
