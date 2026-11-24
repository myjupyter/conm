package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadConm_table(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		toml        string
		connType    ConnType
		wantColumns []string
		wantColor   string
	}{
		{
			name:        "no setting shows every column of the type",
			connType:    PostgresConnType,
			wantColumns: []string{"name", "username", "host", "port", "database", "ping"},
			wantColor:   "postgres",
		},
		{
			name:        "ssh has no database column",
			connType:    SSHConnType,
			wantColumns: []string{"name", "username", "host", "port", "ping"},
			wantColor:   "ssh",
		},
		{
			name:        "order and colour are kept",
			toml:        "[[conm.connection]]\ntype = \"postgres\"\ncolumns = [\"host\", \"name\", \"ping\"]\ntab_color = \"auto\"\n",
			connType:    PostgresConnType,
			wantColumns: []string{"host", "name", "ping"},
			wantColor:   "auto",
		},
		{
			name:        "unknown, repeated and unavailable columns are dropped, name is kept",
			toml:        "[[conm.connection]]\ntype = \"ssh\"\ncolumns = [\"host\", \"host\", \"database\", \"shadow\"]\ntab_color = \"neon\"\n",
			connType:    SSHConnType,
			wantColumns: []string{"name", "host"},
			wantColor:   "ssh",
		},
		{
			name:        "tags are shown when listed",
			toml:        "[[conm.connection]]\ntype = \"redis\"\ncolumns = [\"name\", \"tags\"]\n",
			connType:    RedisConnType,
			wantColumns: []string{"name", "tags"},
			wantColor:   "redis",
		},
		{
			name:        "another type's colour is kept",
			toml:        "[[conm.connection]]\ntype = \"mysql\"\ntab_color = \"redis\"\n",
			connType:    MySQLConnType,
			wantColumns: []string{"name", "username", "host", "port", "database", "ping"},
			wantColor:   "redis",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ReadConm(writeConm(t, tt.toml))
			require.NoError(t, err)
			conn, ok := got.Connection(tt.connType)
			require.True(t, ok)
			assert.Equal(t, tt.wantColumns, conn.VisibleColumns())
			assert.Equal(t, tt.wantColor, conn.ResolvedTabColor())
		})
	}
}

func TestConnectionSettings_ValidateTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		s       ConnectionSettings
		wantErr bool
	}{
		{name: "defaults", s: ConnectionSettings{Type: PostgresConnType}},
		{name: "reordered", s: ConnectionSettings{Type: PostgresConnType, Columns: []string{"ping", "name"}, TabColor: "#ff0000"}},
		{name: "auto colour", s: ConnectionSettings{Type: RedisConnType, TabColor: "auto"}},
		{name: "type colour", s: ConnectionSettings{Type: RedisConnType, TabColor: "postgres"}},
		{name: "old tab keyword", s: ConnectionSettings{Type: RedisConnType, TabColor: "tab"}, wantErr: true},
		{name: "old type keyword", s: ConnectionSettings{Type: RedisConnType, TabColor: "type"}, wantErr: true},
		{name: "without name", s: ConnectionSettings{Type: PostgresConnType, Columns: []string{"host"}}, wantErr: true},
		{name: "database on ssh", s: ConnectionSettings{Type: SSHConnType, Columns: []string{"name", "database"}}, wantErr: true},
		{name: "twice", s: ConnectionSettings{Type: PostgresConnType, Columns: []string{"name", "name"}}, wantErr: true},
		{name: "gradient colour", s: ConnectionSettings{Type: PostgresConnType, TabColor: "teal>pink"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.s.ValidateTable()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
