package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestParsePostgres(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		syntax   Syntax
		client   string
		conn     config.Postgres
		warnings []string
	}{
		{
			name:   "uri with a host list",
			input:  `psql "postgres://user@host1:5432,host2:5432/mydb?target_session_attrs=primary"`,
			syntax: URISyntax,
			client: "psql",
			conn: config.Postgres{
				Hostname:   "host1",
				PortNumber: 5432,
				User:       "user",
				DBName:     "mydb",
				SSLMode:    config.PostgresSSLModePrefer,
			},
			warnings: []string{
				"kept only the first of 2 hosts: host1:5432",
				"ignored parameter target_session_attrs=primary",
			},
		},
		{
			name:   "keyword value conninfo",
			input:  `psql "host=localhost port=5432 dbname=mydb user=me sslmode=require"`,
			syntax: KeywordSyntax,
			client: "psql",
			conn: config.Postgres{
				Hostname:   "localhost",
				PortNumber: 5432,
				User:       "me",
				DBName:     "mydb",
				SSLMode:    config.PostgresSSLModeRequire,
			},
		},
		{
			name:   "pgcli long flags",
			input:  `pgcli --host pgbouncer-dev.db.prod --port 6404 --dbname master --user me`,
			syntax: FlagSyntax,
			client: "pgcli",
			conn: config.Postgres{
				Hostname:   "pgbouncer-dev.db.prod",
				PortNumber: 6404,
				User:       "me",
				DBName:     "master",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "short flags with attached values and a positional",
			input:  `psql -hlocalhost -p5433 -Ume mydb`,
			syntax: FlagSyntax,
			client: "psql",
			conn: config.Postgres{
				Hostname:   "localhost",
				PortNumber: 5433,
				User:       "me",
				DBName:     "mydb",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "positional dbname and username",
			input:  `psql mydb me`,
			syntax: FlagSyntax,
			client: "psql",
			conn: config.Postgres{
				PortNumber: 5432,
				User:       "me",
				DBName:     "mydb",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "valued flag does not swallow the positional",
			input:  `psql -c "select 1" -w mydb`,
			syntax: FlagSyntax,
			client: "psql",
			conn: config.Postgres{
				PortNumber: 5432,
				DBName:     "mydb",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "uri behind -d, flags win over the uri",
			input:  `psql -d postgres://user:pw@example.com/mydb -p 6432`,
			syntax: URISyntax,
			client: "psql",
			conn: config.Postgres{
				Hostname:   "example.com",
				PortNumber: 6432,
				User:       "user",
				Password:   "literal:pw",
				DBName:     "mydb",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "uri with sslmode and a search path",
			input:  `usql "postgresql://me@db:5432/app?sslmode=verify-full&options=-c%20search_path%3Danalytics"`,
			syntax: URISyntax,
			client: "usql",
			conn: config.Postgres{
				Hostname:   "db",
				PortNumber: 5432,
				User:       "me",
				DBName:     "app",
				SchemaName: "analytics",
				SSLMode:    config.PostgresSSLModeVerifyFull,
			},
		},
		{
			name:   "env assignment prefix",
			input:  `PGPASSWORD=s3cr3t psql -h db.example.com -U reader -d metrics`,
			syntax: FlagSyntax,
			client: "psql",
			conn: config.Postgres{
				Hostname:   "db.example.com",
				PortNumber: 5432,
				User:       "reader",
				Password:   "literal:s3cr3t",
				DBName:     "metrics",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "bare uri without a client",
			input:  `postgres://u:p@h.local/db`,
			syntax: URISyntax,
			conn: config.Postgres{
				Hostname:   "h.local",
				PortNumber: 5432,
				User:       "u",
				Password:   "literal:p",
				DBName:     "db",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "bare keyword conninfo without a client",
			input:  `host=localhost dbname=mydb user=me`,
			syntax: KeywordSyntax,
			conn: config.Postgres{
				Hostname:   "localhost",
				PortNumber: 5432,
				User:       "me",
				DBName:     "mydb",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "quoted keyword value and an unknown keyword",
			input:  `psql "host=localhost options='-c search_path=reporting' application_name=conm"`,
			syntax: KeywordSyntax,
			client: "psql",
			conn: config.Postgres{
				Hostname:   "localhost",
				PortNumber: 5432,
				SchemaName: "reporting",
				SSLMode:    config.PostgresSSLModePrefer,
			},
			warnings: []string{"ignored parameter application_name=conm"},
		},
		{
			name:   "ipv6 host",
			input:  `psql "postgres://[::1]:5433/db"`,
			syntax: URISyntax,
			client: "psql",
			conn: config.Postgres{
				Hostname:   "::1",
				PortNumber: 5433,
				DBName:     "db",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "percent encoded credentials",
			input:  `psql "postgres://my%20user:p%40ss@h/my%2Ddb"`,
			syntax: URISyntax,
			client: "psql",
			conn: config.Postgres{
				Hostname:   "h",
				PortNumber: 5432,
				User:       "my user",
				Password:   "literal:p@ss",
				DBName:     "my-db",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name:   "unusable port and sslmode are reported, not kept",
			input:  `psql "host=h port=abc sslmode=maybe"`,
			syntax: KeywordSyntax,
			client: "psql",
			conn: config.Postgres{
				Hostname:   "h",
				PortNumber: 5432,
				SSLMode:    config.PostgresSSLModePrefer,
			},
			warnings: []string{
				`port "abc" is not a number, kept 5432`,
				"invalid SSL mode: maybe, kept prefer",
			},
		},
		{
			name:   "client alone yields the defaults only",
			input:  `psql`,
			syntax: FlagSyntax,
			client: "psql",
			conn: config.Postgres{
				PortNumber: 5432,
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			res, err := Parse(config.PostgresConnType, test.input)
			must.NoError(err)

			is.Equal(test.syntax, res.Syntax)
			is.Equal(test.client, res.Client)
			is.Equal(test.warnings, res.Warnings)

			conn, ok := res.Conn.(config.Postgres)
			must.Truef(ok, "connection is a %T, want config.Postgres", res.Conn)
			is.Equal(test.conn, conn)
		})
	}
}

func TestParseRoundTripsConnectionString(t *testing.T) {
	want := config.Postgres{
		Hostname:   "db.example.com",
		PortNumber: 6432,
		User:       "reader",
		DBName:     "metrics",
		SchemaName: "analytics",
		SSLMode:    config.PostgresSSLModeVerifyCA,
	}

	res, err := Parse(config.PostgresConnType, want.ConnectionString(""))
	require.NoError(t, err)

	got, ok := res.Conn.(config.Postgres)
	require.Truef(t, ok, "connection is a %T, want config.Postgres", res.Conn)
	assert.Equal(t, want, got)
}

func TestParseErrors(t *testing.T) {
	tests := []struct {
		name  string
		kind  config.ConnType
		input string
		want  error
	}{
		{
			name:  "empty",
			kind:  config.PostgresConnType,
			input: "   ",
			want:  ErrEmptyInput,
		},
		{
			name:  "unterminated quote",
			kind:  config.PostgresConnType,
			input: `psql "postgres://h/db`,
			want:  ErrUnterminated,
		},
		{
			name:  "a client of another database",
			kind:  config.PostgresConnType,
			input: `mysql -h localhost -u me mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a connection string of another database",
			kind:  config.PostgresConnType,
			input: `mysql://me@localhost/mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "another database behind an attached flag",
			kind:  config.PostgresConnType,
			input: `psql -dredis://localhost:6379`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a type with no parser",
			kind:  config.ConnType(0),
			input: `-h localhost`,
			want:  ErrUnsupported,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.kind, test.input)
			assert.ErrorIs(t, err, test.want)
		})
	}
}

func TestParseArgs(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want config.Postgres
	}{
		{
			name: "bare flags need no client name",
			args: []string{"-h", "localhost", "-p", "6432", "-U", "me", "mydb"},
			want: config.Postgres{
				Hostname:   "localhost",
				PortNumber: 6432,
				User:       "me",
				DBName:     "mydb",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name: "tokens keep a value that would not survive rejoining",
			args: []string{"psql", "host=localhost options='-c search_path=reporting'"},
			want: config.Postgres{
				Hostname:   "localhost",
				PortNumber: 5432,
				SchemaName: "reporting",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
		{
			name: "a client name is still stripped",
			args: []string{"pgcli", "--host", "db", "--dbname", "app"},
			want: config.Postgres{
				Hostname:   "db",
				PortNumber: 5432,
				DBName:     "app",
				SSLMode:    config.PostgresSSLModePrefer,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := ParseArgs(config.PostgresConnType, test.args)
			require.NoError(t, err)

			conn, ok := res.Conn.(config.Postgres)
			require.Truef(t, ok, "connection is a %T, want config.Postgres", res.Conn)
			assert.Equal(t, test.want, conn)
		})
	}
}
