package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

func TestParseMySQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		syntax   Syntax
		client   string
		conn     config.MySQL
		warnings []string
	}{
		{
			name:   "a url is the positional the client takes",
			input:  `mycli mysql://app:s3cr3t@db.internal:3307/shop`,
			syntax: URISyntax,
			client: "mycli",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3307,
				User:       "app",
				Password:   secret.Ref(secret.Literal, "s3cr3t"),
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name:   "short flags with attached values",
			input:  `mysql -hlocalhost -P3307 -uroot -psecret -Dshop`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				Hostname:   "localhost",
				PortNumber: 3307,
				User:       "root",
				Password:   secret.Ref(secret.Literal, "secret"),
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name:   "the stock client reads a lone -p as a prompt",
			input:  `mysql -h db.internal -u root -p shop`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				User:       "root",
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name:   "mycli reads -p as a value",
			input:  `mycli -h db.internal -u root -p secret shop`,
			syntax: FlagSyntax,
			client: "mycli",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				User:       "root",
				Password:   secret.Ref(secret.Literal, "secret"),
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name:   "an ssl mode becomes the tls mode",
			input:  `mysql --host db.internal --ssl-mode=VERIFY_CA app`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				DBName:     "app",
				TLSMode:    config.MySQLTLSModeVerify,
			},
		},
		{
			name:   "an inline environment fills the password",
			input:  `MYSQL_PWD=fromenv mysql -h db.internal shop`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				Password:   secret.Ref(secret.Literal, "fromenv"),
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name:   "a flag beats the environment it follows",
			input:  `MYSQL_PWD=fromenv MYSQL_HOST=fromenv mysql -h db.internal -psecret`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				Password:   secret.Ref(secret.Literal, "secret"),
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name:   "flags win over the url they follow",
			input:  `mycli mysql://app@db.internal:3307/shop -h replica.internal -D reports`,
			syntax: URISyntax,
			client: "mycli",
			conn: config.MySQL{
				Hostname:   "replica.internal",
				PortNumber: 3307,
				User:       "app",
				DBName:     "reports",
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name:   "usql takes a url of its own alias",
			input:  `usql my://app@db.internal/shop`,
			syntax: URISyntax,
			client: "usql",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				User:       "app",
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name:   "an executed statement is not a connection parameter",
			input:  `mysql -h db.internal -e "select 1" shop`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name:   "a host list keeps the first host",
			input:  `mycli mysql://node1:3306,node2:3306/shop`,
			syntax: URISyntax,
			client: "mycli",
			conn: config.MySQL{
				Hostname:   "node1",
				PortNumber: 3306,
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
			warnings: []string{"kept only the first of 2 hosts: node1:3306"},
		},
		{
			name:   "a url parameter with no field is warned about",
			input:  `mycli "mysql://db.internal/shop?tls=skip-verify&parseTime=true"`,
			syntax: URISyntax,
			client: "mycli",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModeSkipVerify,
			},
			warnings: []string{"ignored parameter parseTime=true"},
		},
		{
			name:   "a unix socket has nowhere to land",
			input:  `mysql -S /var/run/mysqld/mysqld.sock shop`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				PortNumber: 3306,
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
			warnings: []string{`unix socket "/var/run/mysqld/mysqld.sock" has no host and port to keep`},
		},
		{
			name:   "a login path is a named dsn",
			input:  `mysql --login-path=staging`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				PortNumber: 3306,
				TLSMode:    config.MySQLTLSModePreferred,
			},
			warnings: []string{`named dsn "staging" can't be resolved from here`},
		},
		{
			name:   "a defaults file is not read from here",
			input:  `mysql --defaults-file=/etc/my.cnf -h db.internal`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				TLSMode:    config.MySQLTLSModePreferred,
			},
			warnings: []string{`defaults file "/etc/my.cnf" can't be read from here`},
		},
		{
			name:   "an unreadable port is warned about",
			input:  `mysql -h db.internal -P three`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				TLSMode:    config.MySQLTLSModePreferred,
			},
			warnings: []string{`port "three" is not a number, kept 3306`},
		},
		{
			name:   "an unknown ssl mode is warned about",
			input:  `mysql -h db.internal --ssl-mode=SORT_OF`,
			syntax: FlagSyntax,
			client: "mysql",
			conn: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				TLSMode:    config.MySQLTLSModePreferred,
			},
			warnings: []string{"invalid TLS mode: SORT_OF, kept preferred"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			res, err := Parse(config.MySQLConnType, test.input)
			must.NoError(err)

			is.Equal(test.syntax, res.Syntax)
			is.Equal(test.client, res.Client)
			is.Equal(test.warnings, res.Warnings)

			conn, ok := res.Conn.(config.MySQL)
			must.Truef(ok, "connection is a %T, want config.MySQL", res.Conn)
			is.Equal(test.conn, conn)
		})
	}
}

func TestParseMySQLRoundTripsConnectionString(t *testing.T) {
	want := config.MySQL{
		Hostname:   "db.example.com",
		PortNumber: 3307,
		User:       "app",
		DBName:     "shop",
		TLSMode:    config.MySQLTLSModeSkipVerify,
	}

	res, err := Parse(config.MySQLConnType, want.ConnectionString(""))
	require.NoError(t, err)

	got, ok := res.Conn.(config.MySQL)
	require.Truef(t, ok, "connection is a %T, want config.MySQL", res.Conn)
	assert.Equal(t, want, got)
}

func TestParseMySQLErrors(t *testing.T) {
	tests := []struct {
		name  string
		kind  config.ConnType
		input string
		want  error
	}{
		{
			name:  "a client of another database",
			kind:  config.MySQLConnType,
			input: `psql -h localhost mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a connection string of another database",
			kind:  config.MySQLConnType,
			input: `redis://cache.internal:6379/0`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "another database behind the database flag",
			kind:  config.MySQLConnType,
			input: `mysql -Dpostgres://localhost:5432/mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a mysql url handed to another database",
			kind:  config.RedisConnType,
			input: `mysql://app@db.internal:3306/shop`,
			want:  ErrTypeMismatch,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.kind, test.input)
			assert.ErrorIs(t, err, test.want)
		})
	}
}

func TestParseArgsMySQL(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want config.MySQL
	}{
		{
			name: "bare flags need no client name",
			args: []string{"-h", "db.internal", "-P", "3307", "-D", "shop"},
			want: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3307,
				DBName:     "shop",
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
		{
			name: "a client name is still stripped",
			args: []string{"mysql", "-h", "db.internal", "-psecret"},
			want: config.MySQL{
				Hostname:   "db.internal",
				PortNumber: 3306,
				Password:   secret.Ref(secret.Literal, "secret"),
				TLSMode:    config.MySQLTLSModePreferred,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := ParseArgs(config.MySQLConnType, test.args)
			require.NoError(t, err)

			conn, ok := res.Conn.(config.MySQL)
			require.Truef(t, ok, "connection is a %T, want config.MySQL", res.Conn)
			assert.Equal(t, test.want, conn)
		})
	}
}
