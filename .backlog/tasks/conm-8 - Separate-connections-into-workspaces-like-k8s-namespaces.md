---
id: CONM-8
title: 'Separate connections into workspaces, like k8s namespaces'
status: To Do
assignee: []
created_date: '2026-09-08 20:15'
updated_date: '2026-09-08 20:17'
labels: []
dependencies: []
references:
  - ROADMAP.md
  - internal/repository/workspace.go
  - internal/config/xdg.go
type: feature
ordinal: 7000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Everything conm manages lives in one flat set: one list per database type, one secret store, one set of enabled databases. Working across contexts (prod / staging / personal) means seeing all of them at once, and `prod-main` one row away from `local` is how the wrong `psql` gets opened.

A workspace is a named partition over all of it, like a k8s namespace: you are in exactly one, and it owns its connections, its secrets and its enabled database types.

Two things not recoverable from the code:

- `repository.Workspace` already means the open repositories for the enabled databases. One of the two names has to go.
- ROADMAP.md D2 keeps arbitrary grouping (`tagged staging`) out of the model as a view-only label. A workspace is the deliberate exception — one at a time, decides which files are opened, a property of the store. Not the same feature as tags, not D2's connection→connection edge.

Open: field on each record in the existing TOML files, or a config tree per workspace (`$XDG_CONFIG_HOME/conm/<workspace>/…`). The tree fits "owns its enabled database types"; the field is the smaller diff. Pick one in the plan, with the migration path for existing flat configs. Check conm-1/conm-2 first — both add structure to the same model.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 One workspace is active at a time; the TUI shows only its connections, secrets and enabled database types, and names it on screen
- [ ] #2 Workspaces can be created, renamed and removed from the TUI, with the fate of a removed workspace's resources made explicit
- [ ] #3 Switching workspace takes effect without restarting conm
- [ ] #4 An existing flat config keeps working and lands in a default workspace, with no user action
- [ ] #5 Names are unique per workspace, not globally
- [ ] #6 The clash with repository.Workspace is resolved by a rename; no package holds both meanings
- [ ] #7 Tests cover per-workspace uniqueness, switching, and reading a pre-workspace config
- [ ] #8 docs/ and CLAUDE.md's layer map reflect the new concept
<!-- AC:END -->
