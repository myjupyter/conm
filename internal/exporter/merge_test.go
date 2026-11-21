package exporter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/importer"
)

func Test_iniMerged(t *testing.T) {
	orders := importer.Section{Name: "orders", Params: []importer.Param{{Key: "host", Value: "db.internal"}, {Key: "port", Value: "5432"}}}
	billing := importer.Section{Name: "billing", Params: []importer.Param{{Key: "host", Value: "pay.internal"}}}

	type args struct {
		raw      string
		sections []importer.Section
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "an empty file gets the sections, a blank line between them",
			args: args{sections: []importer.Section{orders, billing}},
			want: "[orders]\nhost=db.internal\nport=5432\n\n[billing]\nhost=pay.internal\n",
		},
		{
			name: "what the file holds is kept byte for byte",
			args: args{raw: "# mine\n[old]\nhost = a  ; kept", sections: []importer.Section{billing}},
			want: "# mine\n[old]\nhost = a  ; kept\n\n[billing]\nhost=pay.internal\n",
		},
		{
			name:    "a value with a line break would start a line of its own",
			args:    args{sections: []importer.Section{{Name: "orders", Params: []importer.Param{{Key: "password", Value: "a\n[other]"}}}}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := iniMerged("", []byte(tt.args.raw), tt.args.sections)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func Test_sshMerged(t *testing.T) {
	bastion := importer.Section{Name: "bastion", Params: []importer.Param{{Key: "HostName", Value: "bastion.internal"}, {Key: "Port", Value: "2222"}}}

	got, err := sshMerged("", []byte("Host *\n  User ops\n"), []importer.Section{bastion})
	require.NoError(t, err)
	assert.Equal(t, "Host *\n  User ops\n\nHost bastion\n    HostName bastion.internal\n    Port 2222\n", string(got))
}

func Test_held(t *testing.T) {
	type args struct {
		held    func(string) ([]string, error)
		content string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "ssh: every alias of every Host line, patterns included",
			args: args{held: sshHeld, content: "Host web bastion\n  User ops\nHost *.internal\n"},
			want: []string{"*", "web", "bastion", "*.internal"},
		},
		{
			name: "sqlconfig: endpoints, contexts and users, since a new context names all three after itself",
			args: args{held: sqlConfigHeld, content: "endpoints:\n  - name: local\ncontexts:\n  - name: dev\nusers:\n  - name: sa\n"},
			want: []string{"local", "dev", "sa"},
		},
		{
			name: "ini: the sections",
			args: args{held: heldBy(importer.ReadINI), content: "[orders]\nhost=a\n[billing]\n"},
			want: []string{"orders", "billing"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "file")
			require.NoError(t, os.WriteFile(path, []byte(tt.args.content), 0o600))

			got, err := tt.args.held(path)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			absent, err := tt.args.held(filepath.Join(t.TempDir(), "absent"))
			require.NoError(t, err)
			assert.Empty(t, absent)
		})
	}
}

func Test_clickhouseMerged(t *testing.T) {
	added := []importer.Section{
		{Name: "orders", Params: []importer.Param{{Key: "hostname", Value: "ch.internal"}, {Key: "password", Value: "a<b&c"}}},
		{Name: "nulled", Params: []importer.Param{{Key: "password", Value: "null"}, {Key: "secure", Value: "true"}}},
	}
	old := importer.Section{Name: "old", Params: []importer.Param{{Key: "port", Value: "9440"}}}

	type args struct {
		file string
		raw  string
	}
	tests := []struct {
		name         string
		args         args
		want         []importer.Section
		wantContains string
		wantErr      bool
	}{
		{
			name:         "a new file is xml",
			args:         args{file: "config.xml"},
			want:         added,
			wantContains: "<config>\n    <connections_credentials>\n        <connection>\n            <name>orders</name>",
		},
		{
			name:         "a new yaml file by its extension",
			args:         args{file: "config.yaml"},
			want:         added,
			wantContains: "connections_credentials:\n  connection:\n    - name: orders",
		},
		{
			name: "xml: added to the connections there, the rest of the file untouched",
			args: args{file: "config.xml", raw: "<clickhouse>\n  <!-- mine -->\n  <connections_credentials>\n" +
				"    <connection><name>old</name><port>9440</port></connection>\n  </connections_credentials>\n</clickhouse>\n"},
			want:         append([]importer.Section{old}, added...),
			wantContains: "<clickhouse>\n  <!-- mine -->\n  <connections_credentials>\n    <connection><name>old</name><port>9440</port></connection>\n      <connection>",
		},
		{
			name:         "xml without the section gets one under its root",
			args:         args{file: "config.xml", raw: "<clickhouse>\n  <user>app</user>\n</clickhouse>\n"},
			want:         added,
			wantContains: "<user>app</user>\n    <connections_credentials>",
		},
		{
			name:    "xml whose section only closes inside a comment is left alone",
			args:    args{file: "config.xml", raw: "<config><user>app</user></config>\n<!-- </connections_credentials> -->\n"},
			wantErr: true,
		},
		{
			name:         "yaml: a single connection becomes a list, comments kept",
			args:         args{file: "config.yaml", raw: "# mine\nuser: app\nconnections_credentials:\n  connection:\n    name: old\n    port: 9440\n"},
			want:         append([]importer.Section{old}, added...),
			wantContains: "# mine\nuser: app\n",
		},
		{
			name: "yaml: added to the list",
			args: args{file: "config.yaml", raw: "connections_credentials:\n  connection:\n    - name: old\n      port: 9440\n"},
			want: append([]importer.Section{old}, added...),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			path := filepath.Join(t.TempDir(), tt.args.file)

			merged, err := clickhouseMerged(path, []byte(tt.args.raw), added)
			if tt.wantErr {
				must.Error(err)
				return
			}
			must.NoError(err)
			is.Contains(string(merged), tt.wantContains)

			must.NoError(os.WriteFile(path, merged, 0o600))
			got, err := importer.ReadClickHouseConnections(path)
			must.NoError(err)
			is.Equal(tt.want, got)
		})
	}
}

func Test_sqlConfigMerged(t *testing.T) {
	orders := importer.Section{Name: "orders", Params: []importer.Param{
		{Key: importer.SQLConfigAddress, Value: "sql.internal"},
		{Key: importer.SQLConfigPort, Value: "1433"},
		{Key: importer.SQLConfigUsername, Value: "sa"},
		{Key: importer.SQLConfigPassword, Value: "czNjcmV0"},
		{Key: importer.SQLConfigPasswordEncryption, Value: "none"},
	}}
	anonymous := importer.Section{Name: "anonymous", Params: []importer.Param{{Key: importer.SQLConfigAddress, Value: "localhost"}}}
	dev := importer.Section{Name: "dev", Params: []importer.Param{
		{Key: importer.SQLConfigAddress, Value: "localhost"},
		{Key: importer.SQLConfigPort, Value: "1433"},
	}}

	const existing = "# mine\nversion: v1\nendpoints:\n  - name: local\n    endpoint:\n      address: localhost\n      port: 1433\n" +
		"contexts:\n  - name: dev\n    context:\n      endpoint: local\ncurrentcontext: dev\n"

	type args struct {
		raw      string
		sections []importer.Section
	}
	tests := []struct {
		name         string
		args         args
		want         []importer.Section
		wantContains string
	}{
		{
			name:         "a new file gets its version and the three lists",
			args:         args{sections: []importer.Section{orders, anonymous}},
			want:         []importer.Section{orders, anonymous},
			wantContains: "version: v1\nendpoints:\n  - name: orders\n    endpoint:\n      address: sql.internal\n      port: 1433\n",
		},
		{
			name:         "contexts are added to the ones there, comments and other keys kept",
			args:         args{raw: existing, sections: []importer.Section{orders}},
			want:         []importer.Section{dev, orders},
			wantContains: "# mine\nversion: v1\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			merged, err := sqlConfigMerged("", []byte(tt.args.raw), tt.args.sections)
			must.NoError(err)
			is.Contains(string(merged), tt.wantContains)
			is.Contains(string(merged), "authentication-type: basic\n")

			path := filepath.Join(t.TempDir(), "sqlconfig")
			must.NoError(os.WriteFile(path, merged, 0o600))
			got, err := importer.ReadSQLConfig(path)
			must.NoError(err)
			is.Equal(tt.want, got)
		})
	}
}
