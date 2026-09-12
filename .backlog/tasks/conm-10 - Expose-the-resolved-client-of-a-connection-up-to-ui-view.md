---
id: CONM-10
title: Expose the resolved client of a connection up to ui/view
status: To Do
assignee: []
created_date: '2026-09-11 20:28'
labels:
  - ready-to-work
milestone: m-0
dependencies: []
documentation:
  - docs/cli.md
  - docs/repository.md
  - docs/view.md
priority: low
type: enhancement
ordinal: 6000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`cli.Launcher.name` is unexported and nothing above `internal/network` can ask which client a connection actually launches. Carry it up as a `cli.Info` — the value, not a binary name a screen spells itself: `Launcher.Info()` → `repository.Connections.Client()` → `view.Connections.Client()`. Needed by the info screen's CLI block (CONM-5).
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 `cli.Launcher` reports its resolved client as a `cli.Info` (name, path, installed)
- [ ] #2 `repository.Connections` exposes that client, served by `ConnectionRepository` from its own launcher
- [ ] #3 `view.Connections` exposes the client of the cursor's database, so a screen never holds a repository to get it
- [ ] #4 No layer above `internal/cli` names or spells a client binary to obtain this
- [ ] #5 docs/cli.md, docs/repository.md and docs/view.md record the accessor at each layer
<!-- AC:END -->
