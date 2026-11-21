# `internal/secret` — password material behind a scheme

## Purpose

Resolves, stores and removes password material behind a scheme (`literal:`, `none:`,
`filepath:`, `keyring:`). **It knows schemes and material, not connections**: anything that
can hand over a raw reference string is a `Reference`. A leaf.

May import (internal): *(nothing)*.

## Files

| File | Holds |
|---|---|
| `schema.go` | `Scheme`, the four scheme constants, `IsStore`, `IsMaterial` |
| `ref.go` | `Reference`, `Ref(scheme, location)`, `ParseRef` |
| `provider.go` | `Provider` |
| `resolver.go` | `Resolver`, `DefaultFor(workspace)` |
| `literal.go`, `none.go`, `filepath.go`, `keyring.go` | one `Provider` each |

## Main entities

```go
type Reference interface{ SecretRef() string }

type Provider interface {
    Scheme() Scheme
    Resolve(ctx context.Context, ref Reference) (string, error)
    Store(ctx context.Context, ref Reference, password string) error
    Remove(ctx context.Context, ref Reference) error
}
```

- A reference is a bare literal or `"<scheme>:<location>"`. `Ref` and `ParseRef` spell and
  split it here; the form-side `spec.JoinSecret`/`SplitSecret` are the only callers that
  assemble one from user input.
- `Resolver` is itself a `Provider`: it dispatches by scheme and falls back to the literal
  provider for a value with no known scheme, so a bare password containing a colon stays a
  password. `DefaultFor(workspace)` registers the four providers.
- `LiteralProvider` returns the value itself; `NoneProvider` resolves to the empty string;
  `FilepathProvider` resolves to the path (with `~` expanded) and never opens the file;
  `KeyringProvider` reads and writes the system keyring under a per-workspace service
  name, so workspaces do not see each other's secrets.

## Rules

- `IsStore(scheme)` is the one test for "this reference points into a store" — only
  `keyring` today. Never write `!= secret.Literal`.
- `IsMaterial(scheme)` is the one test for "this value is the password itself" — what a form
  masks and keeps untrimmed.
- The `none` reference is stored as `"none:"`; a bare `""` is an unset literal.
- `Store`/`Remove` are no-ops for the schemes that hold nothing.

## Adding a secret store

One new `Provider` registered in `DefaultFor`, and its scheme constant in `schema.go`.
Nothing else in this package changes. The rest of the seam — the `config.Secret` record, the
repository constructor, the form spec, `spec.SecretProvidersOrder`, the view — is listed in
`internal/AGENTS.md`.
