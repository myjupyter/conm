package importer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestRead_clickhouse(t *testing.T) {
	type args struct {
		file    string
		content string
	}
	type want struct {
		name     string
		conn     config.Connection
		password string
		warnings int
		err      bool
	}
	local := want{
		name: "local",
		conn: config.ClickHouse{
			Metadata:   config.ConnMeta{Name: "local"},
			Hostname:   "localhost",
			PortNumber: 9000,
			User:       "app",
			Password:   "keyring:clickhouse/local",
			DBName:     "orders",
		},
		password: "s3cret",
	}
	tests := []struct {
		name string
		args args
		want []want
	}{
		{
			name: "an xml connection with every field",
			args: args{
				file: "config.xml",
				content: `<config><connections_credentials><connection>
					<name>local</name><hostname>localhost</hostname><port>9000</port>
					<user>app</user><password>s3cret</password><database>orders</database>
				</connection></connections_credentials></config>`,
			},
			want: []want{local},
		},
		{
			name: "a yaml list of connections",
			args: args{
				file:    "config.yaml",
				content: "connections_credentials:\n  connection:\n    - name: local\n      hostname: localhost\n      port: 9000\n      user: app\n      password: s3cret\n      database: orders\n",
			},
			want: []want{local},
		},
		{
			name: "a single yaml connection is a map, not a list",
			args: args{
				file:    "config.yml",
				content: "connections_credentials:\n  connection:\n    name: local\n    hostname: localhost\n    port: 9000\n    user: app\n    password: s3cret\n    database: orders\n",
			},
			want: []want{local},
		},
		{
			name: "the name is the host when there is none, and the client's own user and database fill in",
			args: args{
				file: "config.xml",
				content: `<config><connections_credentials><connection>
					<name>analytics.internal</name><secure>1</secure><prompt>a></prompt>
				</connection></connections_credentials></config>`,
			},
			want: []want{{
				name: "analytics.internal",
				conn: config.ClickHouse{
					Metadata:   config.ConnMeta{Name: "analytics.internal"},
					Hostname:   "analytics.internal",
					PortNumber: 9440,
					User:       "default",
					DBName:     "default",
					Secure:     true,
				},
				warnings: 3,
			}},
		},
		{
			name: "a connection that turns certificate checks off fails alone",
			args: args{
				file: "config.xml",
				content: `<config><connections_credentials>
					<connection><name>cloud</name><accept-invalid-certificate>true</accept-invalid-certificate></connection>
					<connection><name>local</name><hostname>localhost</hostname><port>9000</port>
						<user>app</user><password>s3cret</password><database>orders</database></connection>
				</connections_credentials></config>`,
			},
			want: []want{{name: "cloud", err: true}, local},
		},
		{
			name: "a config without connections",
			args: args{file: "config.xml", content: `<config><user>app</user></config>`},
			want: []want{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.args.file)
			require.NoError(t, os.WriteFile(path, []byte(tt.args.content), 0o600))

			entries, err := Read(config.ClickHouseConnType, path)
			require.NoError(t, err)

			got := make([]want, 0, len(entries))
			for _, e := range entries {
				got = append(got, want{
					name:     e.Name,
					conn:     e.Conn,
					password: e.Password,
					warnings: len(e.Warnings),
					err:      e.Err != nil,
				})
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRead_clickhouseBrokenFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.xml")
	require.NoError(t, os.WriteFile(path, []byte("<config><connections_credentials>"), 0o600))

	_, err := Read(config.ClickHouseConnType, path)
	assert.Error(t, err)
}
