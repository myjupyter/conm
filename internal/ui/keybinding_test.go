package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/utils/keys"
)

func Test_newKeymap(t *testing.T) {
	t.Parallel()

	type args struct {
		settings config.KeybindingSettings
	}
	tests := []struct {
		name             string
		args             args
		wantUpHint       string
		wantMoveVertical string
		wantMatchesX     bool
		wantCopyHint     string
	}{
		{
			name:             "defaults",
			args:             args{settings: config.DefaultKeybindings()},
			wantUpHint:       "↑/k",
			wantMoveVertical: "↑↓/k/j",
			wantCopyHint:     "yy",
		},
		{
			name: "an override",
			args: args{settings: func() config.KeybindingSettings {
				k := config.DefaultKeybindings()
				k.Navigation.Up = []string{"up", "k", "x"}
				k.General.Copy = []string{"y", "ctrl+y"}
				return k
			}()},
			wantUpHint:       "↑/k/x",
			wantMoveVertical: "↑↓/k/x/j",
			wantMatchesX:     true,
			wantCopyHint:     "yy/ctrl+y ctrl+y",
		},
		{
			name: "invalid settings fall back to the defaults",
			args: args{settings: func() config.KeybindingSettings {
				k := config.DefaultKeybindings()
				k.Navigation.Up = []string{"up", "x"}
				k.Connection.Ping = []string{"d"}
				return k
			}()},
			wantUpHint:       "↑/k",
			wantMoveVertical: "↑↓/k/j",
			wantCopyHint:     "yy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newKeymap(tt.args.settings)

			is := assert.New(t)
			is.Equal(tt.wantUpHint, got.Up.hint)
			is.Equal(tt.wantMoveVertical, got.MoveVertical.hint)
			is.Equal(tt.wantMatchesX, got.MoveVertical.matches("x"))
			is.Equal(tt.wantCopyHint, got.Copy.hint)
		})
	}
}

func Test_keystroke(t *testing.T) {
	t.Parallel()

	type args struct {
		raw string
	}
	tests := []struct {
		name    string
		args    args
		pressed tea.Key
	}{
		{name: "letter", args: args{raw: "k"}, pressed: tea.Key{Code: 'k', Text: "k"}},
		{name: "symbol", args: args{raw: "?"}, pressed: tea.Key{Code: '?', Text: "?"}},
		{name: "space", args: args{raw: "space"}, pressed: tea.Key{Code: tea.KeySpace, Text: " "}},
		{name: "uppercase letter", args: args{raw: "R"}, pressed: tea.Key{Code: 'r', ShiftedCode: 'R', Mod: tea.ModShift, Text: "R"}},
		{name: "shift letter", args: args{raw: "shift+r"}, pressed: tea.Key{Code: 'r', ShiftedCode: 'R', Mod: tea.ModShift, Text: "R"}},
		{name: "shift non-latin letter", args: args{raw: "shift+ж"}, pressed: tea.Key{Code: 'ж', ShiftedCode: 'Ж', Mod: tea.ModShift, Text: "Ж"}},
		{name: "ctrl letter", args: args{raw: "ctrl+a"}, pressed: tea.Key{Code: 'a', Mod: tea.ModCtrl}},
		{name: "ctrl uppercase letter", args: args{raw: "ctrl+A"}, pressed: tea.Key{Code: 'a', Mod: tea.ModCtrl | tea.ModShift}},
		{name: "modifiers out of order", args: args{raw: "shift+alt+ctrl+a"}, pressed: tea.Key{Code: 'a', Mod: tea.ModCtrl | tea.ModAlt | tea.ModShift}},
		{name: "ctrl plus", args: args{raw: "ctrl++"}, pressed: tea.Key{Code: '+', Mod: tea.ModCtrl}},
		{name: "shift named key", args: args{raw: "shift+tab"}, pressed: tea.Key{Code: tea.KeyTab, Mod: tea.ModShift}},
		{name: "alt named key", args: args{raw: "alt+enter"}, pressed: tea.Key{Code: tea.KeyEnter, Mod: tea.ModAlt}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.pressed.String(), keystroke(tt.args.raw))
		})
	}
}

func TestNamed_spelledByBubbleTea(t *testing.T) {
	t.Parallel()

	codes := []rune{tea.KeyBackspace, tea.KeyTab, tea.KeyEnter, tea.KeyEscape, tea.KeySpace}
	for code := tea.KeyUp; code <= tea.KeyIsoLevel5Shift; code++ {
		codes = append(codes, code)
	}
	spelled := map[string]bool{}
	for _, code := range codes {
		spelled[tea.Key{Code: code}.String()] = true
	}

	for _, name := range keys.Named() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, spelled[name])
		})
	}
}
