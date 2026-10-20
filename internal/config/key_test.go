package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_validateKey(t *testing.T) {
	t.Parallel()

	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{name: "letter", args: args{key: "k"}},
		{name: "uppercase letter", args: args{key: "R"}},
		{name: "punctuation", args: args{key: "?"}},
		{name: "plus alone", args: args{key: "+"}},
		{name: "non-latin letter", args: args{key: "ж"}},
		{name: "named key", args: args{key: "enter"}},
		{name: "function key", args: args{key: "f12"}},
		{name: "ctrl letter", args: args{key: "ctrl+a"}},
		{name: "ctrl plus", args: args{key: "ctrl++"}},
		{name: "alt punctuation", args: args{key: "alt+?"}},
		{name: "shift named key", args: args{key: "shift+tab"}},
		{name: "shift space", args: args{key: "shift+space"}},
		{name: "every modifier", args: args{key: "ctrl+alt+shift+a"}},
		{name: "ctrl shift letter", args: args{key: "ctrl+shift+a"}},
		{name: "shift letter", args: args{key: "shift+r"}},
		{name: "uppercase letter with ctrl", args: args{key: "ctrl+A"}},
		{name: "modifiers in any order", args: args{key: "alt+ctrl+a"}},
		{name: "empty", args: args{key: ""}, wantErr: errUnknownKey},
		{name: "garbage", args: args{key: "skdfjs;adjf"}, wantErr: errUnknownKey},
		{name: "typo in named key", args: args{key: "entr"}, wantErr: errUnknownKey},
		{name: "capitalised named key", args: args{key: "Enter"}, wantErr: errUnknownKey},
		{name: "literal space", args: args{key: " "}, wantErr: errUnknownKey},
		{name: "control character", args: args{key: "\t"}, wantErr: errUnknownKey},
		{name: "dangling modifier", args: args{key: "ctrl+"}, wantErr: errUnknownKey},
		{name: "keypad key", args: args{key: "plus"}, wantErr: errUnknownKey},
		{name: "unsupported function key", args: args{key: "f13"}, wantErr: errUnknownKey},
		{name: "unsupported media key", args: args{key: "mediaplay"}, wantErr: errUnknownKey},
		{name: "typo in modifier", args: args{key: "crtl+c"}, wantErr: errUnknownModifier},
		{name: "capitalised modifier", args: args{key: "Ctrl+c"}, wantErr: errUnknownModifier},
		{name: "unsupported modifier", args: args{key: "meta+a"}, wantErr: errUnknownModifier},
		{name: "empty modifier", args: args{key: "ctrl++a"}, wantErr: errUnknownModifier},
		{name: "repeated modifier", args: args{key: "ctrl+ctrl+a"}, wantErr: errRepeatedModifier},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := validateKey(tt.args.key)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
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
				k.Navigation.Up = []string{"up", "skdfjs;adjf"}
			}),
			wantErr:  errUnknownKey,
			wantPath: keybindingPathPrefix + navigationUpPath,
		},
		{
			name: "two actions share a key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Action.Ping = []string{"d"}
			}),
			wantErr:  errSharedKey,
			wantPath: keybindingPathPrefix + actionDeletePath,
		},
		{
			name: "an action shares a key with navigation",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Action.Ping = []string{"j"}
			}),
			wantErr:  errSharedKey,
			wantPath: keybindingPathPrefix + actionPingPath,
		},
		{
			name: "a pair that may share keys",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Screen.Cancel = []string{"esc", "q"}
			}),
		},
		{
			name: "a key repeated within one field",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Action.Ping = []string{"p", "p"}
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
