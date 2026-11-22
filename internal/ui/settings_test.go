package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/cli"
	"github.com/myjupyter/conm/internal/config"
)

func newSettingsFixture(current string) settingsModel {
	m := settingsModel{
		conm: config.Conm{Connections: []config.ConnectionSettings{
			{Type: config.PostgresConnType, CLI: current},
			{Type: config.MySQLConnType},
			{Type: config.MSSQLConnType},
			{Type: config.RedisConnType},
		}},
		clients: map[config.ConnType][]cli.Info{
			config.PostgresConnType: {
				{Name: cli.Psql, Installed: true},
				{Name: cli.Pgcli},
				{Name: cli.USQL, Installed: true},
			},
		},
		versions: map[string]string{},
	}
	m.window.Resize(settingsSectionWindow)

	return m
}

func TestSettingsModel_nextInstalled(t *testing.T) {
	tests := []struct {
		name    string
		current string
		step    int
		want    string
		wantOK  bool
	}{
		{name: "forward skips a missing client", current: cli.Psql, step: 1, want: cli.USQL, wantOK: true},
		{name: "forward wraps around", current: cli.USQL, step: 1, want: cli.Psql, wantOK: true},
		{name: "backward skips a missing client", current: cli.USQL, step: -1, want: cli.Psql, wantOK: true},
		{name: "unset counts from the first installed", current: "", step: 1, want: cli.USQL, wantOK: true},
		{name: "unset backward counts from the first installed", current: "", step: -1, want: cli.USQL, wantOK: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSettingsFixture(tt.current)
			db, ok := m.database()
			require.True(t, ok)

			got, gotOK := m.nextInstalled(db, tt.step)
			assert.Equal(t, tt.wantOK, gotOK)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSettingsModel_nextInstalledWithoutAlternative(t *testing.T) {
	m := newSettingsFixture(cli.Psql)
	m.clients[config.PostgresConnType][2].Installed = false
	db, ok := m.database()
	require.True(t, ok)

	_, gotOK := m.nextInstalled(db, 1)
	assert.False(t, gotOK)
}

func TestSettingsModel_move(t *testing.T) {
	tests := []struct {
		name      string
		steps     []int
		wantAt    int
		wantAbove int
		wantBelow int
	}{
		{name: "opens at the top", wantAt: 0, wantAbove: 0, wantBelow: 1},
		{name: "scrolls past the window", steps: []int{1, 1, 1}, wantAt: 3, wantAbove: 1, wantBelow: 0},
		{name: "wraps to the bottom", steps: []int{-1}, wantAt: 3, wantAbove: 1, wantBelow: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSettingsFixture(cli.Psql)
			for _, step := range tt.steps {
				m = m.move(step)
			}

			n := len(m.conm.Connections)
			assert.Equal(t, tt.wantAt, m.cursor)
			assert.Equal(t, tt.wantAbove, m.window.Above())
			assert.Equal(t, tt.wantBelow, m.window.Below(n))
		})
	}
}

func TestSettingsModel_cycleAppearance(t *testing.T) {
	tests := []struct {
		name    string
		current config.AppearanceSettings
		row     int
		step    int
		want    config.AppearanceSettings
	}{
		{
			name:    "border steps forward",
			current: config.AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode"},
			row:     0,
			step:    1,
			want:    config.AppearanceSettings{Border: "rounded", Accent: "green", Glyphs: "unicode"},
		},
		{
			name:    "accent wraps backward",
			current: config.AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode"},
			row:     1,
			step:    -1,
			want:    config.AppearanceSettings{Border: "single", Accent: "amber", Glyphs: "unicode"},
		},
		{
			name:    "glyphs wrap forward",
			current: config.AppearanceSettings{Border: "heavy", Accent: "blue", Glyphs: "ascii"},
			row:     2,
			step:    1,
			want:    config.AppearanceSettings{Border: "heavy", Accent: "blue", Glyphs: "unicode"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var saved config.Conm
			m := newSettingsFixture(cli.Psql)
			m.conm.Appearance = tt.current
			m.appear.cursor = tt.row
			m.save = func(c config.Conm) error { saved = c; return nil }

			_, cmd := m.cycleAppearance(tt.step)
			require.NotNil(t, cmd)
			msg, ok := cmd().(settingsSavedMsg)
			require.True(t, ok)
			require.NoError(t, msg.err)
			assert.Equal(t, tt.want, msg.conm.Appearance)
			assert.Equal(t, tt.want, saved.Appearance)
		})
	}
}

func TestSettingsModel_cycleAppearanceLockedBorder(t *testing.T) {
	m := newSettingsFixture(cli.Psql)
	m.conm.Appearance = config.AppearanceSettings{Border: "single", Accent: "green", Glyphs: config.ASCIIGlyphs}
	m.save = func(config.Conm) error { return nil }

	model, cmd := m.cycleAppearance(1)
	assert.Nil(t, cmd)
	got, ok := model.(settingsModel)
	require.True(t, ok)
	assert.Equal(t, kindWarn, got.statusKind)
}

func TestSettingsModel_hexAccent(t *testing.T) {
	tests := []struct {
		name      string
		typed     []string
		wantSaved string
		wantEdit  bool
	}{
		{name: "a full colour is saved", typed: []string{"ff", "88", "00"}, wantSaved: "#ff8800"},
		{name: "non-hex characters are dropped", typed: []string{"zz12", "34 5", "6"}, wantSaved: "#123456"},
		{name: "input stops at seven characters", typed: []string{"abcdef99"}, wantSaved: "#abcdef"},
		{name: "a short colour keeps the input open", typed: []string{"abc"}, wantEdit: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSettingsFixture(cli.Psql)
			m.conm.Appearance = config.AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode"}
			m.save = func(config.Conm) error { return nil }
			m.appear.cursor = 1

			model, _, handled := m.handleAppearanceKey(keyMap.Settings.EditValue.keys[0])
			require.True(t, handled)
			m, ok := model.(settingsModel)
			require.True(t, ok)
			require.True(t, m.appear.editing)
			for _, text := range tt.typed {
				m = m.typeHex(text)
			}

			model, cmd := m.saveHexDraft()
			m, ok = model.(settingsModel)
			require.True(t, ok)
			assert.Equal(t, tt.wantEdit, m.appear.editing)
			if tt.wantEdit {
				assert.Nil(t, cmd)
				return
			}
			require.NotNil(t, cmd)
			msg, ok := cmd().(settingsSavedMsg)
			require.True(t, ok)
			require.NoError(t, msg.err)
			assert.Equal(t, tt.wantSaved, msg.conm.Appearance.Accent)
		})
	}
}
