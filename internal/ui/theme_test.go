package ui

import (
	"image/color"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/myjupyter/conm/internal/config"
	"github.com/myjupyter/conm/internal/ui/view"
)

func TestThemePresetsAreComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		names    []string
		presets  map[string]map[string]string
		elements []string
		vars     map[string]*string
		valid    func(string) bool
	}{
		{name: "border", names: config.BorderStyles, presets: borderPresets, elements: config.BorderElements, vars: borderVars, valid: config.IsBorderRune},
		{name: "glyphs", names: config.GlyphSets, presets: glyphPresets, elements: config.GlyphElements, vars: glyphVars, valid: config.IsGlyph},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			is.Len(tt.presets, len(tt.names))
			is.Len(tt.vars, len(tt.elements))
			for _, element := range tt.elements {
				is.Contains(tt.vars, element)
			}
			for _, name := range tt.names {
				is.Len(tt.presets[name], len(tt.elements), name)
				for _, element := range tt.elements {
					is.True(tt.valid(tt.presets[name][element]), name+" "+element)
				}
			}
		})
	}
}

func TestThemeAccentsAreComplete(t *testing.T) {
	t.Parallel()

	assert.Len(t, accentColors, len(config.AccentColors))
	for _, name := range config.AccentColors {
		assert.Contains(t, accentColors, name)
	}
}

func TestBorderRunes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    config.AppearanceSettings
		want map[string]string
	}{
		{
			name: "preset with an override",
			a:    config.AppearanceSettings{Border: "single", Glyphs: "unicode", BorderRunes: map[string]string{"top_left": "╭"}},
			want: map[string]string{"top_left": "╭", "top_right": "┐", "top": "─", "left": "│"},
		},
		{
			name: "ascii glyphs replace preset and overrides",
			a:    config.AppearanceSettings{Border: "double", Glyphs: config.ASCIIGlyphs, BorderRunes: map[string]string{"top_left": "╭"}},
			want: map[string]string{"top_left": "+", "top_right": "+", "top": "-", "right": "|"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := borderRunes(tt.a)
			for element, want := range tt.want {
				assert.Equal(t, want, got[element], element)
			}
		})
	}
}

func TestTextPresetsAreComplete(t *testing.T) {
	t.Parallel()

	is := assert.New(t)
	is.Len(textPresets, len(config.TextStyles))
	is.Len(textVars, len(config.TextElements))
	for _, name := range config.TextStyles {
		is.Len(textPresets[name], len(config.TextElements), name)
		for _, element := range config.TextElements {
			is.True(config.IsTextStyle(textPresets[name][element]), name+" "+element)
		}
	}
}

func TestParseTextStyle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value string
		want  textStyle
	}{
		{value: "plain", want: textStyle{}},
		{value: "bold", want: textStyle{bold: true}},
		{value: "italic+faint", want: textStyle{italic: true, faint: true}},
		{value: "bold+italic+underline+faint", want: textStyle{bold: true, italic: true, underline: true, faint: true}},
	}
	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, parseTextStyle(tt.value))
		})
	}
}

func TestBlockBorderSplitsTheLines(t *testing.T) {
	t.Parallel()

	got := borderRunes(config.AppearanceSettings{Border: "block", Glyphs: "unicode"})
	assert.Equal(t, map[string]string{
		"top_left": "▛", "top_right": "▜", "bottom_left": "▙", "bottom_right": "▟",
		"tee_left": "▌", "tee_right": "▐", "top": "▀", "bottom": "▄", "rule": "▀", "left": "▌", "right": "▐",
	}, got)
}

func TestFrameMeasuresWideGlyphs(t *testing.T) {
	t.Parallel()

	const w = 60
	tests := []struct {
		name string
		line string
		want string
	}{
		{name: "tab note", line: frameTabs(w, []string{"one", "two"}, 0, "🔽 24 below"), want: "🔽 24 below"},
		{name: "select box", line: frameLine(w, frameSelectBox(30, selectBox{value: "✅", at: 1, n: 2}, nil), nil), want: "✅"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plain := ansi.Strip(tt.line)
			assert.Contains(t, plain, tt.want)
			assert.Equal(t, w+2, ansi.StringWidth(plain))
		})
	}
}

func TestColorPalettesAreComplete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		names    []string
		palettes map[string]map[string]string
		elements []string
		valid    func(string) bool
	}{
		{name: "border", names: config.BorderPalettes, palettes: borderPalettes, elements: config.BorderElements, valid: config.IsColorOrGradient},
		{name: "glyphs", names: config.GlyphPalettes, palettes: glyphPalettes, elements: config.GlyphColorElements, valid: config.IsColor},
		{name: "text", names: config.TextPalettes, palettes: textPalettes, elements: config.TextColorElements, valid: config.IsColorOrGradient},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			is := assert.New(t)
			is.Len(tt.palettes, len(tt.names))
			for _, name := range tt.names {
				is.Len(tt.palettes[name], len(tt.elements), name)
				for _, element := range tt.elements {
					is.True(tt.valid(tt.palettes[name][element]), name+" "+element)
				}
			}
		})
	}
}

func TestNamedColorResolvesEveryName(t *testing.T) {
	t.Parallel()

	for _, name := range config.ColorNames {
		if name == config.AutoColor {
			assert.Nil(t, namedColor(name))
			continue
		}
		assert.NotNil(t, namedColor(name), name)
	}
	assert.NotNil(t, namedColor("#123456"))
}

func TestGlyphColorsKeepStatusColours(t *testing.T) {
	t.Parallel()

	for _, palette := range config.GlyphPalettes {
		for _, element := range []string{"status_ok", "status_warn", "status_error", "status_pending", "dot_fail", "field_bad"} {
			assert.Equal(t, config.AutoColor, glyphPalettes[palette][element], palette+" "+element)
		}
	}
}

func TestApplyAppearanceColoursTextRoles(t *testing.T) {
	t.Cleanup(func() {
		applyAppearance(config.AppearanceSettings{
			Border: "single", Accent: "green", Glyphs: "unicode", Text: "default",
			BorderColor: "default", GlyphColor: "default", TextColor: "default",
		})
	})

	applyAppearance(config.AppearanceSettings{
		Border: "single", Accent: "green", Glyphs: "unicode", Text: "default",
		BorderColor: "default", GlyphColor: "default", TextColor: "accent",
		TextColors: map[string]string{"note": "#123456"},
	})

	is := assert.New(t)
	is.Equal(cAccent, tTitle.fg)
	is.Equal(lipgloss.Color("#123456"), tNote.fg)
	is.Nil(tValue.fg)
	is.True(tTitle.bold)
}

func withTrueColor(t *testing.T, on bool, a config.AppearanceSettings) {
	t.Helper()
	detected := supportsTrueColor
	t.Cleanup(func() {
		supportsTrueColor = detected
		applyAppearance(config.AppearanceSettings{
			Border: "single", Accent: "green", Glyphs: "unicode", Text: "default",
			BorderColor: "default", GlyphColor: "default", TextColor: "default",
		})
	})
	supportsTrueColor = func() bool { return on }
	applyAppearance(a)
}

func gradientLook() config.AppearanceSettings {
	return config.AppearanceSettings{
		Border: "single", Accent: "green", Glyphs: "unicode", Text: "default",
		BorderColor: "default", GlyphColor: "default", TextColor: "default",
		BorderColors: map[string]string{"top": "#ff0000>#0000ff", "left": "#ff0000>#0000ff"},
		TextColors:   map[string]string{"title": "#ff0000>#0000ff"},
	}
}

func distinctSGR(rendered string) int {
	seen := map[string]bool{}
	for part := range strings.SplitSeq(rendered, "\x1b[") {
		if code, _, ok := strings.Cut(part, "m"); ok && strings.Contains(code, "38;2;") {
			seen[code] = true
		}
	}
	return len(seen)
}

func TestGradientsFollowTheTerminal(t *testing.T) {
	tests := []struct {
		name     string
		on       bool
		wantMany bool
	}{
		{name: "no 24-bit colour draws the first stop", on: false},
		{name: "24-bit colour spreads the stops", on: true, wantMany: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withTrueColor(t, tt.on, gradientLook())

			top := edge("top", strings.Repeat(gLineTop, 10))
			title := span{text: "connections", style: tTitle}.render()

			if tt.wantMany {
				assert.Greater(t, distinctSGR(top), 2)
				assert.Greater(t, distinctSGR(title), 2)
				return
			}
			assert.Equal(t, 1, distinctSGR(top))
			assert.Equal(t, 1, distinctSGR(title))
		})
	}
}

func TestSideGradientsRepaintOnlyTheSides(t *testing.T) {
	withTrueColor(t, true, gradientLook())

	lines := []string{frameTop(10, "x", ""), frameLine(10, nil, nil), frameLine(10, nil, nil), frameBottom(10)}
	got := sideGradients(lines)

	is := assert.New(t)
	is.Equal(lines[0], got[0])
	is.Equal(lines[3], got[3])
	is.NotEqual(lines[1], got[1])
	is.NotEqual(got[1], got[2])
	for i := range lines {
		is.Equal(ansi.Strip(lines[i]), ansi.Strip(got[i]))
	}
}

func TestBodyAndInputColours(t *testing.T) {
	tests := []struct {
		name       string
		colors     map[string]string
		wantBody   color.Color
		wantInput  color.Color
		wantColumn color.Color
	}{
		{name: "auto keeps the base colours", wantBody: baseFg, wantInput: baseInput, wantColumn: cDim},
		{
			name:       "overrides replace them",
			colors:     map[string]string{"body": "#112233", "input": "teal", "column": "pink"},
			wantBody:   lipgloss.Color("#112233"),
			wantInput:  accentColors["teal"],
			wantColumn: accentColors["pink"],
		},
		{
			name:       "a gradient draws its first stop",
			colors:     map[string]string{"body": "#ff0000>#0000ff"},
			wantBody:   lipgloss.Color("#ff0000"),
			wantInput:  baseInput,
			wantColumn: cDim,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			look := gradientLook()
			look.BorderColors, look.TextColors = nil, tt.colors
			withTrueColor(t, true, look)

			assert.Equal(t, tt.wantBody, cFg)
			assert.Equal(t, tt.wantInput, cInput)
			assert.Equal(t, tt.wantColumn, cColumn)
		})
	}
}

func distinctBackgrounds(rendered string) int {
	seen := map[string]bool{}
	for part := range strings.SplitSeq(rendered, "\x1b[") {
		if code, _, ok := strings.Cut(part, "m"); ok {
			if _, bg, found := strings.Cut(code, "48;2;"); found {
				seen[bg] = true
			}
		}
	}
	return len(seen)
}

func TestAccentGradient(t *testing.T) {
	tests := []struct {
		name     string
		on       bool
		wantMany bool
	}{
		{name: "no 24-bit colour keeps one background", on: false},
		{name: "24-bit colour spreads the background", on: true, wantMany: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			look := gradientLook()
			look.BorderColors, look.TextColors, look.Accent = nil, nil, "teal>pink"
			withTrueColor(t, tt.on, look)

			row := frameLine(30, []span{{text: " selected row", fg: cInvFg, bg: cAccent}}, cAccent)
			tabs := frameTabs(30, []string{"one", "two"}, 0, "")

			is := assert.New(t)
			is.True(sameColor(accentColors["teal"], cAccent))
			if tt.wantMany {
				is.Greater(distinctBackgrounds(row), 10)
				is.Greater(distinctBackgrounds(tabs), 2)
			} else {
				is.Equal(1, distinctBackgrounds(row))
				is.Equal(1, distinctBackgrounds(tabs))
			}
			is.Equal(32, ansi.StringWidth(ansi.Strip(row)))
		})
	}
}

func TestGradientsStayOnTheirElement(t *testing.T) {
	tabs := Model.tabsLine
	row := func(m Model) string { return m.rowLine(0) }
	tests := []struct {
		name          string
		accent        string
		tabBackground string
		tabColor      string
		line          func(Model) string
		wantMany      bool
	}{
		{name: "a flat accent row ignores a tab gradient from the accent", accent: "green", tabBackground: "accent>pink", line: row},
		{name: "a flat accent row ignores a tab gradient from its colour", accent: "green", tabBackground: "green>pink", line: row},
		{name: "a flat tab colour ignores a tab gradient from its colour", accent: "green", tabBackground: "teal>pink", tabColor: "teal", line: tabs},
		{name: "a flat tab colour ignores an accent gradient from its colour", accent: "green>blue", tabColor: "green", line: tabs},
		{name: "an auto tab follows the tab gradient", accent: "green", tabBackground: "teal>pink", tabColor: "auto", line: tabs, wantMany: true},
		{name: "an auto tab follows the accent gradient", accent: "green>blue", tabColor: "auto", line: tabs, wantMany: true},
		{name: "the selected row follows the accent gradient", accent: "green>blue", tabBackground: "green>pink", line: row, wantMany: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			look := gradientLook()
			look.BorderColors, look.TextColors, look.Accent = nil, nil, tt.accent
			if tt.tabBackground != "" {
				look.TextBackgrounds = map[string]string{"tab": tt.tabBackground}
			}
			withTrueColor(t, true, look)
			applyConm(config.Conm{
				Connections: []config.ConnectionSettings{{Type: config.PostgresConnType, TabColor: tt.tabColor}},
				Appearance:  look,
			})
			t.Cleanup(func() { applyConm(config.Conm{}); applyAppearance(plainAppearance) })
			conns := view.NewConnections(stubConnections{confs: []config.Connection{config.Postgres{}}})

			got := distinctBackgrounds(tt.line(New(conns, nil, nil)))

			if tt.wantMany {
				assert.Greater(t, got, 5)
				return
			}
			assert.Equal(t, 1, got)
		})
	}
}

func TestTabAndHeadingBackgrounds(t *testing.T) {
	tests := []struct {
		name           string
		backgrounds    map[string]string
		trueColor      bool
		wantTab        color.Color
		wantHead       color.Color
		wantHeadColors int
	}{
		{name: "auto keeps today's colours", wantTab: accentColors["green"], wantHead: baseHeadBg, wantHeadColors: 1},
		{
			name:           "flat overrides",
			backgrounds:    map[string]string{"tab": "pink", "heading": "#202020"},
			wantTab:        accentColors["pink"],
			wantHead:       lipgloss.Color("#202020"),
			wantHeadColors: 1,
		},
		{
			name:           "a heading gradient spreads across the band",
			backgrounds:    map[string]string{"heading": "#202020>#604080"},
			trueColor:      true,
			wantTab:        accentColors["green"],
			wantHead:       lipgloss.Color("#202020"),
			wantHeadColors: 10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			look := gradientLook()
			look.BorderColors, look.TextColors, look.TextBackgrounds = nil, nil, tt.backgrounds
			withTrueColor(t, tt.trueColor, look)

			is := assert.New(t)
			is.Equal(tt.wantTab, tabBg())
			is.True(sameColor(tt.wantHead, cHeadBg))
			band := frameHead(30, "section", "")[1]
			if tt.wantHeadColors > 1 {
				is.Greater(distinctBackgrounds(band), tt.wantHeadColors)
				return
			}
			is.Equal(1, distinctBackgrounds(band))
		})
	}
}

func TestSubheadingRole(t *testing.T) {
	tests := []struct {
		name            string
		styles          map[string]string
		colors          map[string]string
		backgrounds     map[string]string
		trueColor       bool
		wantBackgrounds int
		wantBold        bool
	}{
		{name: "default is bold with no band", wantBold: true},
		{
			name:            "style, colour and background are its own",
			styles:          map[string]string{"subheading": "italic"},
			colors:          map[string]string{"subheading": "pink"},
			backgrounds:     map[string]string{"subheading": "#202020"},
			wantBackgrounds: 1,
		},
		{
			name:            "a background gradient spreads across the line",
			backgrounds:     map[string]string{"subheading": "#202020>#604080"},
			trueColor:       true,
			wantBackgrounds: 10,
			wantBold:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			look := gradientLook()
			look.BorderColors, look.TextStyle, look.TextColors, look.TextBackgrounds = nil, tt.styles, tt.colors, tt.backgrounds
			withTrueColor(t, tt.trueColor, look)

			is := assert.New(t)
			is.Equal(tt.wantBold, tSubheading.bold)
			is.Equal(textStyle{bold: true}, textStyle{bold: tHeading.bold}, "the heading keeps its own style")
			line := frameSubhead(30, "top left", "")
			switch {
			case tt.wantBackgrounds > 1:
				is.Greater(distinctBackgrounds(line), tt.wantBackgrounds)
			default:
				is.Equal(tt.wantBackgrounds, distinctBackgrounds(line))
			}
			if tt.colors != nil {
				is.Equal(accentColors["pink"], tSubheading.fg)
				is.Nil(tHeading.fg)
			}
		})
	}
}

func TestAppearanceRegressions(t *testing.T) {
	base := config.AppearanceSettings{
		Border: "single", Accent: "green", Glyphs: "unicode", Text: "default",
		BorderColor: "default", GlyphColor: "default", TextColor: "default",
	}
	t.Cleanup(func() { applyAppearance(base) })

	tests := []struct {
		name   string
		before func(*config.AppearanceSettings)
		after  func(*config.AppearanceSettings)
		check  func(t *testing.T)
	}{
		{
			name:   "emoji glyphs keep the table row inside the frame",
			before: func(*config.AppearanceSettings) {},
			after:  func(a *config.AppearanceSettings) { a.Glyphs = "emoji" },
			check: func(t *testing.T) {
				t.Helper()
				require.Equal(t, 2, glyphCell)
				got := 2 * glyphCell
				for _, c := range config.DefaultColumns {
					got += layoutWidth(config.DefaultColumns, c, 0) + 1
				}
				assert.Equal(t, tableInner, got)
			},
		},
		{
			name:   "body back to fg drops the previous colour",
			before: func(a *config.AppearanceSettings) { a.TextColors = map[string]string{"body": "#ff0000"} },
			after:  func(a *config.AppearanceSettings) { a.TextColors = map[string]string{"body": "fg"} },
			check: func(t *testing.T) {
				t.Helper()
				assert.True(t, sameColor(baseFg, cFg))
			},
		},
		{
			name:   "accent glyphs keep the caret visible on an accent row",
			before: func(*config.AppearanceSettings) {},
			after:  func(a *config.AppearanceSettings) { a.GlyphColor = "accent" },
			check: func(t *testing.T) {
				t.Helper()
				assert.False(t, sameColor(caretFg(cInvFg, cAccent), cAccent))
				assert.True(t, sameColor(cAccent, caretFg(cInvFg, nil)))
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before, after := base, base
			tt.before(&before)
			tt.after(&after)
			applyAppearance(before)
			applyAppearance(after)
			tt.check(t)
		})
	}
}
