package ui

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"

	"github.com/myjupyter/conm/internal/config"
)

func TestThemeRunesAreOneCellWide(t *testing.T) {
	t.Parallel()

	sets := map[string]any{"ascii box": asciiBox}
	for name, box := range borderStyles {
		sets["border "+name] = box
	}
	for name, glyphs := range glyphSets {
		sets["glyphs "+name] = glyphs
	}

	for name, set := range sets {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := reflect.ValueOf(set)
			for i := range v.NumField() {
				assert.Equal(t, 1, ansi.StringWidth(v.Field(i).String()), v.Type().Field(i).Name)
			}
		})
	}
}

func TestThemeCoversEveryConfigName(t *testing.T) {
	t.Parallel()

	for _, name := range config.BorderStyles {
		assert.Contains(t, borderStyles, name)
	}
	for _, name := range config.AccentColors {
		assert.Contains(t, accentColors, name)
	}
	for _, name := range config.GlyphSets {
		assert.Contains(t, glyphSets, name)
	}
}
