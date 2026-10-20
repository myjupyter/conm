package config

import (
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
		{name: "ctrl letter", args: args{raw: "ctrl+a"}, want: Key{Ctrl: true, Name: "a"}},
		{name: "ctrl plus", args: args{raw: "ctrl++"}, want: Key{Ctrl: true, Name: "+"}},
		{name: "alt punctuation", args: args{raw: "alt+?"}, want: Key{Alt: true, Name: "?"}},
		{name: "ctrl shift symbol", args: args{raw: "ctrl+shift+/"}, want: Key{Ctrl: true, Shift: true, Name: "/"}},
		{name: "shift named key", args: args{raw: "shift+tab"}, want: Key{Shift: true, Name: "tab"}},
		{name: "shift space", args: args{raw: "shift+space"}, want: Key{Shift: true, Name: "space"}},
		{name: "every modifier", args: args{raw: "ctrl+alt+shift+a"}, want: Key{Ctrl: true, Alt: true, Shift: true, Name: "a"}},
		{name: "modifiers in any order", args: args{raw: "shift+alt+ctrl+a"}, want: Key{Ctrl: true, Alt: true, Shift: true, Name: "a"}},
		{name: "uppercase letter is shift", args: args{raw: "R"}, want: Key{Shift: true, Name: "r"}},
		{name: "shift letter", args: args{raw: "shift+r"}, want: Key{Shift: true, Name: "r"}},
		{name: "shift uppercase letter", args: args{raw: "shift+R"}, want: Key{Shift: true, Name: "r"}},
		{name: "uppercase letter with ctrl", args: args{raw: "ctrl+A"}, want: Key{Ctrl: true, Shift: true, Name: "a"}},
		{name: "uppercase non-latin letter", args: args{raw: "Ж"}, want: Key{Shift: true, Name: "ж"}},
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
		{name: "unsupported modifier", args: args{raw: "meta+a"}, wantErr: errUnknownModifier},
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
			name: "an uppercase letter and its shift spelling are one key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Action.Ping = []string{"shift+r"}
			}),
			wantErr:  errSharedKey,
			wantPath: keybindingPathPrefix + actionRetryPath,
		},
		{
			name: "modifier order does not make a new key",
			settings: withKeybindings(func(k *KeybindingSettings) {
				k.Action.Ping = []string{"alt+ctrl+a"}
				k.Action.Add = []string{"ctrl+alt+a"}
			}),
			wantErr:  errSharedKey,
			wantPath: keybindingPathPrefix + actionAddPath,
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
