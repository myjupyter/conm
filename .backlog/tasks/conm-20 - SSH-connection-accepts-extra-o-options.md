---
id: CONM-20
title: SSH connection accepts extra -o options
status: To Do
assignee: []
created_date: '2026-10-06 19:22'
labels:
  - ssh
dependencies: []
type: feature
ordinal: 16000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
An escape hatch for options conm has no field for (StrictHostKeyChecking, ConnectTimeout, Compression, …). Options conm sets itself (PreferredAuthentications, PubkeyAuthentication, …) must not be overridable.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 An SSH connection stores extra -o options and passes them to ssh
- [ ] #2 Options conm manages itself are refused at validation
- [ ] #3 Unknown -o options in a pasted command are kept instead of ignored
- [ ] #4 Covered by tests
<!-- AC:END -->
