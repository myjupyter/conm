---
id: CONM-5
title: Add a read-only info screen for a connection
status: Done
assignee:
  - myjupyter
created_date: '2026-09-06 14:30'
updated_date: '2026-09-12 14:08'
labels:
  - ready-to-work
milestone: m-0
dependencies:
  - CONM-6
  - CONM-9
  - CONM-10
  - CONM-11
documentation:
  - docs/ui.md
  - docs/view.md
  - docs/forms.md
  - dist/DB Connections TUI.dc.html
modified_files:
  - internal/ui/info.go
  - internal/ui/info_view.go
  - internal/ui/info_commands.go
  - internal/ui/info_test.go
  - internal/ui/keybinding.go
  - internal/ui/keyhint_table.go
  - internal/ui/paste.go
  - internal/ui/table.go
  - internal/ui/table_commands.go
  - internal/ui/theme.go
  - docs/ui.md
priority: low
type: feature
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A connection's details are only visible by opening the edit form, one keystroke away from mutating the record just to look at it. A read-only info screen gives a safe place to inspect a connection, with mutation behind an explicit `e` jump to the edit form.

Up to three tabs:

- **general** — meta: name, description, tags; CLI: cli (binary name), version, path.
- **endpoint** — host, port, database, username, sslmode, dsn (password masked as `•••`).
- **links** — one row per resource link, its name (grafana, logs, runbook…) and full url; `enter` opens the highlighted one. Shown only when the connection has links; an unlinked connection cycles between the other two.

No secret tab, no password displayed or resolved.

## Focus

- Info opens with no cursor — plain sheet until you reach for it.
- First `j`/`k` creates it: `j` on the first field, `k` on the last. Steps field to field, wraps, skips section headers.
- Focused row gets the accent background, a `❯` marker, and a hint of what it accepts (`yy yanks`, `enter opens · yy yanks`).
- `yy` yanks the value; `enter` opens a url or connects; `y` without a cursor asks you to move it first.
- `tab`/`←→` switches page and drops the cursor; `esc` clears it, second `esc` leaves info.

## Scroll

- Only the links page scrolls; it renders at most 8 rows.
- The cursor drags the window: moving past the last visible link shifts it by one, same upward.
- The section header counts what is hidden — `↑ 2 above · ↓ 1 below`; no arrows, no scrollbar.
- Switching page resets both cursor and window to the top.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 The screen has up to three tabs — general, endpoint and, only when the connection has links, links — with the active tab visibly marked
- [x] #2 A connection with no links shows no links tab: the tab strip, the tab/←→ cycle and the footer's page list all read the same list of shown pages
- [x] #3 The general tab shows a meta block (name, description, tags) and a CLI block (cli binary name, version, path)
- [x] #4 The endpoint tab shows host, port, database, username, sslmode and the dsn, with the password rendered as `•••`
- [x] #5 The links tab shows one row per resource link with its name and full url, and `enter` opens the highlighted link
- [x] #6 No password material, password reference, secret provider or secret location appears anywhere on the screen
- [x] #7 No key on the screen edits, deletes or otherwise mutates the connection or its secret
- [x] #8 Pressing `e` leaves the info screen and opens the edit form for the same connection
- [x] #9 The keyhint footer lists the screen's keys and follows the shared footer/theme/keybinding primitives, with no raw rune, colour or key string in the screen
- [x] #10 The screen reads entities through `ui/view` only, with no file I/O and no hand-built network connection
- [x] #11 docs/ui.md documents the new screen, its tabs and its place in the screen set
- [x] #12 Pressing `yy` with a field focused copies that field's displayed value to the clipboard and confirms it in the status line; only displayed values are copyable
- [x] #13 The screen's layout, framing and colours follow the terminal-spec section of `dist/DB Connections TUI.dc.html`, expressed through the existing `ui/frame.go` and `ui/theme.go` primitives
- [x] #14 Pressing `i` on the connections table opens the info screen for the selected connection; `i` is added to the shared keymap as an Info binding
- [x] #15 `q`, `i` and `esc` with no cursor return to the connections table with the same row still selected
- [x] #16 The screen opens with no cursor; the first `j`/`k` creates it on the first/last field respectively
- [x] #17 The cursor steps field to field, wraps at both ends and skips section headers
- [x] #18 The focused row is marked with the accent background, a `❯` marker and a hint of what it accepts
- [x] #19 `yy` yanks the focused value, `enter` opens a url or connects, and `y` with no cursor asks the user to move it first
- [x] #20 `tab` and `←`/`→` switch page and drop the cursor; `esc` clears the cursor and a second `esc` leaves the info screen
- [x] #21 Only the links page scrolls, rendering at most 8 rows, and the window shifts by one as the cursor moves past either edge
- [x] #22 The links section header counts what is hidden as `↑ N above · ↓ N below`, with no arrows or scrollbar
- [x] #23 Switching page resets both the cursor and the scroll window to the top
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
## 1. Inputs from their own tasks

- `view.Scroll` — the links window (CONM-11).
- `view.Connections.Client()` → `cli.Info` (CONM-10) and the version probe (CONM-9), the probe run as a `tea.Cmd` and rendered as unknown until it answers.

## 2. Rows — no per-database branch

Built from `spec.FormSpecs[kind]` (`SeedFunc` values + field labels):

- **general** — the metadata section's fields, then `cli` / `version` / `path`.
- **endpoint** — the connection section's fields minus `SecretProviderKey`/`SecretValueKey`, then `dsn` = `cfg.ConnectionString(strings.Repeat(gInputMask, 3))`.
- **links** — `cfg.Meta().Links`.

Row: `{label, value, tone, head, link bool}`.

## 3. Screen — `ui/info.go`, `ui/info_view.go`, `ui/info_commands.go`

- State: `tab`, `cursor *int` (nil until first `j`/`k`), `view.Scroll{size: 8}` for links, pending `y`.
- Keys: `keyMap.Info` (`i`), `keyMap.Yank` (`y`); `j`/`k` wrap and skip heads; `tab`/`←→` switch tab and reset cursor + window; `esc` clears the cursor, else leaves; `e` closes with an edit flag; `enter` opens a link row (`open`/`xdg-open`) else connects; `p` pings.
- Entry: table's `i` → `tea.Exec` nested program (`runInfo`, like `runSecretTable`); `infoClosedMsg{edit}` chains the existing `editCmd`.
- Render: `frameTop` crumb `connection info`, `frameBadge`, tab line, `frameHead` per section, wrapped value lines, cursor row on `cAccent`/`cInvFg` with `gCaret`; hints via `keyhint_table.go`.

## 4. Yank — `ui/paste.go`

Add the write side beside the readers: `pbcopy` / `wl-copy` / `xclip -i` / `xsel -b -i`; status `yanked <label>`.

## 5. Docs and tests

- `docs/ui.md` — the screen, its tabs and its place in the set.
- `ui/info_test.go` — lazy cursor wraps and skips heads; rows never carry a secret key or password.

## Deviation from the plan

- **No `p` ping on the info screen.** No acceptance criterion asks for it and it would pull `network`, a spinner and an error panel onto a read-only sheet. `enter` off a link still connects (AC #18); pinging stays on the table.
- Rows are built unwindowed and the links page marks out-of-window rows `hidden`, so the cursor keeps counting links it cannot see; `view.Scroll` decides the window and the header's `↑ N above · ↓ N below` note.
- General/endpoint rows come from `spec.FormSpecs[kind]`: the metadata section is the general page, every other section is the endpoint page, and `SecretProviderKey`/`SecretValueKey` are dropped at that one point.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Screen lives in `ui/info.go` / `ui/info_view.go` / `ui/info_commands.go`, opened from the table's `i` as a nested program behind `tea.Exec` (`runInfo`) and reporting only whether the user left through `e`; `Model.applyInfoClosed` turns that into the table's existing `editCmd`, so the write still belongs to the screen that owns it.

Two things the render forced:

- **The dsn mask had to go in unescaped.** `ConnectionString` builds through `net/url`, which percent-encodes what it is handed, so `strings.Repeat(gInputMask, 3)` came out as `%E2%80%A2%E2%80%A2%E2%80%A2`. `infoDSN` now passes an ASCII placeholder and swaps it for the mask afterwards, giving `postgresql://svc_api:•••@host:5432/db?sslmode=prefer`.
- **Labels are lowercased spec labels.** `specRows` takes `strings.ToLower(FormField.Label)` so the sheet reads `host` / `sslmode` the way the design spells it, without a second label table to keep in step.

Rendered all three pages against the terminal-spec section of `dist/DB Connections TUI.dc.html` before finishing: crumb, title line, tab row, banded section heads, accent cursor row with `❯` and the right-pinned accepts-hint, and the `↓ 4 below` count on a twelve-link page all match.

No comments in the new code per CLAUDE.md; the reasoning above lives in `docs/ui.md`.

`go build ./... && go vet ./... && go test ./... && make lint` all clean.
<!-- SECTION:NOTES:END -->

## Comments

<!-- COMMENTS:BEGIN -->
author: claude
created: 2026-09-12 14:08
---
Follow-up after the sheet shipped: a connection with no links no longer shows a links tab. `infoModel.tabs()` is now the single list of shown pages — the tab strip, the `tab`/`←→` cycle and the footer's page list all read it — so an unlinked connection cycles between general and endpoint and never lands on an empty sheet. AC #1 was reworded and AC #2 added for it; `linkRows` lost its empty-list branch, and `TestInfoHidesTheLinksTabWithoutLinks` covers both directions.
---
<!-- COMMENTS:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Adds the read-only connection info sheet (`i` from the connections table) with three pages — general, endpoint, links — a cursor that does not exist until `j`/`k` reaches for it, `yy` yanking the focused value, and `enter` opening a link or connecting. `e` is the only way to mutation, and it leaves for the table's edit form rather than editing in place.

- **Pages are not per-database.** `specRows` reads them off `spec.FormSpecs[kind]`: the metadata section is `general`, every other section is `endpoint`, and `SecretProviderKey`/`SecretValueKey` are dropped at that one point — so a new database type appears here with no change and there is no second place the sheet could start naming a password.
- **Only the links page scrolls**, through the `view.Scroll` from CONM-11: at most eight rows, the cursor dragging the window, and the section head counting what is out of sight as `↑ N above · ↓ N below`. Rows outside the window are built and still counted, only `hidden`, so the cursor keeps its place in the whole list.
- **Yank is the second `y`.** The clipboard writers (`pbcopy` / `wl-copy` / `xclip` / `xsel`) sit beside the existing readers in `ui/paste.go`.
- `keyMap` gains `Info` (`i`), `Yank` (`yy`) and the merged `Back` (`i/q/esc`); the client version probe runs as a command, so the cli block renders as unknown until it answers.

Deviation from the plan: **no `p` ping on the sheet.** No acceptance criterion asks for it and it would pull `network`, a spinner and an error panel onto a read-only screen; `enter` off a link still connects. Pinging stays on the table.

Tests in `ui/info_test.go` cover the lazy cursor wrapping and skipping heads, the scroll window shifting with the cursor and resetting on a page switch, and that no page carries password material.
<!-- SECTION:FINAL_SUMMARY:END -->
