package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

func TestParseClickHouse(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		syntax   Syntax
		client   string
		conn     config.ClickHouse
		warnings []string
	}{
		{
			name:   "a connection string is the positional the client takes",
			input:  `clickhouse client clickhouse://app:s3cr3t@ch.internal:9000/metrics`,
			syntax: URISyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
				User:       "app",
				Password:   secret.Ref(secret.Literal, "s3cr3t"),
				DBName:     "metrics",
			},
		},
		{
			name:   "the secure scheme carries its own default port",
			input:  `clickhouse client clickhouses://ch.internal/metrics`,
			syntax: URISyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9440,
				DBName:     "metrics",
				Secure:     true,
			},
		},
		{
			name:   "the client binary may be spelled clickhouse-client",
			input:  `clickhouse-client --host ch.internal --user app --database metrics`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
				User:       "app",
				DBName:     "metrics",
			},
		},
		{
			name:   "short flags with attached values",
			input:  `clickhouse client -hch.internal -uapp -dmetrics --port 9001`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9001,
				User:       "app",
				DBName:     "metrics",
			},
		},
		{
			name:   "an attached password is kept",
			input:  `clickhouse client --host ch.internal --password=s3cr3t -d metrics`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
				Password:   secret.Ref(secret.Literal, "s3cr3t"),
				DBName:     "metrics",
			},
		},
		{
			name:   "a lone password flag is a prompt, not the next token",
			input:  `clickhouse client --host ch.internal --password -d metrics`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
				DBName:     "metrics",
			},
		},
		{
			name:   "the secure flag takes no value",
			input:  `clickhouse client -s -h ch.internal -d metrics`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9440,
				DBName:     "metrics",
				Secure:     true,
			},
		},
		{
			name:   "an inline environment fills the password",
			input:  `CLICKHOUSE_PASSWORD=fromenv clickhouse-client -h ch.internal -d metrics`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
				Password:   secret.Ref(secret.Literal, "fromenv"),
				DBName:     "metrics",
			},
		},
		{
			name:   "a flag beats the environment it follows",
			input:  `CLICKHOUSE_HOST=fromenv clickhouse client -h ch.internal -d metrics`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
				DBName:     "metrics",
			},
		},
		{
			name:   "flags win over the connection string they follow",
			input:  `clickhouse client clickhouse://app@ch.internal:9000/metrics --host replica.internal -d reports`,
			syntax: URISyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "replica.internal",
				PortNumber: 9000,
				User:       "app",
				DBName:     "reports",
			},
		},
		{
			name:   "usql takes a url of its own alias",
			input:  `usql ch://app@ch.internal/metrics`,
			syntax: URISyntax,
			client: "usql",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
				User:       "app",
				DBName:     "metrics",
			},
		},
		{
			name:   "a query is not a connection parameter",
			input:  `clickhouse client -h ch.internal -d metrics -q "SELECT 1" --format JSON`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
				DBName:     "metrics",
			},
		},
		{
			name:   "a host list keeps the first host",
			input:  `clickhouse client clickhouse://node1:9000,node2:9000/metrics`,
			syntax: URISyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "node1",
				PortNumber: 9000,
				DBName:     "metrics",
			},
			warnings: []string{"kept only the first of 2 hosts: node1:9000"},
		},
		{
			name:   "a named connection can't be resolved from here",
			input:  `clickhouse client --connection staging`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				PortNumber: 9000,
			},
			warnings: []string{`named dsn "staging" can't be resolved from here`},
		},
		{
			name:   "a parameter with no field is warned about",
			input:  `clickhouse client "clickhouse://ch.internal/metrics?secure=true&connection_timeout=5"`,
			syntax: URISyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9440,
				DBName:     "metrics",
				Secure:     true,
			},
			warnings: []string{"ignored parameter connection_timeout=5"},
		},
		{
			name:   "an unreadable port is warned about",
			input:  `clickhouse client -h ch.internal --port nine`,
			syntax: FlagSyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
			},
			warnings: []string{`port "nine" is not a number, kept 9000`},
		},
		{
			name:   "an unreadable secure parameter is warned about",
			input:  `clickhouse client "clickhouse://ch.internal/metrics?secure=sort-of"`,
			syntax: URISyntax,
			client: "clickhouse",
			conn: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9000,
				DBName:     "metrics",
			},
			warnings: []string{`secure "sort-of" is not a boolean, kept false`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			res, err := Parse(config.ClickHouseConnType, test.input)
			must.NoError(err)

			is.Equal(test.syntax, res.Syntax)
			is.Equal(test.client, res.Client)
			is.Equal(test.warnings, res.Warnings)

			conn, ok := res.Conn.(config.ClickHouse)
			must.Truef(ok, "connection is a %T, want config.ClickHouse", res.Conn)
			is.Equal(test.conn, conn)
		})
	}
}

func TestParseClickHouseRoundTripsConnectionString(t *testing.T) {
	want := config.ClickHouse{
		Hostname:   "ch.example.com",
		PortNumber: 9440,
		User:       "app",
		DBName:     "metrics",
		Secure:     true,
	}

	res, err := Parse(config.ClickHouseConnType, want.ConnectionString(""))
	require.NoError(t, err)

	got, ok := res.Conn.(config.ClickHouse)
	require.Truef(t, ok, "connection is a %T, want config.ClickHouse", res.Conn)
	assert.Equal(t, want, got)
}

func TestParseClickHouseErrors(t *testing.T) {
	tests := []struct {
		name  string
		kind  config.ConnType
		input string
		want  error
	}{
		{
			name:  "a client of another database",
			kind:  config.ClickHouseConnType,
			input: `psql -h localhost mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a connection string of another database",
			kind:  config.ClickHouseConnType,
			input: `redis://cache.internal:6379/0`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "another database behind the client name",
			kind:  config.ClickHouseConnType,
			input: `clickhouse client mysql://app@db.internal/shop`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a clickhouse url handed to another database",
			kind:  config.RedisConnType,
			input: `clickhouse://app@ch.internal:9000/metrics`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "an accepted invalid certificate is not the secure flag",
			kind:  config.ClickHouseConnType,
			input: `clickhouse client -h ch.internal --accept-invalid-certificate`,
			want:  ErrUnsupportedTLS,
		},
		{
			name:  "a skipped verification has no field to keep it in",
			kind:  config.ClickHouseConnType,
			input: `clickhouse client "clickhouses://ch.internal/metrics?skip_verify=true"`,
			want:  ErrUnsupportedTLS,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.kind, test.input)
			assert.ErrorIs(t, err, test.want)
		})
	}
}

func TestParseArgsClickHouse(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want config.ClickHouse
	}{
		{
			name: "bare flags need no client name",
			args: []string{"-h", "ch.internal", "--port", "9001", "-d", "metrics"},
			want: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9001,
				DBName:     "metrics",
			},
		},
		{
			name: "the client subcommand is stripped too",
			args: []string{"clickhouse", "client", "-h", "ch.internal", "-s"},
			want: config.ClickHouse{
				Hostname:   "ch.internal",
				PortNumber: 9440,
				Secure:     true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := ParseArgs(config.ClickHouseConnType, test.args)
			require.NoError(t, err)

			conn, ok := res.Conn.(config.ClickHouse)
			require.Truef(t, ok, "connection is a %T, want config.ClickHouse", res.Conn)
			assert.Equal(t, test.want, conn)
		})
	}
}
