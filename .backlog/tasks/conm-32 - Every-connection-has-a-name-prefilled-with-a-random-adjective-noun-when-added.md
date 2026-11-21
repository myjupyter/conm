---
id: CONM-32
title: 'Every connection has a name, prefilled with a random adjective-noun when added'
status: In Progress
assignee:
  - '@claude'
created_date: '2026-11-09 17:02'
updated_date: '2026-11-09 17:06'
labels: []
dependencies: []
type: enhancement
ordinal: 28000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
The name is the identifier every named command uses (conm <type> <name>, edit, remove, ping, exec), yet it is optional, so a connection saved without one can't be reached by any of them. The owner wants the name required, and the add form to open with a Docker-style random name (adjective-noun, e.g. calm-otter) already filled in: keep it or change it. Names built from database + host were rejected because long hosts truncate into near-identical names told apart only by a meaningless counter.
<!-- SECTION:DESCRIPTION:END -->

## Acceptance Criteria
<!-- AC:BEGIN -->
- [ ] #1 The add form of every connection type (postgres, mysql, mssql, clickhouse, redis, mongodb, ssh) opens with the name field holding a random adjective-noun name
- [ ] #2 A form prefilled from a command line that carries no name also gets a random name
- [ ] #3 The edit form keeps the stored name and never generates one
- [ ] #4 The name field is marked required and the form refuses to save with it empty
- [ ] #5 The store refuses a connection with an empty name
- [ ] #6 A generated name is a valid connection name: lowercase words joined by '-', never reserved
- [ ] #7 Tests cover the generator, the add/edit/prefill cases and the empty-name refusal
<!-- AC:END -->

## Implementation Plan

<!-- SECTION:PLAN:BEGIN -->
1. internal/ui/spec/name.go: RandomName() — adjective + "-" + noun from two short word lists, math/rand/v2. Lives in spec (form default logic; spec may only import config/cli/secret, so no utils package).
2. spec.FormField gains DefaultFunc func() string; the name field of all 7 connection specs sets DefaultFunc: RandomName and drops OptionalFieldProperty.
3. ui/form.go newFormModel: when not editing and the seeded value is empty, a field with DefaultFunc gets a generated value (covers empty add, command-line prefill without a name, and a refused-save reopen with the name cleared).
4. config: rename ValidateReservedName -> ValidateConnectionName (only the 7 connection validators call it) and make it refuse an empty name, so the store refuses unnamed records too.
5. Tests (gotests tables, testify): RandomName shape + validity; add form prefills, prefill with empty name prefills, edit keeps stored name; ValidateConnectionName refuses empty. Fix existing tests that relied on empty names.
6. go vet, go test, make lint.

Owner asked to keep name generation out of ui: generator moved to the new leaf internal/utils/names (names.Random), ui/spec imports it; arrow and table row added to CLAUDE.md. gosec G404 excluded for that path only in .golangci.yaml.
<!-- SECTION:PLAN:END -->

## Implementation Notes

<!-- SECTION:NOTES:BEGIN -->
Implemented, uncommitted, awaiting owner review of the diff. go vet, go test ./..., make lint (0 issues) all pass. Checked in the real binary via VHS: `conm postgres add "psql -h localhost -U me mydb"` opens with name * = tidy-delta. Existing test TestAddAcceptsADifferentEndpoint relied on unnamed connections; given names. TestValidateReservedName became TestValidateConnectionName in gotests shape. Repository tests in unique_test.go are free-form; the new TestAddRejectsAnEmptyName follows that file's shape.
<!-- SECTION:NOTES:END -->
