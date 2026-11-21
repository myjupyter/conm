---
id: CONM-30
title: >-
  TUI screens stay inside the frame and the window at any size above the minimum
  width
status: Done
assignee:
  - '@claude'
created_date: '2026-11-04 08:48'
updated_date: '2026-11-04 09:02'
labels:
  - ui
dependencies: []
priority: medium
type: bug
ordinal: 26000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
A review of the adaptive-screen work (branch fix/make-adaptive-ui) rendered every screen across window sizes. Frames broke their right border on long values and scroll notes, the links hint and focused fields could not be scrolled into view in short windows, the keybinding list scrolled on every key near its end, and the help table and settings ran past the bottom of the window. Terminals narrower than a screen's minimum width are deliberately out of scope: the user decided to leave that for later.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [x] #1 No rendered line is wider than the frame, including long names, hosts, wide (CJK) characters and scroll notes
- [x] #2 The workspaces top border keeps the scroll note and stays in the frame down to its minimum width
- [x] #3 The reset prompt's y / n answer is readable at the minimum width of the settings screen
- [x] #4 The links hint in the connection form can be reached in a short window and the scroll note then clears
- [x] #5 The focused field stays visible when its section head does not fit
- [x] #6 Moving up from the end of the keybinding list does not scroll until the cursor leaves the window
- [x] #7 With help open, or in a very short window, the frame never exceeds the window height and keeps the help prompt and bottom border
- [x] #8 Settings keep the row the keys act on visible when one section is taller than the window
- [x] #9 Tests cover each of the above
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. frame.go: measure spans and truncPad in terminal cells (charmbracelet/x/ansi, promoted to a direct dependency); frameLine and frameTop clip anything wider than the frame; a top note drops its leading parts first so the scroll position survives.
2. screen_size.go: follow takes a focusSpan (from/row/to) and ends on the focused row, so a span taller than the window shows the row rather than the head; fit retries from the settled scroll position so it stops moving the window; join caps a frame at the window height and keeps the prompt and bottom border; keepEnd trims a section taller than the budget from the top.
3. form_view.go: the empty links head belongs to the last field's span.
4. settings keybind: the reset prompt lives in the status line only; the chips box shrinks by two cells so the longest default note fits.
5. Every screen renders through join. Tests in screen_size_test.go, frame_test.go.
Out of scope by user decision: terminals narrower than a screen's minimum width.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented per plan. go test ./..., go vet and golangci-lint are clean. Each fix was reverted in turn through a go test overlay and a test failed each time. A randomized key and resize run over 10 screens reports no wide line, overflow or lost cursor. Not run in a real terminal. Known ceiling: when help does not fit, the rows that do not fit are hidden and cannot be scrolled to. Waiting for the user's review before commit.
<!-- SECTION:NOTES:END -->

## Final Summary

<!-- SECTION:FINAL_SUMMARY:BEGIN -->
Screens stay inside the frame and the window at any size above their minimum width.

- Frame: spans and truncPad measure terminal cells; frameLine and frameTop clip anything wider than the frame; a top note drops its leading parts first so the scroll position survives.
- Scrolling: follow ends on the focused row; fit retries from the settled position, so the keybinding list no longer scrolls inside its window; the empty links head belongs to the last field.
- Height: every screen renders through screenSize.join, which caps the frame at the window and keeps the help prompt and bottom border; settings trim a too-tall section from the top.
- Settings: the reset prompt lives in the status line only, and the chips box is two cells narrower.

Tests: go test ./..., go vet, golangci-lint; each fix was reverted in turn and a test failed; a randomized key and resize run over 10 screens was clean. The user reviewed the result and approved the commit.

Risks and follow-ups: help rows that do not fit are hidden and cannot be scrolled to, and a very short window can cut the bottom of the error panel. Terminals narrower than a screen's minimum width remain out of scope by user decision.
<!-- SECTION:FINAL_SUMMARY:END -->
