# `internal/cli` — the clients conm launches

## Purpose

The binaries conm shells out to: which clients each connection type offers, what a client
name in `conm.toml` is validated against, which are installed, how each spells its argv and
env, the `exec` that hands one the terminal, the ssh process behind a transport, and the
clipboard reader/writer. **The only package that names a binary.**

May import (internal): `config`.

## Files

| File | Holds |
|---|---|
| `clients.go` | one constant per binary, the per-type lists behind `Clients(t)`, `Validate(t, name)` |
| `command.go` | `commands` (client → argv/env builder), `versionFlags`, `urlCommand`, `Version` |
| `mysql.go`, `mssql.go`, `clickhouse.go`, `ssh.go` | the builders of the clients that take flags rather than a URL |
| `query.go` | `queryArgs` — how each client is handed a single query (`conm <type> exec`) |
| `launcher.go` | `Launcher`: `For(conm, t)`, `Run`, `Info` |
| `detect.go` | `Info`, `Detect`, `Installed`, `Resolve` — the `PATH` lookup |
| `transport.go` | `TransportLauncher` and `transports` — the long-lived client a transport kind runs |
| `clipboard.go` | `ReadClipboard`, `WriteClipboard` |
| `errors.go` | every refusal: `ErrNoClient`, `ErrUnknownClient`, `ErrNoQuery`, `ErrConnType`, `ErrTransport`, `ErrUnknownTransport`, the two clipboard errors |

## Main entities

- `Launcher` — the client the user picked for one connection type, frozen from `conm.toml`.
  `Run(ctx, cfg, password, query...)` execs it; `Info(ctx)` says which client it resolved to.
- `Info{Name, Path, Installed, Version}` — a client as the layers above carry it; no layer
  above spells a binary name to get one.
- `commands` / `versionFlags` / `queryArgs` — the per-client tables; a client missing from
  the first two fails a completeness test.
- `TransportLauncher` — runs the process a `config.Transport` needs (ssh for
  `SSHTransportKind`) until its context ends; `internal/transport` holds it behind its own
  `launcher` interface.

## How it works

`network` knows it launches *a* client; it does not know *which*. `clients.go` is the
vocabulary: one constant per binary, named after the binary, and the
per-type lists behind `Clients(t)`. Order matters — the first entry found on the host is the one used until the user picks
another in settings, so the preferred client goes first. `Validate(t, name)` is what
a write to `conm.toml` is checked against; `config` holds the `ConnectionSettings.CLI` field but
does not know what may go in it.

## The launcher

`For(conm, t)` reads the client `conm.toml` names for that type — or, when it names none,
the first installed one (`Resolve`) — and freezes it into a `Launcher`. `Run(ctx, cfg, password)` takes the connection and the already-resolved secret
— never a pre-built DSN, because a client is free to spell the connection its own way. It
re-checks the name against `Clients(t)` before it execs: `conm.toml` is a file a user
edits, and nothing validates the client on read.

`Info(ctx)` is the same launcher answering *which* client it resolved to, as the value the
layers above carry: the name, its path, whether it is installed and its version. A launcher
with nothing configured returns the zero `Info`, and a name that is not a client of its
database returns the name alone, never installed — the refusal `Run` makes, so a
hand-edited `conm.toml` cannot render as ready on the strength of a coincidental binary on
`PATH`. `repository.Connections.ClientInfo(ctx)` and `view.Connections.ClientInfo(ctx)` are the
same value one and two layers up; nothing above this package spells a binary name to get it.

`Detect(t)` is the `PATH` lookup behind the type picker and the settings screen: one `Info` per candidate name in
`Clients(t)`, with the resolved path and whether it is installed.

## One entry per client, no default

`commands` in `command.go` maps a client's binary name to the function that turns a
connection into argv and env. **Every client conm can launch is in it** — a name with no
entry is `ErrUnknownClient`, never a guess. Clients that want nothing but the URL register
against the shared `urlCommand` rather than falling into it, so the map reads as the whole
list of what conm knows how to start, and `TestEveryClientHasACommand` fails the moment a
name is added to `clients.go` without one:

| Client | What it is handed |
|---|---|
| `psql`, `pgcli`, `usql`, `mongosh`, `mongo`, `mycli` | `cfg.ConnectionString(password)` as one argument |
| `redis-cli`, `valkey-cli` | `-u <url>` |
| `iredis` | `--url <url>` |
| `mysql` | flags off `config.MySQL`, password via `MYSQL_PWD` |
| `sqlcmd` | flags off `config.MSSQL`, password via `SQLCMDPASSWORD` |
| `clickhouse` | the `client` subcommand plus flags off `config.ClickHouse`, password via `CLICKHOUSE_PASSWORD` |
| `ssh` | flags off `config.SSH` (`-p`, `-J` — or a `ProxyCommand` when the password goes via askpass —, `-A`, `-o ServerAliveInterval=`, `-L`), then `user@host`; for key auth the resolved secret is `-i <path>`, or, when it is the key material itself (`-----BEGIN …`), a temp file removed once ssh exits; for password auth it goes via `SSH_ASKPASS`, with public-key auth switched off and `StrictHostKeyChecking=accept-new` so the forced askpass is never asked the host-key question |

A flag-spelling builder reads the fields off the concrete config type
(`cfg.(config.MySQL)`), not off a re-parsed URL: the structured values are what config
persisted, and going through the URL only to split it apart again loses the port and the
TLS mode into strings. A multi-call binary names its subcommand as the first argument —
`clickhouse` is what conm looks for on `PATH`, `client` is what it passes.

**Prefer an out-of-band password over one in argv** where the client supports it: an
environment variable keeps the secret out of the process table, while a client invoked
with a URL necessarily carries it, exactly as `psql` does. `passwordEnv` appends to
`os.Environ()`, so a client that reads a variable still inherits the rest of the
environment.

`ssh` never reads a password from argv or from a variable of its own, so `sshCommand` goes
through the one door OpenSSH leaves open: `SSH_ASKPASS_REQUIRE=force` makes ssh 8.4+ run
the program in `SSH_ASKPASS` instead of prompting. `sshAskpass` makes a fresh 0700
directory per run holding two files: `password` (0600) and a fixed two-line `askpass` script
that `cat`s its sibling. The command's `cleanup` removes the directory once ssh exits, as it
removes a temp key file. The password never travels in an environment variable — only the
two `SSH_ASKPASS*` variables do, which name the program and carry no secret. `agent` and
`key` auth set no askpass at all — the agent, or ssh's own passphrase prompt, is the
credential.

**The askpass reaches the target only.** `-J` makes ssh start a child ssh for each jump
host, and that child inherits the environment — so with the askpass forced, every prompt a
bastion raises (its password, a key passphrase, the `known_hosts` yes/no) would be answered
with the target's password. When the askpass is on and a jump is set, `sshJumpWithoutAskpass`
spells the jump as the `ProxyCommand` `-J` stands for — `ssh [-J earlier hops] -W %h:%p
ssh://<last hop>` — behind `env -u SSH_ASKPASS -u SSH_ASKPASS_REQUIRE`,
so the jump hosts prompt on the terminal as with a plain `ssh -J`. The last hop goes as an
`ssh://` URI because a plain destination cannot carry a `:port`. The hops reach a shell
unquoted, which is safe only because `config.ValidateSSHJump` admits nothing but
`[user@]host[:port]` of `[A-Za-z0-9._-]`. Without the askpass `-J` is kept as is.

**The password enters this package as an argument and leaves only inside `argv`/`env`.**
Nothing here puts it in an error, and `Command` never reaches a caller: `Launcher.Run`
builds it, execs it and returns.

## The version probe

`Version(ctx, name)` runs the client once and reads the version off the first line it
prints. It is the only thing in this package that starts a subprocess without the user
asking to connect, so it is deliberately kept off the listing path: `Detect(t)` stays a
`PATH` lookup and leaves `Info.Version` empty, and a caller that wants versions asks for
them per client, after it knows which ones exist. `Launcher.Info(ctx)` is the one path that
fills it — a single client, already resolved, already known to be installed — which is why
it takes a context and every accessor carrying it up does too.

`versionFlags` in `command.go` is how each client spells the request, one entry per client
beside `commands`. Every database client answers to `--version`; `ssh` answers to `-V`
and prints to stderr, which is why the probe falls back to stderr when stdout is empty. A client
that disagrees is one entry in that map rather than a branch in the probe, and `TestEveryClientHasAVersionFlag`
fails the moment a name is added to `clients.go` without one.

Only the first line is parsed, and only the first `N(.N)*` run on it survives, so the
product name and the trailing build noise both drop: `psql (PostgreSQL) 16.2` and
`redis-cli 7.2.4 (git:0)` yield `16.2` and `7.2.4`. What a client calls its version is what
comes back — `mysql  Ver 14.14 Distrib 5.7.44` is a client whose first number is the tool's
own, not the server's.

**There is no error.** An unknown name, a binary that is not installed, a non-zero exit, a
cancelled context and a line with no number in it all return the empty string; the caller
renders that as unknown. The child's stdin is closed, so a
client that would rather prompt than answer reads EOF instead of hanging, and the context
kills anything that still waits.

## The clipboard

`ReadClipboard()` and `WriteClipboard(text)` are the other binaries conm shells out to, and
they are here for the same reason the database clients are: **naming a binary is this
package's job**, so `ui` asks for the clipboard rather than deciding that a Wayland desktop
spells it `wl-copy`. `pbpaste`/`pbcopy` on darwin; elsewhere the first of `wl-paste`/
`wl-copy`, `xclip`, `xsel` that is on `PATH`, through the same `lookup` the client detection
uses. A host with none of them gets `ErrNoClipboardReader`/`ErrNoClipboardWriter`, which the
screen renders as "install wl-clipboard, xclip or xsel" rather than failing silently.

These two are not in `clients.go`, `commands` or `versionFlags`: they are not clients of a
database, nothing validates a `conm.toml` against them, and they never take a connection or
a password. What comes back is the raw clipboard payload — flattening it into something a
field can hold stays in `ui/clipboard_commands.go`, since what is a legal value is the form's question.

## A single query — `query.go`

`Launcher.Run` takes an optional query (`conm <type> exec <name> -- <query>`). `queryArgs`
maps a client to how it is handed one: `flagQuery("-c")` joins the words behind the client's
flag (`-c`, `-e`, `-Q`, `--query`, `--eval`), `plainQuery` appends them as they are
(`redis-cli GET key`, `ssh host uptime`). The words are appended after the argv the client's
builder produced, so the builder never learns a query exists. A client with no entry
refuses with `ErrNoQuery`.

## The transport process — `transport.go`

`TransportLauncher.Run` starts the long-lived process a `config.Transport` needs and blocks
until its context ends; `transports` maps a `config.TransportKind` to the client and its
builder, the way `commands` does for interactive clients. For `ssh` that is
`sshTransportCommand`: the same login as `sshCommand` (`sshLogin` — port, identity or
askpass, jump), then `-N`, `ExitOnForwardFailure=yes`, the transport's connect timeout and
keepalive, and `-L <local>:<target host>:<target port>`. `BatchMode=yes` is added whenever
no askpass is in play, so a tunnel that cannot authenticate fails instead of prompting on
a terminal it does not own.

The process runs in its own process group, is asked to stop with `SIGTERM` when the context
is cancelled and killed after `transportStopGrace`. A cancelled context is a normal end and
returns nil; any other exit is `ErrTransport` carrying what the process wrote to stderr.

## The refusals — `errors.go`

`network` maps the configuration failures among them to `EINVALID` through `cliErrorCode`
in `network/error.go` before the driver's own matcher sees the error.

- `ErrNoClient` — no client is configured for that connection type and none is installed.
- `ErrUnknownClient` — the configured name is not a client of that type, or has no
  entry in `commands`. The first half is a hand-edited `conm.toml`; the second cannot
  survive `TestEveryClientHasACommand`.
- `ErrConnType` — a builder was handed a connection it cannot spell. Since `Run` has
  already checked the name against the type, this only fires when a `config.Connection`
  disagrees with its own `ConnType()` — a registration bug in a new type, and it fires on
  that type's first run.
- `ErrNoQuery` — the client cannot run a single query.
- `ErrTransport` / `ErrUnknownTransport` — the transport process failed, or the transport
  kind has no entry in `transports`.
- `ErrNoClipboardReader` / `ErrNoClipboardWriter` — no clipboard binary on the host.
