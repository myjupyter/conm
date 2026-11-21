---
id: CONM-31
title: >-
  New users can install and learn conm from a README and a short sequential
  guide
status: In Progress
assignee:
  - '@claude'
created_date: '2026-11-04 19:49'
updated_date: '2026-11-04 19:51'
labels: []
dependencies: []
type: docs
ordinal: 27000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The repository has no README and docs/ only serves contributors, so a new user has no way to learn what conm does or how to use it. The owner wants docs for people: short, use cases only, read in order. docs/ contributor pages stay where they are. This is a draft for the owner's review; nothing is committed until they approve the diff.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 README.md says what conm is, how to install it, a quick start, the supported databases and clients, and links to the guide
- [ ] #2 docs/user/ holds numbered pages that read in order, each organised by use case, and each linking to the next
- [ ] #3 Every command, key and file path in the docs matches the current binary and code
- [ ] #4 The owner has reviewed the draft
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
Write README.md plus docs/user/1-getting-started, 2-daily-use, 3-passwords, 4-ssh, 5-workspaces, 6-keys-and-files. Facts come from the code and from driving the binary in a PTY against scratch config. Leave the demo GIF out until the owner approves installing VHS. Do not commit.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Draft written, uncommitted, awaiting owner review. Checked in a PTY against scratch config: first run, add form, table, transports/secrets/settings/workspaces screens, CLI list and ping failure, keybinding TOML. Not exercised: successful client session, real tunnel, keyring save, go install. Found while checking: a connection saves with an empty name, which no named command can reach.
<!-- SECTION:NOTES:END -->
