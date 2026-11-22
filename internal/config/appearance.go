package config

import (
	"errors"
	"fmt"
	"regexp"
	"slices"
)

const ASCIIGlyphs = "ascii"

var (
	BorderStyles = []string{"single", "rounded", "double", "heavy"}
	AccentColors = []string{"green", "lime", "teal", "blue", "purple", "pink", "red", "orange", "amber"}
	GlyphSets    = []string{"unicode", "heavy", "geometric", "nerd", ASCIIGlyphs}
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type AppearanceSettings struct {
	Border string `toml:"border"`
	Accent string `toml:"accent"`
	Glyphs string `toml:"glyphs"`
}

func IsHexColor(value string) bool {
	return hexColor.MatchString(value)
}

func (a AppearanceSettings) Validate() error {
	var errs []error
	if !slices.Contains(BorderStyles, a.Border) {
		errs = append(errs, fmt.Errorf("unknown border %q", a.Border))
	}
	if !slices.Contains(AccentColors, a.Accent) && !IsHexColor(a.Accent) {
		errs = append(errs, fmt.Errorf("accent %q is neither a preset nor a #rrggbb colour", a.Accent))
	}
	if !slices.Contains(GlyphSets, a.Glyphs) {
		errs = append(errs, fmt.Errorf("unknown glyph set %q", a.Glyphs))
	}
	return errors.Join(errs...)
}

func (a *AppearanceSettings) fillDefaults() {
	a.Border = knownOr(a.Border, BorderStyles)
	if !IsHexColor(a.Accent) {
		a.Accent = knownOr(a.Accent, AccentColors)
	}
	a.Glyphs = knownOr(a.Glyphs, GlyphSets)
}

func knownOr(value string, options []string) string {
	if slices.Contains(options, value) {
		return value
	}
	return options[0]
}
