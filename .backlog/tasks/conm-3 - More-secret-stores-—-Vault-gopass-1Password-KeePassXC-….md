---
id: CONM-3
title: 'More secret stores — Vault, gopass, 1Password, KeePassXC, …'
status: To Do
assignee:
  - myjupyter
created_date: '2026-09-06 14:00'
updated_date: '2026-09-06 14:56'
labels:
  - not-ready
milestone: m-0
dependencies: []
priority: medium
type: enhancement
ordinal: 3000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
`keyring` is the only store in the tree. The seam for adding one already exists — a
`secret.Provider` registered in `secret.Default()`, a `config.Secret` record with its
wrapper, a repository constructor, a named spec in `secretFormSpecs`, the scheme in
`spec.SecretProvidersOrder` and the repository passed to `view.NewSecrets` — and no form or
screen code changes. What each candidate adds beyond that seam:

- **HashiCorp Vault** — a network service, not a local store: an address, a mount, a path
  and a key inside the secret, plus its own auth (token, AppRole, OIDC) that has to be
  obtained and renewed. `config.Secret.Params()` already carries per-record settings; the
  auth material is the part that needs thinking.
- **gopass / pass** — a GPG-backed tree; an external binary, a store path and a key name.
  Resolving may block on a GPG pinentry prompt while the TUI holds the terminal.
- **1Password** — through the `op` CLI or the Connect API: a vault, an item and a field.
  Resolving can require biometric or SSO approval, so it is interactive and slow.
- **KeePassXC** — a local `.kdbx` file plus an entry path; unlocking needs a master
  password or the browser/CLI integration, and the unlocked state is per session.
- **Keycloak** — not really a secret store. It is an identity provider: it issues tokens,
  it does not keep a database password for you. It fits only if conm grows *token-based
  database auth* (an OIDC access token used as the credential, as Postgres and some managed
  engines allow), which is a different axis from a store and belongs in its own item.

Cross-cutting concerns none of the current code handles:

1. **Resolving can be slow or interactive.** `secret.Resolver.Resolve` is called on every
   ping and every run; a provider that prompts, unlocks or does a network round trip needs a
   timeout, and the UI needs to say it is waiting rather than freezing the row.
2. **Caching and its lifetime.** A short-lived Vault lease or an unlocked KeePass database
   is session state that today has nowhere to live — the resolver is stateless by design.
3. **A record that is not a name and a location.** `config.Keyring` is service plus account;
   Vault and 1Password address a *field inside* an item, which the `Params()` list can carry
   but the secrets screen and the form need to display honestly.
<!-- SECTION:DESCRIPTION:END -->
