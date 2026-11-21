package importer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestRead_mysql(t *testing.T) {
	type args struct {
		options string
	}
	type want struct {
		name     string
		conn     config.Connection
		password string
		err      bool
	}
	tests := []struct {
		name string
		args args
		want []want
	}{
		{
			name: "the client and mysql groups are one connection, server groups are none",
			args: args{options: "[client]\nhost=db.internal\nuser=app\npassword=\"s3cret\"\n[mysqld]\nport=3310\n[mysql]\ndatabase=orders\n"},
			want: []want{{
				name: "client",
				conn: config.MySQL{
					Metadata:   config.ConnMeta{Name: "client"},
					Hostname:   "db.internal",
					PortNumber: 3306,
					User:       "app",
					Password:   "keyring:mysql/client",
					DBName:     "orders",
					TLSMode:    config.MySQLTLSModePreferred,
				},
				password: "s3cret",
			}},
		},
		{
			name: "a suffixed group is its own connection, inherits from the client group and can take a value back",
			args: args{options: "[client]\nuser=app\npassword=inherited\n[client_staging]\nhost=staging.internal\nport=3307\npassword=\n[mysql_staging]\ndatabase=orders\nssl_mode=REQUIRED\n"},
			want: []want{
				{
					name: "client",
					conn: config.MySQL{
						Metadata:   config.ConnMeta{Name: "client"},
						PortNumber: 3306,
						User:       "app",
						Password:   "keyring:mysql/client",
						TLSMode:    config.MySQLTLSModePreferred,
					},
					password: "inherited",
				},
				{
					name: "staging",
					conn: config.MySQL{
						Metadata:   config.ConnMeta{Name: "staging"},
						Hostname:   "staging.internal",
						PortNumber: 3307,
						User:       "app",
						DBName:     "orders",
						TLSMode:    config.MySQLTLSModeSkipVerify,
					},
				},
			},
		},
		{
			name: "a group that needs certificate files fails alone, includes are not options",
			args: args{options: "!includedir /etc/mysql/conf.d/\n[client-prod]\nhost=prod.internal\n!include /etc/extra.cnf\nssl-ca=/ca.pem\n[client-dev]\nhost=localhost\nuser=root\ndatabase=dev\n"},
			want: []want{
				{name: "prod", err: true},
				{
					name: "dev",
					conn: config.MySQL{
						Metadata:   config.ConnMeta{Name: "dev"},
						Hostname:   "localhost",
						PortNumber: 3306,
						User:       "root",
						DBName:     "dev",
						TLSMode:    config.MySQLTLSModePreferred,
					},
				},
			},
		},
		{
			name: "a file with server groups only holds no connection",
			args: args{options: "[mysqld]\nport=3306\n[client-server]\nsocket=/tmp/mysql.sock\n"},
			want: []want{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "my.cnf")
			require.NoError(t, os.WriteFile(path, []byte(tt.args.options), 0o600))

			entries, err := Read(config.MySQLConnType, path)
			require.NoError(t, err)

			got := make([]want, 0, len(entries))
			for _, e := range entries {
				got = append(got, want{name: e.Name, conn: e.Conn, password: e.Password, err: e.Err != nil})
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
