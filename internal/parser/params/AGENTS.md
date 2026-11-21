# `internal/parser/params` — a client command line back into a connection

## Purpose

The inverse of `ConnectionString`: takes what a user would type at a shell —
`psql "postgres://…"`, `psql "host=… dbname=…"`, `pgcli --host … --port …` — and returns the
`config.Connection` it describes, so a form opens filled in instead of empty. No I/O, no
screen, no knowledge of forms: a pure function of its input.

May import (internal): `config`, `cli`, `secret`.

## Files

| File | Holds |
|---|---|
| `parser.go` | `Syntax`, `Result`, the sentinel errors, `Parse`/`ParseArgs`, the `request` a command line is reduced to, the registries `parsers` (`config.ConnType` → `parseFunc`) and `schemes` (URI scheme → `config.ConnType`), the flags and warnings every parser shares |
| `lexer.go` | the text half shared by every type: `splitArgs`, `scanFlags` (driven by `flagArity`), `splitKeywords`, the host/port/percent-decoding helpers |
| `postgres.go`, `mysql.go`, `mssql.go`, `clickhouse.go`, `redis.go`, `mongodb.go`, `ssh.go` | one type each: its flag, env and keyword tables, its URI applier and `build<Type>` |
| `examples.go` | `Examples(t)` — the sample command lines `conm <type> add --help` prints |

## Main entities

- `Parse(t, input)` — one quoted command line, tokenised here. `ParseArgs(t, args)` — tokens
  the shell already split. Not interchangeable: rejoining tokens loses the quoting.
- `Result{Syntax, Client, Conn, Warnings}` — the connection, how it was spelled, the client
  named in front of it, and what was dropped on the way.
- Errors: `ErrEmptyInput`, `ErrTypeMismatch`, `ErrUnsupported`, `ErrUnterminated`,
  `ErrMalformedKeyword`, `ErrUnsupportedTLS`.

## Rules

**It knows nothing about forms.** A `Result` carries a `config.Connection`, and turning that
into form values is already `spec.FormSpecs[conn.ConnType()].SeedFunc` — the caller that owns
both, in `cmd` or `ui`, spells that call. Reaching for `ui/spec` from here would point an
arrow outward for a one-line lookup the form layer already owns.

**A further type adds its own file and nothing central moves** — `dsnFlag`/`hostFlag`/
`portFlag`/`userFlag`, `namedDSNWarning`, `socketWarning`, `hostListWarning` and `tlsRefusal`
in `parser.go` are what the parsers already share, because four clients spell `--host` the
same way and a dropped socket, a seed list or a refused CA reads the same whoever dropped it.

**A flag's arity is three-valued, not two.** `scanFlags` asks its caller for a `flagArity` —
`noValueFlag`, `valueFlag` or `attachedValueFlag` — because `mysql -p` takes its password
only when it is attached: `mysql -h db -u root -p shop` prompts for the password and opens
`shop`, so reading `shop` as the password would both lose the database and invent a
credential. `attachedValueFlag` keeps `-psecret` and `--password=secret` and leaves the next
token positional. mycli parses the same flag with click, where `-p secret` does take its
value, so `mysqlArity` closes over `req.client` for that one decision the way
`applyRedisFlags` does for `-d`.

**`conm <type> add [command line]` ends on the connections, not in the shell.** Someone
adding a connection is on their way to opening one, so a form that saved hands over to
`ui.RunOn(conm, t)` — the connections table of that database. A cancelled form is the one
exit: nothing was added, so there is nothing to connect to.

**Two entry points, because a command line reaches conm in two shapes.** `Parse` takes it as
one quoted string and tokenises it; `ParseArgs` takes tokens the shell already split. They
are not interchangeable: rejoining tokens with a space loses the quoting, so
`options='-c search_path=x'` would come back as two keywords instead of one. `cmd` picks by
argument count — one argument is a quoted command line, several are its tokens.

**The database is told, never guessed.** `Parse(t, input)` and `ParseArgs(t, args)` both
take the `config.ConnType` as their first argument, and `parsers[t]` is the only thing that
decides which parser runs. Every type in `config.ConnTypes` has an entry, so `ErrUnsupported`
answers only a `ConnType` that is not one of them — a new type between step 1
and step 7 of the `add-new-database-type` skill. Guessing would make the same command line mean different things
on different hosts, so recognition exists only to *contradict* the caller, never to choose
for them:

- a leading client binary is dropped when it is one of `cli.Clients(t)` (so `psql …`,
  `pgcli …` and a bare `-h … -U …` all work), and rejected with `ErrTypeMismatch` when it
  belongs to another database. `knownClients` is built from `cli.Clients` so the table is
  never restated, and a name several databases share, like `usql`, passes for each of them.
  `clientAliases` is the one exception: `clickhouse-client` is the same program as the
  `clickhouse` binary `cli.Clients` names, but it is not a client conm may be
  told to run, so it is resolved to the canonical name here rather than widened into
  `cli`.
- a connection string of another database is rejected the same way. The check is
  `foreignScheme` against the final `dbname`, not against the raw tokens: by then `scanFlags`
  has already unpacked `-d x`, `-dx` and `--dbname=x` into one value, and no valid postgres
  database name can contain `://`, so one check covers every spelling.

*Which syntax* is still inferred, because that is a property of the string and not of the
user's intent: `Result.Syntax` is `URISyntax`, `KeywordSyntax` or `FlagSyntax`, decided per
database, because a conninfo string is not only a positional argument — `-d` takes one too,
and a bare `host=… port=…` arrives as several tokens that are one conninfo between them.

Rules a parser follows, which are the client's and not ours to change — postgres spells them
against libpq, redis against `redis-cli`/`valkey-cli`/`iredis`:

- **Precedence is env < positional < conninfo < flags.** Every layer only writes the keys it
  actually carries, so a partially given command line stays partially filled rather than
  overwriting an earlier layer with a blank.
- **Only an inline `VAR=value` prefix is read as an environment.** `PGPASSWORD=s psql …` and
  `REDISCLI_AUTH=s redis-cli …` fill the password; the *process* environment is never consulted, since conm parses a
  command line the user quotes at it, not one it is about to run.
- **A flag that takes a value must be known even when its value is discarded.** That is what
  `postgresIgnoredFlags` is for: without `-c` in it, `psql -c "select 1" mydb` would read the
  query as the database name.
- **The first host wins.** `config.Connection` holds one address, so a seed list or a
  `host=h1,h2` keyword keeps `h1` and says so in `Result.Warnings`.
- **Nothing is dropped silently.** A parameter with no field to land in, a host list, an
  unusable port, ssl mode or database index, a unix socket there is no host and port to keep,
  and a named `--dsn` that only pgcli's or iredis's own config could resolve all become a
  warning. The caller shows them; the connection is still returned. Output formatting is not
  a connection parameter, so a discarded `psql -F` or `redis-cli -d` says nothing.
- **TLS beyond the mode is a refusal, not a warning** — `ErrUnsupportedTLS`. A connection
  carries a TLS *mode* and nothing else, so a CA bundle, a client certificate, a client key,
  a key passphrase, a cipher list, a protocol version and an SNI host all have no field —
  and unlike a dropped `-F`, their absence changes whether the connection works. A verifying
  mode without its CA checks the system root store instead, so a form seeded from
  `mysql --ssl-mode=VERIFY_CA --ssl-ca=/etc/rds-ca.pem` would save a connection that cannot
  ping while the command line it came from does; a pinned `--tls-version` or a cipher list
  is the same bargain against a server that requires it. Each parser owns the table of its
  own spellings — `mysqlTLSRefusals`, `redisTLSRefusals`, `postgresTLSRefusals` — and each
  entry is still scanned as taking a value, so the refusal happens on the flag rather than
  on its path being read as a host or a database name. This is provisional: once the
  connection can carry the material, these tables become appliers and the
  refusal goes away.
- **A parsed password is written as `literal:<password>`, not bare.** The explicit scheme is
  what keeps a password that itself contains a colon from reading back as a store reference.
- **`port` and `sslmode` fall back to `5432` and `prefer`.** They are libpq's defaults and the
  form's, and both fields are seeded from a typed value, so an absent one has to be spelled.
  Redis spells `6379`, database `0` and TLS `disable` for the same reason.

Rules that are one database's alone, because its clients are not libpq's — redis first:

- **The syntax a client does not accept is never returned.** Redis has no conninfo keyword
  form and takes no connection string as a positional — a URL reaches `redis-cli` behind
  `-u` and `iredis` behind `-u`/`--url`, so `parseRedis` returns only `URISyntax` and
  `FlagSyntax`. A leading positional is read as a URL **only when no client was named**,
  which is the `conm redis add "redis://…"` a user pastes; after `redis-cli` a positional is
  the command to run and is left alone.
- **`rediss://` and `--tls` are the TLS mode**, since `config.Redis` carries `TLSMode` rather
  than a scheme. `--cacert`, `--cacertdir`, `--cert`, `--key`, `--sni` and iredis's
  `--verify-ssl` are the TLS refusal above.

`-d` is the one flag whose meaning depends on the client: iredis's named DSN and
`redis-cli`'s output delimiter. `applyRedisFlags` takes `req.client` for that single
decision, and warns only for `iredis`; both spellings consume their value either way, so
scanning never reads one as a host.

And mongodb's own:

- **A positional is a *db address*, not just a database.** `mongosh` accepts `records`,
  `mongo.internal/records`, `mongo.internal:27017/records` and a `mongodb://` URI in the one
  argument, so `applyMongoDBAddress` reads it as a database when it carries neither `/` nor
  `:`, and as hosts plus an optional database otherwise. What it never does is invent the
  `localhost` the client would assume — no other parser fills a host the user did not
  write, and an empty host field is the form asking for one.
- **A replica set is a name in front of the hosts**, both in `--host rs0/h1,h2` and in
  `?replicaSet=rs0`, and it has no field: `config.MongoDB` holds one address. Both spellings
  reach `replicaSetWarning`, and the seed list after it is the usual "first host wins".
- **`mongodb+srv://` carries neither hosts nor a port.** The name is a DNS record the driver
  resolves into a seed list, which a pure function of a string cannot do, so the seed name is
  kept as the host with the default port and the warning says so. The scheme also implies
  TLS, which is the one part of it conm can honour.
- **`-p` is the mysql split again.** The legacy `mongo` shell takes its password only
  attached, `mongosh` takes the next token; `mongodb://` and `mongodb+srv://` are the only
  schemes, since no `usql` alias reaches mongo.
- **There is no environment table.** Neither shell reads a connection out of the
  environment, so there is nothing an inline `VAR=value` prefix could fill — an absent table
  rather than an invented one.
- **Parameter names are matched lowercased**, as the driver reads them, and the TLS refusal
  covers both the `--tls*`/`--ssl*` file flags and their URI spellings; `tls`, `ssl` and
  `authSource` are not refused, because those have fields.

And mssql's own:

- **The address is one field with a grammar of its own.** `-S [protocol:]host[\instance][,port]`
  and `SQLCMDSERVER` spell the same thing, so `setMSSQLServer` is the only place it is taken
  apart: the port is after a comma rather than a colon, a `\instance` and a transport that
  is not `tcp:` become warnings because conm keeps one TCP host and port, and brackets are
  trimmed so `tcp:[::1],1433` keeps `::1`. Both callers route through it, which is why env
  and a flag can disagree and the flag still wins.
- **`\` is an escape before `splitArgs` ever sees an instance.** `sqlcmd -S host\INSTANCE`
  pasted unquoted loses the backslash, exactly as a POSIX shell would; quoted with `'…'` it
  survives, and `ParseArgs` never sees the problem because the tokens arrive already split.
  That is `splitArgs` being a shell, not mssql being special.
- **The url path is an instance and the database is a parameter.** `sqlserver://` puts the
  database in `?database=`, which is what `config.MSSQL.ConnectionString` writes, so a path
  is the instance and gets the same warning as `\instance`. Parameter names are matched
  lowercased because the driver reads them case-insensitively.
- **`-N` takes only an attached value and means *mandatory* alone.** `-N`, `-Ns`, `-Nm` and
  `-No` are the client's spelling of require/strict/mandatory/optional, translated by
  `mssqlEncryptModes` into the four `config.MSSQLEncryptMode` values; a bare `-N` is
  `true`, matching what `network/mssql.go` emits in the other direction. `-r` and `-X` share
  the optional-value shape and are listed with it so neither swallows the flag after it.
- **`-E` and `-G` are refused nothing but warned about.** A trusted connection and Entra
  authentication carry no username or password, so the form would open with an empty user
  that `ValidateMSSQLUsername` rejects; the warning says why rather than leaving it blank.
- **`certificate`, `hostnameincertificate` and `tlsmin` are the TLS refusal.** `encrypt` and
  `trustservercertificate` are not — both have fields.

And clickhouse's own:

- **The binary is not the whole client.** `clickhouse client …` reaches conm as the binary
  plus a subcommand, so `clickhouseArgs` drops a leading `client` once the binary itself has
  been stripped; `clickhouse-client …` is the same program under one name and is resolved by
  `clientAliases`. Both end at the same `Result.Client`, the `clickhouse` that
  `cli.Clients` and `network` already agree on.
- **`--password` takes only an attached value**, the same `attachedValueFlag` mysql needs
  and for the same reason: the client parses it with an implicit value, so
  `--password -d metrics` prompts and opens `metrics` rather than sending `-d` as a
  password.
- **A positional is a connection string or a query, never a database.** Only a first
  positional containing `://` is a conninfo; everything else is SQL the client would run,
  which is no more a connection parameter than `psql -c` is.
- **The default port follows `secure`** — `9440` when the connection string said
  `clickhouses://`, the query said `secure=true` or a `-s`/`--secure` flag was given, `9000`
  otherwise. Those are the native protocol's two ports, and defaulting to `9000` on a secure
  connection would save a connection that cannot dial. `Secure` is a bool rather than a
  mode, so an unreadable `secure=` value is a warning and `false`.

And mysql's own:

- **There is no keyword form and no conninfo flag.** A URL reaches mycli and usql as the
  leading positional and nothing else carries one, so `parseMySQL` returns `URISyntax` or
  `FlagSyntax`. A positional that is not a URL is the database, which is what `mysql shop`
  means.
- **`--ssl-mode` is not the field's vocabulary.** The client spells `DISABLED`, `PREFERRED`,
  `REQUIRED`, `VERIFY_CA` and `VERIFY_IDENTITY`; `config.MySQL.TLSMode` holds the driver's
  `false`/`preferred`/`skip-verify`/`true`, which is what a `mysql://…?tls=` URL already
  carries. `mysqlSSLModes` translates the flag and leaves anything it does not know for
  `buildMySQL` to reject, so an unreadable mode is one warning and not two vocabularies in
  one field. `port` and TLS fall back to `3306` and `preferred`, the form's defaults.
- **A `--login-path` or a `--defaults-file` is a credential conm cannot follow.** Both name
  a file or a section that only the client's own configuration could resolve, so both become
  a warning rather than a silently empty form. They are a warning and not the TLS refusal
  because what they carry — a host, a user, a password — does have fields; only this
  particular spelling of them is out of reach.
- **usql's aliases are mysql's schemes.** `my://`, `maria://`, `mariadb://`, `percona://`
  and `aurora://` all reach `usql` as the same database, so `schemes` lists them beside
  `mysql://` and `foreignScheme` accepts each of them.

`params` carries most of the tests under `internal` (`postgres_test.go`, `redis_test.go`,
`mysql_test.go`, `clickhouse_test.go`, `mssql_test.go`, `mongodb_test.go`,
`lexer_test.go`),
because it is the one package that is a pure function of a string. The others follow the same
principle: `utils/unique_test.go` over a generic with no domain, and
`repository/unique_test.go`, which drives `Add`/`Edit` over a real config file in a
`t.TempDir()` to guard the uniqueness rule end to end.

And ssh's own:

- **Flags stop at the remote command.** ssh reads options before the destination and right
  after it, then everything from the next positional on is the command, so `ssh host ls -la`
  must not read `-la` as `-l a`. `sshCommandStart` finds that boundary before `scanFlags`
  runs, because `scanFlags` itself treats a dash anywhere as a flag.
- **The first value wins, as in `ssh_config`.** `sshValues.set` never overwrites, and the
  destination is applied last, so `-l`/`-o User` beat `user@`, `-o HostName` beats the
  destination alias, and the first of two `-o User` is kept. `-p` is the exception ssh makes
  too: it overwrites, so it beats `-o Port` and a url port wherever it stands.
- **`-o` takes `Key=Value`, `Key Value` and `Key = Value`**, the spellings `ssh_config`
  accepts; the key is matched lowercased. The split is `config.SplitSSHOption`, the one
  `config.ReadSSHConfig` reads a line with, so a quoted value loses its quotes here too.
- **A second forward or identity file is a warning.** ssh adds up `LocalForward` and
  `IdentityFile` where it keeps the first of everything else, and `config.SSH` holds one of
  each: `sshValues.keep` keeps the first and names the one it dropped.
- **A `LocalForward` option is rewritten to the `-L` spelling.** `ssh_config` separates the
  listening port from the target with a space (`LocalForward 15432 localhost:5432`), `-L` and
  `config.SSH.LocalForward` with a colon, so the option's fields are joined with `:`.
