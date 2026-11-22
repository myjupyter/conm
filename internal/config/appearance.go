package config

import (
	"errors"
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/charmbracelet/x/ansi"
	"github.com/rivo/uniseg"
)

const (
	ASCIIGlyphs   = "ascii"
	PlainText     = "plain"
	AutoColor     = "auto"
	DefaultPreset = "default"
	AccentColor   = "accent"

	GradientSeparator = ">"
)

var (
	BorderStyles = []string{
		"single", "rounded", "double", "heavy", "dashed", "dotted", "mixed", "block", "solid", "shade", "ascii", "hidden",
	}
	AccentColors = []string{"green", "lime", "teal", "blue", "purple", "pink", "red", "orange", "amber"}
	GlyphSets    = []string{"unicode", "heavy", "geometric", "minimal", "emoji", "nerd", ASCIIGlyphs}
	TextStyles   = []string{DefaultPreset, PlainText, "emphasis", "soft"}

	BorderElements = []string{
		"top_left", "top_right", "bottom_left", "bottom_right",
		"tee_left", "tee_right", "top", "bottom", "rule", "left", "right",
	}
	GlyphElements = []string{
		"status_idle", "status_ok", "status_pending", "status_warn", "status_error",
		"import_new", "caret", "dot_on", "dot_off", "dot_fail", "field_bad",
		"input_cursor", "input_mask", "empty", "ellipsis", "hint_arrow", "scroll_up", "scroll_down",
		"select_prev", "select_next",
	}
	TextElements      = []string{"title", "tab", "heading", "selection", "key", "value", "note"}
	TextColorElements = append(slices.Clone(TextElements), "body", "input")
	TextAttributes    = []string{"bold", "italic", "underline", "faint"}

	BorderPalettes     = []string{DefaultPreset, AccentColor, "soft", "bright", "duotone", "sunset", "ocean"}
	GlyphPalettes      = []string{DefaultPreset, AccentColor, "bright", "muted"}
	TextPalettes       = []string{DefaultPreset, AccentColor, "muted", "sunset", "ocean"}
	GlyphColorElements = []string{
		"status_idle", "status_ok", "status_pending", "status_warn", "status_error",
		"import_new", "caret", "dot_on", "dot_off", "dot_fail", "field_bad",
		"input_cursor", "hint_arrow", "select_prev", "select_next",
	}
	ColorNames = append([]string{AutoColor, AccentColor, "border", "dim", "soft", "fg", "ok", "warn", "error"}, AccentColors...)
)

var hexColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

type AppearanceSettings struct {
	Gradients   bool              `toml:"gradients"`
	Border      string            `toml:"border"`
	BorderRunes map[string]string `toml:"border_runes,omitempty"`
	Accent      string            `toml:"accent"`
	Glyphs      string            `toml:"glyphs"`
	GlyphRunes  map[string]string `toml:"glyph_runes,omitempty"`
	Text        string            `toml:"text"`
	TextStyle   map[string]string `toml:"text_style,omitempty"`

	BorderColor  string            `toml:"border_color"`
	BorderColors map[string]string `toml:"border_colors,omitempty"`
	GlyphColor   string            `toml:"glyph_color"`
	GlyphColors  map[string]string `toml:"glyph_colors,omitempty"`
	TextColor    string            `toml:"text_color"`
	TextColors   map[string]string `toml:"text_colors,omitempty"`
}

func IsColor(value string) bool {
	return slices.Contains(ColorNames, value) || IsHexColor(value)
}

func IsGradient(value string) bool {
	stops := strings.Split(value, GradientSeparator)
	if len(stops) < 2 {
		return false
	}
	for _, stop := range stops {
		if stop == AutoColor || !IsColor(stop) {
			return false
		}
	}
	return true
}

func IsColorOrGradient(value string) bool {
	return IsColor(value) || IsGradient(value)
}

func IsHexColor(value string) bool {
	return hexColor.MatchString(value)
}

func IsBorderRune(value string) bool {
	return uniseg.GraphemeClusterCount(value) == 1 && ansi.StringWidth(value) == 1
}

func IsGlyph(value string) bool {
	width := ansi.StringWidth(value)
	return uniseg.GraphemeClusterCount(value) == 1 && width >= 1 && width <= 2
}

func IsTextStyle(value string) bool {
	if value == PlainText {
		return true
	}
	parts := strings.Split(value, "+")
	for i, part := range parts {
		if !slices.Contains(TextAttributes, part) || slices.Contains(parts[:i], part) {
			return false
		}
	}
	return true
}

func (a AppearanceSettings) Clone() AppearanceSettings {
	a.BorderRunes = maps.Clone(a.BorderRunes)
	a.GlyphRunes = maps.Clone(a.GlyphRunes)
	a.TextStyle = maps.Clone(a.TextStyle)
	a.BorderColors = maps.Clone(a.BorderColors)
	a.GlyphColors = maps.Clone(a.GlyphColors)
	a.TextColors = maps.Clone(a.TextColors)
	return a
}

type overrideKind struct {
	name     string
	elements []string
	valid    func(string) bool
	problem  string
}

var (
	borderOverrides = overrideKind{"border", BorderElements, IsBorderRune, "is not one character one column wide"}
	glyphOverrides  = overrideKind{"glyph", GlyphElements, IsGlyph, "is not one character one or two columns wide"}
	textOverrides   = overrideKind{"text", TextElements, IsTextStyle, "is not plain or a +-joined list of " + strings.Join(TextAttributes, ", ")}

	colorProblem         = "is neither a colour name nor #rrggbb"
	gradientProblem      = "is neither a colour name, #rrggbb nor a gradient like teal>pink"
	borderColorOverrides = overrideKind{"border colour", BorderElements, IsColorOrGradient, gradientProblem}
	glyphColorOverrides  = overrideKind{"glyph colour", GlyphColorElements, IsColor, colorProblem}
	textColorOverrides   = overrideKind{"text colour", TextColorElements, IsColorOrGradient, gradientProblem}
)

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
	if !slices.Contains(TextStyles, a.Text) {
		errs = append(errs, fmt.Errorf("unknown text style %q", a.Text))
	}
	if !slices.Contains(BorderPalettes, a.BorderColor) {
		errs = append(errs, fmt.Errorf("unknown border colour %q", a.BorderColor))
	}
	if !slices.Contains(GlyphPalettes, a.GlyphColor) {
		errs = append(errs, fmt.Errorf("unknown glyph colour %q", a.GlyphColor))
	}
	if !slices.Contains(TextPalettes, a.TextColor) {
		errs = append(errs, fmt.Errorf("unknown text colour %q", a.TextColor))
	}
	errs = append(errs, borderOverrides.validate(a.BorderRunes)...)
	errs = append(errs, glyphOverrides.validate(a.GlyphRunes)...)
	errs = append(errs, textOverrides.validate(a.TextStyle)...)
	errs = append(errs, borderColorOverrides.validate(a.BorderColors)...)
	errs = append(errs, glyphColorOverrides.validate(a.GlyphColors)...)
	errs = append(errs, textColorOverrides.validate(a.TextColors)...)
	return errors.Join(errs...)
}

func (k overrideKind) validate(overrides map[string]string) []error {
	var errs []error
	for _, element := range slices.Sorted(maps.Keys(overrides)) {
		switch value := overrides[element]; {
		case !slices.Contains(k.elements, element):
			errs = append(errs, fmt.Errorf("unknown %s element %q", k.name, element))
		case !k.valid(value):
			errs = append(errs, fmt.Errorf("%s %s %q %s", k.name, element, value, k.problem))
		}
	}
	return errs
}

func (k overrideKind) keepValid(overrides map[string]string) map[string]string {
	maps.DeleteFunc(overrides, func(element, value string) bool {
		return !slices.Contains(k.elements, element) || !k.valid(value)
	})
	if len(overrides) == 0 {
		return nil
	}
	return overrides
}

func (a *AppearanceSettings) fillDefaults() {
	a.Border = knownOr(a.Border, BorderStyles)
	if !IsHexColor(a.Accent) {
		a.Accent = knownOr(a.Accent, AccentColors)
	}
	a.Glyphs = knownOr(a.Glyphs, GlyphSets)
	a.Text = knownOr(a.Text, TextStyles)
	a.BorderRunes = borderOverrides.keepValid(a.BorderRunes)
	a.GlyphRunes = glyphOverrides.keepValid(a.GlyphRunes)
	a.TextStyle = textOverrides.keepValid(a.TextStyle)
	a.BorderColor = knownOr(a.BorderColor, BorderPalettes)
	a.GlyphColor = knownOr(a.GlyphColor, GlyphPalettes)
	a.BorderColors = borderColorOverrides.keepValid(a.BorderColors)
	a.GlyphColors = glyphColorOverrides.keepValid(a.GlyphColors)
	a.TextColor = knownOr(a.TextColor, TextPalettes)
	a.TextColors = textColorOverrides.keepValid(a.TextColors)
}

func knownOr(value string, options []string) string {
	if slices.Contains(options, value) {
		return value
	}
	return options[0]
}
