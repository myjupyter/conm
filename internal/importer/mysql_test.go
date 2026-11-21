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
		name string
		conn config.Connection
		err  bool
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
					Password:   "literal:s3cret",
					DBName:     "orders",
					TLSMode:    config.MySQLTLSModePreferred,
				},
			}},
		},
		{
			name: "a suffixed group is its own connection, inherits from the client group and can take a value back",
			args: args{options: "[client]\nuser=app\npassword=inherited\n[client_staging]\nhost=staging.internal\nport=3307\npassword=\n[mysql_staging]\ndatabase=orders\nssl_mode=REQUIRED\n"},
			want: []want{
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
			name: "a client group without a host is a connection only when it is the only one",
			args: args{options: "[client]\nuser=app\npassword=s3cret\n"},
			want: []want{{
				name: "client",
				conn: config.MySQL{
					Metadata:   config.ConnMeta{Name: "client"},
					PortNumber: 3306,
					User:       "app",
					Password:   "literal:s3cret",
					TLSMode:    config.MySQLTLSModePreferred,
				},
			}},
		},
		{
			name: "an empty value takes the inherited one back and leaves the next option alone",
			args: args{options: "[client]\nuser=app\ndatabase=shop\n[client_reports]\ndatabase=\nhost=reports.internal\n"},
			want: []want{{
				name: "reports",
				conn: config.MySQL{
					Metadata:   config.ConnMeta{Name: "reports"},
					Hostname:   "reports.internal",
					PortNumber: 3306,
					User:       "app",
					TLSMode:    config.MySQLTLSModePreferred,
				},
			}},
		},
		{
			name: "a suffix whose short name is taken or empty keeps its own spelling",
			args: args{options: "[client_a]\nhost=one\nuser=u\ndatabase=d\n[client-a]\nhost=two\nuser=u\ndatabase=d\n[client_]\nhost=three\nuser=u\ndatabase=d\n"},
			want: []want{
				{name: "a", conn: config.MySQL{Metadata: config.ConnMeta{Name: "a"}, Hostname: "one", PortNumber: 3306, User: "u", DBName: "d", TLSMode: config.MySQLTLSModePreferred}},
				{name: "-a", conn: config.MySQL{Metadata: config.ConnMeta{Name: "-a"}, Hostname: "two", PortNumber: 3306, User: "u", DBName: "d", TLSMode: config.MySQLTLSModePreferred}},
				{name: "_", conn: config.MySQL{Metadata: config.ConnMeta{Name: "_"}, Hostname: "three", PortNumber: 3306, User: "u", DBName: "d", TLSMode: config.MySQLTLSModePreferred}},
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
				got = append(got, want{name: e.Name, conn: e.Conn, err: e.Err != nil})
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
