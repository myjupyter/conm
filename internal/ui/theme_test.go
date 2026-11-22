package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
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
		{name: "text", names: config.TextPalettes, palettes: textPalettes, elements: config.TextElements, valid: config.IsColorOrGradient},
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

func withGradients(t *testing.T, a config.AppearanceSettings) {
	t.Helper()
	t.Cleanup(func() {
		applyAppearance(config.AppearanceSettings{
			Border: "single", Accent: "green", Glyphs: "unicode", Text: "default",
			BorderColor: "default", GlyphColor: "default", TextColor: "default",
		})
	})
	applyAppearance(a)
}

func gradientLook(on bool) config.AppearanceSettings {
	return config.AppearanceSettings{
		Gradients: on, Border: "single", Accent: "green", Glyphs: "unicode", Text: "default",
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

func TestGradientsFollowTheSwitch(t *testing.T) {
	tests := []struct {
		name     string
		on       bool
		wantMany bool
	}{
		{name: "off draws the first stop", on: false},
		{name: "on spreads the stops", on: true, wantMany: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			withGradients(t, gradientLook(tt.on))

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
	withGradients(t, gradientLook(true))

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
