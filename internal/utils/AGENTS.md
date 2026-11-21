# `internal/utils` — domain-free helpers

## Purpose

Leaves with no knowledge of conm's domain. A helper that must name a connection, a secret or
a screen does not belong here. This file covers the package and its three subpackages; none
of them imports anything internal.

## Packages

| Package | File | Holds |
|---|---|---|
| `utils` | `unique.go` | `FirstDuplicate(seq, key, of)` — the first element of a sequence sharing a key with `of`; the walk behind the repository's uniqueness rule |
| `utils/keys` | `keys.go` | a keystroke as text: `Parse` into a `Key` (name + `Mod` bitmask), `Named()`, `JoinSequence`/`SplitSequence` for multi-stroke bindings. Knows keys, not actions — what a key is bound to lives in `config` |
| `utils/pathname` | `pathname.go` | `Validate(name)` — whether a name is a legal folder name on Linux, macOS and Windows at once: forbidden characters, trailing dot/space, Windows reserved device names |
| `utils/names` | `names.go` | `Random()` — an adjective-noun name (`calm-otter`), what a new connection's form opens with |

## Who uses what

`repository` → `utils`; `config`, `ui` → `utils/keys`; `config` → `utils/pathname`;
`ui/spec` → `utils/names`.
