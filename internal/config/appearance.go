package config

import "slices"

var (
	BorderStyles = []string{"single", "rounded", "double", "heavy"}
	AccentColors = []string{"green", "amber", "blue", "pink"}
	GlyphSets    = []string{"unicode", "ascii"}
)

type AppearanceSettings struct {
	Border string `toml:"border"`
	Accent string `toml:"accent"`
	Glyphs string `toml:"glyphs"`
}

func (a *AppearanceSettings) fillDefaults() {
	a.Border = knownOr(a.Border, BorderStyles)
	a.Accent = knownOr(a.Accent, AccentColors)
	a.Glyphs = knownOr(a.Glyphs, GlyphSets)
}

func knownOr(value string, options []string) string {
	if slices.Contains(options, value) {
		return value
	}
	return options[0]
}
