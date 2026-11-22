package ui

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

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
	section := appearanceRows[appearanceRowAt(t, row)].section
	for page, name := range appearancePages {
		if name == section {
			m.tab = page
		}
	}
	m.appear.cursor = slices.IndexFunc(m.pageRows(), func(r appearanceRow) bool { return r.name() == row })
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

var plainAppearance = config.AppearanceSettings{
	Border: "single", Accent: "green", Glyphs: "unicode", Text: "default",
	BorderColor: "default", GlyphColor: "default", TextColor: "default",
}

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
			row:     "colour accent",
			step:    -1,
			want:    withAppearance(func(a *config.AppearanceSettings) { a.Accent = "orange>red" }),
		},
		{
			name:    "glyph set wraps forward",
			current: withAppearance(func(a *config.AppearanceSettings) { a.Glyphs = "ascii" }),
			row:     "glyphs set",
			step:    1,
			want:    plainAppearance,
		},
		{
			name:    "text style steps forward",
			current: plainAppearance,
			row:     "text style",
			step:    1,
			want:    withAppearance(func(a *config.AppearanceSettings) { a.Text = "plain" }),
		},
		{
			name:    "a text element steps through the style choices",
			current: plainAppearance,
			row:     "text note",
			step:    1,
			want:    withAppearance(func(a *config.AppearanceSettings) { a.TextStyle = map[string]string{"note": "bold"} }),
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
		{name: "element", row: "border left"},
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
			name:  "a hex accent is saved",
			row:   "colour accent",
			typed: []string{"#FF", "88", "00"},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.Accent = "#ff8800" }),
		},
		{
			name:  "an accent gradient is saved",
			row:   "colour accent",
			typed: []string{"Teal > ", "#FF00AA>pink"},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.Accent = "teal>#ff00aa>pink" }),
		},
		{
			name:     "a short hex accent keeps the input open",
			row:      "colour accent",
			typed:    []string{"#abc"},
			wantEdit: true,
		},
		{
			name:     "an accent through a colour name keeps the input open",
			row:      "colour accent",
			typed:    []string{"fg>pink"},
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
			row:   "border top",
			typed: []string{"="},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.BorderRunes = map[string]string{"top": "="} }),
		},
		{
			name:     "a wide border character keeps the input open",
			row:      "border right",
			typed:    []string{"好"},
			wantEdit: true,
		},
		{
			name:  "an emoji glyph is accepted",
			row:   "glyphs status ok",
			typed: []string{"👍🏽"},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.GlyphRunes = map[string]string{"status_ok": "👍🏽"} }),
		},
		{
			name:  "a text style is typed",
			row:   "text note",
			typed: []string{"Italic", "+", "under line"},
			want:  withAppearance(func(a *config.AppearanceSettings) { a.TextStyle = map[string]string{"note": "italic+underline"} }),
		},
		{
			name:     "an unknown text attribute keeps the input open",
			row:      "text title",
			typed:    []string{"blink"},
			wantEdit: true,
		},
		{
			name:     "nothing typed keeps the input open",
			row:      "border left",
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
				a.BorderRunes = map[string]string{"top_left": "╭", "left": "┃"}
			}),
			row:     "border top left",
			want:    withAppearance(func(a *config.AppearanceSettings) { a.BorderRunes = map[string]string{"left": "┃"} }),
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
		{name: "accent has none", current: plainAppearance, row: "colour accent"},
		{name: "untouched border", current: plainAppearance, row: "border style"},
		{
			name:    "one glyph customised",
			current: withAppearance(func(a *config.AppearanceSettings) { a.GlyphRunes = map[string]string{"caret": "»"} }),
			row:     "glyphs set",
			want:    "1 change to the preset",
		},
		{
			name: "two border elements customised",
			current: withAppearance(func(a *config.AppearanceSettings) {
				a.BorderRunes = map[string]string{"top_left": "╭", "left": "┃"}
			}),
			row:  "border style",
			want: "2 changes to the preset",
		},
		{
			name: "rune and colour changes add up",
			current: withAppearance(func(a *config.AppearanceSettings) {
				a.GlyphRunes = map[string]string{"caret": "»"}
				a.GlyphColors = map[string]string{"caret": "pink", "dot_on": "#ff0000"}
			}),
			row:  "glyphs set",
			want: "3 changes to the preset",
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

func TestSettingsRenderStaysAligned(t *testing.T) {
	looks := []config.AppearanceSettings{
		withAppearance(func(a *config.AppearanceSettings) { a.Glyphs = "emoji" }),
		withAppearance(func(a *config.AppearanceSettings) { a.Glyphs, a.Border = "emoji", "block" }),
		withAppearance(func(a *config.AppearanceSettings) { a.Border, a.Text = "hidden", "emphasis" }),
		withAppearance(func(a *config.AppearanceSettings) { a.Glyphs = config.ASCIIGlyphs }),
		withAppearance(func(a *config.AppearanceSettings) {
			a.Glyphs, a.BorderColor, a.GlyphColor = "emoji", "duotone", "bright"
			a.BorderColors = map[string]string{"left": "#ff0000"}
		}),
	}
	t.Cleanup(func() { applyAppearance(plainAppearance) })

	for _, look := range looks {
		for tab := range settingsPages {
			t.Run(look.Glyphs+"/"+look.Border+"/"+strconv.Itoa(tab), func(t *testing.T) {
				applyAppearance(look)
				m := newSettingsFixture(cli.Psql)
				m.conm.Appearance = look
				m.tab = tab
				m.appear.window.Resize(appearanceWindow)
				model, _ := m.Update(tea.WindowSizeMsg{Width: 110, Height: 40})
				rendered, ok := model.(settingsModel)
				require.True(t, ok)

				lines := strings.Split(ansi.Strip(rendered.render()), "\n")
				require.NotEmpty(t, lines)
				selector := -1
				for _, line := range lines {
					assert.Equal(t, ansi.StringWidth(lines[0]), ansi.StringWidth(line), line)
					before, _, found := strings.Cut(line, gSelectPrev+" ")
					if !found || !strings.Contains(line, gSelectNext) {
						continue
					}
					if selector < 0 {
						selector = ansi.StringWidth(before)
					}
					assert.Equal(t, selector, ansi.StringWidth(before), line)
				}
			})
		}
	}
}

func TestSettingsModel_colorRows(t *testing.T) {
	tests := []struct {
		name    string
		current config.AppearanceSettings
		row     string
		act     func(settingsModel) (tea.Model, tea.Cmd)
		want    config.AppearanceSettings
	}{
		{
			name:    "border palette steps forward",
			current: plainAppearance,
			row:     "border palette",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.cycleAppearance(1) },
			want:    withAppearance(func(a *config.AppearanceSettings) { a.BorderColor = "accent" }),
		},
		{
			name:    "glyph palette wraps backward",
			current: plainAppearance,
			row:     "glyphs palette",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.cycleAppearance(-1) },
			want:    withAppearance(func(a *config.AppearanceSettings) { a.GlyphColor = "muted" }),
		},
		{
			name:    "text palette steps forward",
			current: plainAppearance,
			row:     "text palette",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.cycleAppearance(1) },
			want:    withAppearance(func(a *config.AppearanceSettings) { a.TextColor = "accent" }),
		},
		{
			name:    "a border element colour steps off auto",
			current: plainAppearance,
			row:     "border top colour",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.cycleAppearance(1) },
			want:    withAppearance(func(a *config.AppearanceSettings) { a.BorderColors = map[string]string{"top": "accent"} }),
		},
		{
			name:    "a text role colour steps off auto",
			current: plainAppearance,
			row:     "text note colour",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.cycleAppearance(1) },
			want:    withAppearance(func(a *config.AppearanceSettings) { a.TextColors = map[string]string{"note": "accent"} }),
		},
		{
			name:    "a glyph colour is typed as hex",
			current: plainAppearance,
			row:     "glyphs caret colour",
			act: func(m settingsModel) (tea.Model, tea.Cmd) {
				m = m.startAppearanceInput()
				m = m.typeAppearance("#FF0088")
				return m.saveAppearanceDraft()
			},
			want: withAppearance(func(a *config.AppearanceSettings) { a.GlyphColors = map[string]string{"caret": "#ff0088"} }),
		},
		{
			name:    "the accent cycles into the gradient choices",
			current: withAppearance(func(a *config.AppearanceSettings) { a.Accent = "amber" }),
			row:     "colour accent",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.cycleAppearance(1) },
			want:    withAppearance(func(a *config.AppearanceSettings) { a.Accent = "teal>pink" }),
		},
		{
			name:    "a border gradient is typed",
			current: plainAppearance,
			row:     "border top colour",
			act: func(m settingsModel) (tea.Model, tea.Cmd) {
				m = m.startAppearanceInput()
				m = m.typeAppearance("Teal > Pink")
				return m.saveAppearanceDraft()
			},
			want: withAppearance(func(a *config.AppearanceSettings) { a.BorderColors = map[string]string{"top": "teal>pink"} }),
		},
		{
			name:    "a text gradient is typed",
			current: plainAppearance,
			row:     "text title colour",
			act: func(m settingsModel) (tea.Model, tea.Cmd) {
				m = m.startAppearanceInput()
				m = m.typeAppearance("#ff8800>accent")
				return m.saveAppearanceDraft()
			},
			want: withAppearance(func(a *config.AppearanceSettings) { a.TextColors = map[string]string{"title": "#ff8800>accent"} }),
		},
		{
			name:    "the tab background steps off auto",
			current: plainAppearance,
			row:     "text tab background",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.cycleAppearance(1) },
			want:    withAppearance(func(a *config.AppearanceSettings) { a.TextBackgrounds = map[string]string{"tab": "accent"} }),
		},
		{
			name:    "a heading background gradient is typed",
			current: plainAppearance,
			row:     "text heading background",
			act: func(m settingsModel) (tea.Model, tea.Cmd) {
				m = m.startAppearanceInput()
				m = m.typeAppearance("#202020>#3a2a40")
				return m.saveAppearanceDraft()
			},
			want: withAppearance(func(a *config.AppearanceSettings) {
				a.TextBackgrounds = map[string]string{"heading": "#202020>#3a2a40"}
			}),
		},
		{
			name:    "the input colour steps off auto",
			current: plainAppearance,
			row:     "text input colour",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.cycleAppearance(1) },
			want:    withAppearance(func(a *config.AppearanceSettings) { a.TextColors = map[string]string{"input": "accent"} }),
		},
		{
			name: "a colour override is reset",
			current: withAppearance(func(a *config.AppearanceSettings) {
				a.GlyphColors = map[string]string{"dot_on": "pink", "caret": "teal"}
			}),
			row:  "glyphs dot on colour",
			act:  func(m settingsModel) (tea.Model, tea.Cmd) { return m.resetElement() },
			want: withAppearance(func(a *config.AppearanceSettings) { a.GlyphColors = map[string]string{"caret": "teal"} }),
		},
		{
			name:    "the ascii lock leaves border colours free",
			current: withAppearance(func(a *config.AppearanceSettings) { a.Glyphs = config.ASCIIGlyphs }),
			row:     "border left colour",
			act:     func(m settingsModel) (tea.Model, tea.Cmd) { return m.cycleAppearance(1) },
			want: withAppearance(func(a *config.AppearanceSettings) {
				a.Glyphs = config.ASCIIGlyphs
				a.BorderColors = map[string]string{"left": "accent"}
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newAppearanceFixture(t, tt.current, tt.row)

			_, cmd := tt.act(m)

			assert.Equal(t, tt.want, savedAppearance(t, cmd))
		})
	}
}

func TestAppearanceRowsHaveColourRows(t *testing.T) {
	tests := []struct {
		row       string
		wantColor bool
	}{
		{row: "text title", wantColor: true},
		{row: "border bottom right", wantColor: true},
		{row: "glyphs status ok", wantColor: true},
		{row: "glyphs ellipsis"},
		{row: "glyphs input mask"},
	}
	for _, tt := range tests {
		t.Run(tt.row, func(t *testing.T) {
			i := appearanceRowAt(t, tt.row)
			hasColor := i+1 < len(appearanceRows) && appearanceRows[i+1].name() == tt.row+" colour"
			assert.Equal(t, tt.wantColor, hasColor)
		})
	}
}

func TestSettingsModel_appearanceSubheads(t *testing.T) {
	tests := []struct {
		page    string
		subhead string
		rows    []string
	}{
		{page: "text style", subhead: "TITLE", rows: []string{"style", "colour"}},
		{page: "text style", subhead: "TAB", rows: []string{"style", "colour", "background"}},
		{page: "text style", subhead: "HEADING", rows: []string{"style", "colour", "background"}},
		{page: "text style", subhead: "SUBHEADING", rows: []string{"style", "colour", "background"}},
		{page: "border style", subhead: "TOP LEFT", rows: []string{"character", "colour"}},
		{page: "glyphs set", subhead: "STATUS OK", rows: []string{"glyph", "colour"}},
		{page: "glyphs set", subhead: "ELLIPSIS", rows: []string{"glyph"}},
		{page: "text style", subhead: "BODY", rows: []string{"colour"}},
		{page: "text style", subhead: "INPUT", rows: []string{"colour"}},
	}
	for _, tt := range tests {
		t.Run(tt.subhead, func(t *testing.T) {
			m := newAppearanceFixture(t, plainAppearance, tt.page)
			m.size.observe(tea.WindowSizeMsg{Width: 110, Height: 40})
			lines := m.appearanceLines()
			plain := make([]string, len(lines))
			for i, line := range lines {
				plain[i] = strings.TrimSpace(strings.Trim(ansi.Strip(line), gLineLeft+gLineRight))
			}

			at := slices.Index(plain, tt.subhead)
			require.GreaterOrEqual(t, at, 0)
			assert.Equal(t, 1, strings.Count(strings.Join(plain, "\n"), "\n"+tt.subhead+"\n"))
			for i, row := range tt.rows {
				require.Less(t, at+1+i, len(plain))
				assert.True(t, strings.HasPrefix(plain[at+1+i], row), plain[at+1+i])
			}
			if next := at + 1 + len(tt.rows); next < len(plain) {
				assert.Empty(t, plain[next], "a blank line closes the element")
			}
		})
	}
}

func TestSettingsModel_glyphColourRefusesGradient(t *testing.T) {
	m := newAppearanceFixture(t, plainAppearance, "glyphs caret colour")
	m = m.startAppearanceInput()
	m = m.typeAppearance("teal>pink")

	model, cmd := m.saveAppearanceDraft()
	assert.Nil(t, cmd)
	got, ok := model.(settingsModel)
	require.True(t, ok)
	assert.True(t, got.appear.editing)
	assert.Equal(t, kindWarn, got.statusKind)
}

func TestSettingsModel_flatGradientHint(t *testing.T) {
	tests := []struct {
		name      string
		trueColor bool
		wantHint  bool
	}{
		{name: "no 24-bit colour", wantHint: true},
		{name: "24-bit colour", trueColor: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			current := withAppearance(func(a *config.AppearanceSettings) { a.Accent = "amber" })
			withTrueColor(t, tt.trueColor, current)
			m := newAppearanceFixture(t, current, "colour accent")

			_, cmd := m.cycleAppearance(1)
			require.NotNil(t, cmd)
			msg, ok := cmd().(settingsSavedMsg)
			require.True(t, ok)
			assert.Equal(t, tt.wantHint, strings.Contains(msg.status, flatNote))
		})
	}
}

func TestSettingsModel_nestedTabs(t *testing.T) {
	tests := []struct {
		name       string
		keys       []string
		wantTab    int
		wantNested bool
	}{
		{name: "tab walks the top tabs", keys: []string{"tab", "tab"}, wantTab: settingsAccentTab},
		{name: "tab past appearance wraps to client", keys: []string{"tab", "tab", "tab"}, wantTab: settingsClientTab},
		{name: "enter opens the appearance tabs", keys: []string{"tab", "tab", "enter"}, wantTab: settingsAccentTab, wantNested: true},
		{name: "tab inside walks the appearance tabs", keys: []string{"tab", "tab", "enter", "tab", "tab"}, wantTab: settingsGlyphsTab, wantNested: true},
		{name: "shift+tab inside wraps to table", keys: []string{"tab", "tab", "enter", "shift+tab"}, wantTab: settingsTableTab, wantNested: true},
		{name: "esc goes back to the top tabs", keys: []string{"tab", "tab", "enter", "tab", "esc"}, wantTab: settingsTextTab},
		{name: "appearance remembers its tab", keys: []string{"tab", "tab", "enter", "tab", "esc", "tab", "shift+tab"}, wantTab: settingsTextTab},
		{name: "enter elsewhere does not nest", keys: []string{"enter"}, wantTab: settingsClientTab},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newSettingsFixture(cli.Psql)
			for _, key := range tt.keys {
				model, _ := m.handleSettingsKey(key)
				next, ok := model.(settingsModel)
				require.True(t, ok)
				m = next
			}

			is := assert.New(t)
			is.Equal(tt.wantTab, m.tab)
			is.Equal(tt.wantNested, m.nested)
		})
	}
}
