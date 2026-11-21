# AGENTS.md

The entry point of this repository: what it is, what to do after a change, and where the
rest is written down. **The architecture lives in `internal/AGENTS.md` and the detail of
each module in the `AGENTS.md` inside that module — put new detail there rather than here.**

## What this is

`conm` is a terminal connection manager for databases (Postgres, MySQL, SQL Server,
ClickHouse, Redis, MongoDB) and SSH hosts: a Cobra CLI whose main screen is a Bubble Tea TUI
over saved connections — ping them, open the interactive client (`psql`/`pgcli`/`usql`,
`mysql`, `sqlcmd`, `clickhouse client`, `redis-cli`, `mongosh`, `ssh`, …), add/edit/remove
entries, keep their passwords in a secret store, reach a database through an SSH transport,
and keep separate sets of all of that in workspaces. Everything is persisted as TOML.

- Module `github.com/myjupyter/conm`; the Go version is the one in `go.mod`.
- The TUI is Bubble Tea **v2** under the `charm.land` import path
  (`charm.land/bubbletea/v2`, `bubbles/v2`, `lipgloss/v2`) — not the v1 API or the old
  `github.com/charmbracelet` path. Every other dependency is in `go.mod`.

Runtime files — paths come from `internal/config/xdg.go` (`config.ConmPath()`,
`config.PostgresPath()`, …), never hardcode them:

- `$XDG_CONFIG_HOME/conm/conm.toml` — global settings: the client per connection type, the
  keybinding overrides.
- `$XDG_CONFIG_HOME/conm/<workspace>/<type>.toml` — the connections of that type;
  `transport.toml` beside them.
- `$XDG_DATA_HOME/conm/<workspace>/secret.toml` — secret records: a location and metadata,
  never the password material.

## After changing code

Run all four, in this order, and fix what they report before calling the work done:

```sh
make tools      # install the pinned golangci-lint, lefthook, … into bin/tools
make build      # -> bin/<os>-<arch>/conm
make lint       # go vet + golangci-lint run ./...  (config in .golangci.yaml)
make test       # go vet + go test -race ./...
```

`make fmt` (`go fix` + `golangci-lint fmt`) fixes formatting. The same four targets run on
`pre-push` through `lefthook.yml`; commit messages follow Conventional Commits and branches
are `<type>/<lowercase-description>`.

Then the documentation, in the same piece of work:

- **A new feature or any change to user-visible behaviour populates the user
  documentation** — the page under `docs/user/` that owns it, `docs/user/commands.md` for a
  command, the demo GIF that shows it. How: `docs/user/AGENTS.md`.
- **A change to a module's files, entities, rules or seams updates that module's
  `AGENTS.md`**; a new package gets one. How: `docs/agents/AGENTS.md`.

## Code style

**Do not write comments** — inline, block, doc, `// TODO` or section headers. The code must
read on its own: pick names that say what a value is, split a function that needs an
explanation, let the types carry the intent. The non-test code holds none; keep it that
way. If something genuinely cannot be expressed in code — an external constraint, a
non-obvious protocol detail — it belongs in the module's `AGENTS.md`, not next to the
statement.

Keep everything gofmt-clean, vet-clean and lint-clean. Tests are table-driven and assert
through **`testify`** — `require` for the preconditions a failure would make the rest of the
test meaningless past, `assert` for the verifications, bound as `is`/`must` where a subtest
makes more than a couple. Do not hand-roll `reflect.DeepEqual` or `if got != want`;
`testifylint` guards argument order and the assert/require split.

## Architecture

The layers, the import rules between them, the hard rules that cross modules, the connection
lifecycle and the extension seams are in `internal/AGENTS.md`. **Read it before changing
code anywhere, `cmd` included** — then the `AGENTS.md` of the module you are about to touch.

## Documentation

| Kind | Where | Guide |
|---|---|---|
| For users | `README.md`, `docs/user/`, GIFs in `docs/demo/` | `docs/user/AGENTS.md` |
| For agents | this file, `internal/AGENTS.md`, one `AGENTS.md` per module | `docs/agents/AGENTS.md` |

## Task tracking

Tasks are tracked with Backlog.md under `.backlog/`. **Read `.backlog/AGENTS.md` before
creating, picking up or closing a task.**
