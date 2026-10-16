---
id: CONM-26
title: Saving a config file never leaves it truncated or corrupted
status: Done
assignee:
  - '@claude'
created_date: '2026-10-11 16:00'
updated_date: '2026-10-16 19:33'
labels:
  - release
dependencies: []
references:
  - internal/config/config.go
  - internal/config/conm.go
modified_files:
  - internal/config/config.go
  - internal/config/conm.go
  - internal/config/config_test.go
priority: medium
type: bug
ordinal: 22000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
conm rewrites its TOML files in place: the connection files are held open for the session and saved by seek-to-start, write, truncate; conm.toml is rewritten whole. A crash, kill or full disk in the middle of a save leaves a half-written file, and the user's connections are lost or unparseable on the next start.

Related, to decide in the plan: two conm instances running at once each keep their own in-memory copy, and the later save silently overwrites the earlier one's changes. Whether this task also guards against that, or only makes a single save atomic, is open.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 An interrupted save leaves either the previous or the new file content on disk, never a partial one
- [x] #2 File permissions stay 0600 after a save
- [x] #3 Applies to every file conm writes: conm.toml, the per-type connection files and the secret records
- [x] #4 A decision on concurrent instances is recorded in the task: either handled, or explicitly out of scope
- [x] #5 Tests cover a successful save and a save that fails mid-way leaving the old content intact
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/config: one helper `writeFileAtomic(path, data)` — resolve symlinks, CreateTemp in the target's dir (0600), write, fsync, close, rename over the target; remove the temp on any failure.
2. `Config` keeps the path instead of an open *os.File: OpenConfig reads with os.ReadFile (missing file = empty), Save marshals and calls writeFileAtomic, Close is a no-op kept for the File interface.
3. `CreateConm` writes through writeFileAtomic too. This covers conm.toml, every per-type file, transports and secret.toml — all go through OpenConfig/CreateConm.
4. Concurrent instances: out of scope. Atomic save guarantees a whole file, not merge; last writer wins as before. Revisit with locking/mtime check if it becomes a real problem.
5. Tests in config: successful save (content + 0600), failed save (read-only dir, target replaced by dir) leaves old content and no temp files.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented writeFileAtomic in internal/config/config.go (temp in same dir → write → fsync → rename; symlink resolved first so a dotfiles link survives). Config no longer holds an open file; Close is a no-op. CreateConm uses the same helper. Concurrent instances: explicitly out of scope — last writer still wins, but always with a whole file. Tests: internal/config/config_test.go TestWriteFileAtomic. go test ./... and make lint clean.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Config saves are atomic. `writeFileAtomic` (internal/config/config.go) writes to a temp file in the target's directory, fsyncs, and renames it over the target; the temp file is removed on any failure. Symlinks are resolved first so a linked dotfile stays a link. `Config` no longer holds an open file (Close is a no-op kept for the File interface); `Save` and `CreateConm` both go through the helper, covering conm.toml, every per-type connection file, the transports file and secret.toml. Files are always written 0600.

Concurrent instances: out of scope here, tracked as a follow-up task (last writer still wins, but always with a whole file).

Tests: TestWriteFileAtomic — replace, create, write through symlink, temp-creation failure and rename failure (old content intact, no temp left). go test ./... and make lint clean.
<!-- SECTION:FINAL_SUMMARY:END -->
