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
func (hostOnly) Identity() string               { return "" }

var _ Connection = hostOnly{}

func TestReadConm(t *testing.T) {
	t.Parallel()

	withKeybindings := func(override func(*KeybindingSettings)) KeybindingSettings {
		k := DefaultKeybindings()
		override(&k)
		return k
	}

	type args struct {
		toml string
	}
	tests := []struct {
		name            string
		args            args
		wantConnection  ConnectionSettings
		wantKeybindings KeybindingSettings
		wantErr         error
	}{
		{
			name:            "connection row is kept",
			args:            args{toml: "[[conm.connection]]\ntype = \"postgres\"\ncli = \"psql\"\nenabled = true\n"},
			wantConnection:  ConnectionSettings{Type: PostgresConnType, CLI: "psql"},
			wantKeybindings: DefaultKeybindings(),
		},
		{
			name:            "old database section is dropped",
			args:            args{toml: "[[conm.database]]\ntype = \"postgres\"\ncli = \"psql\"\nenabled = true\n"},
			wantConnection:  ConnectionSettings{Type: PostgresConnType},
			wantKeybindings: DefaultKeybindings(),
		},
		{
			name:            "set key is kept, the rest take defaults",
			args:            args{toml: "[conm.keybinding.navigation]\nup = [\"w\"]\n"},
			wantConnection:  ConnectionSettings{Type: PostgresConnType},
			wantKeybindings: withKeybindings(func(k *KeybindingSettings) { k.Navigation.Up = []string{"w"} }),
		},
		{
			name:            "empty list takes the default",
			args:            args{toml: "[conm.keybinding.screen]\nquit = []\n"},
			wantConnection:  ConnectionSettings{Type: PostgresConnType},
			wantKeybindings: DefaultKeybindings(),
		},
		{
			name:    "key shared with a default is refused",
			args:    args{toml: "[conm.keybinding.action]\nping = [\"d\"]\n"},
			wantErr: errSharedKey,
		},
		{
			name:    "unknown key is refused",
			args:    args{toml: "[conm.keybinding.navigation]\nup = [\"up\", \"skdfjs;adjf\"]\n"},
			wantErr: errUnknownKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ReadConm(writeConm(t, tt.args.toml))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)

			is := assert.New(t)
			is.Len(got.Connections, len(ConnTypes))
			connection, ok := got.Connection(PostgresConnType)
			is.True(ok)
			is.Equal(tt.wantConnection, connection)
			is.Equal(tt.wantKeybindings, got.Keybindings)
		})
	}
}

func TestConmMarshalsConnectionSection(t *testing.T) {
	raw, err := toml.Marshal(ConmConfigWrapper{Conm: Conm{Connections: []ConnectionSettings{
		{Type: PostgresConnType, CLI: "psql"},
	}}})
	require.NoError(t, err)

	assert.Contains(t, string(raw), "[[conm.connection]]")
	assert.Contains(t, string(raw), `type = 'postgres'`)
}

func writeConm(t *testing.T, content string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "conm.toml")
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}
