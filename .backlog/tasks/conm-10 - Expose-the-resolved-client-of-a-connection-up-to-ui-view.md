---
id: CONM-10
title: Expose the resolved client of a connection up to ui/view
status: Done
assignee: []
created_date: '2026-09-11 20:28'
updated_date: '2026-09-12 13:40'
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
- [ ] #1 `cli.Launcher` reports its resolved client as a `cli.Info` (name, path, installed, version)
- [ ] #2 `repository.Connections` exposes that client, served by `ConnectionRepository` from its own launcher
- [ ] #3 `view.Connections` exposes the client of the cursor's database, so a screen never holds a repository to get it
- [ ] #4 No layer above `internal/cli` names or spells a client binary to obtain this
- [ ] #5 docs/cli.md, docs/repository.md and docs/view.md record the accessor at each layer
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. `internal/cli/detect.go` — `Info` becomes the value carried up: `Exists` → `Installed` (the AC's word), plus a `Version string`. Split the `PATH` lookup out as `lookup(name string) Info` and let `Detect` build its candidates from it. **`Detect` still leaves `Version` empty** — the setup listing stays a `PATH` lookup with no subprocess, exactly as `docs/cli.md` promises; the version is filled only where a single, already-resolved client is asked about.
2. `internal/cli/launcher.go` — `func (l Launcher) Info(ctx context.Context) Info`: no client configured → zero `Info`; a name that is not in `Clients(l.kind)` → `Info{Name: l.name}` and nothing else, the same refusal `Run` makes, so a hand-edited `conm.toml` never renders as installed on the strength of a coincidental binary; otherwise `lookup(l.name)` and, when installed, `Version(ctx, l.name)`. This is the only accessor in the chain that starts a subprocess — which is why `ctx` travels with it through every layer below.
3. `internal/ui/init_table.go` (`installedClient`) and `internal/ui/init_table_view.go` (client span colour) — the two `cli.Exists` reads become `cli.Installed`. No other caller of `Info` exists.
4. `internal/repository/connection.go` — `ClientInfo(ctx context.Context) cli.Info` on the `Connections` interface, implemented by `ConnectionRepository` as `r.launcher.Info(ctx)`. No lock: the launcher is frozen at `openConnections` and never reassigned. `ConnectionRepository` is the only implementer, so no mock or fake follows.
5. `internal/ui/view/connection.go` — `ClientInfo(ctx context.Context) (cli.Info, bool)`, served by the repository of the **active** kind: the visible rows are always the active kind's rows, so the cursor's database *is* `Active()`. One method, not the usual `At`/`For` pair — the client is a property of the database, not of a row, so there is nothing a `ConnRef` would disambiguate.
6. `CLAUDE.md` — the layer table and the arrow diagram gain `cli` to `internal/ui/view`'s allowed imports. `cli` sits below `view`, so the new arrow still points inward and no rule is reversed; returning the `cli.Info` value rather than a re-declared copy of its fields is what keeps AC #4 true.
7. `internal/cli/launcher_test.go` — one table-driven test over `Launcher.Info`: no client configured → zero `Info`; a name outside `Clients(kind)` → `Name` set, `Installed` false, `Path`/`Version` empty; a real configured client → `Name` echoed and `Installed == (Path != "")`, which holds whether or not the host has it. Nothing new is tested in `repository` or `view` — both are pass-throughs.
8. Docs — `docs/cli.md`: `Info()` beside `Run()` in the launcher section, and a line in the version-probe section saying which of the two paths fills `Version` and why `Detect` still does not. `docs/repository.md`: `ClientInfo(ctx)` in the `Connections` surface. `docs/view.md`: `ClientInfo(ctx)` in the `Connections` accessor list, with the note that it is keyed by the active kind rather than by a `Ref`.
<!-- SECTION:PLAN:END -->
