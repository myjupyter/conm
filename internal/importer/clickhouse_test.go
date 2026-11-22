package importer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
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
			Password:   "literal:s3cret",
			DBName:     "orders",
		},
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

func TestReadClickHouseConnections(t *testing.T) {
	want := []Section{
		{Name: "local", Params: []Param{{Key: "hostname", Value: "localhost"}, {Key: "secure", Value: "true"}}},
		{Name: "other", Params: []Param{{Key: "port", Value: "9440"}}},
	}

	type args struct {
		file    string
		content string
	}
	tests := []struct {
		name    string
		args    args
		want    []Section
		wantErr bool
	}{
		{
			name: "xml, under any root element",
			args: args{file: "config.xml", content: "<clickhouse><connections_credentials>" +
				"<connection><name>local</name><hostname> localhost </hostname><secure>true</secure></connection>" +
				"<connection><name>other</name><port>9440</port></connection>" +
				"</connections_credentials></clickhouse>"},
			want: want,
		},
		{
			name: "yaml",
			args: args{file: "config.yaml", content: "connections_credentials:\n  connection:\n" +
				"    - name: local\n      hostname: localhost\n      secure: true\n    - name: other\n      port: 9440\n"},
			want: want,
		},
		{
			name: "the format is told by the content, not the file name",
			args: args{file: "clickhouse-client", content: "<config><connections_credentials>" +
				"<connection><name>other</name><port>9440</port></connection>" +
				"</connections_credentials></config>"},
			want: want[1:],
		},
		{
			name: "a yaml null is no value at all",
			args: args{file: "config.yaml", content: "connections_credentials:\n  connection:\n" +
				"    name: other\n    port: 9440\n    password: ~\n    user: null\n"},
			want: want[1:],
		},
		{
			name: "yaml without the section",
			args: args{file: "config.yml", content: "user: app\n"},
		},
		{
			name:    "broken xml",
			args:    args{file: "config.xml", content: "<config>"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.args.file)
			require.NoError(t, os.WriteFile(path, []byte(tt.args.content), 0o600))

			got, err := ReadClickHouseConnections(path)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClickHouseClientPath(t *testing.T) {
	type args struct {
		files []string
	}
	tests := []struct {
		name     string
		args     args
		want     string
		wantUser string
	}{
		{
			name:     "nothing there falls back to the home config",
			want:     filepath.Join("home", ".clickhouse-client", "config.xml"),
			wantUser: filepath.Join("home", ".clickhouse-client", "config.xml"),
		},
		{
			name:     "the home config, in whatever format it is",
			args:     args{files: []string{filepath.Join("home", ".clickhouse-client", "config.yaml")}},
			want:     filepath.Join("home", ".clickhouse-client", "config.yaml"),
			wantUser: filepath.Join("home", ".clickhouse-client", "config.yaml"),
		},
		{
			name: "the xdg config wins over the home one",
			args: args{files: []string{
				filepath.Join("home", ".clickhouse-client", "config.xml"),
				filepath.Join("xdg", "clickhouse", "config.yml"),
			}},
			want:     filepath.Join("xdg", "clickhouse", "config.yml"),
			wantUser: filepath.Join("xdg", "clickhouse", "config.yml"),
		},
		{
			name:     "the system config is read, never written: the user's own is",
			args:     args{files: []string{filepath.Join("etc", "config.xml")}},
			want:     filepath.Join("etc", "config.xml"),
			wantUser: filepath.Join("home", ".clickhouse-client", "config.xml"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			t.Setenv("HOME", filepath.Join(root, "home"))
			t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "xdg"))
			xdg.Reload()
			t.Cleanup(xdg.Reload)
			t.Chdir(root)

			systemDir := clickhouseSystemDir
			clickhouseSystemDir = filepath.Join(root, "etc")
			t.Cleanup(func() { clickhouseSystemDir = systemDir })

			for _, file := range tt.args.files {
				path := filepath.Join(root, file)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
				require.NoError(t, os.WriteFile(path, nil, 0o600))
			}

			assert.Equal(t, filepath.Join(root, tt.want), ClickHouseClientPath())
			assert.Equal(t, filepath.Join(root, tt.wantUser), ClickHouseUserPath())
		})
	}
}
