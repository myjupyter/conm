package config

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/utils/keys"
)

func TestKeybindingSettings_Validate(t *testing.T) {
	t.Parallel()

	withKeybindings := func(override func(*KeybindingSettings)) KeybindingSettings {
		k := DefaultKeybindings()
		override(&k)
		return k
	}

	tests := []struct {
		name     string
		settings KeybindingSettings
		wantErr  error
		wantPath string
	}{
		{
			name:     "defaults",
			settings: DefaultKeybindings(),
		},
		{
			name: "unknown key names its field",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Navigation.Up = []string{"up", "k", "skdfjs;adjf"}
			}),
			wantErr:  keys.ErrUnknownKey,
			wantPath: "conm.keybinding.navigation.up",
		},
		{
			name: "two actions on one screen share a key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Connection.Ping = []string{"d"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.connection.ping",
		},
		{
			name: "a screen opener and a connection action share a key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Connection.Ping = []string{"t"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.connection.ping",
		},
		{
			name: "an action shares a key with navigation",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Secret.Add = []string{"j"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.secret.add",
		},
		{
			name: "an action shares a key with a general key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Form.Secret = []string{"/"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.form.secret",
		},
		{
			name: "a form action shares a key with the connection ping it uses",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Form.EditField = []string{"p"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.connection.ping",
		},
		{
			name: "a form action may share a key with a connection action the form does not use",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Form.Secret = []string{"e"}
			}),
		},
		{
			name: "connect shares a key with another connection action",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Connection.Connect = []string{"e"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.connection.edit",
		},
		{
			name: "a form action shares a key with confirm",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Form.EditField = []string{"enter"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.general.confirm",
		},
		{
			name: "a secret action shares a key with confirm",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Secret.Add = []string{"enter"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.general.confirm",
		},
		{
			name: "actions on different screens may share a key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Secret.Add = []string{"x"}
				k.Transport.Add = []string{"x"}
				k.Connection.Ping = []string{"x"}
			}),
		},
		{
			name: "an uppercase letter and its shift spelling are one key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Connection.Ping = []string{"shift+r"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.connection.retry",
		},
		{
			name: "modifier order does not make a new key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Connection.Ping = []string{"alt+ctrl+a"}
				k.Connection.Info = []string{"ctrl+alt+a"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.connection.ping",
		},
		{
			name: "a sequence may start with another action's key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Screen.Transport = []string{"s t"}
			}),
		},
		{
			name: "two actions share a sequence",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Screen.Transport = []string{"s s"}
			}),
			wantErr:  errSharedKey,
			wantPath: "conm.keybinding.screen.settings",
		},
		{
			name: "a sequence longer than two keys",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Screen.Transport = []string{"s t u"}
			}),
			wantErr:  errLongSequence,
			wantPath: "conm.keybinding.screen.transport",
		},
		{
			name: "a sequence with an unknown key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Screen.Transport = []string{"s Enter"}
			}),
			wantErr:  keys.ErrUnknownKey,
			wantPath: "conm.keybinding.screen.transport",
		},
		{
			name: "a sequence outside the connections table",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Secret.Add = []string{"g a"}
			}),
			wantErr:  errNoSequence,
			wantPath: "conm.keybinding.secret.add",
		},
		{
			name: "prompt answers may reuse a key used elsewhere",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Prompt.Yes = []string{"e"}
			}),
		},
		{
			name: "a locked binding is changed",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.General.Interrupt = []string{"ctrl+x"}
			}),
			wantErr:  errLocked,
			wantPath: "conm.keybinding.general.interrupt",
		},
		{
			name: "a locked binding gains a key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.General.Backspace = []string{"backspace", "ctrl+h"}
			}),
			wantErr:  errLocked,
			wantPath: "conm.keybinding.general.backspace",
		},
		{
			name: "an append-only binding loses a default key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Navigation.Down = []string{"down", "ctrl+n"}
			}),
			wantErr:  errDefaultRemoved,
			wantPath: "conm.keybinding.navigation.down",
		},
		{
			name: "an append-only binding gains a key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Navigation.Down = []string{"down", "j", "ctrl+n"}
			}),
		},
		{
			name: "a key repeated within one field",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Connection.Ping = []string{"p", "p"}
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.settings.Validate()
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.ErrorContains(t, err, tt.wantPath)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestKeybindingSettings_Set(t *testing.T) {
	withKeybindings := func(override func(*KeybindingSettings)) KeybindingSettings {
		k := DefaultKeybindings()
		override(&k)
		return k
	}

	type args struct {
		path string
		keys []string
	}
	tests := []struct {
		name string
		args args
		want KeybindingSettings
	}{
		{
			name: "a known path takes the keys",
			args: args{path: "navigation.down", keys: []string{"down", "ctrl+n"}},
			want: withKeybindings(func(k *KeybindingSettings) { k.Navigation.Down = []string{"down", "ctrl+n"} }),
		},
		{
			name: "an unknown path changes nothing",
			args: args{path: "navigation.sideways", keys: []string{"x"}},
			want: DefaultKeybindings(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultKeybindings()
			got.Set(tt.args.path, tt.args.keys)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKeybindingSettings_Bindings(t *testing.T) {
	tests := []struct {
		name     string
		settings KeybindingSettings
		wantPath string
		want     Keybinding
	}{
		{
			name:     "an unrestricted binding",
			settings: DefaultKeybindings(),
			wantPath: "connection.ping",
			want:     Keybinding{Path: "connection.ping", Keys: []string{"p"}, Sequence: true},
		},
		{
			name:     "an append-only binding",
			settings: DefaultKeybindings(),
			wantPath: "navigation.down",
			want:     Keybinding{Path: "navigation.down", Keys: []string{"down", "j"}, OnlyAppend: true},
		},
		{
			name:     "a locked binding",
			settings: DefaultKeybindings(),
			wantPath: "general.interrupt",
			want:     Keybinding{Path: "general.interrupt", Keys: []string{"ctrl+c"}, Locked: true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.settings.Bindings()

			require.Len(t, got, len(tt.settings.fields()))
			i := slices.IndexFunc(got, func(b Keybinding) bool { return b.Path == tt.wantPath })
			require.GreaterOrEqual(t, i, 0)
			assert.Equal(t, tt.want, got[i])

			got[i].Keys[0] = "x"
			assert.Equal(t, tt.want.Keys, tt.settings.Bindings()[i].Keys, "the keys handed out are a copy")
		})
	}
}
