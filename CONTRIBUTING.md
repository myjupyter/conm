# Contributing to conm

Thanks for taking the time to contribute. Bug reports, ideas, documentation fixes and code
are all welcome.

## How to contribute

### Report a bug

Open an [issue](https://github.com/myjupyter/conm/issues) with:

1. **Your system** — the output of `uname -a`, `conm --version`, your terminal emulator, and
   the version of the client involved (`psql --version`, `ssh -V`, …) if the bug is about
   connecting.
2. **Steps to reproduce** — numbered, one action per step, starting from launching `conm`.
3. **What you expected** and **what happened instead** — include the error text or a
   screenshot.

Remove passwords, hostnames and anything else private from what you paste.

### Suggest an enhancement or a feature

1. Search the [existing issues](https://github.com/myjupyter/conm/issues?q=is%3Aissue) first,
   closed ones included. If your idea is already there, add your use case to that issue.
2. If nothing matches, open a new issue describing the idea: the problem it solves, how you
   imagine using it, and any alternatives you considered.

### Improve the documentation

Fix typos, clarify setup steps or add examples directly in a pull request — no issue needed.
User documentation lives in `README.md` and `docs/user/`.

### Submit code

Code changes should be connected to an issue. Pick an existing one or open one first, and
say in it that you are working on it — this is the moment to agree on the approach before
you write the code. Reference the issue in the pull request (`Closes #123`).

## Local development

1. Fork the repository and clone your fork:

   ```sh
   git clone git@github.com:<your-username>/conm.git
   cd conm
   ```

2. Install [Go](https://go.dev/doc/install) — the version in `go.mod` or newer — then the
   pinned development tools and the git hooks:

   ```sh
   make tools   # golangci-lint, lefthook, commitlint, … into bin/tools
   make hooks   # checks commit messages and branch names, runs the checks on push
   ```

3. Build it and check that the binary runs:

   ```sh
   make build
   ./bin/<os>-<arch>/conm --version   # e.g. ./bin/darwin-arm64/conm --version
   ```

4. Before pushing, make sure the checks pass:

   ```sh
   make lint
   make test
   ```

   `make fmt` fixes formatting.

## Commits

Commit messages follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

```
<type>[optional scope]: <description>
```

- `type` says what kind of change it is: `feat` for a new feature, `fix` for a bug fix,
  `docs`, `refactor`, `test`, `chore`, …
- `scope` optionally names the part of the code that changed.
- `description` is a short summary in the imperative mood — "add", not "added".
- A breaking change is marked with `!` before the colon.

```
feat: add alias for database types
fix(ssh): close the tunnel when the client exits
docs(README): add homebrew installation line
```

Write clear commit messages, and do not put everything into one commit when the change
covers several features — one commit per feature is easier to review and to revert.

## Creating a pull request

1. Create a branch from `main` named by the rules of
   [Conventional Branch](https://conventionalbranch.org/) — `<type>/<lowercase-description>`:

   ```sh
   git checkout -b feat/redis-cluster-support main
   git checkout -b fix/ssh-tunnel-leak main
   ```

   The accepted types are `feature`/`feat`, `bugfix`/`fix`, `hotfix`, `release` and `chore`;
   the description uses lowercase letters, digits and hyphens.

2. Commit your changes and push the branch to your fork.
3. Open a pull request against `main`. Describe what changed and why, and link the issue.
4. Keep the pull request focused on one thing; unrelated changes go in a separate one.
