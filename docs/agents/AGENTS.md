# Writing agent documentation

For an agent — or a person — about to change conm's code. It lives next to the code it
describes, so it is created, changed and deleted together with it.

## Where each fact goes

| File | Holds | Loaded |
|---|---|---|
| `AGENTS.md` (root) | what conm is, what to run after a change, code style, where the rest lives | every session |
| `internal/AGENTS.md` | the architecture: the layer map, the import table, the rules that cross modules, the connection lifecycle, the extension seams | on demand, when working on code |
| `<module>/AGENTS.md` | one module: its purpose, files, entities, own rules, extension seam | on demand, when working in that module |
| `docs/user/AGENTS.md`, `docs/agents/AGENTS.md` | how to write the two kinds of documentation | on demand |

A fact has one home. About one module: that module's file, and `internal/AGENTS.md` at most
links to it. Constraining several modules — an import direction, who may name a binary —
`internal/AGENTS.md`. The root holds only what applies to every change wherever it lands.
Never copy a paragraph between files.

Every `AGENTS.md` has a `CLAUDE.md` beside it holding the single line `@AGENTS.md`, so tools
that look for either name reach the same text. Create it together with the file, and never
put anything else in it.

## Which modules get one

Every Go package with code in it: `cmd` and each directory under `internal/`. The exception
is `internal/utils`, whose three single-file subpackages share the parent's file. An empty
directory gets none; a package that grows a subpackage with entities of its own gives that
subpackage its own file (`internal/ui/view`, `internal/ui/spec`).

## The shape of a module file

```markdown
# `<import path>` — <what it is, in a few words>

## Purpose
Two to four sentences: what the module is responsible for and what it must not know.
"May import (internal): …" — the allowed list, matching the table in internal/AGENTS.md.

## Files
A table: file (or group of files sharing a pattern) → what it holds.

## Main entities
The interfaces as Go, then a bullet per entity: what it is, who implements it, who
consumes it. Types and functions a reader needs to navigate — not every exported name.

## Rules / How it works
What cannot be read off the code: invariants, deliberate choices and their reasons,
external constraints, per-driver or per-client quirks.

## Adding a <thing>
Only for a module that is an extension seam: the registration points, in order.
```

The first three sections are mandatory and stay short. The rest is free-form and may be
long — it is only read by someone already working in that module.

## What to write, and what not to

- **Write what the code cannot say**: why a rule exists, what breaks without it, which
  alternative was rejected. The repository forbids code comments, so this is the only place
  such a fact can live.
- **Do not restate the code.** No signature lists beyond the interfaces layers meet at, no
  step-by-step paraphrase of a function.
- Name things as the code does, in backticks — `repository.Session`, `form_loop.go` — so a
  rename can be found by search. A name in an `AGENTS.md` that no longer exists in the code
  is a bug in the documentation.
- State rules in the imperative and bold the rule itself, the way `internal/AGENTS.md` does.
- No history ("used to", "no longer", "was moved"): describe what is, not how it got there.
- Refer to another module by the path of its file: `internal/cli/AGENTS.md`.

## When to update

In the same change as the code:

- a file added, removed or renamed in a module → its **Files** table;
- an interface or a main type changed → **Main entities**;
- a rule introduced, lifted or bent → the rules section of the module that owns it;
- a package added → a new `AGENTS.md` and its `CLAUDE.md`, a row in the package table of
  `internal/AGENTS.md`, an arrow in its diagram;
- an import added between internal packages → the table and diagram in `internal/AGENTS.md`
  first; if the arrow points outward, the change is wrong, not the table;
- a registration point of a connection type moved → the `add-new-database-type` skill too.
