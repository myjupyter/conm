---
id: CONM-5
title: Add a read-only info screen for a connection
status: To Do
assignee:
  - myjupyter
created_date: '2026-09-06 14:30'
updated_date: '2026-09-06 15:12'
labels:
  - not-ready
milestone: m-0
dependencies:
  - CONM-6
documentation:
  - docs/ui.md
  - docs/view.md
  - docs/forms.md
  - dist/DB Connections TUI.dc.html
priority: low
type: feature
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Everything known about a connection is currently only visible by opening the edit form, which puts the user one keystroke away from mutating the record just to look at it. There is also no place that shows the derived DSN, the secret's location/provider, or record metadata at all — that information is spread across the connection file, the secret store and the form spec, and never presented together. A dedicated read-only screen gives a safe place to inspect a connection before pinging or launching a client, and keeps mutation behind the explicit `e` jump to the edit form.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The screen shows connection details, the connection's DSN, secret information and metadata, grouped into the tabs: connection details, secret, metadata
- [ ] #2 The password material never appears on the screen or in the DSN it displays
- [ ] #3 Focus can be moved between tabs and the focused tab's content can be scrolled/navigated, with the active tab visibly marked
- [ ] #4 No key on the screen edits, deletes or otherwise mutates the connection or its secret
- [ ] #5 Pressing `e` leaves the info screen and opens the edit form for the same connection
- [ ] #6 The keyhint footer lists the screen's keys and follows the shared footer/theme/keybinding primitives, with no raw rune, colour or key string in the screen
- [ ] #7 The screen reads entities through `ui/view` only, with no file I/O and no hand-built network connection
- [ ] #8 docs/ui.md documents the new screen and its place in the screen set
- [ ] #9 Pressing `yy` with a field focused copies that field's displayed value to the clipboard and confirms it in the status line; only displayed values are copyable, never resolved password material
- [ ] #10 The screen's layout, framing and colours follow the terminal-spec section of `dist/DB Connections TUI.dc.html`, expressed through the existing `ui/frame.go` and `ui/theme.go` primitives
- [ ] #11 Pressing `i` on the connections table opens the info screen for the selected connection; `i` is added to the shared keymap as an Info binding
- [ ] #12 On the info screen `q`, `esc` and `i` all return to the connections table with the same row still selected
<!-- AC:END -->
