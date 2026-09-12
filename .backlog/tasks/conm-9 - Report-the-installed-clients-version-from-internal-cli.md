---
id: CONM-9
title: Report the installed client's version from internal/cli
status: To Do
assignee: []
created_date: '2026-09-11 20:27'
labels:
  - ready-to-work
milestone: m-0
dependencies: []
documentation:
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
- [ ] #1 `internal/cli` exposes a version probe that takes a client name and a context and returns the version string it printed
- [ ] #2 The version is parsed off the first line of the client's output, with the surrounding noise trimmed
- [ ] #3 A client that is not installed, fails, or prints nothing parseable yields no version and no error above `internal/cli` — callers render it as unknown
- [ ] #4 The probe is never run from `Detect`; it is an explicit per-client call, so listing clients stays free of subprocesses
- [ ] #5 The version flag lives in a table beside the other per-client spellings, so a client that disagrees with `--version` is one entry, not a branch
- [ ] #6 The probe honours its context and does not hang on a client that waits for input
- [ ] #7 A completeness test asserts every client in `clients` has a version flag entry
- [ ] #8 docs/cli.md documents the probe, the flag table and the unknown-version fallback
<!-- AC:END -->
