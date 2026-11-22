package e2e

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

type e2eFormat struct {
	name      string
	kind      config.ConnType
	fixture   string
	file      string
	imported  []string
	refused   []string
	passwords map[string]string
	foreign   string
	broken    string
	conn      func(name string) config.Connection
	routed    func(name, transport string) config.Connection
}

var e2eFormats = []e2eFormat{
	{
		name:      "postgres",
		kind:      config.PostgresConnType,
		fixture:   "postgres/pg_service.conf",
		file:      "pg_service.conf",
		imported:  []string{"local", "local-noauth", "reports"},
		refused:   []string{"incomplete", "cloud"},
		passwords: map[string]string{"local": "pg-s3cret"},
		foreign:   "# mine\n[other]\nhost=elsewhere\nuser=app\ndbname=app\n",
		conn: func(name string) config.Connection {
			return config.Postgres{Metadata: config.ConnMeta{Name: name}, Hostname: "db.internal", PortNumber: 5432, User: "app", DBName: "orders"}
		},
		routed: func(name, transport string) config.Connection {
			return config.Postgres{Metadata: config.ConnMeta{Name: name}, Hostname: "db.internal", PortNumber: 5432, User: "app", DBName: "orders", Transport: transport}
		},
	},
	{
		name:      "mysql",
		kind:      config.MySQLConnType,
		fixture:   "mysql/my.cnf",
		file:      "my.cnf",
		imported:  []string{"client", "noauth", "reports"},
		refused:   []string{"cloud"},
		passwords: map[string]string{"client": "my-s3cret", "reports": "my-s3cret"},
		foreign:   "# mine\n[client_other]\nhost=elsewhere\nuser=app\n",
		conn: func(name string) config.Connection {
			return config.MySQL{Metadata: config.ConnMeta{Name: name}, Hostname: "my.internal", PortNumber: 3306, User: "app", DBName: "orders", TLSMode: config.MySQLTLSModePreferred}
		},
		routed: func(name, transport string) config.Connection {
			return config.MySQL{Metadata: config.ConnMeta{Name: name}, Hostname: "my.internal", PortNumber: 3306, User: "app", DBName: "orders", TLSMode: config.MySQLTLSModePreferred, Transport: transport}
		},
	},
	{
		name:      "clickhouse-xml",
		kind:      config.ClickHouseConnType,
		fixture:   "clickhouse/config.xml",
		file:      "config.xml",
		imported:  []string{"analytics.internal", "local"},
		refused:   []string{"cloud"},
		passwords: map[string]string{"local": "ch-s3cret"},
		foreign: "<clickhouse>\n    <!-- mine -->\n    <connections_credentials>\n        <connection>\n" +
			"            <name>other</name>\n        </connection>\n    </connections_credentials>\n</clickhouse>\n",
		broken: "<config>\n    <connections_credentials>\n",
		conn: func(name string) config.Connection {
			return config.ClickHouse{Metadata: config.ConnMeta{Name: name}, Hostname: "ch.internal", PortNumber: 9000, User: "default", DBName: "default"}
		},
		routed: func(name, transport string) config.Connection {
			return config.ClickHouse{Metadata: config.ConnMeta{Name: name}, Hostname: "ch.internal", PortNumber: 9000, User: "default", DBName: "default", Transport: transport}
		},
	},
	{
		name:      "clickhouse-yaml",
		kind:      config.ClickHouseConnType,
		fixture:   "clickhouse/config.yaml",
		file:      "config.yaml",
		imported:  []string{"analytics.internal", "local"},
		refused:   []string{"cloud"},
		passwords: map[string]string{"local": "ch-s3cret"},
		foreign:   "# mine\nconnections_credentials:\n  connection:\n    - name: other\n      hostname: elsewhere\n",
		broken:    "connections_credentials: [\n",
		conn: func(name string) config.Connection {
			return config.ClickHouse{Metadata: config.ConnMeta{Name: name}, Hostname: "ch.internal", PortNumber: 9000, User: "default", DBName: "default"}
		},
		routed: func(name, transport string) config.Connection {
			return config.ClickHouse{Metadata: config.ConnMeta{Name: name}, Hostname: "ch.internal", PortNumber: 9000, User: "default", DBName: "default", Transport: transport}
		},
	},
	{
		name:      "mssql",
		kind:      config.MSSQLConnType,
		fixture:   "mssql/sqlconfig",
		file:      "sqlconfig",
		imported:  []string{"local"},
		refused:   []string{"reports", "orphan"},
		passwords: map[string]string{"local": "ms-s3cret"},
		foreign: "# mine\nversion: v1\nendpoints:\n  - name: other\n    endpoint:\n      address: elsewhere\n" +
			"contexts:\n  - name: other\n    context:\n      endpoint: other\ncurrentcontext: other\n",
		broken: "endpoints: [\n",
		conn: func(name string) config.Connection {
			return config.MSSQL{Metadata: config.ConnMeta{Name: name}, Hostname: "sql.internal", PortNumber: 1433, User: "sa", DBName: "master", EncryptMode: config.MSSQLEncryptLogin}
		},
		routed: func(name, transport string) config.Connection {
			return config.MSSQL{Metadata: config.ConnMeta{Name: name}, Hostname: "sql.internal", PortNumber: 1433, User: "sa", DBName: "master", EncryptMode: config.MSSQLEncryptLogin, Transport: transport}
		},
	},
	{
		name:      "ssh",
		kind:      config.SSHConnType,
		fixture:   "ssh/ssh_config",
		file:      "ssh_config",
		imported:  []string{"bastion", "build.internal", "db", "web"},
		refused:   []string{"nouser"},
		passwords: map[string]string{},
		foreign:   "Host *\n    User admin\n",
		conn: func(name string) config.Connection {
			return config.SSH{Metadata: config.ConnMeta{Name: name}, Hostname: "ssh.internal", PortNumber: 22, User: "ops", Auth: config.SSHAuthAgent}
		},
	},
}

func (f e2eFormat) importFixture(t *testing.T, e *e2eEnv) {
	t.Helper()

	e.mustRun(f.kind.String(), "import", "--skip-errors", "--path", e.write("fixture/"+f.file, fixture(t, f.fixture)))
}

func seedTransport(t *testing.T, e *e2eEnv, name string) {
	t.Helper()

	e.seed(config.SSHConnType, config.SSH{Metadata: config.ConnMeta{Name: name}, Hostname: "bastion.internal", PortNumber: 22, User: "ops", Auth: config.SSHAuthAgent})
	require.NoError(t, os.WriteFile(config.TransportPath(), []byte("[[transport]]\nname = \""+name+"\"\nkind = \"ssh\"\nvia = \""+name+"\"\n"), 0o600))
}

func TestExport_roundTrip(t *testing.T) {
	type args struct {
		format e2eFormat
	}
	type test struct {
		name string
		args args
	}
	tests := make([]test, 0, len(e2eFormats))
	for _, format := range e2eFormats {
		tests = append(tests, test{name: format.name, args: args{format: format}})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)
			f := tt.args.format
			kind := f.kind.String()

			first := newE2EEnv(t)
			f.importFixture(t, first)
			conns := first.connections(f.kind, config.DefaultWorkspace)
			must.Equal(f.imported, names(conns))
			is.Equal(f.passwords, first.passwords(f.kind, config.DefaultWorkspace))

			bare := first.path("bare/" + f.file)
			first.mustRun(kind, "export", "--path", bare)
			got := first.read(bare)
			is.Equal(golden(t, filepath.Join(filepath.Dir(f.fixture), "exported-"+f.file), got), got)
			for _, password := range f.passwords {
				is.NotContains(got, password)
			}

			withPasswords := first.path("with-passwords/" + f.file)
			first.mustRun(kind, "export", "--with-passwords", "--path", withPasswords)
			exported := first.read(withPasswords)
			is.Equal(golden(t, filepath.Join(filepath.Dir(f.fixture), "exported-with-passwords-"+f.file), exported), exported)

			second := newE2EEnv(t)
			second.mustRun(kind, "import", "--path", withPasswords)
			is.Equal(conns, second.connections(f.kind, config.DefaultWorkspace))
			is.Equal(f.passwords, second.passwords(f.kind, config.DefaultWorkspace))

			again := second.path("again/" + f.file)
			second.mustRun(kind, "export", "--with-passwords", "--path", again)
			is.Equal(exported, second.read(again))
		})
	}
}

func TestImport(t *testing.T) {
	type args struct {
		flags []string
		names func(f e2eFormat) []string
		path  func(t *testing.T, e *e2eEnv, f e2eFormat) string
	}
	fixturePath := func(t *testing.T, e *e2eEnv, f e2eFormat) string {
		t.Helper()
		return e.write("fixture/"+f.file, fixture(t, f.fixture))
	}
	none := func(e2eFormat) []string { return []string{} }
	imported := func(f e2eFormat) []string { return f.imported }
	first := func(f e2eFormat) []string { return f.imported[:1] }

	tests := []struct {
		name       string
		args       args
		prepare    func(t *testing.T, e *e2eEnv, f e2eFormat)
		want       func(f e2eFormat) []string
		wantOutput func(f e2eFormat) []string
		wantErr    bool
	}{
		{
			name: "show lists what would be imported and saves nothing",
			args: args{flags: []string{"--show"}, path: fixturePath},
			want: none,
			wantOutput: func(f e2eFormat) []string {
				listed := make([]string, 0, len(f.imported))
				for _, name := range f.imported {
					listed = append(listed, "] "+name+" ")
				}
				return listed
			},
		},
		{
			name:    "one refused entry stops a strict import",
			args:    args{path: fixturePath},
			want:    none,
			wantErr: true,
		},
		{
			name: "skip-errors imports the rest and marks the refused",
			args: args{flags: []string{"--skip-errors"}, path: fixturePath},
			want: imported,
			wantOutput: func(f e2eFormat) []string {
				marked := make([]string, 0, len(f.refused))
				for _, name := range f.refused {
					marked = append(marked, "[✗] "+name)
				}
				return marked
			},
		},
		{
			name:       "a second import finds them already there",
			args:       args{flags: []string{"--skip-errors"}, path: fixturePath},
			prepare:    func(t *testing.T, e *e2eEnv, f e2eFormat) { t.Helper(); f.importFixture(t, e) },
			want:       imported,
			wantOutput: func(e2eFormat) []string { return []string{"[!] "} },
		},
		{
			name: "only the named one",
			args: args{names: first, path: fixturePath},
			want: first,
		},
		{
			name:    "a name the file does not have",
			args:    args{names: func(e2eFormat) []string { return []string{"nobody"} }, path: fixturePath},
			want:    none,
			wantErr: true,
		},
		{
			name: "a file that is not there",
			args: args{path: func(_ *testing.T, e *e2eEnv, f e2eFormat) string {
				return e.path("absent/" + f.file)
			}},
			want:    none,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		for _, f := range e2eFormats {
			t.Run(f.name+"/"+tt.name, func(t *testing.T) {
				is, must := assert.New(t), require.New(t)

				e := newE2EEnv(t)
				if tt.prepare != nil {
					tt.prepare(t, e, f)
				}

				args := append([]string{f.kind.String(), "import", "--path", tt.args.path(t, e, f)}, tt.args.flags...)
				if tt.args.names != nil {
					args = append(args, tt.args.names(f)...)
				}

				out, err := e.run(args...)
				if tt.wantErr {
					must.Error(err)
				} else {
					must.NoError(err, out)
				}
				is.Equal(tt.want(f), names(e.connections(f.kind, config.DefaultWorkspace)))
				if tt.wantOutput != nil {
					for _, line := range tt.wantOutput(f) {
						is.Contains(out, line)
					}
				}
			})
		}
	}
}

func TestExport(t *testing.T) {
	type args struct {
		flags []string
		names func(f e2eFormat) []string
		path  string
	}
	imported := func(f e2eFormat) []string { return f.imported }
	first := func(f e2eFormat) []string { return f.imported[:1] }
	exportFirst := func(t *testing.T, e *e2eEnv, f e2eFormat, path string) {
		t.Helper()
		e.mustRun(f.kind.String(), "export", "--path", path, f.imported[0])
	}
	tests := []struct {
		name       string
		args       args
		prepare    func(t *testing.T, e *e2eEnv, f e2eFormat, path string)
		skip       func(f e2eFormat) bool
		want       func(f e2eFormat) []string
		wantKept   bool
		wantOutput string
		wantErr    bool
	}{
		{
			name: "show writes nothing",
			args: args{flags: []string{"--show"}, path: "out/"},
		},
		{
			name: "only the named one",
			args: args{names: first, path: "out/"},
			want: first,
		},
		{
			name:    "a name conm does not have",
			args:    args{names: func(e2eFormat) []string { return []string{"nobody"} }, path: "out/"},
			wantErr: true,
		},
		{
			name: "a file with other entries keeps them",
			args: args{path: "out/"},
			prepare: func(t *testing.T, e *e2eEnv, f e2eFormat, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte(f.foreign), 0o600))
			},
			want: func(f e2eFormat) []string {
				if f.kind == config.SSHConnType {
					return f.imported
				}
				return slices.Sorted(slices.Values(append([]string{"other"}, f.imported...)))
			},
			wantKept: true,
		},
		{
			name:    "an entry the file already has stops the export",
			args:    args{path: "out/"},
			prepare: exportFirst,
			want:    first,
			wantErr: true,
		},
		{
			name:    "unless the errors are skipped",
			args:    args{flags: []string{"--skip-errors"}, path: "out/"},
			prepare: exportFirst,
			want:    imported,
		},
		{
			name: "a second export of the same connections is refused",
			args: args{path: "out/"},
			prepare: func(t *testing.T, e *e2eEnv, f e2eFormat, path string) {
				t.Helper()
				e.mustRun(f.kind.String(), "export", "--path", path)
			},
			want:    imported,
			wantErr: true,
		},
		{
			name: "a name the file cannot hold stops the export",
			args: args{path: "out/"},
			prepare: func(t *testing.T, e *e2eEnv, f e2eFormat, _ string) {
				t.Helper()
				e.seed(f.kind, f.conn("two words"))
			},
			wantErr: true,
		},
		{
			name: "unless the errors are skipped, then the rest is written",
			args: args{flags: []string{"--skip-errors"}, path: "out/"},
			prepare: func(t *testing.T, e *e2eEnv, f e2eFormat, _ string) {
				t.Helper()
				e.seed(f.kind, f.conn("two words"))
			},
			want: imported,
		},
		{
			name: "a connection through a transport is written as a direct one",
			args: args{flags: []string{"--skip-errors"}, names: func(e2eFormat) []string { return []string{"routed"} }, path: "out/"},
			prepare: func(t *testing.T, e *e2eEnv, f e2eFormat, _ string) {
				t.Helper()
				seedTransport(t, e, "jump")
				e.seed(f.kind, f.routed("routed", "ssh:jump"))
			},
			skip:       func(f e2eFormat) bool { return f.routed == nil },
			want:       func(e2eFormat) []string { return []string{"routed"} },
			wantOutput: "transport ssh:jump is not exported: the client would connect directly",
		},
		{
			name: "a file the format cannot read is left alone",
			args: args{path: "out/"},
			prepare: func(t *testing.T, e *e2eEnv, f e2eFormat, path string) {
				t.Helper()
				require.NoError(t, os.WriteFile(path, []byte(f.broken), 0o600))
			},
			skip:    func(f e2eFormat) bool { return f.broken == "" },
			wantErr: true,
		},
	}
	for _, tt := range tests {
		for _, f := range e2eFormats {
			t.Run(f.name+"/"+tt.name, func(t *testing.T) {
				if tt.skip != nil && tt.skip(f) {
					t.Skip("the format has nothing to test here")
				}
				is, must := assert.New(t), require.New(t)

				e := newE2EEnv(t)
				f.importFixture(t, e)

				path := e.path(tt.args.path + f.file)
				must.NoError(os.MkdirAll(filepath.Dir(path), 0o700))
				if tt.prepare != nil {
					tt.prepare(t, e, f, path)
				}
				before, readErr := os.ReadFile(path)

				args := append([]string{f.kind.String(), "export", "--path", path}, tt.args.flags...)
				if tt.args.names != nil {
					args = append(args, tt.args.names(f)...)
				}

				out, err := e.run(args...)
				if tt.wantErr {
					must.Error(err, out)
					after, afterErr := os.ReadFile(path)
					is.Equal(readErr == nil, afterErr == nil)
					is.Equal(string(before), string(after))
					return
				}
				must.NoError(err, out)

				if tt.want == nil {
					is.NoFileExists(path)
					return
				}
				is.Equal(tt.want(f), fileNames(t, f.kind, path))
				if tt.wantKept {
					is.True(linesKept(f.foreign, e.read(path)), e.read(path))
				}
				is.Contains(out, tt.wantOutput)
			})
		}
	}
}

func TestExport_files(t *testing.T) {
	type args struct {
		format e2eFormat
	}
	type test struct {
		name    string
		args    args
		prepare func(t *testing.T, e *e2eEnv, f e2eFormat) (path, target string)
	}
	tests := make([]test, 0, 2*len(e2eFormats))
	for _, f := range e2eFormats {
		tests = append(tests,
			test{
				name: f.name + "/a file in a folder that does not exist yet is created readable by its owner alone",
				args: args{format: f},
				prepare: func(_ *testing.T, e *e2eEnv, f e2eFormat) (string, string) {
					path := e.path("new/folder/" + f.file)
					return path, path
				},
			},
			test{
				name: f.name + "/a symlinked file is written through the link, which stays",
				args: args{format: f},
				prepare: func(t *testing.T, e *e2eEnv, f e2eFormat) (string, string) {
					t.Helper()
					target := e.write("dotfiles/"+f.file, f.foreign)
					path := e.path("linked/" + f.file)
					require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
					require.NoError(t, os.Symlink(target, path))
					return path, target
				},
			},
		)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)
			f := tt.args.format

			e := newE2EEnv(t)
			f.importFixture(t, e)
			path, target := tt.prepare(t, e, f)

			e.mustRun(f.kind.String(), "export", "--path", path)

			info, err := os.Stat(target)
			must.NoError(err)
			is.Equal(os.FileMode(0o600), info.Mode().Perm())

			link, err := os.Lstat(path)
			must.NoError(err)
			is.Equal(path != target, link.Mode()&os.ModeSymlink != 0)
			if path == target {
				dir, err := os.Stat(filepath.Dir(path))
				must.NoError(err)
				is.Equal(os.FileMode(0o700), dir.Mode().Perm())
			}

			is.Subset(fileNames(t, f.kind, target), f.imported)
		})
	}
}

func TestImportExport_workspace(t *testing.T) {
	type args struct {
		format e2eFormat
	}
	type test struct {
		name string
		args args
	}
	tests := make([]test, 0, len(e2eFormats))
	for _, f := range e2eFormats {
		tests = append(tests, test{name: f.name, args: args{format: f}})
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)
			f := tt.args.format
			kind := f.kind.String()

			e := newE2EEnv(t)
			e.mustRun("workspace", "add", "staging")
			e.mustRun(kind, "import", "--skip-errors", "--workspace", "staging", "--path", e.write("fixture/"+f.file, fixture(t, f.fixture)))

			is.Empty(e.connections(f.kind, config.DefaultWorkspace))
			must.Equal(f.imported, names(e.connections(f.kind, "staging")))

			staging := e.path("staging/" + f.file)
			e.mustRun(kind, "export", "--workspace", "staging", "--path", staging)
			is.Equal(f.imported, fileNames(t, f.kind, staging))

			fromDefault := e.path("default/" + f.file)
			e.mustRun(kind, "export", "--path", fromDefault)
			is.NoFileExists(fromDefault)
		})
	}
}
