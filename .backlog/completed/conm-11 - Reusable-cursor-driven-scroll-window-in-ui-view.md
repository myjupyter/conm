---
id: CONM-11
title: Reusable cursor-driven scroll window in ui/view
status: Done
assignee:
  - myjupyter
created_date: '2026-09-11 20:30'
updated_date: '2026-09-12 12:57'
labels:
  - ready-to-work
milestone: m-0
dependencies: []
documentation:
  - docs/view.md
modified_files:
  - internal/ui/view/scroll.go
  - internal/ui/view/scroll_test.go
  - docs/view.md
priority: low
type: enhancement
ordinal: 1000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
No screen scrolls today — every list renders all of its rows. The info screen's links page needs a window (CONM-5), and the connections and secrets tables will want the same one, so it lands in `ui/view` as a domain-free primitive rather than inside a screen.

`view.Scroll` counts rows and knows nothing about what they are:

```go
type Scroll struct{ size, start int }
func (s *Scroll) Resize(size int)
func (s *Scroll) Follow(cursor, total int)
func (s *Scroll) Range(total int) (start, end int)
func (s *Scroll) Above() int
func (s *Scroll) Below(total int) int
func (s *Scroll) Top()
```
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 `view.Scroll` holds a window size and a start offset and never names a connection, secret or screen
- [x] #2 `Follow` drags the window by one row when the cursor moves past either edge and leaves it alone while the cursor is inside
- [x] #3 `Range` returns the visible half-open bounds, clamped so it never runs past the total or below zero
- [x] #4 `Above` and `Below` report how many rows are hidden on each side, for a caller to render as a count
- [x] #5 `Top` resets the window, and a size or total that shrinks below the current start clamps rather than panics
- [x] #6 A total that fits inside the window keeps the start at zero and hides nothing
- [x] #7 docs/view.md documents `Scroll`, its contract and that cursor and window are the view's, not a screen's
- [x] #8 A table-driven test covers follow in both directions, the clamps, the hidden counts and the reset
<!-- AC:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
`Scroll` is a bare `{size, start}` pair in `internal/ui/view/scroll.go`; no screen wires it yet — CONM-5's links page is the first caller.

`Range`, `Follow` and `Below` all run the same `clamp(total)` first, so every entry point normalises `start` against the total it is handed rather than each caller re-checking. `size == 0` and `total <= size` share one branch (`windowed`): start pinned to 0, `Range` returns `0, total`, nothing hidden — a screen that never calls `Resize` renders every row.

`Follow` clamps rather than stepping, so a one-row move drags the window by one and a jump (`Bottom`, a kind switch) re-homes it in the same call.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Added `view.Scroll` — a domain-free cursor-driven window in `ui/view`, with a table-driven test and a `docs/view.md` section.
<!-- SECTION:FINAL_SUMMARY:END -->
