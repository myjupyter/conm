package ui

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
	"github.com/myjupyter/conm/internal/secret"
)

func TestImportConnections(t *testing.T) {
	orders := config.Postgres{
		Metadata:   config.ConnMeta{Name: "orders"},
		Hostname:   "db.internal",
		PortNumber: 5432,
		User:       "app",
		DBName:     "orders",
	}
	secured := orders
	secured.Password = "literal:s3cret"

	renamed := orders
	renamed.Metadata.Name = "orders-again"

	invalid := orders
	invalid.Metadata.Name, invalid.Hostname = "no-host", ""

	dsn := orders.ConnectionString("")

	type args struct {
		entries []importer.Entry
		mode    ImportMode
	}
	tests := []struct {
		name         string
		args         args
		want         []ImportResult
		wantNames    []string
		wantPassword string
	}{
		{
			name:         "a clean entry is saved and its password goes to the keyring",
			args:         args{entries: []importer.Entry{{Name: "orders", Conn: secured}}},
			want:         []ImportResult{{Name: "orders", DSN: dsn, Status: ImportDone}},
			wantNames:    []string{"orders"},
			wantPassword: "s3cret",
		},
		{
			name: "the same target twice is left alone the second time and is not an error",
			args: args{entries: []importer.Entry{
				{Name: "orders", Conn: orders},
				{Name: "orders-again", Conn: renamed},
			}},
			want: []ImportResult{
				{Name: "orders", DSN: dsn, Status: ImportDone},
				{Name: "orders-again", DSN: dsn, Status: ImportPresent, Note: "orders"},
			},
			wantNames: []string{"orders"},
		},
		{
			name: "one refused entry keeps every other out",
			args: args{entries: []importer.Entry{
				{Name: "orders", Conn: orders},
				{Name: "orders-again", Conn: renamed},
				{Name: "no-host", Conn: invalid},
				{Name: "broken", Err: errors.New("unreadable")},
			}},
			want: []ImportResult{
				{Name: "orders", DSN: dsn, Status: ImportNew},
				{Name: "orders-again", DSN: dsn, Status: ImportNew, Note: `same connection as "orders"`},
				{Name: "no-host", DSN: invalid.ConnectionString(""), Status: ImportFailed, Note: "validation error: hostname can't be empty"},
				{Name: "broken", Status: ImportFailed, Note: "unreadable"},
			},
			wantNames: []string{},
		},
		{
			name: "skipping errors saves the rest",
			args: args{
				entries: []importer.Entry{{Name: "orders", Conn: orders}, {Name: "no-host", Conn: invalid}},
				mode:    ImportSkipErrors,
			},
			want: []ImportResult{
				{Name: "orders", DSN: dsn, Status: ImportDone},
				{Name: "no-host", DSN: invalid.ConnectionString(""), Status: ImportFailed, Note: "validation error: hostname can't be empty"},
			},
			wantNames: []string{"orders"},
		},
		{
			name: "showing saves nothing and passes the warnings on",
			args: args{
				entries: []importer.Entry{{Name: "orders", Conn: orders, Warnings: []string{"ignored parameter x=1"}}},
				mode:    ImportShow,
			},
			want: []ImportResult{
				{Name: "orders", DSN: dsn, Status: ImportNew, Warnings: []string{"ignored parameter x=1"}},
			},
			wantNames: []string{},
		},
		{
			name: "importing passes the warnings on too",
			args: args{entries: []importer.Entry{{Name: "orders", Conn: orders, Warnings: []string{"ignored parameter x=1"}}}},
			want: []ImportResult{
				{Name: "orders", DSN: dsn, Status: ImportDone, Warnings: []string{"ignored parameter x=1"}},
			},
			wantNames: []string{"orders"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conm := emptyWorkspace(t)

			got, err := ImportConnections(context.Background(), conm, config.PostgresConnType, tt.args.entries, tt.args.mode)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)

			conns, err := Connections(conm, config.PostgresConnType)
			require.NoError(t, err)
			names := make([]string, 0, len(conns))
			for _, c := range conns {
				names = append(names, c.Meta().Name)
			}
			assert.Equal(t, tt.wantNames, names)

			if tt.wantPassword != "" {
				workspace, err := config.SecretWorkspace()
				require.NoError(t, err)
				stored, err := keyring.Get(workspace, "postgres/orders")
				require.NoError(t, err)
				assert.Equal(t, tt.wantPassword, stored)

				require.Len(t, conns, 1)
				assert.Equal(t, "keyring:postgres/orders", secretRefOf(t, conns[0]))
			}
		})
	}
}

func secretRefOf(t *testing.T, conn config.Connection) string {
	t.Helper()

	ref, ok := conn.(secret.Reference)
	require.True(t, ok)

	return ref.SecretRef()
}

func emptyWorkspace(t *testing.T) config.Conm {
	t.Helper()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("XDG_DATA_HOME", t.TempDir())
	xdg.Reload()
	t.Cleanup(xdg.Reload)
	keyring.MockInit()
	require.NoError(t, config.UseWorkspace(config.DefaultWorkspace))

	conm, err := config.ReadConm(config.ConmPath())
	require.NoError(t, err)

	return conm
}

func TestImportConnections_replacesTheSecretAConnectionLeftBehind(t *testing.T) {
	conm := emptyWorkspace(t)
	ctx := context.Background()

	for _, password := range []string{"before", "rotated"} {
		orders := config.Postgres{
			Metadata:   config.ConnMeta{Name: "orders"},
			Hostname:   "db.internal",
			PortNumber: 5432,
			User:       "app",
			Password:   "literal:" + password,
			DBName:     "orders",
		}
		entries := []importer.Entry{{Name: "orders", Conn: orders}}

		got, err := ImportConnections(ctx, conm, config.PostgresConnType, entries, ImportStrict)
		require.NoError(t, err)
		require.Equal(t, []ImportResult{{Name: "orders", DSN: orders.ConnectionString(""), Status: ImportDone}}, got)
		assert.Equal(t, password, storedPassword(t, "postgres/orders"))

		require.NoError(t, RemoveConnection(conm, config.PostgresConnType, "orders"))
	}
}

func storedPassword(t *testing.T, location string) string {
	t.Helper()

	workspace, err := config.SecretWorkspace()
	require.NoError(t, err)
	stored, err := keyring.Get(workspace, location)
	require.NoError(t, err)

	return stored
}

func Test_keyringBound(t *testing.T) {
	meta := config.ConnMeta{Name: "orders", Description: "the orders database", Tags: []string{"prod"}}

	tests := []struct {
		name         string
		conn         config.Connection
		want         config.Connection
		wantPassword string
	}{
		{
			name:         "postgres",
			conn:         config.Postgres{Metadata: meta, Hostname: "db", PortNumber: 5433, User: "app", Password: "literal:s3:cret", DBName: "orders", SchemaName: "public", SSLMode: config.PostgresSSLModeRequire},
			want:         config.Postgres{Metadata: meta, Hostname: "db", PortNumber: 5433, User: "app", Password: "keyring:postgres/orders", DBName: "orders", SchemaName: "public", SSLMode: config.PostgresSSLModeRequire},
			wantPassword: "s3:cret",
		},
		{
			name:         "mysql",
			conn:         config.MySQL{Metadata: meta, Hostname: "db", PortNumber: 3307, User: "app", Password: "literal:s3cret", DBName: "orders", TLSMode: config.MySQLTLSModeSkipVerify},
			want:         config.MySQL{Metadata: meta, Hostname: "db", PortNumber: 3307, User: "app", Password: "keyring:mysql/orders", DBName: "orders", TLSMode: config.MySQLTLSModeSkipVerify},
			wantPassword: "s3cret",
		},
		{
			name:         "clickhouse",
			conn:         config.ClickHouse{Metadata: meta, Hostname: "db", PortNumber: 9440, User: "app", Password: "literal:s3cret", DBName: "orders", Secure: true},
			want:         config.ClickHouse{Metadata: meta, Hostname: "db", PortNumber: 9440, User: "app", Password: "keyring:clickhouse/orders", DBName: "orders", Secure: true},
			wantPassword: "s3cret",
		},
		{
			name:         "mssql",
			conn:         config.MSSQL{Metadata: meta, Hostname: "db", PortNumber: 1434, User: "sa", Password: "literal:s3cret", DBName: "master"},
			want:         config.MSSQL{Metadata: meta, Hostname: "db", PortNumber: 1434, User: "sa", Password: "keyring:mssql/orders", DBName: "master"},
			wantPassword: "s3cret",
		},
		{
			name: "a key file is not a password and stays where it is",
			conn: config.SSH{Metadata: meta, Hostname: "db", PortNumber: 22, User: "ops", Auth: config.SSHAuthKey, Password: "filepath:~/.ssh/id"},
			want: config.SSH{Metadata: meta, Hostname: "db", PortNumber: 22, User: "ops", Auth: config.SSHAuthKey, Password: "filepath:~/.ssh/id"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, write, err := keyringBound(tt.conn.ConnType(), tt.conn)
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.wantPassword, write.Password)
		})
	}
}

func TestImportConnections_fixesARefusedEntryInTheForm(t *testing.T) {
	conm := emptyWorkspace(t)
	ctx := context.Background()

	existing := config.Postgres{
		Metadata:   config.ConnMeta{Name: "orders"},
		Hostname:   "db.internal",
		PortNumber: 5432,
		User:       "app",
		Password:   "literal:kept",
		DBName:     "orders",
		SSLMode:    config.PostgresSSLModePrefer,
	}
	_, err := ImportConnections(ctx, conm, config.PostgresConnType, []importer.Entry{{Name: "orders", Conn: existing}}, ImportStrict)
	require.NoError(t, err)

	clashing := existing
	clashing.Hostname, clashing.Password = "eu.internal", "literal:fresh"

	opened := make(chan struct{}, 1)
	host := &screenHost{
		program: tea.NewProgram(
			screenStack{},
			tea.WithInput(nil),
			tea.WithOutput(io.Discard),
			tea.WithoutSignals(),
			tea.WithWindowSize(160, 50),
			tea.WithFilter(func(_ tea.Model, msg tea.Msg) tea.Msg {
				if _, pushed := msg.(pushScreenMsg); pushed {
					opened <- struct{}{}
				}

				return msg
			}),
		),
		closed: make(chan struct{}),
	}
	go func() {
		_, _ = host.program.Run()
		close(host.closed)
	}()
	activeScreenHost.Store(host)
	t.Cleanup(func() {
		activeScreenHost.Store(nil)
		host.program.Quit()
		<-host.closed
	})

	tab, enter, esc := tea.KeyPressMsg{Code: tea.KeyTab}, tea.KeyPressMsg{Code: tea.KeyEnter}, tea.KeyPressMsg{Code: tea.KeyEscape}
	renameAndSave := []tea.KeyPressMsg{tab, tab, char('i'), char('-'), char('e'), char('u'), esc, enter}
	go func() {
		<-opened
		for _, key := range renameAndSave {
			host.program.Send(key)
		}
	}()

	type outcome struct {
		results []ImportResult
		err     error
	}
	done := make(chan outcome, 1)
	go func() {
		results, err := ImportConnections(ctx, conm, config.PostgresConnType, []importer.Entry{{Name: "orders", Conn: clashing}}, ImportInteractive)
		done <- outcome{results: results, err: err}
	}()

	var got outcome
	select {
	case got = <-done:
	case <-time.After(10 * time.Second):
		require.FailNow(t, "the form was never saved")
	}
	require.NoError(t, got.err)

	saved := clashing
	saved.Metadata.Name, saved.Password = "orders-eu", "keyring:postgres/orders-eu"
	assert.Equal(t, []ImportResult{{Name: "orders-eu", DSN: saved.ConnectionString(""), Status: ImportDone}}, got.results)

	assert.Equal(t, "kept", storedPassword(t, "postgres/orders"))
	assert.Equal(t, "fresh", storedPassword(t, "postgres/orders-eu"))

	conns, err := Connections(conm, config.PostgresConnType)
	require.NoError(t, err)
	require.Len(t, conns, 2)
	assert.Equal(t, saved, conns[1])
}

func TestRenderImport(t *testing.T) {
	type args struct {
		results []ImportResult
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "one line per result, names aligned, notes underneath",
			args: args{results: []ImportResult{
				{Name: "orders", DSN: "postgresql://app@db:5432/orders", Status: ImportDone},
				{Name: "new", DSN: "postgresql://app@db:5432/new", Status: ImportNew},
				{Name: "same", DSN: "postgresql://app@db:5432/same", Status: ImportPresent, Note: "same"},
				{Name: "copy", DSN: "postgresql://app@db:5432/orders", Status: ImportPresent, Note: "orders"},
				{Name: "broken", Status: ImportFailed, Note: "unreadable"},
				{Name: "warned", DSN: "postgresql://app@db:5432/warned", Status: ImportNew, Warnings: []string{"ignored parameter x=1"}},
			}},
			want: "[✓] orders  postgresql://app@db:5432/orders\n" +
				"[○] new     postgresql://app@db:5432/new\n" +
				"[!] same    postgresql://app@db:5432/same\n" +
				"[!] copy    postgresql://app@db:5432/orders\n" +
				"  - already there as \"orders\"\n" +
				"[✗] broken  \n" +
				"  - unreadable\n" +
				"[○] warned  postgresql://app@db:5432/warned\n" +
				"  - ignored parameter x=1\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out strings.Builder
			RenderImport(&out, tt.args.results)
			assert.Equal(t, tt.want, out.String())
		})
	}
}
