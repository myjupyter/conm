---
id: CONM-21
title: >-
  List and launch connections from the shell so an external picker like fzf can
  search them
status: To Do
assignee: []
created_date: '2026-10-06 19:44'
labels:
  - cli
  - search
dependencies: []
priority: medium
type: feature
ordinal: 17000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Today a connection can only be found and opened from inside the TUI, and its search is plain substring. A list printed as one line per connection plus a command that runs one by key lets the user pipe the set into fzf (or skim, peco, grep) for fuzzy search, a ping preview and a one-keystroke launch — without conm depending on or naming any picker. The same list feeds shell completion and scripts.

Example: `conm run "$(conm ls | fzf --preview 'conm ping {1}' | cut -f1)"`.

Constraint: `cmd` may not import `repository` or `network` (depguard), so these commands go through `ui`/`ui/view`, not around them.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `conm ls` prints one tab-separated line per connection: a stable key `<type>/<name>` first, then the same fields the TUI search indexes
- [ ] #2 No line ever contains a password or resolved secret
- [ ] #3 `conm run <key>` hands the terminal to the connection's client exactly as `enter` does in the TUI
- [ ] #4 `conm ping <key>` prints the ping result and exits non-zero on failure
- [ ] #5 An unknown key fails with a clear error and non-zero exit
- [ ] #6 `conm run` and `conm ping` complete keys in the shell via Cobra completion
- [ ] #7 Tests cover the line format, secret absence and unknown-key handling
- [ ] #8 docs/ describe the commands and the fzf one-liner
<!-- AC:END -->
