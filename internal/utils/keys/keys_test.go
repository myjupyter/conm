package keys

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParse(t *testing.T) {
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
		{name: "empty", args: args{raw: ""}, wantErr: ErrUnknownKey},
		{name: "garbage", args: args{raw: "skdfjs;adjf"}, wantErr: ErrUnknownKey},
		{name: "typo in named key", args: args{raw: "entr"}, wantErr: ErrUnknownKey},
		{name: "capitalised named key", args: args{raw: "Enter"}, wantErr: ErrUnknownKey},
		{name: "literal space", args: args{raw: " "}, wantErr: ErrUnknownKey},
		{name: "control character", args: args{raw: "\t"}, wantErr: ErrUnknownKey},
		{name: "dangling modifier", args: args{raw: "ctrl+"}, wantErr: ErrUnknownKey},
		{name: "keypad key", args: args{raw: "plus"}, wantErr: ErrUnknownKey},
		{name: "unsupported function key", args: args{raw: "f13"}, wantErr: ErrUnknownKey},
		{name: "unsupported media key", args: args{raw: "mediaplay"}, wantErr: ErrUnknownKey},
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

			got, err := Parse(tt.args.raw)
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
