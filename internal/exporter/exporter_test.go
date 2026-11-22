package exporter_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/exporter"
	"github.com/myjupyter/conm/internal/importer"
)

func TestPlan_roundTrip(t *testing.T) {
	t.Setenv("PGPASSFILE", filepath.Join(t.TempDir(), "absent"))

	meta := config.ConnMeta{Name: "orders"}

	type args struct {
		file     string
		conn     config.Connection
		password string
	}
	tests := []struct {
		name         string
		args         args
		want         config.Connection
		wantWarnings []string
	}{
		{
			name: "postgres",
			args: args{conn: config.Postgres{
				Metadata: meta, Hostname: "db.internal", PortNumber: 6543, User: "app", DBName: "orders",
				SchemaName: "sales", SSLMode: config.PostgresSSLModeRequire, Password: "keyring:postgres/orders",
			}, password: "s3cret"},
			want: config.Postgres{
				Metadata: meta, Hostname: "db.internal", PortNumber: 6543, User: "app", DBName: "orders",
				SchemaName: "sales", SSLMode: config.PostgresSSLModeRequire, Password: "literal:s3cret",
			},
		},
		{
			name: "postgres without the password, the transport left behind",
			args: args{conn: config.Postgres{
				Metadata: meta, Hostname: "db.internal", PortNumber: 5432, User: "app", DBName: "orders",
				SSLMode: config.PostgresSSLModePrefer, Password: "keyring:postgres/orders", Transport: "bastion",
			}},
			want: config.Postgres{
				Metadata: meta, Hostname: "db.internal", PortNumber: 5432, User: "app", DBName: "orders",
				SSLMode: config.PostgresSSLModePrefer,
			},
			wantWarnings: []string{"transport bastion is not exported: the client would connect directly"},
		},
		{
			name: "mysql",
			args: args{conn: config.MySQL{
				Metadata: meta, Hostname: "my.internal", PortNumber: 3307, User: "app", DBName: "orders",
				TLSMode: config.MySQLTLSModeSkipVerify, Password: "keyring:mysql/orders",
			}, password: `pa#s"s\ word`},
			want: config.MySQL{
				Metadata: meta, Hostname: "my.internal", PortNumber: 3307, User: "app", DBName: "orders",
				TLSMode: config.MySQLTLSModeSkipVerify, Password: `literal:pa#s"s\ word`,
			},
		},
		{
			name: "clickhouse, xml",
			args: args{file: "config.xml", conn: config.ClickHouse{
				Metadata: meta, Hostname: "ch.internal", PortNumber: 9440, User: "app", DBName: "orders", Secure: true,
			}, password: "a<b&c"},
			want: config.ClickHouse{
				Metadata: meta, Hostname: "ch.internal", PortNumber: 9440, User: "app", DBName: "orders", Secure: true,
				Password: "literal:a<b&c",
			},
		},
		{
			name: "clickhouse, yaml",
			args: args{file: "config.yaml", conn: config.ClickHouse{
				Metadata: meta, Hostname: "ch.internal", PortNumber: 9000, User: "app", DBName: "orders",
			}, password: "null"},
			want: config.ClickHouse{
				Metadata: meta, Hostname: "ch.internal", PortNumber: 9000, User: "app", DBName: "orders",
				Password: "literal:null",
			},
		},
		{
			name: "mssql",
			args: args{conn: config.MSSQL{
				Metadata: meta, Hostname: "sql.internal", PortNumber: 1434, User: "sa", DBName: "orders",
				EncryptMode: config.MSSQLEncryptStrict,
			}, password: "s3cret"},
			want: config.MSSQL{
				Metadata: meta, Hostname: "sql.internal", PortNumber: 1434, User: "sa", DBName: "master",
				EncryptMode: config.MSSQLEncryptLogin, Password: "literal:s3cret",
			},
			wantWarnings: []string{
				"database orders is not exported: a context holds none",
				"encryption settings are not exported: a context holds none",
			},
		},
		{
			name: "mssql without a user leaves the password out",
			args: args{conn: config.MSSQL{
				Metadata: meta, Hostname: "sql.internal", PortNumber: 1433, DBName: "master",
			}, password: "s3cret"},
			want: config.MSSQL{
				Metadata: meta, Hostname: "sql.internal", PortNumber: 1433, DBName: "master", EncryptMode: config.MSSQLEncryptLogin,
			},
			wantWarnings: []string{"the password is not exported: a context without a user holds none"},
		},
		{
			name: "ssh with a key file",
			args: args{conn: config.SSH{
				Metadata: meta, Hostname: "bastion.internal", PortNumber: 2222, User: "ops",
				Auth: config.SSHAuthKey, Password: "filepath:~/.ssh/my key", Jump: "ops@gate:22, gate2",
				ForwardAgent: true, KeepAlive: 30, LocalForward: "5432:localhost:5432",
			}},
			want: config.SSH{
				Metadata: meta, Hostname: "bastion.internal", PortNumber: 2222, User: "ops",
				Auth: config.SSHAuthKey, Password: "filepath:~/.ssh/my key", Jump: "ops@gate:22,gate2",
				ForwardAgent: true, KeepAlive: 30, LocalForward: "5432:localhost:5432",
			},
		},
		{
			name: "ssh with a password comes back as agent auth",
			args: args{conn: config.SSH{
				Metadata: meta, Hostname: "bastion.internal", PortNumber: 22, User: "ops",
				Auth: config.SSHAuthPassword, Password: "keyring:ssh/orders",
			}, password: "s3cret"},
			want: config.SSH{
				Metadata: meta, Hostname: "bastion.internal", PortNumber: 22, User: "ops", Auth: config.SSHAuthAgent,
			},
			wantWarnings: []string{"the password is not exported: an ssh config holds none"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			kind := tt.args.conn.ConnType()
			path := filepath.Join(t.TempDir(), "sub", tt.args.file+"file")
			if tt.args.file != "" {
				path = filepath.Join(t.TempDir(), tt.args.file)
			}

			planned, err := exporter.Plan(kind, path, []config.Connection{tt.args.conn}, []string{tt.args.password})
			must.NoError(err)
			must.Len(planned, 1)
			must.NoError(planned[0].Err)
			is.Equal(tt.wantWarnings, planned[0].Warnings)

			must.NoError(exporter.Write(kind, path, planned))

			read, err := importer.Read(kind, path)
			must.NoError(err)
			must.Len(read, 1)
			must.NoError(read[0].Err)
			is.Equal("orders", read[0].Name)
			is.Equal(tt.want, read[0].Conn)

			again, err := exporter.Plan(kind, path, []config.Connection{tt.args.conn}, []string{""})
			must.NoError(err)
			is.Error(again[0].Err)
		})
	}
}

func TestPlan_refusals(t *testing.T) {
	named := func(name string) config.Connection {
		return config.Postgres{Metadata: config.ConnMeta{Name: name}, Hostname: "db.internal", PortNumber: 5432, User: "app", DBName: "orders"}
	}

	path := filepath.Join(t.TempDir(), "pg_service.conf")
	require.NoError(t, os.WriteFile(path, []byte("[taken]\nhost=a\n"), 0o600))

	conns := []config.Connection{named("fresh"), named("taken"), named("two words"), named("fresh"), named("broken")}
	passwords := []string{"", "", "", "", "a\n[other]"}
	planned, err := exporter.Plan(config.PostgresConnType, path, conns, passwords)
	require.NoError(t, err)
	require.Len(t, planned, 5)

	require.NoError(t, planned[0].Err)
	require.Error(t, planned[1].Err)
	require.Error(t, planned[2].Err)
	require.Error(t, planned[3].Err)
	require.Error(t, planned[4].Err)

	require.NoError(t, exporter.Write(config.PostgresConnType, path, planned))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, "[taken]\nhost=a\n\n[fresh]\nhost=db.internal\nport=5432\nuser=app\ndbname=orders\n", string(got))
}

func TestPlan_unsupported(t *testing.T) {
	assert.False(t, exporter.Supported(config.RedisConnType))

	_, err := exporter.Plan(config.RedisConnType, "anywhere", nil, nil)
	assert.Error(t, err)
}
