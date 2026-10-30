package config

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseKey(t *testing.T) {
	t.Parallel()

	type args struct {
		raw string
	}
	tests := []struct {
		name    string
		args    args
		want    Key
		wantErr error
	}{
		{name: "letter", args: args{raw: "k"}, want: Key{Name: "k"}},
		{name: "punctuation", args: args{raw: "?"}, want: Key{Name: "?"}},
		{name: "plus alone", args: args{raw: "+"}, want: Key{Name: "+"}},
		{name: "non-latin letter", args: args{raw: "ж"}, want: Key{Name: "ж"}},
		{name: "named key", args: args{raw: "enter"}, want: Key{Name: "enter"}},
		{name: "function key", args: args{raw: "f12"}, want: Key{Name: "f12"}},
		{name: "ctrl letter", args: args{raw: "ctrl+a"}, want: Key{Mod: ModCtrl, Name: "a"}},
		{name: "ctrl plus", args: args{raw: "ctrl++"}, want: Key{Mod: ModCtrl, Name: "+"}},
		{name: "alt punctuation", args: args{raw: "alt+?"}, want: Key{Mod: ModAlt, Name: "?"}},
		{name: "ctrl shift symbol", args: args{raw: "ctrl+shift+/"}, want: Key{Mod: ModCtrl | ModShift, Name: "/"}},
		{name: "shift named key", args: args{raw: "shift+tab"}, want: Key{Mod: ModShift, Name: "tab"}},
		{name: "shift space", args: args{raw: "shift+space"}, want: Key{Mod: ModShift, Name: "space"}},
		{name: "meta letter", args: args{raw: "meta+a"}, want: Key{Mod: ModMeta, Name: "a"}},
		{name: "hyper letter", args: args{raw: "hyper+a"}, want: Key{Mod: ModHyper, Name: "a"}},
		{name: "super letter", args: args{raw: "super+a"}, want: Key{Mod: ModSuper, Name: "a"}},
		{name: "every modifier", args: args{raw: "ctrl+alt+shift+meta+hyper+super+a"}, want: Key{Mod: ModCtrl | ModAlt | ModShift | ModMeta | ModHyper | ModSuper, Name: "a"}},
		{name: "modifiers in any order", args: args{raw: "shift+alt+ctrl+a"}, want: Key{Mod: ModCtrl | ModAlt | ModShift, Name: "a"}},
		{name: "uppercase letter is shift", args: args{raw: "R"}, want: Key{Mod: ModShift, Name: "r"}},
		{name: "shift letter", args: args{raw: "shift+r"}, want: Key{Mod: ModShift, Name: "r"}},
		{name: "shift uppercase letter", args: args{raw: "shift+R"}, want: Key{Mod: ModShift, Name: "r"}},
		{name: "uppercase letter with ctrl", args: args{raw: "ctrl+A"}, want: Key{Mod: ModCtrl | ModShift, Name: "a"}},
		{name: "uppercase non-latin letter", args: args{raw: "Ж"}, want: Key{Mod: ModShift, Name: "ж"}},
		{name: "empty", args: args{raw: ""}, wantErr: errUnknownKey},
		{name: "garbage", args: args{raw: "skdfjs;adjf"}, wantErr: errUnknownKey},
		{name: "typo in named key", args: args{raw: "entr"}, wantErr: errUnknownKey},
		{name: "capitalised named key", args: args{raw: "Enter"}, wantErr: errUnknownKey},
		{name: "literal space", args: args{raw: " "}, wantErr: errUnknownKey},
		{name: "control character", args: args{raw: "\t"}, wantErr: errUnknownKey},
		{name: "dangling modifier", args: args{raw: "ctrl+"}, wantErr: errUnknownKey},
		{name: "keypad key", args: args{raw: "plus"}, wantErr: errUnknownKey},
		{name: "unsupported function key", args: args{raw: "f13"}, wantErr: errUnknownKey},
		{name: "unsupported media key", args: args{raw: "mediaplay"}, wantErr: errUnknownKey},
		{name: "typo in modifier", args: args{raw: "crtl+c"}, wantErr: errUnknownModifier},
		{name: "capitalised modifier", args: args{raw: "Ctrl+c"}, wantErr: errUnknownModifier},
		{name: "unsupported modifier", args: args{raw: "cmd+a"}, wantErr: errUnknownModifier},
		{name: "empty modifier", args: args{raw: "ctrl++a"}, wantErr: errUnknownModifier},
		{name: "repeated modifier", args: args{raw: "ctrl+ctrl+a"}, wantErr: errRepeatedModifier},
		{name: "shift symbol", args: args{raw: "shift+/"}, wantErr: errShiftedSymbol},
		{name: "shift digit", args: args{raw: "shift+1"}, wantErr: errShiftedSymbol},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := ParseKey(tt.args.raw)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMod_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		m    Mod
		want string
	}{
		{name: "no modifier", m: 0, want: ""},
		{name: "one modifier", m: ModAlt, want: "alt"},
		{name: "every modifier in a fixed order", m: ModSuper | ModHyper | ModMeta | ModShift | ModAlt | ModCtrl, want: "ctrl+alt+shift+meta+hyper+super"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.m.String())
		})
	}
}

func TestKeybindingSettings_validate(t *testing.T) {
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
			wantErr:  errUnknownKey,
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
			wantErr:  errUnknownKey,
			wantPath: "conm.keybinding.screen.transport",
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

			err := tt.settings.validate()
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
			want:     Keybinding{Path: "connection.ping", Keys: []string{"p"}},
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
