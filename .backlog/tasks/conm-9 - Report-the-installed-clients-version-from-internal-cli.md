---
id: CONM-9
title: Report the installed client's version from internal/cli
status: Done
assignee:
  - Kirill Tkachuk
created_date: '2026-09-11 20:27'
updated_date: '2026-09-12 13:09'
labels:
  - ready-to-work
milestone: m-0
dependencies: []
documentation:
  - docs/cli.md
modified_files:
  - internal/cli/command.go
  - internal/cli/command_test.go
  - docs/cli.md
priority: low
type: enhancement
ordinal: 5000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`cli.Detect` reports a client's name, path and whether it exists — never its version, so nothing above `internal/cli` can say which `psql` is on the PATH. The info screen's CLI block (CONM-5) needs it, and `conm init` can use it too.

Add a version probe to `internal/cli`: run the binary once and read the version off its first output line. `--version` is assumed for all six clients; a client that spells it differently gets its own entry in the same table the argv builders live beside.

The probe shells out, so it stays off the detection path: callers ask for it explicitly, per client, and a failure degrades to "unknown" rather than an error on screen.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `internal/cli` exposes a version probe that takes a client name and a context and returns the version string it printed
- [x] #2 The version is parsed off the first line of the client's output, with the surrounding noise trimmed
- [x] #3 A client that is not installed, fails, or prints nothing parseable yields no version and no error above `internal/cli` — callers render it as unknown
- [x] #4 The probe is never run from `Detect`; it is an explicit per-client call, so listing clients stays free of subprocesses
- [x] #5 The version flag lives in a table beside the other per-client spellings, so a client that disagrees with `--version` is one entry, not a branch
- [x] #6 The probe honours its context and does not hang on a client that waits for input
- [x] #7 A completeness test asserts every client in `clients` has a version flag entry
- [x] #8 docs/cli.md documents the probe, the flag table and the unknown-version fallback
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. `internal/cli/command.go` — add `versionFlags map[string]string` next to `commands` (the per-client spelling table), every client mapped to `--version`.
2. Same file — `Version(ctx, name) string`: look the flag up, `exec.CommandContext(...).Output()` (stdin left nil so a client that waits for input reads EOF), hand stdout to `parseVersion`. Every failure path — unknown name, not installed, non-zero exit, context cancelled, nothing parseable — returns `""`; no error crosses the package boundary.
3. `parseVersion(out string) string` — first line only, first `\d+(\.\d+)*` token off it, so `psql (PostgreSQL) 16.2` and `redis-cli 7.2.4 (git:0)` both yield the number.
4. `Detect` untouched: no subprocess on the listing path.
5. `internal/cli/command_test.go` — table-driven `TestParseVersion` over the six real `--version` spellings plus empty/garbage, and `TestEveryClientHasAVersionFlag` mirroring `TestEveryClientHasACommand`.
6. `docs/cli.md` — new section documenting the probe, the flag table and the unknown fallback.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
`Version(ctx, name) string` and `versionFlags` live in `command.go`, beside `commands` — the two per-client spelling tables read together. All twelve clients map to the `versionFlag` constant (golangci's `goconst` refuses twelve copies of the literal); a client that spells it otherwise is one entry.

Parsing is the first line, first `\d+(\.\d+)*` run on it. Verified against the binaries actually installed here: `psql -> 17.11`, `redis-cli -> 8.10.1`, `mongosh -> 2.9.2`, `mysql -> 14.14`, and `usql` (not installed) -> `""`.

`mysql` reporting 14.14 is the client's own version, not the server's — that is what `mysql --version` prints first (`mysql  Ver 14.14 Distrib 5.7.44`). Documented rather than special-cased: the probe reports what the client says about itself.

No error crosses the boundary: unknown name, missing binary, non-zero exit, cancelled context and an unparseable line all return `""`. `Output()` leaves the child's stdin nil, so a client that would prompt reads EOF instead of hanging, and `CommandContext` kills the rest.

`Detect` is untouched — listing clients still starts no subprocess.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added a version probe to `internal/cli`.

`cli.Version(ctx, name)` runs the client once with its version flag and returns the version it printed, or `""` for anything that does not work out — unknown client, not installed, non-zero exit, cancelled context, no number on the first line. Callers render the empty string as unknown; no error reaches above `internal/cli`.

The flag is a table (`versionFlags`) beside the argv-builder table in `command.go`, one entry per client, guarded by `TestEveryClientHasAVersionFlag` the same way `commands` is guarded by `TestEveryClientHasACommand`. Parsing takes the first line and the first `\d+(\.\d+)*` run on it, so `psql (PostgreSQL) 16.2` and `redis-cli 7.2.4 (git:0)` both reduce to the number; `TestParseVersion` covers the six real spellings plus the empty and unparseable cases.

`Detect` is unchanged: listing clients remains a pure `PATH` lookup with no subprocess. `docs/cli.md` gains a "The version probe" section covering the probe, the flag table and the unknown fallback.

`make lint`, `go vet ./...` and `go test ./...` are clean.
<!-- SECTION:FINAL_SUMMARY:END -->
