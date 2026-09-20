package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/pelletier/go-toml/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hostOnly struct{}

func (hostOnly) Meta() ConnMeta                 { return ConnMeta{} }
func (hostOnly) Username() string               { return "" }
func (hostOnly) Host() string                   { return "" }
func (hostOnly) Port() int                      { return 0 }
func (hostOnly) ConnectionString(string) string { return "" }
func (hostOnly) ConnType() ConnType             { return 0 }
func (hostOnly) IsValid() bool                  { return false }
func (hostOnly) Validate() []error              { return nil }

var _ Connection = hostOnly{}

func TestReadConm(t *testing.T) {
	tests := map[string]struct {
		toml string
		want ConnectionSettings
	}{
		"kind kept": {
			toml: "[[conm.connection]]\nkind = \"database\"\ntype = \"postgres\"\ncli = \"psql\"\nenabled = true\n",
			want: ConnectionSettings{Kind: DatabaseConnKind, Type: PostgresConnType, CLI: "psql", Enabled: true},
		},
		"kind defaults to database": {
			toml: "[[conm.connection]]\ntype = \"postgres\"\ncli = \"psql\"\nenabled = true\n",
			want: ConnectionSettings{Kind: DatabaseConnKind, Type: PostgresConnType, CLI: "psql", Enabled: true},
		},
		"old database section is dropped": {
			toml: "[[conm.database]]\ntype = \"postgres\"\ncli = \"psql\"\nenabled = true\n",
			want: ConnectionSettings{Kind: DatabaseConnKind, Type: PostgresConnType},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "conm.toml")
			require.NoError(t, os.WriteFile(path, []byte(tt.toml), 0o600))

			conm, err := ReadConm(path)
			require.NoError(t, err)

			assert.Len(t, conm.Connections, len(Databases))
			got, ok := conm.Connection(PostgresConnType)
			require.True(t, ok)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConmMarshalsKindAsText(t *testing.T) {
	raw, err := toml.Marshal(ConmConfigWrapper{Conm: Conm{Connections: []ConnectionSettings{
		{Kind: DatabaseConnKind, Type: PostgresConnType, CLI: "psql", Enabled: true},
	}}})
	require.NoError(t, err)

	assert.Contains(t, string(raw), "[[conm.connection]]")
	assert.Contains(t, string(raw), `kind = 'database'`)
}
