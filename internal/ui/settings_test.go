package ui

import (
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"

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

func appearanceRowAt(t *testing.T, name string) int {
	t.Helper()
	i := slices.IndexFunc(appearanceRows, func(r appearanceRow) bool { return r.name() == name })
	require.GreaterOrEqual(t, i, 0, name)
	return i
}

func newAppearanceFixture(t *testing.T, a config.AppearanceSettings, row string) settingsModel {
	t.Helper()
	m := newSettingsFixture(cli.Psql)
	m.conm.Appearance = a
	m.save = func(config.Conm) error { return nil }
	m.appear.cursor = appearanceRowAt(t, row)
	return m
}

func savedAppearance(t *testing.T, cmd tea.Cmd) config.AppearanceSettings {
	t.Helper()
	require.NotNil(t, cmd)
	msg, ok := cmd().(settingsSavedMsg)
	require.True(t, ok)
	require.NoError(t, msg.err)
	return msg.conm.Appearance
}

var plainAppearance = config.AppearanceSettings{Border: "single", Accent: "green", Glyphs: "unicode"}

func withAppearance(override func(*config.AppearanceSettings)) config.AppearanceSettings {
	a := plainAppearance.Clone()
	override(&a)
	return a
}

func TestSettingsModel_cycleAppearance(t *testing.T) {
	tests := []struct {
		name    string
		current config.AppearanceSettings
		row     string
		step    int
		want    config.AppearanceSettings
	}{
		{
			name:    "border style steps forward",
			current: plainAppearance,
			row:     "border style",
			step:    1,
			want:    withAppearance(func(a *config.AppearanceSettings) { a.Border = "rounded" }),
		},
		{
			name:    "accent wraps backward",
			current: plainAppearance,
			row:     "accent colour",
			step:    -1,
			want:    withAppearance(func(a *config.AppearanceSettings) { a.Accent = "amber" }),
		},
		{
			name:    "glyph set wraps forward",
			current: withAppearance(func(a *config.AppearanceSettings) { a.Glyphs = "ascii" }),
			row:     "glyphs set",
			step:    1,
			want:    plainAppearance,
		},
		{
			name:    "an element steps to another preset's rune",
			current: plainAppearance,
			row:     "border top left",
			step:    1,
			want:    withAppearance(func(a *config.AppearanceSettings) { a.BorderRunes = map[string]string{"top_left": "╭"} }),
		},
		{
			name:    "an element back on the preset's rune drops the override",
			current: withAppearance(func(a *config.AppearanceSettings) { a.GlyphRunes = map[string]string{"status_ok": "✔"} }),
			row:     "glyphs status ok",
			step:    -1,
			want:    withAppearance(func(a *config.AppearanceSettings) { a.GlyphRunes = map[string]string{} }),
		},
		{
			name:    "a custom element steps onto the first variant",
			current: withAppearance(func(a *config.AppearanceSettings) { a.GlyphRunes = map[string]string{"caret": "»"} }),
			row:     "glyphs caret",
			step:    1,
			want:    withAppearance(func(a *config.AppearanceSettings) { a.GlyphRunes = map[string]string{} }),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAppearanceFixture(t, tt.current, tt.row)
			before := m.conm.Appearance.Clone()

			_, cmd := m.cycleAppearance(tt.step)

			assert.Equal(t, tt.want, savedAppearance(t, cmd))
			assert.Equal(t, before, m.conm.Appearance)
		})
	}
}

func TestSettingsModel_lockedBorder(t *testing.T) {
	ascii := withAppearance(func(a *config.AppearanceSettings) { a.Glyphs = config.ASCIIGlyphs })
	tests := []struct {
		name string
		row  string
	}{
		{name: "style", row: "border style"},
		{name: "element", row: "border vertical"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAppearanceFixture(t, ascii, tt.row)

			model, cmd := m.cycleAppearance(1)
			assert.Nil(t, cmd)
			got, ok := model.(settingsModel)
			require.True(t, ok)
			assert.Equal(t, kindWarn, got.statusKind)

			got = m.startAppearanceInput()
			assert.False(t, got.appear.editing)
		})
	}
}

func TestSettingsModel_appearanceInput(t *testing.T) {
	tests := []struct {
		name     string
		current  config.AppearanceSettings
		row      string
		typed    []string
		want     config.AppearanceSettings
		wantEdit bool
	}{
		{
			name:  "a full hex colour is saved",
			row:   "accent colour",
			typed: []string{"ff", "88", "00"},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.Accent = "#ff8800" }),
		},
		{
			name:  "non-hex characters are dropped",
			row:   "accent colour",
			typed: []string{"zz12", "34 5", "6"},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.Accent = "#123456" }),
		},
		{
			name:  "hex input stops at seven characters",
			row:   "accent colour",
			typed: []string{"abcdef99"},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.Accent = "#abcdef" }),
		},
		{
			name:     "a short hex colour keeps the input open",
			row:      "accent colour",
			typed:    []string{"abc"},
			wantEdit: true,
		},
		{
			name:  "an element keeps the last character typed",
			row:   "glyphs caret",
			typed: []string{"x", "»"},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.GlyphRunes = map[string]string{"caret": "»"} }),
		},
		{
			name:  "a border element is overridden",
			row:   "border horizontal",
			typed: []string{"="},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.BorderRunes = map[string]string{"horizontal": "="} }),
		},
		{
			name:     "a wide character keeps the input open",
			row:      "glyphs caret",
			typed:    []string{"好"},
			wantEdit: true,
		},
		{
			name:     "nothing typed keeps the input open",
			row:      "border vertical",
			wantEdit: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := tt.current
			if current.Border == "" {
				current = plainAppearance
			}
			m := newAppearanceFixture(t, current, tt.row)

			model, _, handled := m.handleAppearanceKey(keyMap.Settings.EditValue.keys[0])
			require.True(t, handled)
			m, ok := model.(settingsModel)
			require.True(t, ok)
			require.True(t, m.appear.editing)
			for _, text := range tt.typed {
				m = m.typeAppearance(text)
			}

			model, cmd := m.saveAppearanceDraft()
			m, ok = model.(settingsModel)
			require.True(t, ok)
			assert.Equal(t, tt.wantEdit, m.appear.editing)
			if tt.wantEdit {
				assert.Nil(t, cmd)
				return
			}
			assert.Equal(t, tt.want, savedAppearance(t, cmd))
		})
	}
}

func TestSettingsModel_resetElement(t *testing.T) {
	tests := []struct {
		name    string
		current config.AppearanceSettings
		row     string
		want    config.AppearanceSettings
		wantCmd bool
	}{
		{
			name: "an override is dropped",
			current: withAppearance(func(a *config.AppearanceSettings) {
				a.BorderRunes = map[string]string{"top_left": "╭", "vertical": "┃"}
			}),
			row:     "border top left",
			want:    withAppearance(func(a *config.AppearanceSettings) { a.BorderRunes = map[string]string{"vertical": "┃"} }),
			wantCmd: true,
		},
		{
			name:    "an element on the preset stays",
			current: plainAppearance,
			row:     "glyphs caret",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAppearanceFixture(t, tt.current, tt.row)

			_, cmd := m.resetElement()
			if !tt.wantCmd {
				assert.Nil(t, cmd)
				return
			}
			assert.Equal(t, tt.want, savedAppearance(t, cmd))
		})
	}
}

func TestSettingsModel_sectionNote(t *testing.T) {
	tests := []struct {
		name    string
		current config.AppearanceSettings
		row     string
		want    string
	}{
		{name: "accent has none", current: plainAppearance, row: "accent colour"},
		{name: "untouched border", current: plainAppearance, row: "border style"},
		{
			name:    "one glyph customised",
			current: withAppearance(func(a *config.AppearanceSettings) { a.GlyphRunes = map[string]string{"caret": "»"} }),
			row:     "glyphs set",
			want:    "1 element customised",
		},
		{
			name: "two border elements customised",
			current: withAppearance(func(a *config.AppearanceSettings) {
				a.BorderRunes = map[string]string{"top_left": "╭", "vertical": "┃"}
			}),
			row:  "border style",
			want: "2 elements customised",
		},
		{
			name:    "border locked by ascii glyphs",
			current: withAppearance(func(a *config.AppearanceSettings) { a.Glyphs = config.ASCIIGlyphs }),
			row:     "border style",
			want:    "drawn by the ascii glyphs",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAppearanceFixture(t, tt.current, tt.row)
			assert.Equal(t, tt.want, m.sectionNote(m.appearanceRow()))
		})
	}
}
