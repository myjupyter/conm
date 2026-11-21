package ui

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/adrg/xdg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/zalando/go-keyring"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
)

func TestImportConnections(t *testing.T) {
	orders := config.Postgres{
		Metadata:   config.ConnMeta{Name: "orders"},
		Hostname:   "db.internal",
		PortNumber: 5432,
		User:       "app",
		Password:   "keyring:postgres/orders",
		DBName:     "orders",
	}
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
			args:         args{entries: []importer.Entry{{Name: "orders", Conn: orders, Password: "s3cret"}}},
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
				{Name: "no-host", Conn: invalid},
				{Name: "broken", Err: errors.New("unreadable")},
			}},
			want: []ImportResult{
				{Name: "orders", DSN: dsn, Status: ImportNew},
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
			name:      "importing keeps the warnings to itself",
			args:      args{entries: []importer.Entry{{Name: "orders", Conn: orders, Warnings: []string{"ignored parameter x=1"}}}},
			want:      []ImportResult{{Name: "orders", DSN: dsn, Status: ImportDone}},
			wantNames: []string{"orders"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("XDG_CONFIG_HOME", t.TempDir())
			t.Setenv("XDG_DATA_HOME", t.TempDir())
			xdg.Reload()
			t.Cleanup(xdg.Reload)
			keyring.MockInit()
			require.NoError(t, config.UseWorkspace(config.DefaultWorkspace))

			conm, err := config.ReadConm(config.ConmPath())
			require.NoError(t, err)

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
			}
		})
	}
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
