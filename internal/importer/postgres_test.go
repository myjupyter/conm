package importer

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
)

func TestRead_postgres(t *testing.T) {
	type args struct {
		services string
		pgpass   string
		passfile string
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
			name: "a service becomes a connection named after it",
			args: args{services: "[orders]\nhost=db.internal\nport=5433\nuser=app\ndbname=orders\nsslmode=require\n"},
			want: []want{{
				name: "orders",
				conn: config.Postgres{
					Metadata:   config.ConnMeta{Name: "orders"},
					Hostname:   "db.internal",
					PortNumber: 5433,
					User:       "app",
					DBName:     "orders",
					SSLMode:    config.PostgresSSLModeRequire,
				},
			}},
		},
		{
			name: "a password in the service is taken out of the connection",
			args: args{services: "[orders]\nhost=db.internal\nuser=app\ndbname=orders\npassword=it's \\ here\n"},
			want: []want{{
				name: "orders",
				conn: config.Postgres{
					Metadata:   config.ConnMeta{Name: "orders"},
					Hostname:   "db.internal",
					PortNumber: 5432,
					User:       "app",
					Password:   "literal:" + `it's \ here`,
					DBName:     "orders",
					SSLMode:    config.PostgresSSLModePrefer,
				},
			}},
		},
		{
			name: "a service without a password gets it from the password file",
			args: args{
				services: "[orders]\nhost=db.internal\nuser=app\ndbname=orders\n",
				pgpass:   "db.internal:5432:*:app:from-pgpass\n",
			},
			want: []want{{
				name: "orders",
				conn: config.Postgres{
					Metadata:   config.ConnMeta{Name: "orders"},
					Hostname:   "db.internal",
					PortNumber: 5432,
					User:       "app",
					Password:   "literal:from-pgpass",
					DBName:     "orders",
					SSLMode:    config.PostgresSSLModePrefer,
				},
			}},
		},
		{
			name: "a service that names its own password file is looked up there, an absent host as localhost",
			args: args{
				services: "[local]\nuser=app\ndbname=orders\npassfile=own-pgpass\n",
				pgpass:   "localhost:5432:*:app:from-the-shared-file\n",
				passfile: "localhost:5432:*:app:from-its-own-file\n",
			},
			want: []want{{
				name: "local",
				conn: config.Postgres{
					Metadata:   config.ConnMeta{Name: "local"},
					PortNumber: 5432,
					User:       "app",
					Password:   "literal:from-its-own-file",
					DBName:     "orders",
					SSLMode:    config.PostgresSSLModePrefer,
				},
				warnings: 1,
			}},
		},
		{
			name: "the first of a repeated keyword is the one the client uses",
			args: args{services: "[orders]\nhost=first\nuser=app\ndbname=orders\nhost=second\n"},
			want: []want{{
				name: "orders",
				conn: config.Postgres{
					Metadata:   config.ConnMeta{Name: "orders"},
					Hostname:   "first",
					PortNumber: 5432,
					User:       "app",
					DBName:     "orders",
					SSLMode:    config.PostgresSSLModePrefer,
				},
			}},
		},
		{
			name: "an unknown keyword is a warning, an unsupported one fails only its service",
			args: args{services: "[warned]\nhost=a\nuser=u\ndbname=d\nconnect_timeout=5\n[refused]\nhost=b\nsslrootcert=/ca.pem\n"},
			want: []want{
				{
					name: "warned",
					conn: config.Postgres{
						Metadata:   config.ConnMeta{Name: "warned"},
						Hostname:   "a",
						PortNumber: 5432,
						User:       "u",
						DBName:     "d",
						SSLMode:    config.PostgresSSLModePrefer,
					},
					warnings: 1,
				},
				{name: "refused", err: true},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			services, pgpass := filepath.Join(dir, "pg_service.conf"), filepath.Join(dir, "pgpass")
			require.NoError(t, os.WriteFile(services, []byte(tt.args.services), 0o600))
			require.NoError(t, os.WriteFile(pgpass, []byte(tt.args.pgpass), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(dir, "own-pgpass"), []byte(tt.args.passfile), 0o600))
			t.Setenv("PGPASSFILE", pgpass)
			t.Chdir(dir)

			entries, err := Read(config.PostgresConnType, services)
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

func TestRead_postgres_unreadablePasswordFile(t *testing.T) {
	dir := t.TempDir()
	services := filepath.Join(dir, "pg_service.conf")
	require.NoError(t, os.WriteFile(services, []byte("[orders]\nhost=db.internal\nuser=app\ndbname=orders\npassword=own\n"), 0o600))
	t.Setenv("PGPASSFILE", dir)

	entries, err := Read(config.PostgresConnType, services)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	require.NoError(t, entries[0].Err)
	pg, ok := entries[0].Conn.(config.Postgres)
	require.True(t, ok)
	assert.Equal(t, "literal:own", pg.Password)
}
