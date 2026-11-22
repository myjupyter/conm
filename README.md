<p align="center">
  <img src="logo.png" alt="conm logo" width="200">
</p>

[![Go version](https://img.shields.io/github/go-mod/go-version/myjupyter/conm)](go.mod)
[![Build](https://github.com/myjupyter/conm/actions/workflows/go.yml/badge.svg)](https://github.com/myjupyter/conm/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/myjupyter/conm)](LICENSE)

> [!CAUTION]
> conm has no stable release yet. Commands, keys and the config file format can change between
> versions without a migration path.

![conm demo](docs/demo/demo.gif)

## What is this

conm is a terminal connection manager for databases and SSH hosts. It doesn't try to replace
the tools you already use — `psql`, `mysql`, `redis-cli`, `mongosh`, `ssh` and the rest. It
brings them together into one organized place to reach and manage your connections.

## What problem it solves

You live in the terminal and work through your favourite clients, but the number of
connections keeps growing: hosts, ports, users and passwords scattered across shell history,
notes and config files. conm keeps them in one place, so you can move between them, store
them, and keep them isolated from one another.

It also takes over the routine around opening a connection: no bringing up an SSH tunnel by
hand before reaching a database behind it, and no typing or pasting a password or secret
every time — conm resolves it and hands it to the client for you.

- **Databases:** Postgres, MySQL, SQL Server, ClickHouse, Redis and MongoDB.
- **SSH hosts:** kept right alongside your databases.
- **Passwords out of your configs:** they can stay in your system keyring or keychain.
- **Private networks:** tunnel to a database through an SSH host.
- **Workspaces:** keep each project or environment to itself.
- **Links:** keep dashboards, docs and runbooks next to the connection they belong to.
- **Your look:** borders, glyphs, colours and gradients, text styles and the table's columns —
  [make conm yours](docs/user/customisation.md).

## Inspiration

- [sshm](https://github.com/Gu1llaum-3/sshm) — the direct inspiration: a terminal manager for
  SSH hosts. conm grew out of wanting the same for every other kind of connection.
- [Teleport](https://goteleport.com) — solves access to many resources too, but through
  centralized, server-side access control rather than a personal tool over your own clients.
- [sqlit](https://github.com/Maxteabag/sqlit) — works with many databases from one terminal
  UI, but as its own client; conm keeps the clients you already use.

## Documentation

The [user guide](docs/user/README.md) walks through conm in order, from adding the first
connection to tunnels and workspaces.

## Install

#### Homebrew
```sh
brew install myjupyter/conm/conm
```

#### Golang
```sh
go install github.com/myjupyter/conm@latest
```


Needs the Go version in `go.mod`, plus the client for each database you connect to. Then run `conm`: with no
connections yet it asks for the type of the first one.

## Supported

| Type | Clients |
|---|---|
| `postgres` | psql, pgcli, usql |
| `mysql` | mysql, mycli, usql |
| `mssql` | sqlcmd, usql |
| `clickhouse` | clickhouse, usql |
| `redis` | redis-cli, valkey-cli, iredis |
| `mongodb` | mongosh, mongo |
| `ssh` | ssh |



## License

[MIT](LICENSE)
