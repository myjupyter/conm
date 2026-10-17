---
id: CONM-28
title: Two conm instances never silently overwrite each other's saves
status: To Do
assignee: []
created_date: '2026-10-16 19:33'
updated_date: '2026-10-16 19:33'
labels:
  - release
dependencies:
  - CONM-26
references:
  - internal/config/config.go
  - internal/ui/form_loop.go
  - internal/repository/errors.go
priority: medium
type: bug
ordinal: 24000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Each conm instance reads its TOML files once into memory and writes the whole file back on save (atomically since conm-26). With two instances running at once — two terminals, or `conm add` while the TUI is open — the later save overwrites the earlier one's changes without a word: an added connection or secret record disappears.

To decide in the plan: refuse a save when the file changed on disk since it was read (mtime/size or content hash check, typed refusal surfaced through the form loop), an advisory lock held across read-modify-write, or re-read-and-reapply. Rejection over merging matches how conm handles uniqueness.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 A save never discards changes another conm instance wrote to the same file after this instance read it
- [ ] #2 When a save is refused, the user sees why in the status line and stays on the screen they came from
- [ ] #3 Applies to conm.toml, the per-type connection files, the transports file and secret.toml
- [ ] #4 Tests cover two configs opened on the same file where the second save meets the first one's change
<!-- AC:END -->
