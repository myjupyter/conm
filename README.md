# conm

[![Go version](https://img.shields.io/github/go-mod/go-version/myjupyter/conm)](go.mod)
[![Build](https://github.com/myjupyter/conm/actions/workflows/go.yml/badge.svg)](https://github.com/myjupyter/conm/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/myjupyter/conm)](LICENSE)

> [!CAUTION]
> conm has no stable release yet. Commands, keys and the config file format can change between
> versions without a migration path.

conm is a terminal connection manager for databases and SSH hosts. 
It keeps your connections organized

![conm demo](docs/demo/demo.gif)

- **Supported database types:** Postgres, MySQL, SQL Server, ClickHouse, Redis and MongoDB, saved
  and ready to open.
- **SSH hosts:** kept right alongside your databases.
- **Passwords out of your configs:** they can stay in your system keyring or keychain.
- **Private networks:** tunnel to a database through an SSH host.
- **Workspaces:** keep each project or environment to itself.
- **Links:** keep dashboards, docs and runbooks next to the connection they belong to.

## Install

```sh
go install github.com/myjupyter/conm@latest
```

Needs Go 1.26, plus the client for each database you connect to.

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

## Documentation

The [user guide](docs/user/README.md) walks through conm in order, from adding the first
connection to tunnels and workspaces.

## License

[MIT](LICENSE)
