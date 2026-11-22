package ui

import (
	"testing"

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
	}{
		{name: "border", names: config.BorderStyles, presets: borderPresets, elements: config.BorderElements, vars: borderVars},
		{name: "glyphs", names: config.GlyphSets, presets: glyphPresets, elements: config.GlyphElements, vars: glyphVars},
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
					is.Equal(1, ansi.StringWidth(tt.presets[name][element]), name+" "+element)
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
			want: map[string]string{"top_left": "╭", "top_right": "┐", "horizontal": "─"},
		},
		{
			name: "ascii glyphs replace preset and overrides",
			a:    config.AppearanceSettings{Border: "double", Glyphs: config.ASCIIGlyphs, BorderRunes: map[string]string{"top_left": "╭"}},
			want: map[string]string{"top_left": "+", "top_right": "+", "horizontal": "-"},
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
