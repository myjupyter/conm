---
id: CONM-4
title: >-
  Carry the TLS settings a connection needs — CA, certificates, ciphers,
  versions, SNI
status: To Do
assignee:
  - myjupyter
created_date: '2026-09-06 14:01'
updated_date: '2026-09-06 14:56'
due_date: '2026-09-06'
labels:
  - not-ready
milestone: m-0
dependencies: []
priority: medium
type: enhancement
ordinal: 4000
---

## Description

<!-- SECTION:DESCRIPTION:BEGIN -->
Every type's TLS field today is a *mode* and nothing else: `config.Postgres.SSLMode`,
`config.MySQL.TLSMode`, `config.Redis.TLSMode`, `config.MSSQL.EncryptMode` plus `TrustCert`,
`config.ClickHouse.Secure`, `config.MongoDB.TLS`. A mode says how hard to verify. It does
not say *against what* — a private CA, the normal case for RDS, Cloud SQL and anything
self-hosted, or a mutual-TLS client certificate. Nor *with which* — a cipher list or a
protocol-version floor, which a hardened server may require and a FIPS or PCI estate
mandates. Nor *under which name* — the SNI host, which a proxied or multi-tenant endpoint
routes on. None of the three has a field to land in, and none can be smuggled into an
existing one: the TOML keys are fixed per type and there is no free-text parameter bag
anywhere on `config.Connection`.

What that costs today:

- `network` dials with each driver's defaults, so a verifying mode checks the system root
  store. A server behind a private CA fails to ping where the user's own client succeeds,
  and `x509: certificate signed by unknown authority` blames the server rather than the
  missing CA.
- `internal/parser/params` refuses any command line carrying it — mysql's
  `--ssl-ca`/`--ssl-capath`/`--ssl-cert`/`--ssl-key`/`--ssl-crl`, redis's
  `--cacert`/`--cacertdir`/`--cert`/`--key`, postgres's
  `sslrootcert`/`sslcert`/`sslkey`/`sslcrl`/`sslpassword` — with `ErrUnsupportedTLS`, rather
  than seeding a form that would save a connection unable to ping. The same refusal covers
  the TLS settings that are not certificates but are just as unrepresentable: cipher lists
  (`--ssl-cipher`, `--tls-ciphersuites`), protocol versions (`--tls-version`,
  `ssl_min_protocol_version`/`ssl_max_protocol_version`) and SNI (`--sni`, `sslsni`). That
  refusal is this item's placeholder: once the fields exist, each parser's `*TLSRefusals`
  table becomes an applier and `ErrUnsupportedTLS` goes away.
- MySQL's `VERIFY_CA` and `VERIFY_IDENTITY` both collapse onto the driver's `tls=true`.
  Chain-only versus chain-plus-hostname is not representable, and a custom CA, a cipher list
  and a version bound are reachable in `go-sql-driver` only through a *registered*
  `tls.Config` name, never through a DSN value — so this one needs a driver-side seam, not
  just a field.

What it touches:

1. **The model.** Three paths — CA bundle, client certificate, client key — plus the
   passphrase a key may carry, and three settings that are values rather than files: a
   cipher list, a protocol-version floor and ceiling, and an SNI host. A passphrase is
   credential material, so it belongs behind `secret`, and the single `SecretRef()` cannot
   hold a second one; that is the same limit item 2 hits for SSH keys, and the two should be
   decided together. The open question is whether TLS settings live on the connection or in
   a named profile several connections point at — a profile matches how one CA bundle and
   one cipher policy are shared across a fleet, and it is the shape that keeps seven new
   fields off every connection form.
2. **The drivers.** Each takes its material differently: postgres by conninfo keyword,
   `go-sql-driver/mysql` only via `mysql.RegisterTLSConfig` under a generated name,
   clickhouse via `clickhouse.Options.TLS`, mongo through `options.ClientOptions`, sqlserver
   by query parameter. `network` gains one step that builds a single `*tls.Config` and hands
   each client its own spelling of it.
3. **The CLI child process.** `Run` execs the real client, which takes file paths rather
   than a `tls.Config`, so the paths must survive into the argv `network` already builds per
   client — and a path that only conm can read is a failure the child reports, not us.
4. **The form and the parser.** More fields per type in `ui/spec` — wanting a path field
   kind that validates the file is readable, and a plain text one for the cipher list — and
   each parser's `*TLSRefusals` table turns from a refusal into an applier.
<!-- SECTION:DESCRIPTION:END -->
