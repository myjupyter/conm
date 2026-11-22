package e2e

import (
	"bufio"
	"cmp"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/myjupyter/conm/cmd"
	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/secret"
)

const withPasswords = "--with-passwords"

func TestExport_formats(t *testing.T) {
	type args struct {
		kind   config.ConnType
		file   string
		out    string
		input  string
		target string
		flags  []string
	}
	tests := []struct {
		name             string
		args             args
		prepare          func(t *testing.T, e *e2eEnv)
		check            func(t *testing.T, e *e2eEnv, path string)
		want             string
		wantContains     []string
		wantMissing      []string
		wantOutput       []string
		wantImportOutput []string
		wantErr          bool
		noRoundTrip      bool
	}{
		{
			name: "postgres: a password from .pgpass is imported and written back",
			args: args{
				kind:  config.PostgresConnType,
				file:  "pg_service.conf",
				input: "[noauth]\nhost=localhost\nport=5433\nuser=postgres\ndbname=postgres\n",
				flags: []string{withPasswords},
			},
			prepare: func(t *testing.T, e *e2eEnv) {
				t.Helper()
				require.NoError(t, os.WriteFile(os.Getenv("PGPASSFILE"), []byte("localhost:5433:postgres:postgres:from-pgpass\n"), 0o600))
			},
			wantContains: []string{"[noauth]\n", "password=from-pgpass\n"},
		},
		{
			name: "postgres: a service's own passfile wins over .pgpass",
			args: args{
				kind:  config.PostgresConnType,
				file:  "pg_service.conf",
				input: "[noauth]\nhost=localhost\nport=5433\nuser=postgres\ndbname=postgres\npassfile=$WORK/own.pgpass\n",
				flags: []string{withPasswords},
			},
			prepare: func(t *testing.T, e *e2eEnv) {
				t.Helper()
				require.NoError(t, os.WriteFile(os.Getenv("PGPASSFILE"), []byte("*:*:*:*:from-pgpass\n"), 0o600))
				e.write("own.pgpass", "*:*:*:*:from-own\n")
			},
			wantContains: []string{"password=from-own\n"},
			wantMissing:  []string{"from-pgpass"},
		},
		{
			name: "postgres: quotes, backslashes, hashes and spaces in a password survive",
			args: args{
				kind:  config.PostgresConnType,
				file:  "pg_service.conf",
				input: "[odd]\nhost=localhost\nuser=postgres\ndbname=postgres\npassword=it's a \"p\\ss\" #1\n",
				flags: []string{withPasswords},
			},
			wantContains: []string{"password=it's a \"p\\ss\" #1\n"},
		},
		{
			name: "postgres: a password with a line break refuses its connection",
			args: args{
				kind:  config.PostgresConnType,
				file:  "pg_service.conf",
				input: "[fine]\nhost=fine.internal\nuser=postgres\ndbname=postgres\n",
				flags: []string{withPasswords},
			},
			prepare: func(t *testing.T, e *e2eEnv) {
				t.Helper()
				e.seed(config.PostgresConnType, config.Postgres{
					Metadata: config.ConnMeta{Name: "broken"}, Hostname: "localhost", PortNumber: 5432, User: "postgres", DBName: "postgres", Password: "literal:a\n[other]",
				})
			},
			wantOutput: []string{"[✗] broken", "a line break can't be written to this file"},
			wantErr:    true,
		},
		{
			name: "postgres: with --skip-errors the others are still written",
			args: args{
				kind:  config.PostgresConnType,
				file:  "pg_service.conf",
				input: "[fine]\nhost=fine.internal\nuser=postgres\ndbname=postgres\n",
				flags: []string{withPasswords, "--skip-errors"},
			},
			prepare: func(t *testing.T, e *e2eEnv) {
				t.Helper()
				e.seed(config.PostgresConnType, config.Postgres{
					Metadata: config.ConnMeta{Name: "broken"}, Hostname: "localhost", PortNumber: 5432, User: "postgres", DBName: "postgres", Password: "literal:a\n[other]",
				})
			},
			wantContains: []string{"[fine]\n"},
			wantMissing:  []string{"[broken]", "[other]"},
			wantOutput:   []string{"[✗] broken", "[✓] fine"},
		},
		{
			name: "postgres: every sslmode survives",
			args: args{
				kind: config.PostgresConnType,
				file: "pg_service.conf",
				input: "[disable]\nhost=disable.internal\nuser=u\ndbname=d\nsslmode=disable\n" +
					"[allow]\nhost=allow.internal\nuser=u\ndbname=d\nsslmode=allow\n" +
					"[prefer]\nhost=prefer.internal\nuser=u\ndbname=d\nsslmode=prefer\n" +
					"[require]\nhost=require.internal\nuser=u\ndbname=d\nsslmode=require\n" +
					"[verify-ca]\nhost=verify-ca.internal\nuser=u\ndbname=d\nsslmode=verify-ca\n" +
					"[verify-full]\nhost=verify-full.internal\nuser=u\ndbname=d\nsslmode=verify-full\n",
			},
			wantContains: []string{
				"sslmode=disable\n", "sslmode=allow\n", "sslmode=prefer\n", "sslmode=require\n", "sslmode=verify-ca\n", "sslmode=verify-full\n",
			},
		},
		{
			name: "mysql: the connection goes to a client group, its database to a mysql group",
			args: args{
				kind:  config.MySQLConnType,
				file:  "my.cnf",
				input: "[client_orders]\nhost=h\nuser=u\n[mysql_orders]\ndatabase=d\n",
			},
			want: "[client_orders]\nhost=\"h\"\nport=3306\nuser=\"u\"\n\n[mysql_orders]\ndatabase=\"d\"\n",
		},
		{
			name: "mysql: every TLS mode is written as loose-ssl-mode, preferred as none",
			args: args{
				kind: config.MySQLConnType,
				file: "my.cnf",
				input: "[client_off]\nhost=off.internal\nuser=u\nssl-mode=DISABLED\n[mysql_off]\ndatabase=d\n" +
					"[client_required]\nhost=required.internal\nuser=u\nssl-mode=REQUIRED\n[mysql_required]\ndatabase=d\n" +
					"[client_verified]\nhost=verified.internal\nuser=u\nssl-mode=VERIFY_IDENTITY\n[mysql_verified]\ndatabase=d\n" +
					"[client_preferred]\nhost=preferred.internal\nuser=u\nssl-mode=PREFERRED\n[mysql_preferred]\ndatabase=d\n",
			},
			wantContains: []string{
				"[client_off]\nhost=\"off.internal\"\nport=3306\nuser=\"u\"\nloose-ssl-mode=DISABLED\n",
				"[client_required]\nhost=\"required.internal\"\nport=3306\nuser=\"u\"\nloose-ssl-mode=REQUIRED\n",
				"[client_verified]\nhost=\"verified.internal\"\nport=3306\nuser=\"u\"\nloose-ssl-mode=VERIFY_IDENTITY\n",
				"[client_preferred]\nhost=\"preferred.internal\"\nport=3306\nuser=\"u\"\n",
			},
			wantMissing: []string{"PREFERRED", "\nssl-mode"},
		},
		{
			name: "mysql: escapes, hashes and edge spaces survive in values",
			args: args{
				kind:  config.MySQLConnType,
				file:  "my.cnf",
				input: "[client_odd]\nhost=h\nuser=\"DOM\\\\tom\"\npassword=\" pa\\\"ss\\\\word #1 \"\n[mysql_odd]\ndatabase=d\n",
				flags: []string{withPasswords},
			},
			wantContains: []string{"user=\"DOM\\\\tom\"\n", "password=\" pa\\\"ss\\\\word #1 \"\n"},
		},
		{
			name: "mysql: the [client] connection comes back under its name",
			args: args{
				kind:  config.MySQLConnType,
				file:  "my.cnf",
				input: "[client]\nhost=h\nuser=u\n[mysql]\ndatabase=d\n",
			},
			want: "[client_client]\nhost=\"h\"\nport=3306\nuser=\"u\"\n\n[mysql_client]\ndatabase=\"d\"\n",
		},
		{
			name: "clickhouse: a new file is yaml by its extension",
			args: args{
				kind:  config.ClickHouseConnType,
				file:  "config.xml",
				out:   "config.yml",
				input: "<config><connections_credentials><connection><name>orders</name><hostname>h</hostname></connection></connections_credentials></config>",
			},
			wantContains: []string{"connections_credentials:\n  connection:\n    - name: orders\n"},
		},
		{
			name: "clickhouse: a new file without an extension is xml",
			args: args{
				kind:  config.ClickHouseConnType,
				file:  "config.yaml",
				out:   "client-config",
				input: "connections_credentials:\n  connection:\n    name: orders\n    hostname: h\n",
			},
			wantContains: []string{"<config>\n    <connections_credentials>\n        <connection>\n            <name>orders</name>\n"},
		},
		{
			name: "clickhouse: a .yaml file holding xml stays xml",
			args: args{
				kind:   config.ClickHouseConnType,
				file:   "config.yaml",
				input:  "connections_credentials:\n  connection:\n    name: orders\n    hostname: h\n",
				target: "<config>\n    <user>app</user>\n</config>\n",
			},
			wantContains: []string{"<user>app</user>\n", "<name>orders</name>"},
		},
		{
			name: "clickhouse: a commented-out section is not the section",
			args: args{
				kind:   config.ClickHouseConnType,
				file:   "config.xml",
				input:  "<config><connections_credentials><connection><name>orders</name><hostname>h</hostname></connection></connections_credentials></config>",
				target: "<config>\n    <!-- <connections_credentials></connections_credentials> -->\n    <connections_credentials>\n    </connections_credentials>\n</config>\n",
			},
			want: "<config>\n    <!-- <connections_credentials></connections_credentials> -->\n    <connections_credentials>\n        <connection>\n" +
				"            <name>orders</name>\n            <hostname>h</hostname>\n            <port>9000</port>\n            <user>default</user>\n" +
				"            <database>default</database>\n        </connection>\n    </connections_credentials>\n</config>\n",
		},
		{
			name: "clickhouse: an empty self-closing section is opened, not doubled",
			args: args{
				kind:   config.ClickHouseConnType,
				file:   "config.xml",
				input:  "<config><connections_credentials><connection><name>orders</name><hostname>h</hostname></connection></connections_credentials></config>",
				target: "<clickhouse>\n    <connections_credentials/>\n</clickhouse>\n",
			},
			want: "<clickhouse>\n    <connections_credentials>\n        <connection>\n            <name>orders</name>\n            <hostname>h</hostname>\n" +
				"            <port>9000</port>\n            <user>default</user>\n            <database>default</database>\n        </connection>\n" +
				"    </connections_credentials>\n</clickhouse>\n",
		},
		{
			name: "clickhouse: a section nested deeper than the root is not the client's",
			args: args{
				kind:   config.ClickHouseConnType,
				file:   "config.xml",
				input:  "<config><connections_credentials><connection><name>orders</name><hostname>h</hostname></connection></connections_credentials></config>",
				target: "<config>\n    <profile>\n        <connections_credentials>\n        </connections_credentials>\n    </profile>\n</config>\n",
			},
			wantContains: []string{"    </profile>\n    <connections_credentials>\n        <connection>\n            <name>orders</name>\n"},
		},
		{
			name: "clickhouse: a file without the section gets one under its root",
			args: args{
				kind:   config.ClickHouseConnType,
				file:   "config.xml",
				input:  "<config><connections_credentials><connection><name>orders</name><hostname>h</hostname></connection></connections_credentials></config>",
				target: "<clickhouse>\n    <user>app</user>\n</clickhouse>\n",
			},
			wantContains: []string{"<clickhouse>\n    <user>app</user>\n    <connections_credentials>\n        <connection>\n"},
		},
		{
			name: "clickhouse: a single yaml connection becomes a list",
			args: args{
				kind:   config.ClickHouseConnType,
				file:   "config.yaml",
				input:  "connections_credentials:\n  connection:\n    name: orders\n    hostname: h\n",
				target: "connections_credentials:\n  connection:\n    name: other\n    hostname: elsewhere\n",
			},
			wantContains: []string{"  connection:\n    - name: other\n      hostname: elsewhere\n    - name: orders\n"},
		},
		{
			name: "clickhouse: a yaml file with a second document is left alone",
			args: args{
				kind:   config.ClickHouseConnType,
				file:   "config.yaml",
				input:  "connections_credentials:\n  connection:\n    name: orders\n    hostname: h\n",
				target: "user: app\n---\nother: document\n",
			},
			wantOutput: []string{"more than one yaml document"},
			wantErr:    true,
		},
		{
			name: "clickhouse: passwords yaml would read as another type come back as written",
			args: args{
				kind: config.ClickHouseConnType,
				file: "config.xml",
				out:  "config.yaml",
				input: "<config><connections_credentials>" +
					"<connection><name>octal</name><hostname>octal.internal</hostname><password>0123</password></connection>" +
					"<connection><name>boolean</name><hostname>boolean.internal</hostname><password>true</password></connection>" +
					"<connection><name>nulled</name><hostname>nulled.internal</hostname><password>null</password></connection>" +
					"<connection><name>tilde</name><hostname>tilde.internal</hostname><password>~</password></connection>" +
					"</connections_credentials></config>",
				flags: []string{withPasswords},
			},
			wantContains: []string{`password: "0123"`, `password: "true"`, `password: "null"`, `password: "~"`},
		},
		{
			name: "clickhouse: markup in an xml password is escaped",
			args: args{
				kind:  config.ClickHouseConnType,
				file:  "config.yaml",
				out:   "config.xml",
				input: "connections_credentials:\n  connection:\n    name: orders\n    hostname: h\n    password: \"<&>\\\"'\"\n",
				flags: []string{withPasswords},
			},
			wantContains: []string{"<password>&lt;&amp;&gt;&#34;&#39;</password>"},
		},
		{
			name: "mssql: the current context of a file is kept",
			args: args{
				kind:   config.MSSQLConnType,
				file:   "sqlconfig",
				input:  fixtureMSSQL,
				target: "version: v1\ncurrentcontext: other\n",
			},
			wantContains: []string{"currentcontext: other\n", "  - name: local\n"},
			wantMissing:  []string{"currentcontext: local"},
		},
		{
			name: "mssql: an endpoint of the same name is a clash",
			args: args{
				kind:   config.MSSQLConnType,
				file:   "sqlconfig",
				input:  fixtureMSSQL,
				target: "endpoints:\n  - name: local\n    endpoint:\n      address: elsewhere\n",
			},
			wantOutput: []string{"already has local"},
			wantErr:    true,
		},
		{
			name: "mssql: a user of the same name is a clash",
			args: args{
				kind:   config.MSSQLConnType,
				file:   "sqlconfig",
				input:  fixtureMSSQL,
				target: "users:\n  - name: local\n    authentication-type: basic\n    basic-auth:\n      username: other\n",
			},
			wantOutput: []string{"already has local"},
			wantErr:    true,
		},
		{
			name: "mssql: a database and encryption settings are warned about",
			args: args{kind: config.MSSQLConnType, file: "sqlconfig"},
			prepare: func(t *testing.T, e *e2eEnv) {
				t.Helper()
				e.seed(config.MSSQLConnType, config.MSSQL{
					Metadata: config.ConnMeta{Name: "orders"}, Hostname: "sql.internal", PortNumber: 1433, User: "sa", DBName: "orders",
					EncryptMode: config.MSSQLEncryptStrict, TrustCert: true,
				})
			},
			wantOutput: []string{
				"database orders is not exported: a context holds none",
				"encryption settings are not exported: a context holds none",
			},
			noRoundTrip: true,
		},
		{
			name: "mssql: a password encrypted by sqlcmd is left out on import",
			args: args{
				kind: config.MSSQLConnType,
				file: "sqlconfig",
				input: "endpoints:\n  - name: win\n    endpoint:\n      address: h\ncontexts:\n  - name: win\n    context:\n      endpoint: win\n      user: win\n" +
					"users:\n  - name: win\n    authentication-type: basic\n    basic-auth:\n      username: sa\n      password-encryption: dpapi\n      password: AQAAANCM\n",
				flags: []string{withPasswords},
			},
			wantImportOutput: []string{"password is encrypted with dpapi and was left out"},
			wantMissing:      []string{"password"},
		},
		{
			name: "ssh: ssh itself reads the exported values, not those of a Host * already there",
			args: args{
				kind:   config.SSHConnType,
				file:   "ssh_config",
				input:  fixtureSSH,
				target: "Host *\n    User admin\n    Port 2200\n    ForwardAgent no\n",
			},
			check: sshResolves,
		},
		{
			name: "ssh: the blocks go above the global options, which a Host * keeps global",
			args: args{
				kind:   config.SSHConnType,
				file:   "ssh_config",
				input:  "Host one\n    HostName one.internal\n    User ops\n",
				target: "User me\nInclude ~/.ssh/extra\n\nHost *\n    User admin\n",
			},
			want:  "Host one\n    HostName one.internal\n    Port 22\n    User ops\n\nHost *\nUser me\nInclude ~/.ssh/extra\n\nHost *\n    User admin\n",
			check: sshResolves,
		},
		{
			name: "ssh: and above a Match block",
			args: args{
				kind:   config.SSHConnType,
				file:   "ssh_config",
				input:  "Host one\n    HostName one.internal\n    User ops\n",
				target: "Match host legacy\n    User nobody\n",
			},
			want: "Host one\n    HostName one.internal\n    Port 22\n    User ops\n\nMatch host legacy\n    User nobody\n",
		},
		{
			name: "ssh: one alias of a Host line is a clash",
			args: args{
				kind:   config.SSHConnType,
				file:   "ssh_config",
				input:  fixtureSSH,
				target: "Host web other\n    User x\n",
			},
			wantOutput: []string{"already has web"},
			wantErr:    true,
		},
		{
			name: "ssh: a key file with a space is quoted",
			args: args{
				kind:  config.SSHConnType,
				file:  "ssh_config",
				input: "Host k\n    HostName k.internal\n    User ops\n    IdentityFile \"~/.ssh/my key\"\n",
			},
			wantContains: []string{"    IdentityFile \"~/.ssh/my key\"\n"},
		},
		{
			name: "ssh: a password and a stored key are warned about, never resolved",
			args: args{kind: config.SSHConnType, file: "ssh_config", flags: []string{withPasswords}},
			prepare: func(t *testing.T, e *e2eEnv) {
				t.Helper()
				e.seed(config.SSHConnType,
					config.SSH{Metadata: config.ConnMeta{Name: "pw"}, Hostname: "h", PortNumber: 22, User: "ops", Auth: config.SSHAuthPassword, Password: "literal:s3cret"},
					config.SSH{Metadata: config.ConnMeta{Name: "stored"}, Hostname: "stored.internal", PortNumber: 22, User: "ops", Auth: config.SSHAuthKey, Password: "literal:-----BEGIN OPENSSH PRIVATE KEY-----"},
				)
				forgetSecrets(t, e, config.SSHConnType)
			},
			wantContains: []string{"Host pw\n", "Host stored\n"},
			wantMissing:  []string{"s3cret", "PRIVATE KEY", "IdentityFile"},
			wantOutput: []string{
				"the password is not exported: an ssh config holds none",
				"the key is not a file and is not exported: ssh would fall back to its default identities",
			},
			noRoundTrip: true,
		},
		{
			name: "ssh: a jump list with spaces is written as one argument",
			args: args{kind: config.SSHConnType, file: "ssh_config"},
			prepare: func(t *testing.T, e *e2eEnv) {
				t.Helper()
				e.seed(config.SSHConnType, config.SSH{Metadata: config.ConnMeta{Name: "far"}, Hostname: "h", PortNumber: 22, User: "ops", Auth: config.SSHAuthAgent, Jump: "gate1, ops@gate2:2222"})
			},
			wantContains: []string{"    ProxyJump gate1,ops@gate2:2222\n"},
			noRoundTrip:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)
			kind := tt.args.kind.String()

			e := newE2EEnv(t)
			if tt.prepare != nil {
				tt.prepare(t, e)
			}
			if tt.args.input != "" {
				input := strings.ReplaceAll(tt.args.input, "$WORK", e.path(""))
				out := e.mustRun(kind, "import", "--skip-errors", "--path", e.write("in/"+tt.args.file, input))
				for _, line := range tt.wantImportOutput {
					is.Contains(out, line)
				}
			}

			out := cmp.Or(tt.args.out, tt.args.file)
			path := e.path("out/" + out)
			if tt.args.target != "" {
				e.write("out/"+out, tt.args.target)
			}
			before, readErr := os.ReadFile(path)

			output, err := e.run(append([]string{kind, "export", "--path", path}, tt.args.flags...)...)
			if err != nil {
				output += err.Error()
			}
			for _, line := range tt.wantOutput {
				is.Contains(output, line)
			}
			if tt.wantErr {
				must.Error(err, output)
				after, afterErr := os.ReadFile(path)
				is.Equal(readErr == nil, afterErr == nil)
				is.Equal(string(before), string(after))
				return
			}
			must.NoError(err, output)

			got := e.read(path)
			if tt.want != "" {
				is.Equal(tt.want, got)
			}
			for _, part := range tt.wantContains {
				is.Contains(got, part)
			}
			for _, part := range tt.wantMissing {
				is.NotContains(got, part)
			}
			if tt.check != nil {
				tt.check(t, e, path)
			}
			if tt.noRoundTrip {
				return
			}

			conns := e.connections(tt.args.kind, config.DefaultWorkspace)
			passwords := e.passwords(tt.args.kind, config.DefaultWorkspace)
			written := fileNames(t, tt.args.kind, path)

			second := newE2EEnv(t)
			second.mustRun(kind, "import", "--skip-errors", "--path", path)
			back := second.connections(tt.args.kind, config.DefaultWorkspace)
			backPasswords := second.passwords(tt.args.kind, config.DefaultWorkspace)

			for _, conn := range conns {
				name := conn.Meta().Name
				if !slices.Contains(written, name) {
					continue
				}
				i := slices.IndexFunc(back, func(c config.Connection) bool { return c.Meta().Name == name })
				if is.GreaterOrEqual(i, 0, name) {
					is.Equal(conn, back[i])
				}
				if slices.Contains(tt.args.flags, withPasswords) {
					is.Equal(passwords[name], backPasswords[name], name)
				}
			}
		})
	}
}

const (
	fixtureMSSQL = "endpoints:\n  - name: local\n    endpoint:\n      address: localhost\n      port: 1433\n" +
		"contexts:\n  - name: local\n    context:\n      endpoint: local\n      user: sa\n" +
		"users:\n  - name: sa\n    authentication-type: basic\n    basic-auth:\n      username: sa\n"
	fixtureSSH = "Host bastion\n    HostName bastion.example.com\n    User ops\n    Port 2222\n\n" +
		"Host db web\n    User deploy\n    ProxyJump bastion\n    ForwardAgent yes\n"
)

func forgetSecrets(t *testing.T, e *e2eEnv, kind config.ConnType) {
	t.Helper()

	service, err := config.SecretWorkspace()
	require.NoError(t, err)

	for _, conn := range e.connections(kind, config.DefaultWorkspace) {
		ref, ok := conn.(secret.Reference)
		if !ok {
			continue
		}
		if scheme, location, _ := secret.ParseRef(ref.SecretRef()); scheme == secret.Keyring {
			require.NoError(t, keyring.Delete(service, location))
		}
	}
}

func sshResolves(t *testing.T, e *e2eEnv, path string) {
	t.Helper()

	if _, err := exec.LookPath("ssh"); err != nil {
		t.Skip("ssh is not installed")
	}

	for _, conn := range e.connections(config.SSHConnType, config.DefaultWorkspace) {
		ssh, ok := conn.(config.SSH)
		require.True(t, ok)

		raw, err := exec.CommandContext(t.Context(), "ssh", "-G", "-F", path, ssh.Metadata.Name).Output()
		require.NoError(t, err)

		resolved := map[string]string{}
		scanner := bufio.NewScanner(strings.NewReader(string(raw)))
		for scanner.Scan() {
			key, value, _ := strings.Cut(scanner.Text(), " ")
			resolved[key] = value
		}

		forwardAgent := "no"
		if ssh.ForwardAgent {
			forwardAgent = "yes"
		}
		assert.Equal(t, map[string]string{
			"hostname":     ssh.Hostname,
			"user":         ssh.User,
			"port":         strconv.Itoa(ssh.PortNumber),
			"proxyjump":    cmp.Or(ssh.Jump, "none"),
			"forwardagent": forwardAgent,
		}, map[string]string{
			"hostname":     resolved["hostname"],
			"user":         resolved["user"],
			"port":         resolved["port"],
			"proxyjump":    cmp.Or(resolved["proxyjump"], "none"),
			"forwardagent": resolved["forwardagent"],
		}, ssh.Metadata.Name)
	}
}

func TestImportExport_defaultPath(t *testing.T) {
	format := func(name string) e2eFormat {
		i := slices.IndexFunc(e2eFormats, func(f e2eFormat) bool { return f.name == name })
		return e2eFormats[i]
	}
	home := func(parts ...string) func(e *e2eEnv) string {
		return func(e *e2eEnv) string { return filepath.Join(append([]string{e.root, "home"}, parts...)...) }
	}

	type args struct {
		format  e2eFormat
		readAt  func(e *e2eEnv) string
		writeAt func(e *e2eEnv) string
	}
	tests := []struct {
		name    string
		args    args
		prepare func(t *testing.T, e *e2eEnv)
	}{
		{
			name: "postgres: ~/.pg_service.conf",
			args: args{format: format("postgres"), readAt: home(".pg_service.conf"), writeAt: home(".pg_service.conf")},
		},
		{
			name: "postgres: the file PGSERVICEFILE names",
			args: args{
				format:  format("postgres"),
				readAt:  func(e *e2eEnv) string { return e.path("services/pg_service.conf") },
				writeAt: func(e *e2eEnv) string { return e.path("services/pg_service.conf") },
			},
			prepare: func(t *testing.T, e *e2eEnv) {
				t.Helper()
				t.Setenv("PGSERVICEFILE", e.path("services/pg_service.conf"))
			},
		},
		{
			name: "mysql: ~/.my.cnf",
			args: args{format: format("mysql"), readAt: home(".my.cnf"), writeAt: home(".my.cnf")},
		},
		{
			name: "clickhouse: the user's own config",
			args: args{format: format("clickhouse-xml"), readAt: home(".clickhouse-client", "config.xml"), writeAt: home(".clickhouse-client", "config.xml")},
		},
		{
			name: "clickhouse: the xdg config, in yaml",
			args: args{
				format:  format("clickhouse-yaml"),
				readAt:  func(e *e2eEnv) string { return filepath.Join(e.root, "config", "clickhouse", "config.yaml") },
				writeAt: func(e *e2eEnv) string { return filepath.Join(e.root, "config", "clickhouse", "config.yaml") },
			},
		},
		{
			name: "clickhouse: a file in the working directory is read but never written",
			args: args{
				format:  format("clickhouse-xml"),
				readAt:  func(e *e2eEnv) string { return e.path("clickhouse-client.xml") },
				writeAt: home(".clickhouse-client", "config.xml"),
			},
		},
		{
			name: "mssql: ~/.sqlcmd/sqlconfig",
			args: args{format: format("mssql"), readAt: home(".sqlcmd", "sqlconfig"), writeAt: home(".sqlcmd", "sqlconfig")},
		},
		{
			name: "ssh: ~/.ssh/config",
			args: args{format: format("ssh"), readAt: home(".ssh", "config"), writeAt: home(".ssh", "config")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)
			f := tt.args.format
			kind := f.kind.String()

			e := newE2EEnv(t)
			if tt.prepare != nil {
				tt.prepare(t, e)
			}
			readAt, writeAt := tt.args.readAt(e), tt.args.writeAt(e)
			must.NoError(os.MkdirAll(filepath.Dir(readAt), 0o700))
			must.NoError(os.WriteFile(readAt, []byte(fixture(t, f.fixture)), 0o600))

			e.mustRun(kind, "import", "--skip-errors")
			must.Equal(f.imported, names(e.connections(f.kind, config.DefaultWorkspace)))

			if readAt == writeAt {
				must.NoError(os.WriteFile(readAt, nil, 0o600))
			}
			e.mustRun(kind, "export")

			is.Equal(f.imported, fileNames(t, f.kind, writeAt))
			if readAt != writeAt {
				is.Equal(fixture(t, f.fixture), e.read(readAt))
			}
		})
	}
}

func TestImportExport_unsupported(t *testing.T) {
	type args struct {
		kind config.ConnType
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "redis", args: args{kind: config.RedisConnType}},
		{name: "mongodb", args: args{kind: config.MongoDBConnType}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := assert.New(t)

			e := newE2EEnv(t)
			for _, sub := range []string{"import", "export"} {
				found, _, err := cmd.NewRootCmd().Find([]string{tt.args.kind.String(), sub})
				require.NoError(t, err)
				is.NotEqual(sub, found.Name())

				_, err = e.run(tt.args.kind.String(), sub, "--path", e.path("anything"))
				is.Error(err)
			}
		})
	}
}
