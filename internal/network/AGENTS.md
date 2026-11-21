# `internal/network` — live connections and their failures

## Purpose

Turns a `config.Connection` into a live one that can be pinged, opened in its client and
closed. Resolves the secret lazily, per call, and describes every failure as a
`*network.OpError`. Never imports `repository` or `ui`, and never names a client binary.

May import (internal): `config`, `secret`, `cli`, `transport`.

## Files

| File | Holds |
|---|---|
| `conn.go` | `Connection`, `PingResult`, `NewConnection` — the per-type dispatch |
| `error.go` | `Operation`, `ErrorCode`, `OpError`, `transportErrorCode`, `cliErrorCode` |
| `postgres.go`, `mysql.go`, `mssql.go`, `clickhouse.go`, `redis.go`, `mongodb.go`, `ssh.go` | one client each (`PGClient`, `MySQLClient`, …): its driver DSN, its target, its error matcher |
| `no_conn.go` | `NoClient` — the stub for a record that does not validate: displays, errors on `Ping`/`Run` |
| `transported.go` | `Transported`, `NewRouted` — a database connection reached through a transport |

## Main entities

```go
type Connection interface {
    Ping(ctx context.Context) (PingResult, error)
    Run(ctx context.Context, query ...string) error
    Close() error
}
```

- A `Connection` carries no config: a screen that needs a name, a host or a port reads the
  `config.Connection` beside it. `Ping` and `Run` failures wrap `*OpError`; `Run` execs the
  configured client on the terminal, or hands it one query.
- `NewRouted` is the constructor repositories call: a database connection whose config
  carries a `transport.Reference` becomes a `Transported`, which opens the transport, moves
  the config to the local end (`SetHost`/`SetPort`) and builds the plain client over it;
  everything else goes straight to `NewConnection`.

## Connection strings and what a driver takes

**A connection string is one thing; what a driver takes is another.** `ConnectionString`
always emits the URL form, because that is what a config layer can spell without knowing a
driver. Where a driver wants something else, its own file re-spells it and nothing shared
moves:

| Type  | Driver takes |
|-------|--------------|
| Postgres | the URL as-is |
| Redis | the URL, via `redis.ParseURL` |
| MySQL | `user:pass@tcp(addr)/db` — `mysqlDriverDSN` |
| MSSQL | the URL as-is, plus a dial timeout — `mssqlDriverDSN` |
| ClickHouse | the URL via `clickhouse.ParseDSN`, plus a dial timeout — `clickhouseDriverDSN` |
| MongoDB | the URI via `options.Client().ApplyURI`, plus a dial and a server-selection timeout |
| SSH | nothing — there is no driver; `Ping` dials `host:port` and reads the `SSH-` banner line, which needs no credential and no dependency. With a jump chain set it pings the **first hop**, not the target (CONM-15) |

What the *client* takes is not here: `Run` resolves the secret and hands the connection to
a `cli.Launcher`, which owns every binary name, argv and environment variable — see
`internal/cli/AGENTS.md`. The `*DialTimeout` constants are set on the driver because the UI pings with
`context.Background()`, so without them an unreachable host would hang the row until the OS
gave up.

`Slowest`/`Slow`/`Medium`/`Fast` in `conn.go` are the ping-time thresholds
`PingResult.Bars()` turns into the signal bars a row shows.

A redis connection that requires TLS against a server that does not speak it fails as
`ETIMEDOUT`, not `ETLS`: the handshake hangs rather than producing matchable wording.
Real certificate failures still carry `x509`/`certificate` and land on `ETLS`. A mongodb
one behaves the same way, for the same reason.

The mongodb driver reports every dial-time failure as one server selection error whose text
always contains "context deadline exceeded", with the real cause quoted inside it. That
buries the wording `transportErrorCode` orders first, so `mongodbErrorCode` matches the
certificate cause the envelope masks — `x509`/`certificate` — before it falls back to the
shared matcher.

## A connection behind a transport — `transported.go`

`NewRouted` is what a repository calls. A database config that carries a transport
reference becomes a `Transported`; anything else goes straight to `NewConnection`.
`Transported` holds no tunnel of its own: on every `Ping` and `Run` it asks the
`transport.Provider` for the transport, opens it, re-points a copy of the config at the
local end (`SetHost`/`SetPort`), builds the ordinary client over that copy, and closes the
transport when the call returns. The transport is shared and counted in
`internal/transport`, so two calls at once use one tunnel.

Two things keep the tunnel out of what the user reads:

- **A failure names the configured address, not the local one.** `configuredAddress`
  rewrites `127.0.0.1:<port>` back to the connection's own `host:port` in the inner
  `OpError`'s `Target` and `During`, and sets `Through` to the transport reference.
- **A failure of the transport itself is the transport's code**, not the driver's:
  `ETRANSPORT` by default, `ESECRET` for `transport.ErrSecret`, `EADDRINUSE` for
  `transport.ErrLocalPortBusy` — named by `Transported.fail` before any driver runs.

## Failures — `*OpError`

**A client describes its own failure.** `Ping` and `Run` never return a driver's error bare:
they wrap it in a `*network.OpError`, so every layer above reads structured fields instead of
error text.

```go
type OpError struct {
    Op     Operation   // PingOperation | ConnectOperation — what was attempted
    Code    ErrorCode  // ECONNREFUSED · ETIMEDOUT · EHOSTUNREACH · ENOTFOUND · ELOCALNET · EAUTH ·
                       // ESECRET · ETLS · EINVALID · ETRANSPORT · EADDRINUSE · ECONN
    Target  string     // password-free, in whatever form the driver speaks
    During  string     // "dial tcp host:port", "open session on host:port"
    Through string     // the transport reference, when the connection is routed
    Err    error       // the driver's own error; Error() and Unwrap() defer to it
}
```

`error.go` holds the vocabulary — `Operation`, `ErrorCode`, `OpError` and `transportErrorCode`,
the codes any TCP client can hit. `transportErrorCode` reads the error's *type* first
(`errnoErrorCode`: `syscall.ECONNREFUSED`, `EHOSTUNREACH`, a `*net.DNSError` that is
not-found, a timeout) and falls back to wording only when nothing typed matched. On macOS an
`EHOSTUNREACH` towards a private or link-local address is `ELOCALNET`, not
`EHOSTUNREACH`: that is the system's local-network permission refusing the dial, and the
advice is a different one. **The driver-specific half lives in that driver's file**:
`postgres.go` has `pgErrorCode` (auth and sslmode wording, falling back to
`transportErrorCode`), `pgTarget`, and the `fail` method every `PGClient` return path goes
through; `mysql.go` has `mysqlErrorCode`, which reads the server's error **number** off a
`*mysql.MySQLError` (1044/1045/1698 → `EAUTH`, 1049 → `EINVALID`) rather than its wording —
matching a number is always preferable to matching text when the driver offers one —
plus `mysqlTarget` and `mysqlDriverDSN`; `mssql.go` does the same off
`mssql.Error` (18456 → `EAUTH`, 4060/4063/911 → `EINVALID`), as does `clickhouse.go` off
`*clickhouse.Exception` (516/192/497 → `EAUTH`, 81/82 → `EINVALID`); `redis.go` has
`redisErrorCode` (`WRONGPASS`/`NOAUTH`/`NOPERM` wording) and
`redisTarget`; `mongodb.go` has `mongodbErrorCode`, which reads the codes off a
`mongo.ServerError` (18/13/11 → `EAUTH`) and falls back to the auth wording the driver
produces when the handshake fails before any command runs, plus `mongodbTarget`.
`no_conn.go` does the same for the invalid-config stub, with a generic
`host:port` target. A new client type adds its own three pieces and nothing central
changes.

**The code is the call site's to name whenever the call site already knows it.** Matching a
driver's wording is the last resort, not the rule: a `fail` takes the `ErrorCode` as an
argument, so resolving the secret failing is `ESECRET` and a DSN the driver refuses to parse
is `EINVALID` — both known positionally, before any text is read. Only the error that comes
back from the wire goes through the driver's own matcher. An error out of `cli.Launcher`
is read the same way: `cliErrorCode` turns `cli.ErrNoClient`, `cli.ErrConnType` and
`cli.ErrNoQuery` into `EINVALID` before the driver's matcher would try to read a configuration failure as a
transport one.

Above `network`, nobody matches on an error's text: `ui/conn_error.go` recovers the failure
with `errors.AsType[*network.OpError]` and flattens it into `connError`, the display strings
the panel draws. That file also holds `hintFor` — advice keyed by `ErrorCode`, the only part
of a failure the UI owns, since it spells the key to press from `keyMap`. An error that is
*not* an `OpError` (the client could not be built at all) still renders: the screen supplies
the operation it attempted and the code falls back to `ECONN`.
