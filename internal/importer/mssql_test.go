package importer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestRead_mssql(t *testing.T) {
	const endpoints = "endpoints:\n" +
		"  - name: local\n    endpoint:\n      address: localhost\n      port: 1433\n" +
		"  - name: reports\n    endpoint:\n      address: reports.internal\n      port: 1444\n"

	type args struct {
		content string
	}
	type want struct {
		name     string
		conn     config.Connection
		warnings int
		err      bool
	}
	tests := []struct {
		name string
		args args
		want []want
	}{
		{
			name: "a context is its endpoint and its user, the password decoded from base64",
			args: args{content: endpoints +
				"contexts:\n  - name: local\n    context:\n      endpoint: local\n      user: sa\n" +
				"users:\n  - name: sa\n    authentication-type: basic\n    basic-auth:\n      username: sa\n      password-encryption: none\n      password: czNjcmV0\n"},
			want: []want{{
				name: "local",
				conn: config.MSSQL{
					Metadata:    config.ConnMeta{Name: "local"},
					Hostname:    "localhost",
					PortNumber:  1433,
					User:        "sa",
					Password:    "literal:s3cret",
					DBName:      "master",
					EncryptMode: config.MSSQLEncryptLogin,
				},
				warnings: 1,
			}},
		},
		{
			name: "an encrypted password is left out with a warning",
			args: args{content: endpoints +
				"contexts:\n  - name: reports\n    context:\n      endpoint: reports\n      user: reader\n" +
				"users:\n  - name: reader\n    authentication-type: basic\n    basic-auth:\n      username: reader\n      password-encryption: dpapi\n      password: AQAAANCM\n"},
			want: []want{{
				name: "reports",
				conn: config.MSSQL{
					Metadata:    config.ConnMeta{Name: "reports"},
					Hostname:    "reports.internal",
					PortNumber:  1444,
					User:        "reader",
					DBName:      "master",
					EncryptMode: config.MSSQLEncryptLogin,
				},
				warnings: 2,
			}},
		},
		{
			name: "a context without a user has no username, one without an endpoint fails alone",
			args: args{content: endpoints +
				"contexts:\n  - name: anonymous\n    context:\n      endpoint: local\n  - name: orphan\n    context:\n      endpoint: gone\n"},
			want: []want{
				{
					name: "anonymous",
					conn: config.MSSQL{
						Metadata:    config.ConnMeta{Name: "anonymous"},
						Hostname:    "localhost",
						PortNumber:  1433,
						DBName:      "master",
						EncryptMode: config.MSSQLEncryptLogin,
					},
					warnings: 1,
				},
				{name: "orphan", err: true},
			},
		},
		{
			name: "a config without contexts",
			args: args{content: endpoints},
			want: []want{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "sqlconfig")
			require.NoError(t, os.WriteFile(path, []byte(tt.args.content), 0o600))

			entries, err := Read(config.MSSQLConnType, path)
			require.NoError(t, err)

			got := make([]want, 0, len(entries))
			for _, e := range entries {
				got = append(got, want{
					name:     e.Name,
					conn:     e.Conn,
					warnings: len(e.Warnings),
					err:      e.Err != nil,
				})
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
