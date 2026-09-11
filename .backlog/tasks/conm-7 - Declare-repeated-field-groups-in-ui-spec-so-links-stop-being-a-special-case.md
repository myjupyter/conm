---
id: CONM-7
title: Declare repeated field groups in ui/spec so links stop being a special case
status: To Do
assignee: []
created_date: '2026-09-08 19:48'
labels:
  - tech-debt
dependencies:
  - CONM-6
documentation:
  - docs/forms.md
  - docs/ui.md
  - CLAUDE.md
  - internal/ui/spec/links.go
type: enhancement
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A form's fields are declared statically in `internal/ui/spec`, one `FormField` per key, and the form driver walks that list. A list the user grows at runtime cannot be expressed that way, so adding links to a connection carved out an exception instead: `ui/spec/links.go` spells and parses an indexed key format, manufactures `FormField`s while the form is open, and exports the queries the form model uses to ask whether the field under its cursor is a link.

The cost is a layering inversion. Nine of that file's ten exported symbols are consumed by `internal/ui`, so form-driver logic lives in the spec package, and screens that should not care — the renderer, the keyhint footer, the key handler — all know the word "link". `spec` is meant to declare what a form holds; it now also owns one entity's runtime list mechanics.

Links will not be the last list. The `add-new-database-type` fit check already records cluster host lists (ClickHouse, MongoDB replica sets) as a shape the shared connection model cannot express, and a client's key/value connection parameters are the same shape. Under the current design each one copies the links exception, and the second copy is where this stops being affordable.

Worth recording so the shape is not mistaken: nesting is not the missing piece. A nested struct is already expressible as sibling fields — the secret provider/value pair does exactly that with no machinery. Repetition is the thing a spec cannot declare. The two known future shapes differ (a repeated pair for links, a repeated single value for host lists), which is why this was deferred rather than designed from links alone; whichever lands second should be built alongside this so the abstraction is fitted to more than one example.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A form spec declares a repeated group of fields once; how many instances exist is decided at runtime, not in the spec
- [ ] #2 The form model expands, adds and removes instances of a declared group without naming any particular group
- [ ] #3 No identifier in internal/ui is named after links; the only link-specific code left is the mapping between a group's values and config.WebLink
- [ ] #4 The mechanism covers a repeated group of more than one field and a repeated single field, so it is not fitted to the links shape alone
- [ ] #5 Links behave exactly as before: order, TOML round-trip including no key when empty, the add/remove/open keys, pair validation, and a refusal keeping the form open with its values
- [ ] #6 Table-driven testify tests cover group expansion, add/remove with renumbering, and the values-map round trip, written against the generic mechanism rather than against links
- [ ] #7 docs/forms.md, docs/ui.md and CLAUDE.md's internal/ui/spec row describe repeated groups, with links demoted to an example
<!-- AC:END -->
