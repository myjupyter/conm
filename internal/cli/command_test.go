package cli

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestCommandFor(t *testing.T) {
	t.Parallel()

	postgres := config.Postgres{Hostname: "db.local", PortNumber: 5432, User: "me", DBName: "app"}
	mysql := config.MySQL{Hostname: "db.local", PortNumber: 3306, User: "me", DBName: "app", TLSMode: config.MySQLTLSModeSkipVerify}
	mssql := config.MSSQL{Hostname: "db.local", PortNumber: 1433, User: "me", DBName: "app", EncryptMode: config.MSSQLEncryptStrict, TrustCert: true}
	clickhouse := config.ClickHouse{Hostname: "db.local", PortNumber: 9000, User: "me", DBName: "app", Secure: true}
	redis := config.Redis{Hostname: "db.local", PortNumber: 6379, DBIndex: 0}

	tests := []struct {
		name     string
		cli      string
		cfg      config.Connection
		password string
		args     []string
		env      string
	}{
		{
			name: "psql takes the url",
			cli:  Psql,
			cfg:  postgres,
			args: []string{postgres.ConnectionString("s3cret")},

			password: "s3cret",
		},
		{
			name:     "mysql takes flags and the password out of band",
			cli:      MySQL,
			cfg:      mysql,
			password: "s3cret",
			args:     []string{"--protocol=TCP", "--host=db.local", "--user=me", "--port=3306", "--ssl-mode=REQUIRED", "app"},
			env:      "MYSQL_PWD=s3cret",
		},
		{
			name:     "usql takes the mysql url",
			cli:      USQL,
			cfg:      mysql,
			password: "s3cret",
			args:     []string{mysql.ConnectionString("s3cret")},
		},
		{
			name:     "sqlcmd takes flags",
			cli:      SQLCmd,
			cfg:      mssql,
			password: "s3cret",
			args:     []string{"-S", "tcp:db.local,1433", "-U", "me", "-d", "app", "-N", "-C"},
			env:      "SQLCMDPASSWORD=s3cret",
		},
		{
			name:     "clickhouse names its subcommand",
			cli:      ClickHouse,
			cfg:      clickhouse,
			password: "s3cret",
			args:     []string{"client", "--host", "db.local", "--user", "me", "--database", "app", "--port", "9000", "--secure"},
			env:      "CLICKHOUSE_PASSWORD=s3cret",
		},
		{
			name: "redis-cli takes the url behind a flag",
			cli:  RedisCLI,
			cfg:  redis,
			args: []string{"-u", redis.ConnectionString("")},
		},
		{
			name: "iredis spells the same flag differently",
			cli:  IRedis,
			cfg:  redis,
			args: []string{"--url", redis.ConnectionString("")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			build, ok := commands[tt.cli]
			require.True(t, ok)

			cmd, err := build(tt.cfg, tt.password)
			require.NoError(t, err)

			is := assert.New(t)
			is.Equal(tt.args, cmd.args)
			if tt.env == "" {
				is.Nil(cmd.env)
				return
			}
			is.Contains(cmd.env, tt.env)
		})
	}
}

func TestEveryClientHasACommand(t *testing.T) {
	t.Parallel()

	for _, db := range config.Databases {
		for _, name := range Clients(db) {
			_, ok := commands[name]
			assert.True(t, ok, "%s has no command", name)
		}
	}
}

func TestLauncherRefusesForeignClient(t *testing.T) {
	t.Parallel()

	err := Launcher{name: Psql, kind: config.MySQLConnType}.Run(
		context.Background(),
		config.MySQL{Hostname: "db.local", PortNumber: 3306, User: "me", DBName: "app"},
		"s3cret",
	)

	require.ErrorIs(t, err, ErrUnknownClient)
}

func TestCommandRefusesForeignConnection(t *testing.T) {
	t.Parallel()

	postgres := config.Postgres{Hostname: "db.local", PortNumber: 5432, User: "me", DBName: "app"}

	for name, build := range map[string]builder{
		MySQL:      mysqlCommand,
		SQLCmd:     mssqlCommand,
		ClickHouse: clickhouseCommand,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := build(postgres, "s3cret")
			require.ErrorIs(t, err, ErrConnType)
		})
	}
}

func TestLauncherWithoutClient(t *testing.T) {
	t.Parallel()

	err := Launcher{kind: config.PostgresConnType}.Run(
		context.Background(),
		config.Postgres{Hostname: "db.local", PortNumber: 5432, User: "me", DBName: "app"},
		"s3cret",
	)

	require.ErrorIs(t, err, ErrNoClient)
}

func TestParseVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		out  string
		want string
	}{
		{name: "psql names the product first", out: "psql (PostgreSQL) 16.2\n", want: "16.2"},
		{name: "redis-cli trails a git revision", out: "redis-cli 7.2.4 (git:0)\n", want: "7.2.4"},
		{name: "mongo spells a v prefix", out: "MongoDB shell version v5.0.5\n", want: "5.0.5"},
		{name: "clickhouse counts four components", out: "ClickHouse client version 24.1.1.1.\n", want: "24.1.1.1"},
		{name: "mysql buries it in a banner", out: "mysql  Ver 8.0.36 for macos14 on arm64 (Homebrew)\n", want: "8.0.36"},
		{name: "mongosh prints it bare", out: "2.1.1\n", want: "2.1.1"},
		{name: "later lines are ignored", out: "no version here\n9.9.9\n", want: ""},
		{name: "nothing parseable", out: "command not found\n", want: ""},
		{name: "no output at all", out: "", want: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, parseVersion(tt.out))
		})
	}
}

func TestEveryClientHasAVersionFlag(t *testing.T) {
	t.Parallel()

	for _, db := range config.Databases {
		for _, name := range Clients(db) {
			_, ok := versionFlags[name]
			assert.True(t, ok, "%s has no version flag", name)
		}
	}
}

func TestVersionOfUnknownClient(t *testing.T) {
	t.Parallel()

	assert.Empty(t, Version(context.Background(), "nosuchclient"))
}
