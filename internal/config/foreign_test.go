package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeTemp(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "file")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

func TestPGPass_Password(t *testing.T) {
	const file = `# hostname:port:database:username:password
db.internal:5432:orders:app:first
db.internal:5432:orders:app:shadowed
db.internal:*:*:admin:any-db
*:*:*:colon\:user:with\:colon\\slash
too:few:fields
`
	type args struct {
		host     string
		port     string
		database string
		user     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "the first matching row wins",
			args: args{host: "db.internal", port: "5432", database: "orders", user: "app"},
			want: "first",
		},
		{
			name: "wildcards match any value",
			args: args{host: "db.internal", port: "6543", database: "billing", user: "admin"},
			want: "any-db",
		},
		{
			name: "escaped colons and backslashes are part of the field",
			args: args{host: "elsewhere", port: "5432", database: "x", user: "colon:user"},
			want: `with:colon\slash`,
		},
		{
			name: "no row for the connection",
			args: args{host: "db.internal", port: "5432", database: "orders", user: "nobody"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pass, err := ReadPGPass(writeTemp(t, file))
			require.NoError(t, err)
			assert.Equal(t, tt.want, pass.Password(tt.args.host, tt.args.port, tt.args.database, tt.args.user))
		})
	}
}

func TestReadPGPass_missingFile(t *testing.T) {
	pass, err := ReadPGPass(filepath.Join(t.TempDir(), "absent"))
	require.NoError(t, err)
	assert.Empty(t, pass.Password("db.internal", "5432", "orders", "app"))
}

func TestReadSSHConfig(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name string
		args args
		want []ForeignSection
	}{
		{
			name: "an empty file still has the block of options given before any Host",
			want: []ForeignSection{{Name: "*"}},
		},
		{
			name: "options before the first Host apply to every host",
			args: args{content: "User early\nHost one\n  Port 2222\n"},
			want: []ForeignSection{
				{Name: "*", Params: []ForeignParam{{Key: "User", Value: "early"}}},
				{Name: "one", Params: []ForeignParam{{Key: "Port", Value: "2222"}}},
			},
		},
		{
			name: "a Host line keeps all its patterns, a keyword is cut at a space, a tab or an equals sign",
			args: args{content: "# comment\nHost one *.internal !two\n  User ops\n\tPort\t2222\n  HostName=one.example.com\n  IdentityFile = \"~/.ssh/my key\"\n"},
			want: []ForeignSection{
				{Name: "*"},
				{Name: "one *.internal !two", Params: []ForeignParam{
					{Key: "User", Value: "ops"},
					{Key: "Port", Value: "2222"},
					{Key: "HostName", Value: "one.example.com"},
					{Key: "IdentityFile", Value: "~/.ssh/my key"},
				}},
			},
		},
		{
			name: "the options of a Match block are dropped until the next Host",
			args: args{content: "Host one\n  User ops\nMatch host one\n  User matched\nhost two\n  User dev\n"},
			want: []ForeignSection{
				{Name: "*"},
				{Name: "one", Params: []ForeignParam{{Key: "User", Value: "ops"}}},
				{Name: "two", Params: []ForeignParam{{Key: "User", Value: "dev"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadSSHConfig(writeTemp(t, tt.args.content))
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadSQLConfig(t *testing.T) {
	type args struct {
		content string
	}
	tests := []struct {
		name    string
		args    args
		want    []ForeignSection
		wantErr bool
	}{
		{
			name: "a context is joined with the endpoint and the user it names",
			args: args{content: "endpoints:\n  - name: local\n    endpoint:\n      address: localhost\n      port: 1433\n" +
				"contexts:\n  - name: dev\n    context:\n      endpoint: local\n      user: sa\n" +
				"users:\n  - name: sa\n    authentication-type: basic\n    basic-auth:\n      username: sa\n      password-encryption: none\n      password: czNjcmV0\n"},
			want: []ForeignSection{{Name: "dev", Params: []ForeignParam{
				{Key: SQLConfigAddress, Value: "localhost"},
				{Key: SQLConfigPort, Value: "1433"},
				{Key: SQLConfigAuthType, Value: "basic"},
				{Key: SQLConfigUsername, Value: "sa"},
				{Key: SQLConfigPassword, Value: "czNjcmV0"},
				{Key: SQLConfigPasswordEncryption, Value: "none"},
			}}},
		},
		{
			name: "an endpoint or a user that is not there leaves the context without it",
			args: args{content: "contexts:\n  - name: orphan\n    context:\n      endpoint: gone\n      user: nobody\n"},
			want: []ForeignSection{{Name: "orphan"}},
		},
		{
			name:    "not yaml",
			args:    args{content: "contexts: [\n"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadSQLConfig(writeTemp(t, tt.args.content))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReadClickHouseConnections(t *testing.T) {
	want := []ForeignSection{
		{Name: "local", Params: []ForeignParam{{Key: "hostname", Value: "localhost"}, {Key: "secure", Value: "true"}}},
		{Name: "other", Params: []ForeignParam{{Key: "port", Value: "9440"}}},
	}

	type args struct {
		file    string
		content string
	}
	tests := []struct {
		name    string
		args    args
		want    []ForeignSection
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
		name string
		args args
		want string
	}{
		{
			name: "nothing there falls back to the home config",
			want: filepath.Join("home", ".clickhouse-client", "config.xml"),
		},
		{
			name: "the home config, in whatever format it is",
			args: args{files: []string{filepath.Join("home", ".clickhouse-client", "config.yaml")}},
			want: filepath.Join("home", ".clickhouse-client", "config.yaml"),
		},
		{
			name: "the xdg config wins over the home one",
			args: args{files: []string{
				filepath.Join("home", ".clickhouse-client", "config.xml"),
				filepath.Join("xdg", "clickhouse", "config.yml"),
			}},
			want: filepath.Join("xdg", "clickhouse", "config.yml"),
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

			for _, file := range tt.args.files {
				path := filepath.Join(root, file)
				require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
				require.NoError(t, os.WriteFile(path, nil, 0o600))
			}

			assert.Equal(t, filepath.Join(root, tt.want), ClickHouseClientPath())
		})
	}
}
