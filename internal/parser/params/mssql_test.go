package params

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

func TestParseMSSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		syntax   Syntax
		client   string
		conn     config.MSSQL
		warnings []string
	}{
		{
			name:   "the server flag carries host and port",
			input:  `sqlcmd -S tcp:sql.internal,1433 -U sa -P s3cr3t -d app`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				User:        "sa",
				Password:    secret.Ref(secret.Literal, "s3cr3t"),
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
		},
		{
			name:   "a server without a port keeps the default",
			input:  `sqlcmd -S sql.internal -U sa -d app`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				User:        "sa",
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
		},
		{
			name:   "short flags with attached values",
			input:  `sqlcmd -Ssql.internal,1435 -Usa -Ps3cr3t -dapp`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1435,
				User:        "sa",
				Password:    secret.Ref(secret.Literal, "s3cr3t"),
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
		},
		{
			name:   "the encrypt flag alone is mandatory encryption",
			input:  `sqlcmd -S sql.internal -d app -N`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptRequire,
			},
		},
		{
			name:   "an attached encrypt level is translated",
			input:  `sqlcmd -S sql.internal -d app -Ns`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptStrict,
			},
		},
		{
			name:   "the trust flag takes no value",
			input:  `sqlcmd -S sql.internal -d app -C -N`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptRequire,
				TrustCert:   true,
			},
		},
		{
			name:   "a url is the positional usql takes",
			input:  `usql sqlserver://sa:s3cr3t@sql.internal:1433?database=app&encrypt=true`,
			syntax: URISyntax,
			client: "usql",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				User:        "sa",
				Password:    secret.Ref(secret.Literal, "s3cr3t"),
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptRequire,
			},
		},
		{
			name:   "a trusted certificate is a url parameter",
			input:  `usql "sqlserver://sql.internal?database=app&trustservercertificate=true"`,
			syntax: URISyntax,
			client: "usql",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
				TrustCert:   true,
			},
		},
		{
			name:   "flags win over the url they follow",
			input:  `sqlcmd "sqlserver://sa@sql.internal:1433?database=app" -S replica.internal,1435 -d reports`,
			syntax: URISyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "replica.internal",
				PortNumber:  1435,
				User:        "sa",
				DBName:      "reports",
				EncryptMode: config.MSSQLEncryptLogin,
			},
		},
		{
			name:   "an inline environment fills the connection",
			input:  `SQLCMDPASSWORD=fromenv SQLCMDSERVER=tcp:sql.internal,1435 sqlcmd -d app`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1435,
				Password:    secret.Ref(secret.Literal, "fromenv"),
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
		},
		{
			name:   "a flag beats the environment it follows",
			input:  `SQLCMDSERVER=fromenv sqlcmd -S sql.internal -d app`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
		},
		{
			name:   "a query is not a connection parameter",
			input:  `sqlcmd -S sql.internal -d app -Q "SELECT 1" -h -1`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
		},
		{
			name:   "a named instance has no field to keep it in",
			input:  `sqlcmd -S 'sql.internal\SQLEXPRESS' -d app`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
			warnings: []string{`named instance "SQLEXPRESS" has no field to keep it in`},
		},
		{
			name:   "a url path is an instance, not a database",
			input:  `usql "sqlserver://sa@sql.internal/SQLEXPRESS?database=app"`,
			syntax: URISyntax,
			client: "usql",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				User:        "sa",
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
			warnings: []string{`named instance "SQLEXPRESS" has no field to keep it in`},
		},
		{
			name:   "a transport that is not tcp has no field to keep it in",
			input:  `sqlcmd -S np:sql.internal -d app`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
			warnings: []string{`transport "np" has no field to keep it in`},
		},
		{
			name:   "a trusted connection keeps no credentials",
			input:  `sqlcmd -S sql.internal -d app -E`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
			warnings: []string{"a trusted connection keeps no username or password"},
		},
		{
			name:   "a parameter with no field is warned about",
			input:  `usql "sqlserver://sql.internal?database=app&failoverpartner=backup.internal"`,
			syntax: URISyntax,
			client: "usql",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
			warnings: []string{"ignored parameter failoverpartner=backup.internal"},
		},
		{
			name:   "an unreadable port is warned about",
			input:  `sqlcmd -S sql.internal,fifteen -d app`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
			warnings: []string{`port "fifteen" is not a number, kept 1433`},
		},
		{
			name:   "an unknown encrypt mode is warned about",
			input:  `sqlcmd -S sql.internal -d app -Nx`,
			syntax: FlagSyntax,
			client: "sqlcmd",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
			warnings: []string{"invalid encrypt mode: x, kept false"},
		},
		{
			name:   "an unreadable trusted certificate is warned about",
			input:  `usql "sqlserver://sql.internal?database=app&trustservercertificate=sure"`,
			syntax: URISyntax,
			client: "usql",
			conn: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
			warnings: []string{`trustservercertificate "sure" is not a boolean, kept false`},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			res, err := Parse(config.MSSQLConnType, test.input)
			must.NoError(err)

			is.Equal(test.syntax, res.Syntax)
			is.Equal(test.client, res.Client)
			is.Equal(test.warnings, res.Warnings)

			conn, ok := res.Conn.(config.MSSQL)
			must.Truef(ok, "connection is a %T, want config.MSSQL", res.Conn)
			is.Equal(test.conn, conn)
		})
	}
}

func TestParseMSSQLRoundTripsConnectionString(t *testing.T) {
	want := config.MSSQL{
		Hostname:    "sql.example.com",
		PortNumber:  1433,
		User:        "sa",
		DBName:      "app",
		EncryptMode: config.MSSQLEncryptRequire,
		TrustCert:   true,
	}

	res, err := Parse(config.MSSQLConnType, want.ConnectionString(""))
	require.NoError(t, err)

	got, ok := res.Conn.(config.MSSQL)
	require.Truef(t, ok, "connection is a %T, want config.MSSQL", res.Conn)
	assert.Equal(t, want, got)
}

func TestParseMSSQLErrors(t *testing.T) {
	tests := []struct {
		name  string
		kind  config.ConnType
		input string
		want  error
	}{
		{
			name:  "a client of another database",
			kind:  config.MSSQLConnType,
			input: `psql -h localhost mydb`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a connection string of another database",
			kind:  config.MSSQLConnType,
			input: `redis://cache.internal:6379/0`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a sqlserver url handed to another database",
			kind:  config.RedisConnType,
			input: `sqlserver://sa@sql.internal:1433?database=app`,
			want:  ErrTypeMismatch,
		},
		{
			name:  "a ca has no field to keep it in",
			kind:  config.MSSQLConnType,
			input: `usql "sqlserver://sql.internal?database=app&certificate=/etc/ca.pem"`,
			want:  ErrUnsupportedTLS,
		},
		{
			name:  "a protocol floor is refused whatever its case",
			kind:  config.MSSQLConnType,
			input: `usql "sqlserver://sql.internal?database=app&TlsMin=1.2"`,
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

func TestParseArgsMSSQL(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want config.MSSQL
	}{
		{
			name: "bare flags need no client name",
			args: []string{"-S", "tcp:sql.internal,1435", "-U", "sa", "-d", "app"},
			want: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1435,
				User:        "sa",
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
		},
		{
			name: "an instance survives the shell that would have eaten it",
			args: []string{"sqlcmd", "-S", `sql.internal\SQLEXPRESS`, "-d", "app"},
			want: config.MSSQL{
				Hostname:    "sql.internal",
				PortNumber:  1433,
				DBName:      "app",
				EncryptMode: config.MSSQLEncryptLogin,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := ParseArgs(config.MSSQLConnType, test.args)
			require.NoError(t, err)

			conn, ok := res.Conn.(config.MSSQL)
			require.Truef(t, ok, "connection is a %T, want config.MSSQL", res.Conn)
			assert.Equal(t, test.want, conn)
		})
	}
}
