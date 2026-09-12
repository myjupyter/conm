---
id: CONM-5
title: Add a read-only info screen for a connection
status: To Do
assignee:
  - myjupyter
created_date: '2026-09-06 14:30'
updated_date: '2026-09-11 20:31'
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
priority: low
type: feature
ordinal: 8000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A connection's details are only visible by opening the edit form, one keystroke away from mutating the record just to look at it. A read-only info screen gives a safe place to inspect a connection, with mutation behind an explicit `e` jump to the edit form.

Three tabs:

- **general** — meta: name, description, tags; CLI: cli (binary name), version, path.
- **endpoint** — host, port, database, username, sslmode, dsn (password masked as `•••`).
- **links** — one row per resource link, its name (grafana, logs, runbook…) and full url; `enter` opens the highlighted one.

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
- [ ] #1 The screen has exactly three tabs — general, endpoint, links — with the active tab visibly marked
- [ ] #2 The general tab shows a meta block (name, description, tags) and a CLI block (cli binary name, version, path)
- [ ] #3 The endpoint tab shows host, port, database, username, sslmode and the dsn, with the password rendered as `•••`
- [ ] #4 The links tab shows one row per resource link with its name and full url, and `enter` opens the highlighted link
- [ ] #5 No password material, password reference, secret provider or secret location appears anywhere on the screen
- [ ] #6 No key on the screen edits, deletes or otherwise mutates the connection or its secret
- [ ] #7 Pressing `e` leaves the info screen and opens the edit form for the same connection
- [ ] #8 The keyhint footer lists the screen's keys and follows the shared footer/theme/keybinding primitives, with no raw rune, colour or key string in the screen
- [ ] #9 The screen reads entities through `ui/view` only, with no file I/O and no hand-built network connection
- [ ] #10 docs/ui.md documents the new screen, its tabs and its place in the screen set
- [ ] #11 Pressing `yy` with a field focused copies that field's displayed value to the clipboard and confirms it in the status line; only displayed values are copyable
- [ ] #12 The screen's layout, framing and colours follow the terminal-spec section of `dist/DB Connections TUI.dc.html`, expressed through the existing `ui/frame.go` and `ui/theme.go` primitives
- [ ] #13 Pressing `i` on the connections table opens the info screen for the selected connection; `i` is added to the shared keymap as an Info binding
- [ ] #14 `q`, `i` and `esc` with no cursor return to the connections table with the same row still selected
- [ ] #15 The screen opens with no cursor; the first `j`/`k` creates it on the first/last field respectively
- [ ] #16 The cursor steps field to field, wraps at both ends and skips section headers
- [ ] #17 The focused row is marked with the accent background, a `❯` marker and a hint of what it accepts
- [ ] #18 `yy` yanks the focused value, `enter` opens a url or connects, and `y` with no cursor asks the user to move it first
- [ ] #19 `tab` and `←`/`→` switch page and drop the cursor; `esc` clears the cursor and a second `esc` leaves the info screen
- [ ] #20 Only the links page scrolls, rendering at most 8 rows, and the window shifts by one as the cursor moves past either edge
- [ ] #21 The links section header counts what is hidden as `↑ N above · ↓ N below`, with no arrows or scrollbar
- [ ] #22 Switching page resets both the cursor and the scroll window to the top
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
<!-- SECTION:PLAN:END -->
