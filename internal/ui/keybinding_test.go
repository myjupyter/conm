package ui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
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
		wantMatchesW     bool
	}{
		{
			name:             "defaults keep the written hints",
			args:             args{settings: config.DefaultKeybindings()},
			wantUpHint:       "↑/k",
			wantMoveVertical: "↑↓/jk",
		},
		{
			name: "an override spells its own hint",
			args: args{settings: func() config.KeybindingSettings {
				k := config.DefaultKeybindings()
				k.Navigation.Up = []string{"up", "w"}
				return k
			}()},
			wantUpHint:       "↑/w",
			wantMoveVertical: "↑/w ↓/j",
			wantMatchesW:     true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := newKeymap(tt.args.settings)

			is := assert.New(t)
			is.Equal(tt.wantUpHint, got.Up.hint)
			is.Equal(tt.wantMoveVertical, got.MoveVertical.hint)
			is.Equal(tt.wantMatchesW, got.MoveVertical.matches("w"))
		})
	}
}

func TestNamedKeys_spelledByBubbleTea(t *testing.T) {
	t.Parallel()

	codes := []rune{tea.KeyBackspace, tea.KeyTab, tea.KeyEnter, tea.KeyEscape, tea.KeySpace}
	for code := tea.KeyUp; code <= tea.KeyIsoLevel5Shift; code++ {
		codes = append(codes, code)
	}
	spelled := map[string]bool{}
	for _, code := range codes {
		spelled[tea.Key{Code: code}.String()] = true
	}

	for _, name := range config.NamedKeys() {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			assert.True(t, spelled[name])
		})
	}
}
