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
	}
	type want struct {
		name     string
		conn     config.Connection
		password string
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
					Password:   "keyring:postgres/orders",
					DBName:     "orders",
					SSLMode:    config.PostgresSSLModePrefer,
				},
				password: `it's \ here`,
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
					Password:   "keyring:postgres/orders",
					DBName:     "orders",
					SSLMode:    config.PostgresSSLModePrefer,
				},
				password: "from-pgpass",
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
			t.Setenv("PGPASSFILE", pgpass)

			entries, err := Read(config.PostgresConnType, services)
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
