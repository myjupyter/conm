package ui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/importer"
)

func TestExportConnections(t *testing.T) {
	orders := config.Postgres{
		Metadata:   config.ConnMeta{Name: "orders"},
		Hostname:   "db.internal",
		PortNumber: 5432,
		User:       "app",
		Password:   "literal:s3cret",
		DBName:     "orders",
	}
	billing := orders
	billing.Metadata.Name, billing.Hostname, billing.Password = "billing", "pay.internal", ""

	const (
		ordersService  = "[orders]\nhost=db.internal\nport=5432\nuser=app\ndbname=orders\n"
		billingService = "[billing]\nhost=pay.internal\nport=5432\nuser=app\ndbname=orders\n"
		held           = "[billing]\nhost=elsewhere\n"
	)

	type args struct {
		content string
		names   []string
		opts    ExportOptions
	}
	tests := []struct {
		name       string
		args       args
		wantStatus []ImportStatus
		wantFile   string
		wantErr    bool
	}{
		{
			name:       "every connection, the password left in conm",
			wantStatus: []ImportStatus{ImportDone, ImportDone},
			wantFile:   ordersService + "\n" + billingService,
		},
		{
			name:       "only the named one, with its password when asked for",
			args:       args{names: []string{"orders"}, opts: ExportOptions{WithPasswords: true}},
			wantStatus: []ImportStatus{ImportDone},
			wantFile:   ordersService + "password=s3cret\n",
		},
		{
			name:       "show writes nothing",
			args:       args{opts: ExportOptions{Show: true}},
			wantStatus: []ImportStatus{ImportNew, ImportNew},
		},
		{
			name:       "an entry the file already has stops the whole export",
			args:       args{content: held},
			wantStatus: []ImportStatus{ImportNew, ImportFailed},
			wantFile:   held,
		},
		{
			name:       "unless the errors are skipped",
			args:       args{content: held, opts: ExportOptions{SkipErrors: true}},
			wantStatus: []ImportStatus{ImportDone, ImportFailed},
			wantFile:   held + "\n" + ordersService,
		},
		{
			name:    "a name conm does not have",
			args:    args{names: []string{"nobody"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is, must := assert.New(t), require.New(t)

			conm := emptyWorkspace(t)
			ctx := context.Background()

			seed := []importer.Entry{{Name: "orders", Conn: orders}, {Name: "billing", Conn: billing}}
			_, err := ImportConnections(ctx, conm, config.PostgresConnType, seed, ImportStrict)
			must.NoError(err)

			tt.args.opts.Path = filepath.Join(t.TempDir(), "pg_service.conf")
			if tt.args.content != "" {
				must.NoError(os.WriteFile(tt.args.opts.Path, []byte(tt.args.content), 0o600))
			}

			got, err := ExportConnections(ctx, conm, config.PostgresConnType, tt.args.names, tt.args.opts)
			if tt.wantErr {
				must.Error(err)
				return
			}
			must.NoError(err)

			statuses := make([]ImportStatus, 0, len(got))
			for _, res := range got {
				statuses = append(statuses, res.Status)
				is.NotContains(res.DSN, "s3cret")
			}
			is.Equal(tt.wantStatus, statuses)

			file, _ := os.ReadFile(tt.args.opts.Path)
			is.Equal(tt.wantFile, string(file))
		})
	}
}
