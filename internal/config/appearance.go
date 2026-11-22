package config

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/charmbracelet/x/ansi"
)

const ASCIIGlyphs = "ascii"

var (
	BorderStyles = []string{"single", "rounded", "double", "heavy", "dashed", "dotted", "mixed", "ascii", "hidden"}
	AccentColors = []string{"green", "lime", "teal", "blue", "purple", "pink", "red", "orange", "amber"}
	GlyphSets    = []string{"unicode", "heavy", "geometric", "minimal", "nerd", ASCIIGlyphs}

	BorderElements = []string{
		"top_left", "top_right", "bottom_left", "bottom_right",
		"tee_left", "tee_right", "horizontal", "vertical",
	}
	GlyphElements = []string{
		"status_idle", "status_ok", "status_pending", "status_warn", "status_error",
		"import_new", "caret", "dot_on", "dot_off", "dot_fail", "field_bad",
		"input_cursor", "input_mask", "empty", "ellipsis", "hint_arrow", "scroll_up", "scroll_down",
		"select_prev", "select_next",
	}
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type AppearanceSettings struct {
	Border      string            `toml:"border"`
	BorderRunes map[string]string `toml:"border_runes,omitempty"`
	Accent      string            `toml:"accent"`
	Glyphs      string            `toml:"glyphs"`
	GlyphRunes  map[string]string `toml:"glyph_runes,omitempty"`
}

func IsHexColor(value string) bool {
	return hexColor.MatchString(value)
}

func IsOneCell(value string) bool {
	return ansi.StringWidth(value) == 1 && len([]rune(value)) == 1
}

func (a AppearanceSettings) Clone() AppearanceSettings {
	a.BorderRunes = maps.Clone(a.BorderRunes)
	a.GlyphRunes = maps.Clone(a.GlyphRunes)
	return a
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
	errs = append(errs, validateRunes("border", a.BorderRunes, BorderElements)...)
	errs = append(errs, validateRunes("glyph", a.GlyphRunes, GlyphElements)...)
	return errors.Join(errs...)
}

func validateRunes(kind string, runes map[string]string, elements []string) []error {
	var errs []error
	for _, element := range slices.Sorted(maps.Keys(runes)) {
		switch value := runes[element]; {
		case !slices.Contains(elements, element):
			errs = append(errs, fmt.Errorf("unknown %s element %q", kind, element))
		case !IsOneCell(value):
			errs = append(errs, fmt.Errorf("%s %s %q is not one character wide", kind, element, value))
		}
	}
	return errs
}

func (a *AppearanceSettings) fillDefaults() {
	a.Border = knownOr(a.Border, BorderStyles)
	if !IsHexColor(a.Accent) {
		a.Accent = knownOr(a.Accent, AccentColors)
	}
	a.Glyphs = knownOr(a.Glyphs, GlyphSets)
	a.BorderRunes = validRunes(a.BorderRunes, BorderElements)
	a.GlyphRunes = validRunes(a.GlyphRunes, GlyphElements)
}

func validRunes(runes map[string]string, elements []string) map[string]string {
	maps.DeleteFunc(runes, func(element, value string) bool {
		return !slices.Contains(elements, element) || !IsOneCell(value)
	})
	if len(runes) == 0 {
		return nil
	}
	return runes
}

func knownOr(value string, options []string) string {
	if slices.Contains(options, value) {
		return value
	}
	return options[0]
}
