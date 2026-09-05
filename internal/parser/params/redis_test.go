package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

func TestParseRedis(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		syntax   Syntax
		client   string
		conn     config.Redis
		warnings []string
	}{
		{
			name:   "url behind the client flag",
			input:  `redis-cli -u redis://default:s3cr3t@cache.internal:6380/3`,
			syntax: URISyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6380,
				User:       "default",
				Password:   secret.Ref(secret.Literal, "s3cr3t"),
				DBIndex:    3,
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
		{
			name:   "a tls scheme is a tls mode",
			input:  `redis-cli -u rediss://cache.internal/1`,
			syntax: URISyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				DBIndex:    1,
				TLSMode:    config.RedisTLSModeRequire,
			},
		},
		{
			name:   "a bare url needs no client",
			input:  `redis://:letmein@127.0.0.1:6379/2`,
			syntax: URISyntax,
			conn: config.Redis{
				Hostname:   "127.0.0.1",
				PortNumber: 6379,
				Password:   secret.Ref(secret.Literal, "letmein"),
				DBIndex:    2,
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
		{
			name:   "short flags with attached values",
			input:  `redis-cli -hlocalhost -p6380 -n7 -adevpass`,
			syntax: FlagSyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "localhost",
				PortNumber: 6380,
				Password:   secret.Ref(secret.Literal, "devpass"),
				DBIndex:    7,
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
		{
			name:   "iredis long flags, with an empty url to fall through",
			input:  `iredis --host cache.internal --port 6380 --username reader --password ro --url ""`,
			syntax: FlagSyntax,
			client: "iredis",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6380,
				User:       "reader",
				Password:   secret.Ref(secret.Literal, "ro"),
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
		{
			name:   "the tls flag takes no value",
			input:  `valkey-cli --tls -h cache.internal -n 4`,
			syntax: FlagSyntax,
			client: "valkey-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				DBIndex:    4,
				TLSMode:    config.RedisTLSModeRequire,
			},
		},
		{
			name:   "flags win over the url they follow",
			input:  `redis-cli -u redis://cache.internal:6380/3 -h replica.internal -n 5`,
			syntax: URISyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "replica.internal",
				PortNumber: 6380,
				DBIndex:    5,
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
		{
			name:   "an inline environment fills the password",
			input:  `REDISCLI_AUTH=fromenv redis-cli -h cache.internal`,
			syntax: FlagSyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				Password:   secret.Ref(secret.Literal, "fromenv"),
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
		{
			name:   "a flag beats the environment it follows",
			input:  `VALKEYCLI_AUTH=fromenv valkey-cli -h cache.internal -a fromflag`,
			syntax: FlagSyntax,
			client: "valkey-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				Password:   secret.Ref(secret.Literal, "fromflag"),
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
		{
			name:   "a command is not a connection parameter",
			input:  `redis-cli -h cache.internal -t 5 -r 3 GET session:42`,
			syntax: FlagSyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
		{
			name:   "a host list keeps the first host",
			input:  `redis-cli -u redis://node1:6379,node2:6379/0`,
			syntax: URISyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "node1",
				PortNumber: 6379,
				TLSMode:    config.RedisTLSModeDisable,
			},
			warnings: []string{"kept only the first of 2 hosts: node1:6379"},
		},
		{
			name:   "an unusable database index is warned about",
			input:  `redis-cli -u redis://cache.internal/900`,
			syntax: URISyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				TLSMode:    config.RedisTLSModeDisable,
			},
			warnings: []string{"database must be between 0 and 255, kept 0"},
		},
		{
			name:   "an unreadable port is warned about",
			input:  `redis-cli -h cache.internal -p six`,
			syntax: FlagSyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				TLSMode:    config.RedisTLSModeDisable,
			},
			warnings: []string{`port "six" is not a number, kept 6379`},
		},
		{
			name:   "a url parameter with no field is warned about",
			input:  `redis-cli -u "redis://cache.internal/0?db=9&protocol=3"`,
			syntax: URISyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				DBIndex:    9,
				TLSMode:    config.RedisTLSModeDisable,
			},
			warnings: []string{"ignored parameter protocol=3"},
		},
		{
			name:   "a unix socket has nowhere to land",
			input:  `redis-cli -s /var/run/redis.sock -n 2`,
			syntax: FlagSyntax,
			client: "redis-cli",
			conn: config.Redis{
				PortNumber: 6379,
				DBIndex:    2,
				TLSMode:    config.RedisTLSModeDisable,
			},
			warnings: []string{`unix socket "/var/run/redis.sock" has no host and port to keep`},
		},
		{
			name:   "iredis reads -d as a named dsn",
			input:  `iredis -d staging`,
			syntax: FlagSyntax,
			client: "iredis",
			conn: config.Redis{
				PortNumber: 6379,
				TLSMode:    config.RedisTLSModeDisable,
			},
			warnings: []string{`named dsn "staging" can't be resolved from here`},
		},
		{
			name:   "redis-cli reads -d as an output delimiter",
			input:  `redis-cli -h cache.internal -d ";"`,
			syntax: FlagSyntax,
			client: "redis-cli",
			conn: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			res, err := Parse(config.RedisConnType, test.input)
			must.NoError(err)

			is.Equal(test.syntax, res.Syntax)
			is.Equal(test.client, res.Client)
			is.Equal(test.warnings, res.Warnings)

			conn, ok := res.Conn.(config.Redis)
			must.Truef(ok, "connection is a %T, want config.Redis", res.Conn)
			is.Equal(test.conn, conn)
		})
	}
}

func TestParseRedisRoundTripsConnectionString(t *testing.T) {
	want := config.Redis{
		Hostname:   "cache.example.com",
		PortNumber: 6380,
		User:       "reader",
		DBIndex:    3,
		TLSMode:    config.RedisTLSModeRequire,
	}

	res, err := Parse(config.RedisConnType, want.ConnectionString(""))
	require.NoError(t, err)

	got, ok := res.Conn.(config.Redis)
	require.Truef(t, ok, "connection is a %T, want config.Redis", res.Conn)
	assert.Equal(t, want, got)
}

func TestParseRedisErrors(t *testing.T) {
	tests := []struct {
		name  string
		kind  config.ConnType
		input string
		want  error
	}{
		{
			name:  "a client of another database",
			kind:  config.RedisConnType,
			input: `psql -h localhost mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a connection string of another database",
			kind:  config.RedisConnType,
			input: `postgres://me@localhost/mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "another database behind the url flag",
			kind:  config.RedisConnType,
			input: `redis-cli -upostgres://localhost:5432/mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a redis url handed to another database",
			kind:  config.PostgresConnType,
			input: `redis://cache.internal:6379/0`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a ca has no field to keep it in",
			kind:  config.RedisConnType,
			input: `redis-cli --tls --cacert /etc/ca.pem -h cache.internal`,
			want:  ErrUnsupportedTLS,
		},
		{
			name:  "a client certificate has no field to keep it in",
			kind:  config.RedisConnType,
			input: `redis-cli --tls --cert /etc/client.crt --key /etc/client.key -h cache.internal`,
			want:  ErrUnsupportedTLS,
		},
		{
			name:  "an sni host is not the tls mode either",
			kind:  config.RedisConnType,
			input: `redis-cli --tls --sni cache.example.com -h cache.internal`,
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

func TestParseArgsRedis(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want config.Redis
	}{
		{
			name: "bare flags need no client name",
			args: []string{"-h", "cache.internal", "-p", "6380", "-n", "2"},
			want: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6380,
				DBIndex:    2,
				TLSMode:    config.RedisTLSModeDisable,
			},
		},
		{
			name: "a client name is still stripped",
			args: []string{"valkey-cli", "-h", "cache.internal", "--tls"},
			want: config.Redis{
				Hostname:   "cache.internal",
				PortNumber: 6379,
				TLSMode:    config.RedisTLSModeRequire,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := ParseArgs(config.RedisConnType, test.args)
			require.NoError(t, err)

			conn, ok := res.Conn.(config.Redis)
			require.Truef(t, ok, "connection is a %T, want config.Redis", res.Conn)
			assert.Equal(t, test.want, conn)
		})
	}
}
