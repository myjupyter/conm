# `internal/transport` — a route to a connection that is not directly reachable

## Purpose

Opens and closes the thing that stands between conm and a database it cannot dial directly
— an SSH tunnel today — and hands back the local `host:port` to connect to instead. It runs
the tunnel through a launcher it is given; it never names the `ssh` binary itself.

May import (internal): `config`, `secret`.

## Files

| File | Holds |
|---|---|
| `transport.go` | `Transport`, `New` (the per-kind dispatch), `Through`, `shareKey` |
| `ref.go` | `Reference`, `Ref(kind, name)`, `ParseRef` |
| `provider.go` | `Provider`, `Pool` |
| `ssh.go` | `SSH`, the `launcher` interface, the tunnel lifecycle |
| `errors.go` | `ErrSecret`, `ErrLocalPortBusy`, `ErrNotReady` |

## Main entities

```go
type Transport interface {
    Open(ctx context.Context) (host string, port int, err error)
    Close(ctx context.Context) error
}

type Reference interface{ TransportRef() string }

type Provider interface {
    Resolve(target config.Connection) (Transport, error)
}
```

- `Transport` — `Open` brings the route up and returns where its local end listens; `Close`
  releases it. Both are counted: `SSH` keeps one tunnel for all its users and stops it when
  the last one closes.
- `Reference` — a connection config that names a transport as `"<kind>:<name>"`; `Ref` and
  `ParseRef` are the only spelling of that format here (`spec.JoinTransport`/`SplitTransport`
  on the form side).
- `Provider` — "give me the transport this connection goes through".
  `repository.TransportRepository` implements it over the stored records; `network.Transported`
  consumes it.
- `Pool` — shares one `Transport` between every connection with the same record and target,
  and rebuilds it when the record or its `via` connection changed.
- `launcher` — what actually runs the process; `cli.TransportLauncher` is the implementation,
  passed in by `repository`, which is why this package does not import `cli`.
- `Through(kind)` — the `config.ConnType` a transport kind must go through (`ssh` → an SSH
  connection).

## Adding a transport kind

The kind constant and validation in `config/transport.go`; a `Transport` implementation here
plus its cases in `New`, `Through` and `shareKey`; the process it runs in
`cli/transport.go`'s `transports`; the kind in `spec.TransportKindsOrder`.
